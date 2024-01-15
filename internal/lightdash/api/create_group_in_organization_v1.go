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

type CreateGroupInOrganizationV1Request struct {
	Name    string                              `json:"name" validate:"required"`
	Members []CreateGroupInOrganizationV1Member `json:"members" validate:"required"`
}

type CreateGroupInOrganizationV1Member struct {
	UserUUID string `json:"userUuid"`
}

type CreateGroupInOrganizationV1Results struct {
	OrganizationUUID string `json:"organizationUuid"`
	GroupUUID        string `json:"uuid"`
	Name             string `json:"name"`
	CreatedAt        string `json:"createdAt"`
}

type CreateGroupInOrganizationV1Response struct {
	Results CreateGroupInOrganizationV1Results `json:"results,omitempty"`
	Status  string                             `json:"status"`
}

func (c *Client) CreateGroupInOrganizationV1(organizationUuid string, groupName string, members []CreateGroupInOrganizationV1Member) (*CreateGroupInOrganizationV1Results, error) {
	// Create the request body
	data := CreateGroupInOrganizationV1Request{
		Name:    groupName,
		Members: members,
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("impossible to marshall teacher: %s", err)
	}
	// Create the request
	path := fmt.Sprintf("%s/api/v1/org/groups", c.HostUrl)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v, body: %s", err, string(marshalled))
	}
	// Do request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v, body: %s", err, string(marshalled))
	}
	// Marshal the response
	response := CreateGroupInOrganizationV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v, body: %s", err, string(body))
	}
	// Validate that the group UUID is present in the response
	if response.Results.GroupUUID == "" {
		return nil, fmt.Errorf("group UUID is missing in the response")
	}
	return &response.Results, nil
}
