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

type GetOrganizationGroupsV1Results struct {
	OrganizationUUID string `json:"organizationUuid"`
	Name             string `json:"name"`
	GroupUUID        string `json:"uuid"`
	CreatedAt        string `json:"createdAt"`
}

type GetOrganizationGroupsV1Response struct {
	Results []GetOrganizationGroupsV1Results `json:"results,omitempty"`
	Status  string                           `json:"status"`
}

func (c *Client) GetOrganizationGroupsV1() ([]GetOrganizationGroupsV1Results, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/org/groups", c.HostUrl), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	response := GetOrganizationGroupsV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Validate the response results
	for _, group := range response.Results {
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

	return response.Results, nil
}
