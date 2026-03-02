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

type UpdateProjectGroupV2Request struct {
	RoleUUID string `json:"roleUuid"`
}

type ProjectGroupV2Results struct {
	GroupUUID   string `json:"groupUuid"`
	ProjectUUID string `json:"projectUuid"`
	RoleUUID    string `json:"roleUuid"`
}

type UpdateProjectGroupV2Response struct {
	Status  string                `json:"status"`
	Results ProjectGroupV2Results `json:"results,omitempty"`
}

func UpdateProjectGroupV2(c *api.Client, projectUUID, groupUUID, roleUUID string) (*ProjectGroupV2Results, error) {
	data := UpdateProjectGroupV2Request{
		RoleUUID: roleUUID,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v2/projects/%s/groups/%s", c.HostUrl, projectUUID, groupUUID), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var res UpdateProjectGroupV2Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Status != "ok" {
		return nil, fmt.Errorf("error updating project group: %s", res.Status)
	}

	return &res.Results, nil
}

func AddProjectGroupV2(c *api.Client, projectUUID, groupUUID, roleUUID string) (*ProjectGroupV2Results, error) {
	data := map[string]interface{}{
		"groupUuid": groupUUID,
		"roleUuid":  roleUUID,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v2/projects/%s/groups", c.HostUrl, projectUUID), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var res UpdateProjectGroupV2Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Status != "ok" {
		return nil, fmt.Errorf("error adding project group: %s", res.Status)
	}

	return &res.Results, nil
}

func RemoveProjectGroupV2(c *api.Client, projectUUID, groupUUID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v2/projects/%s/groups/%s", c.HostUrl, projectUUID, groupUUID), nil)
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
		return fmt.Errorf("error removing project group: %s", res.Status)
	}

	return nil
}

type GetProjectGroupAccessesV2Results struct {
	GroupUUID   string `json:"groupUuid"`
	ProjectUUID string `json:"projectUuid"`
	RoleUUID    string `json:"roleUuid"`
	RoleName    string `json:"roleName"`
}

type GetProjectGroupAccessesV2Response struct {
	Status  string                             `json:"status"`
	Results []GetProjectGroupAccessesV2Results `json:"results"`
}

func GetProjectGroupAccessesV2(c *api.Client, projectUUID string) ([]GetProjectGroupAccessesV2Results, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v2/projects/%s/groups", c.HostUrl, projectUUID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var res GetProjectGroupAccessesV2Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Status != "ok" {
		return nil, fmt.Errorf("error listing project group accesses: %s", res.Status)
	}

	return res.Results, nil
}
