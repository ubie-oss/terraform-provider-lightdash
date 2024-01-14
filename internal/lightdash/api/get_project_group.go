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

type GetProjectGroupV1Result struct {
	OrganizationUUID string `json:"organizationUuid"`
	GroupUUID        string `json:"uuid"`
	CreatedAt        string `json:"createdAt"`
	Name             string `json:"name"`
}

type GetProjectGroupV1Response struct {
	Results GetProjectGroupV1Result `json:"results,omitempty"`
	Status  string                  `json:"status"`
}

func (c *Client) GetProjectGroupV1(groupUuid string) (*GetProjectGroupV1Result, error) {
	// Validate the arguments
	if strings.TrimSpace(groupUuid) == "" {
		return nil, fmt.Errorf("group UUID is empty")
	}

	// Make a request
	path := fmt.Sprintf("%s/api/v1/groups/%s", c.HostUrl, groupUuid)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	// Do the request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	// Parse the response
	response := GetProjectGroupV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Results, nil
}
