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
	"log"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type UpdateOrganizationMemberV1Request struct {
	OrganizationRole models.OrganizationMemberRole `json:"role"`
}

type UpdateOrganizationMemberV1Results struct {
	OrganizationUUID string                        `json:"organizationUuid"`
	UserUUID         string                        `json:"userUuid"`
	Email            string                        `json:"email"`
	OrganizationRole models.OrganizationMemberRole `json:"role"`
	IsActive         bool                          `json:"isActive"`
	IsInviteExpired  bool                          `json:"isInviteExpired"`
}

type UpdateOrganizationMemberV1Response struct {
	Results UpdateOrganizationMemberV1Results `json:"results,omitempty"`
	Status  string                            `json:"status"`
}

func (c *Client) UpdateOrganizationMemberV1(userUuid string, role models.OrganizationMemberRole) (*UpdateOrganizationMemberV1Results, error) {
	// Create the request body
	data := UpdateOrganizationMemberV1Request{
		OrganizationRole: role,
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("impossible to marshall teacher: %s", err)
	}
	// Create the request
	path := fmt.Sprintf("%s/api/v1/org/users/%s", c.HostUrl, userUuid)
	req, err := http.NewRequest("PATCH", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, err
	}
	// Do request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	// Marshal the response
	response := UpdateOrganizationMemberV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return &response.Results, nil
}
