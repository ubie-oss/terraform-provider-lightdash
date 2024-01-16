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

type GetProjectV1Results struct {
	OrganizationUUID string `json:"organizationUuid"`
	ProjectUUID      string `json:"projectUuid"`
	ProjectName      string `json:"name"`
	ProjectType      string `json:"type"`
}

type GetProjectV1Response struct {
	Results GetProjectV1Results `json:"results,omitempty"`
	Status  string              `json:"status"`
}

func (c *Client) GetProjectV1(projectUuid string) (*GetProjectV1Results, error) {
	path := fmt.Sprintf("%s/api/v1/projects/%s", c.HostUrl, projectUuid)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request for project: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request for project: %w", err)
	}

	response := GetProjectV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling project response: %w", err)
	}

	// Make sure if the organization is not nil
	if response.Results.ProjectUUID == "" {
		return nil, fmt.Errorf("project UUID is nil")
	}

	return &response.Results, nil
}
