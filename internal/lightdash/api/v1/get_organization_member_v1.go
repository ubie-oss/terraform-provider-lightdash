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

type GetOrganizationMemberByUuidV1Response struct {
	Results GetOrganizationMembersV1Results `json:"results,omitempty"`
	Status  string                          `json:"status"`
}

func GetOrganizationMemberByUuidV1(c *api.Client, userUuid string) (*GetOrganizationMembersV1Results, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/org/users/%s", c.HostUrl, userUuid), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new HTTP request: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request to get organization member: %w", err)
	}

	response := GetOrganizationMemberByUuidV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response body: %w", err)
	}

	// Check if each member is valid
	if response.Results.OrganizationUUID == "" {
		return nil, fmt.Errorf("organization is nil")
	}
	if response.Results.UserUUID == "" {
		return nil, fmt.Errorf("user is nil")
	}
	if response.Results.Email == "" {
		return nil, fmt.Errorf("email is nil")
	}

	return &response.Results, nil
}
