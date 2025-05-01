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

type UpdateSpaceV1Request struct {
	Name            string  `json:"name"`
	IsPrivate       bool    `json:"isPrivate"`
	ParentSpaceUUID *string `json:"parentSpaceUuid,omitempty"`
}

type UpdateSpaceV1Results struct {
	OrganizationUUID string  `json:"organizationUuid"`
	ProjectUUID      string  `json:"projectUuid"`
	ParentSpaceUUID  *string `json:"parentSpaceUuid,omitempty"`
	SpaceUUID        string  `json:"uuid"`
	SpaceName        string  `json:"name"`
	IsPrivate        bool    `json:"isPrivate"`
}

type UpdateSpaceV1Response struct {
	Results UpdateSpaceV1Results `json:"results,omitempty"`
	Status  string               `json:"status"`
}

func (c *Client) UpdateSpaceV1(projectUuid string, spaceUuid string, spaceName string, isPrivate bool, parentSpaceUuid *string) (*UpdateSpaceV1Results, error) {
	// Create the request body, including parentSpaceUuid if provided
	data := UpdateSpaceV1Request{
		Name:      spaceName,
		IsPrivate: isPrivate,
	}
	if parentSpaceUuid != nil {
		data.ParentSpaceUUID = parentSpaceUuid
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal UpdateSpaceV1Request: %w", err)
	}
	// Create the request
	path := fmt.Sprintf("%s/api/v1/projects/%s/spaces/%s", c.HostUrl, projectUuid, spaceUuid)
	req, err := http.NewRequest("PATCH", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}
	// Do request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	// Marshal the response
	response := UpdateSpaceV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	// Make sure if the space UUID is not empty
	if response.Results.SpaceUUID == "" {
		return nil, fmt.Errorf("space UUID is nil")
	}
	return &response.Results, nil
}
