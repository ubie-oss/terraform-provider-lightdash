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
)

type GetGroupV1Results struct {
	OrganizationUUID string `json:"organizationUuid"`
	GroupUUID        string `json:"uuid"`
	Name             string `json:"name"`
	CreatedAt        string `json:"createdAt"`
}

type GetGroupV1Response struct {
	Results GetGroupV1Results `json:"results,omitempty"`
	Status  string            `json:"status"`
}

func GetGroupV1(c *api.Client, groupUuid string) (*GetGroupV1Results, error) {
	// Validate the arguments
	if strings.TrimSpace(groupUuid) == "" {
		return nil, fmt.Errorf("group UUID is empty")
	}

	// Make a request
	path := fmt.Sprintf("%s/api/v1/groups/%s", c.HostUrl, groupUuid)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for group UUID '%s': %w", groupUuid, err)
	}
	// Do the request
	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request for group UUID '%s' failed: %w", groupUuid, err)
	}
	// Parse the response
	response := GetGroupV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response for group UUID '%s': %w", groupUuid, err)
	}

	return &response.Results, nil
}
