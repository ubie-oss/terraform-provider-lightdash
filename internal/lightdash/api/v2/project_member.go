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

type UpdateProjectMemberV2Request struct {
	RoleUUID string `json:"roleUuid"`
}

type ProjectMemberV2Results struct {
	UserUUID    string `json:"userUuid"`
	ProjectUUID string `json:"projectUuid"`
	Email       string `json:"email"`
	RoleUUID    string `json:"roleUuid"`
}

type UpdateProjectMemberV2Response struct {
	Status  string                 `json:"status"`
	Results ProjectMemberV2Results `json:"results,omitempty"`
}

func UpdateProjectMemberV2(c *api.Client, projectUUID, userUUID, roleUUID string) (*ProjectMemberV2Results, error) {
	data := UpdateProjectMemberV2Request{
		RoleUUID: roleUUID,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v2/projects/%s/members/%s", c.HostUrl, projectUUID, userUUID), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var res UpdateProjectMemberV2Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Status != "ok" {
		return nil, fmt.Errorf("error updating project member: %s", res.Status)
	}

	return &res.Results, nil
}

func GrantProjectMemberV2(c *api.Client, projectUUID, email, roleUUID string, sendEmail bool) (*ProjectMemberV2Results, error) {
	data := map[string]interface{}{
		"email":     email,
		"roleUuid":  roleUUID,
		"sendEmail": sendEmail,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v2/projects/%s/members", c.HostUrl, projectUUID), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var res UpdateProjectMemberV2Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Status != "ok" {
		return nil, fmt.Errorf("error granting project member: %s", res.Status)
	}

	return &res.Results, nil
}

func RevokeProjectMemberV2(c *api.Client, projectUUID, userUUID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v2/projects/%s/members/%s", c.HostUrl, projectUUID, userUUID), nil)
	if err != nil {
		return err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return err
	}

	var res struct {
		Status string `json:"status"`
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	if res.Status != "ok" {
		return fmt.Errorf("error revoking project member: %s", res.Status)
	}

	return nil
}

type GetProjectMembersV2Results struct {
	UserUUID    string `json:"userUuid"`
	ProjectUUID string `json:"projectUuid"`
	Email       string `json:"email"`
	RoleUUID    string `json:"roleUuid"`
	RoleName    string `json:"roleName"`
}

type GetProjectMembersV2Response struct {
	Status  string                       `json:"status"`
	Results []GetProjectMembersV2Results `json:"results"`
}

func GetProjectMembersV2(c *api.Client, projectUUID string) ([]GetProjectMembersV2Results, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v2/projects/%s/members", c.HostUrl, projectUUID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var res GetProjectMembersV2Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Status != "ok" {
		return nil, fmt.Errorf("error listing project members: %s", res.Status)
	}

	return res.Results, nil
}
