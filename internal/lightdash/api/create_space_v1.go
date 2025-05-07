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

type CreateSpaceV1Request struct {
	Name            string  `json:"name"`
	IsPrivate       bool    `json:"isPrivate"`
	ParentSpaceUUID *string `json:"parentSpaceUuid,omitempty"`
}

type CreateSpaceV1Results struct {
	OrganizationUUID string  `json:"organizationUuid"`
	ProjectUUID      string  `json:"projectUuid"`
	ParentSpaceUUID  *string `json:"parentSpaceUuid,omitempty"`
	SpaceUUID        string  `json:"uuid"`
	SpaceName        string  `json:"name"`
	IsPrivate        bool    `json:"isPrivate"`
}

type CreateSpaceV1Response struct {
	Results CreateSpaceV1Results `json:"results,omitempty"`
	Status  string               `json:"status"`
}

// CreateSpaceV1 creates a new space in the given project. If parentSpaceUUID is nil, the space is created at the root level.
func (c *Client) CreateSpaceV1(projectUUID, spaceName string, isPrivate bool, parentSpaceUUID *string) (*CreateSpaceV1Results, error) {
	data := CreateSpaceV1Request{
		Name:            spaceName,
		IsPrivate:       isPrivate,
		ParentSpaceUUID: parentSpaceUUID,
	}

	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CreateSpaceV1Request: %w", err)
	}

	path := fmt.Sprintf("%s/api/v1/projects/%s/spaces", c.HostUrl, projectUUID)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	var response CreateSpaceV1Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Results.SpaceUUID == "" {
		return nil, fmt.Errorf("space UUID is empty in response")
	}

	return &response.Results, nil
}
