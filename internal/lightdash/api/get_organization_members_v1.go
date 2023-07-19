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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type GetOrganizationMembersV1Results struct {
	OrganizationUUID string                        `json:"organizationUuid"`
	UserUUID         string                        `json:"userUuid"`
	Email            string                        `json:"email"`
	OrganizationRole models.OrganizationMemberRole `json:"role"`
	IsActive         bool                          `json:"isActive"`
	IsInviteExpired  bool                          `json:"isInviteExpired"`
}

type GetOrganizationMembersV1Response struct {
	Results []GetOrganizationMembersV1Results `json:"results,omitempty"`
	Status  string                            `json:"status"`
}

func (c *Client) GetOrganizationMembersV1() ([]GetOrganizationMembersV1Results, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/org/users", c.HostUrl), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	response := GetOrganizationMembersV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Check if each member is valid
	for _, member := range response.Results {
		if member.OrganizationUUID == "" {
			return nil, fmt.Errorf("organization is nil")
		}
		if member.UserUUID == "" {
			return nil, fmt.Errorf("user is nil")
		}
		if member.Email == "" {
			return nil, fmt.Errorf("email is nil")
		}
	}

	return response.Results, nil
}

func (c *Client) GetOrganizationMemberByEmail(email string) (*GetOrganizationMembersV1Results, error) {
	if email == "" {
		return nil, fmt.Errorf("email is nil")
	}

	// Get all members in the organization
	members, err := c.GetOrganizationMembersV1()
	if err != nil {
		return nil, err
	}

	// Check if each member is valid
	for _, member := range members {
		if member.Email == email {
			return &member, nil
		}
	}

	return nil, fmt.Errorf("user is not found")
}
