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

type OAuthClientV1 struct {
	ClientID          string   `json:"clientId"`
	ClientName        string   `json:"clientName"`
	RedirectURIs      []string `json:"redirectUris"`
	Scopes            []string `json:"scopes"`
	CreatedAt         string   `json:"createdAt"`
	CreatedByUserUUID *string  `json:"createdByUserUuid"`
	ClientSecret      string   `json:"clientSecret,omitempty"`
}

type ListOAuthClientsV1Response struct {
	Results []OAuthClientV1 `json:"results,omitempty"`
	Status  string          `json:"status"`
}

func ListOAuthClientsV1(c *api.Client) ([]OAuthClientV1, error) {
	path := fmt.Sprintf("%s/api/v1/oauth/clients", c.HostUrl)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list OAuth clients request: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list OAuth clients request failed: %w", err)
	}

	response := ListOAuthClientsV1Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal list OAuth clients response: %w", err)
	}

	return response.Results, nil
}
