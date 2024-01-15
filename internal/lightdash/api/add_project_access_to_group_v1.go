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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type AddProjectAccessToGroupRequest struct {
	Role string `json:"role"`
}

type AddProjectAccessToGroupResults struct {
	ProjectUUID string `json:"projectUuid"`
	GroupUUID   string `json:"groupUuid"`
	Role        string `json:"role"`
}

type AddProjectAccessToGroupResponse struct {
	Results AddProjectAccessToGroupResults `json:"results,omitempty"`
	Status  string                         `json:"status"`
}

func (c *Client) AddProjectAccessToGroupV1(projectUuid string, groupUuid string, role models.ProjectMemberRole) (*AddProjectAccessToGroupResults, error) {
	// Validate the role
	if !models.IsValidProjectMemberRole(role.String()) {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	// Create the request body
	data := AddProjectAccessToGroupRequest{
		Role: role.String(),
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request data: %s", err)
	}

	// Create the request
	path := fmt.Sprintf("%s/api/v1/groups/%s/projects/%s", c.HostUrl, groupUuid, projectUuid)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create new HTTP request: %v", err)
	}

	// Do request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %v", err)
	}

	// Marshal the response
	response := AddProjectAccessToGroupResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	// Make sure the project and group UUIDs are not empty
	if response.Results.ProjectUUID == "" || response.Results.GroupUUID == "" {
		return nil, fmt.Errorf("project or group UUID is empty")
	}

	return &response.Results, nil
}
