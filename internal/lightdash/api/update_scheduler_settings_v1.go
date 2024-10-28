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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type UpdateSchedulerSettingsV1Request struct {
	SchedulerTimezone string `json:"schedulerTimezone"`
}

type UpdateSchedulerSettingsV1Response struct {
	Results interface{} `json:"results,omitempty"`
	Status  string      `json:"status"`
}

func (c *Client) UpdateSchedulerSettingsV1(projectUuid string, schedulerTimezone string) (*UpdateSchedulerSettingsV1Response, error) {
	// Create the request body
	data := UpdateSchedulerSettingsV1Request{
		SchedulerTimezone: schedulerTimezone,
	}

	// Marshal the request body
	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("impossible to marshal scheduler settings data: %w", err)
	}
	// Create the request
	path := fmt.Sprintf("%s/api/v1/projects/%s/schedulerSettings", c.HostUrl, projectUuid)
	req, err := http.NewRequest("PATCH", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request for updating scheduler settings: %w", err)
	}
	// Do request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request for updating scheduler settings in project (%s) with timezone (%s): %w", projectUuid, schedulerTimezone, err)
	}

	// Marshal the response
	response := UpdateSchedulerSettingsV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body for updating scheduler settings: %w", err)
	}

	// Validate the response status
	if response.Status != "ok" {
		return nil, fmt.Errorf("unexpected response status: %s", response.Status)
	}

	return &response, nil
}
