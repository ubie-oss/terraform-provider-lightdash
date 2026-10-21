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
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type ListAiProviderCredentialsV1Response struct {
	Results models.AiProviderCredentialsList `json:"results"`
	Status  string                           `json:"status"`
}

func ListAiProviderCredentialsV1(c *api.Client) (*models.AiProviderCredentialsList, error) {
	path := fmt.Sprintf("%s/api/v1/ai/provider-credentials", c.HostUrl)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list AI provider credentials request: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list AI provider credentials request failed: %w", err)
	}

	response := ListAiProviderCredentialsV1Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal list AI provider credentials response: %w", err)
	}

	return &response.Results, nil
}
