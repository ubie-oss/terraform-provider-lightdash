// Copyright 2023 Ubie, inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	HTTPClient            *http.Client
	HostUrl               string
	Token                 string
	MaxConcurrentRequests int64
	RequestTimeout        time.Duration
	RetryTimes            int64
	semaphore             chan struct{}
}

func NewClient(host, token *string, maxConcurrentRequests, requestTimeout, retryTimes int64) (*Client, error) {
	c := Client{
		HTTPClient:            &http.Client{Timeout: time.Duration(requestTimeout) * time.Second},
		MaxConcurrentRequests: maxConcurrentRequests,
		RequestTimeout:        time.Duration(requestTimeout) * time.Second,
		RetryTimes:            retryTimes,
		semaphore:             make(chan struct{}, maxConcurrentRequests),
	}

	if host != nil {
		c.HostUrl = *host
	}

	if token != nil {
		c.Token = *token
	}

	// Get the organization for the current token
	// _, err := GetMyOrganizationV1(&c)
	// if err != nil {
	// 	return nil, err
	// }

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	// Acquire a semaphore to limit concurrent requests
	c.semaphore <- struct{}{}
	defer func() { <-c.semaphore }() // Release the semaphore when done

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("ApiKey %s", c.Token))

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer res.Body.Close() // #nosec G307

	var body []byte
	for i := int64(0); i < c.RetryTimes; i++ {
		// Retry logic
		if res.StatusCode >= 500 {
			time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
			continue
		}

		// Successful response codes
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			body, err = io.ReadAll(res.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response body: %v", err)
			}
			return body, nil
		}

		// Error response codes
		body, err = io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %v", err)
		}
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, body)
	}

	return nil, fmt.Errorf("failed after retries: unexpected status code: %d, body: %s", res.StatusCode, body)
}
