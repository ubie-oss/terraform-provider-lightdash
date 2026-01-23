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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type GetProjectAccessListV1Results struct {
	ProjectUUID string                   `json:"projectUuid"`
	UserUUID    string                   `json:"userUuid"`
	Email       string                   `json:"email"`
	ProjectRole models.ProjectMemberRole `json:"role"`
}

type GetProjectAccessListV1Response struct {
	Results []GetProjectAccessListV1Results `json:"results,omitempty"`
	Status  string                          `json:"status"`
}

func GetProjectAccessListV1(c *api.Client, projectUuid string) ([]GetProjectAccessListV1Results, error) {
	// Validate the arguments
	if len(strings.TrimSpace(projectUuid)) == 0 {
		return nil, fmt.Errorf("project UUID is empty")
	}

	// Make a request
	path := fmt.Sprintf("%s/api/v1/projects/%s/access", c.HostUrl, projectUuid)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request for project access list: %w", err)
	}
	// Do the request
	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request for project access list: %w", err)
	}
	// Parse the response
	response := GetProjectAccessListV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling project access list response: %w", err)
	}
	// Make sure if all of the user UUIDs are not empty
	for _, projectMember := range response.Results {
		if len(strings.TrimSpace(projectMember.UserUUID)) == 0 {
			return nil, fmt.Errorf("user UUID is empty")
		}
	}

	return response.Results, nil
}
