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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type GetProjectGroupAccessesV1Results struct {
	ProjectUUID string                   `json:"projectUuid"`
	GroupUUID   string                   `json:"groupUuid"`
	ProjectRole models.ProjectMemberRole `json:"role"`
}

type GetProjectGroupAccessesV1Response struct {
	Results []GetProjectGroupAccessesV1Results `json:"results,omitempty"`
	Status  string                             `json:"status"`
}

func (c *Client) GetProjectGroupAccessesV1(projectUuid string) ([]GetProjectGroupAccessesV1Results, error) {
	// Validate the arguments
	if len(strings.TrimSpace(projectUuid)) == 0 {
		return nil, fmt.Errorf("project UUID is empty")
	}

	// Make a request
	path := fmt.Sprintf("%s/api/v1/projects/%s/groupAccesses", c.HostUrl, projectUuid)
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
	response := GetProjectGroupAccessesV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.Results, nil
}
