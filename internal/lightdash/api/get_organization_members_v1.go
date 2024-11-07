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

type MembersPagination struct {
	Page           int `json:"page"`
	PageSize       int `json:"pageSize"`
	TotalPageCount int `json:"totalPageCount"`
}

type GetOrganizationMembersV1Results struct {
	OrganizationUUID string                        `json:"organizationUuid"`
	UserUUID         string                        `json:"userUuid"`
	Email            string                        `json:"email"`
	FirstName        string                        `json:"firstName"`
	LastName         string                        `json:"lastName"`
	OrganizationRole models.OrganizationMemberRole `json:"role"`
	IsActive         bool                          `json:"isActive"`
	IsInviteExpired  bool                          `json:"isInviteExpired"`
}

type GetOrganizationMembersV1Response struct {
	Results struct {
		Pagination MembersPagination                 `json:"pagination"`
		Data       []GetOrganizationMembersV1Results `json:"data"`
	} `json:"results,omitempty"`
	Status string `json:"status"`
}

func (c *Client) GetOrganizationMembersV1(includeGroups, pageSize, page int, searchQuery string) ([]GetOrganizationMembersV1Results, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/org/users", c.HostUrl), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request for organization members: %w", err)
	}

	q := req.URL.Query()
	if includeGroups != 0 {
		q.Add("includeGroups", fmt.Sprintf("%d", includeGroups))
	}
	if pageSize != 0 {
		q.Add("pageSize", fmt.Sprintf("%d", pageSize))
	}
	if page != 0 {
		q.Add("page", fmt.Sprintf("%d", page))
	}
	if searchQuery != "" {
		q.Add("searchQuery", searchQuery)
	}
	req.URL.RawQuery = q.Encode()

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request for organization members: %w", err)
	}

	response := GetOrganizationMembersV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response for organization members: %w", err)
	}

	// Check if each member is valid
	for _, member := range response.Results.Data {
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

	return response.Results.Data, nil
}
