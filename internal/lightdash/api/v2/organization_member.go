// Copyright 2024 Ubie, inc.
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

package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

type UpdateOrganizationMemberV2Request struct {
	RoleUUID string `json:"roleUuid"`
}

type OrganizationMemberV2Results struct {
	UserUUID         string `json:"userUuid"`
	OrganizationUUID string `json:"organizationUuid"`
	Email            string `json:"email"`
	RoleUUID         string `json:"roleUuid"`
}

type UpdateOrganizationMemberV2Response struct {
	Status  string                      `json:"status"`
	Results OrganizationMemberV2Results `json:"results,omitempty"`
}

func UpdateOrganizationMemberV2(c *api.Client, orgUUID, userUUID, roleUUID string) (*OrganizationMemberV2Results, error) {
	data := UpdateOrganizationMemberV2Request{
		RoleUUID: roleUUID,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v2/orgs/%s/members/%s", c.HostUrl, orgUUID, userUUID), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var res UpdateOrganizationMemberV2Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Status != "ok" {
		return nil, fmt.Errorf("error updating organization member: %s", res.Status)
	}

	return &res.Results, nil
}
