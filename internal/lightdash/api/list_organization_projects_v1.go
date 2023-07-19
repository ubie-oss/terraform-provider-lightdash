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

type ListOrganizationProjectsV1Results struct {
	ProjectUUID string `json:"projectUuid"`
	ProjectName string `json:"name"`
	ProjectType string `json:"type"`
}

type ListOrganizationProjectsV1Response struct {
	Results []ListOrganizationProjectsV1Results `json:"results,omitempty"`
	Status  string                              `json:"status"`
}

func (c *Client) ListOrganizationProjectsV1() ([]ListOrganizationProjectsV1Results, error) {
	path := fmt.Sprintf("%s/api/v1/org/projects", c.HostUrl)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	response := ListOrganizationProjectsV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.Results, nil
}
