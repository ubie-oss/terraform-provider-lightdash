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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type GetGroupMembersV1Result struct {
	LastName  string `json:"lastName"`
	FirstName string `json:"firstName"`
	Email     string `json:"email"`
	UserUUID  string `json:"userUuid"`
}

type GetGroupMembersV1Response struct {
	Results []GetGroupMembersV1Result `json:"results,omitempty"`
	Status  string                    `json:"status"`
}

func (c *Client) GetGroupMembersV1(groupUuid string) ([]GetGroupMembersV1Result, error) {
	// Validate the arguments
	if strings.TrimSpace(groupUuid) == "" {
		return nil, fmt.Errorf("group UUID is empty")
	}

	// Make a request
	path := fmt.Sprintf("%s/api/v1/groups/%s/members", c.HostUrl, groupUuid)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request for group members: %w", err)
	}
	// Do the request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request for group members: %w", err)
	}
	// Parse the response
	response := GetGroupMembersV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling group members response: %w", err)
	}

	return response.Results, nil
}
