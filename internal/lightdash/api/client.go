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
	"net/url"
	"time"
)

type Client struct {
	HTTPClient *http.Client
	HostUrl    string
	Token      string
	Semaphore  chan struct{}
}

func NewClient(host, token *string, maxConcurrentRequests *int64) (*Client, error) {
	var maxRequests int64 = 10
	if maxConcurrentRequests != nil {
		maxRequests = *maxConcurrentRequests
	}

	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Semaphore:  make(chan struct{}, maxRequests),
	}

	if host != nil {
		u, err := url.Parse(*host)
		if err != nil {
			return nil, fmt.Errorf("invalid host URL: %w", err)
		}
		if u.Scheme == "" {
			u.Scheme = "https"
		}
		c.HostUrl = u.String()
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

func (c *Client) DoRequest(req *http.Request) ([]byte, error) {
	if c.Semaphore != nil {
		c.Semaphore <- struct{}{}
		defer func() { <-c.Semaphore }()
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("ApiKey %s", c.Token))

	res, err := c.HTTPClient.Do(req) // #nosec G704
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer res.Body.Close() // #nosec G307

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Successful response codes
	if res.StatusCode == http.StatusOK ||
		res.StatusCode == http.StatusCreated ||
		res.StatusCode == http.StatusAccepted ||
		res.StatusCode == http.StatusNonAuthoritativeInfo ||
		res.StatusCode == http.StatusNoContent ||
		res.StatusCode == http.StatusResetContent ||
		res.StatusCode == http.StatusPartialContent ||
		res.StatusCode == http.StatusMultiStatus ||
		res.StatusCode == http.StatusAlreadyReported ||
		res.StatusCode == http.StatusIMUsed {
		return body, nil
	}

	// Error response codes
	return nil, fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, body)
}
