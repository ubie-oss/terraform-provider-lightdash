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
	"log"
	"net/http"
)

type UpdateGroupInOrganizationV1Request struct {
	Name    string                `json:"name" validate:"required"`
	Members []UpdateGroupV1Member `json:"members,omitempty"`
}

type UpdateGroupV1Member struct {
	UserUUID string `json:"userUuid"`
}

type UpdateGroupV1Results struct {
	OrganizationUUID string `json:"organizationUuid"`
	GroupUUID        string `json:"uuid"`
	Name             string `json:"name"`
	CreatedAt        string `json:"createdAt"`
}

type UpdateGroupV1Response struct {
	Results UpdateGroupV1Results `json:"results,omitempty"`
	Status  string               `json:"status"`
}

func (c *Client) UpdateGroupV1(groupUuid string, groupName string, members []UpdateGroupV1Member) (*UpdateGroupV1Results, error) {
	// Create the request body
	data := UpdateGroupInOrganizationV1Request{
		Name:    groupName,
		Members: members,
	}

	// Marshal the request body
	marshalled, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("impossible to marshal group data: %s", err)
	}
	// Create the request
	path := fmt.Sprintf("%s/api/v1/groups/%s", c.HostUrl, groupUuid)
	req, err := http.NewRequest("PATCH", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request for updating group: %w", err)
	}
	// Do request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request for updating group: %w", err)
	}

	// Marshal the response
	response := UpdateGroupV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body for updating group: %w", err)
	}
	// Validate that the group UUID is present in the response
	if response.Results.GroupUUID == "" {
		return nil, fmt.Errorf("group UUID is missing in the response")
	}

	return &response.Results, nil
}
