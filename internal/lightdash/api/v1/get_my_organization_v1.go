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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

type GetMyOrganizationV1Results struct {
	OrganizationUUID string `json:"organizationUuid"`
	Name             string `json:"name"`
}

type GetMyOrganizationV1Response struct {
	Results GetMyOrganizationV1Results `json:"results,omitempty"`
	Status  string                     `json:"status"`
}

func GetMyOrganizationV1(c *api.Client) (*GetMyOrganizationV1Results, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/org", c.HostUrl), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for organization: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request to get organization failed: %w", err)
	}

	response := GetMyOrganizationV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal organization response: %w", err)
	}

	// Make sure if the organization is not nil
	if response.Results.OrganizationUUID == "" {
		return nil, fmt.Errorf("organization is nil")
	}

	return &response.Results, nil
}
