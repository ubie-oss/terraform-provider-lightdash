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
)

type GroupsPagination struct {
	Page           float64 `json:"page"`
	PageSize       float64 `json:"pageSize"`
	TotalResults   float64 `json:"totalResults"`
	TotalPageCount float64 `json:"totalPageCount"`
}

type GetOrganizationGroupsV1Results struct {
	OrganizationUUID string   `json:"organizationUuid"`
	Name             string   `json:"name"`
	GroupUUID        string   `json:"uuid"`
	CreatedAt        string   `json:"createdAt"`
	MemberUUIDs      []string `json:"memberUuids,omitempty"`
	Members          []struct {
		UserUUID string `json:"userUuid"`
		Email    string `json:"email"`
	} `json:"members,omitempty"`
}

type GetOrganizationGroupsV1Response struct {
	Results struct {
		Pagination GroupsPagination                 `json:"pagination"`
		Data       []GetOrganizationGroupsV1Results `json:"data"`
	} `json:"results"`
	Status string `json:"status"`
}

func (c *Client) GetOrganizationGroupsV1(page float64, pageSize float64, includeMembers float64, searchQuery string) ([]GetOrganizationGroupsV1Results, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/org/groups?page=%f&pageSize=%f&includeMembers=%f&searchQuery=%s", c.HostUrl, page, pageSize, includeMembers, searchQuery), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request for organization groups: %v", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing GET request for organization groups: %v", err)
	}

	response := GetOrganizationGroupsV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response for organization groups: %v", err)
	}

	// Validate the response results
	for _, group := range response.Results.Data {
		if group.OrganizationUUID == "" {
			return nil, fmt.Errorf("organization UUID is empty")
		}
		if group.GroupUUID == "" {
			return nil, fmt.Errorf("group UUID is empty")
		}
		if group.Name == "" {
			return nil, fmt.Errorf("group name is empty")
		}
	}

	return response.Results.Data, nil
}
