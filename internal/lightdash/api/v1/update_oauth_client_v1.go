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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

type UpdateOAuthClientV1Request struct {
	ClientName   string   `json:"clientName"`
	RedirectURIs []string `json:"redirectUris"`
}

type UpdateOAuthClientV1Response struct {
	Results OAuthClientV1 `json:"results,omitempty"`
	Status  string        `json:"status"`
}

func UpdateOAuthClientV1(c *api.Client, clientID string, clientName string, redirectURIs []string) (*OAuthClientV1, error) {
	if strings.TrimSpace(clientID) == "" {
		return nil, fmt.Errorf("client ID is empty")
	}

	data := UpdateOAuthClientV1Request{
		ClientName:   clientName,
		RedirectURIs: redirectURIs,
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update OAuth client request: %w", err)
	}

	path := fmt.Sprintf("%s/api/v1/oauth/clients/%s", c.HostUrl, clientID)
	req, err := http.NewRequest(http.MethodPatch, path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create update OAuth client request: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("update OAuth client request failed: %w", err)
	}

	response := UpdateOAuthClientV1Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal update OAuth client response: %w", err)
	}

	if response.Results.ClientID == "" {
		return nil, fmt.Errorf("client ID is missing in the update OAuth client response")
	}

	return &response.Results, nil
}
