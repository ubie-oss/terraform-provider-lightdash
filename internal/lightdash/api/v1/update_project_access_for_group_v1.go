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

package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type UpdateProjectAccessForGroupRequest struct {
	ProjectRole models.ProjectMemberRole `json:"role"`
}

type UpdateProjectAccessForGroupResults struct {
	ProjectUUID string `json:"projectUuid"`
	GroupUUID   string `json:"groupUuid"`
	Role        string `json:"role"`
}

type UpdateProjectAccessForGroupResponse struct {
	Results UpdateProjectAccessForGroupResults `json:"results"`
	Status  string                             `json:"status"`
}

func UpdateProjectAccessForGroupV1(c *api.Client, projectUuid string, groupUuid string, role models.ProjectMemberRole) (*UpdateProjectAccessForGroupResults, error) {
	// Create the request body
	data := UpdateProjectAccessForGroupRequest{
		ProjectRole: role,
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("impossible to marshal data: %s", err)
	}
	// Create the request
	path := fmt.Sprintf("%s/api/v1/groups/%s/projects/%s", c.HostUrl, groupUuid, projectUuid)
	req, err := http.NewRequest("PATCH", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create new HTTP request for updating project access for group: %v", err)
	}
	// Do request
	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request to update project access for group failed: %v", err)
	}

	// Unmarshal the response into the UpdateProjectAccessForGroupResponse struct
	response := UpdateProjectAccessForGroupResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return &response.Results, nil
}
