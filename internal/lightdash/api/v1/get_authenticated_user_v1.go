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

type GetAuthenticatedUserV1Results struct {
	OrganizationUUID string `json:"organizationUuid"`
	UserUUID         string `json:"userUuid"`
	// We don't add the other fields because we don't need them and they are a little sensitive
	// SEE https://docs.lightdash.com/api/v1/#tag/My-Account/operation/GetAuthenticatedUser
}

type GetAuthenticatedUserV1Response struct {
	Results GetAuthenticatedUserV1Results `json:"results,omitempty"`
	Status  string                        `json:"status"`
}

func GetAuthenticatedUserV1(c *api.Client) (*GetAuthenticatedUserV1Results, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/user", c.HostUrl), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request for authenticated user: %v", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing GET request for authenticated user: %v", err)
	}

	response := GetAuthenticatedUserV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response for authenticated user: %v", err)
	}

	// Make sure if the organization is not nil
	if response.Results.UserUUID == "" {
		return nil, fmt.Errorf("UserUUID is nil")
	}

	return &response.Results, nil
}
