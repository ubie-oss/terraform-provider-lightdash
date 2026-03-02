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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

type OrganizationRole struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type GetOrganizationRolesResponse struct {
	Status  string             `json:"status"`
	Results []OrganizationRole `json:"results"`
}

func GetOrganizationRolesV2(c *api.Client, orgUUID string) ([]OrganizationRole, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v2/orgs/%s/roles", c.HostUrl, orgUUID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var res GetOrganizationRolesResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Status != "ok" {
		return nil, fmt.Errorf("error listing roles: %s", res.Status)
	}

	return res.Results, nil
}
