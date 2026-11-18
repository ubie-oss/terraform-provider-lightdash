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

// UserAttributeV1 is one user attribute returned by GET /api/v1/org/attributes.
// `value` and `attributeDefault` are the deprecated single-value fields. Callers use `values` and `attributeDefaults`.
type UserAttributeV1 struct {
	UUID              string                 `json:"uuid"`
	OrganizationUUID  string                 `json:"organizationUuid"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	AttributeDefault  *string                `json:"attributeDefault"`
	AttributeDefaults []string               `json:"attributeDefaults"`
	Users             []UserAttributeV1User  `json:"users"`
	Groups            []UserAttributeV1Group `json:"groups"`
	CreatedAt         string                 `json:"createdAt"`
}

type UserAttributeV1User struct {
	UserUUID string   `json:"userUuid"`
	Email    string   `json:"email"`
	Value    string   `json:"value"`
	Values   []string `json:"values"`
}

type UserAttributeV1Group struct {
	GroupUUID string   `json:"groupUuid"`
	Value     string   `json:"value"`
	Values    []string `json:"values"`
}

type getUserAttributesV1Response struct {
	Results []UserAttributeV1 `json:"results"`
	Status  string            `json:"status"`
}

func GetUserAttributesV1(c *api.Client) ([]UserAttributeV1, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/org/attributes", c.HostUrl), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for user attributes: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request to list user attributes failed: %w", err)
	}

	var response getUserAttributesV1Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user attributes response: %w", err)
	}
	if response.Status != "ok" {
		return nil, fmt.Errorf("user attributes response status %q", response.Status)
	}
	if response.Results == nil {
		return []UserAttributeV1{}, nil
	}
	return response.Results, nil
}
