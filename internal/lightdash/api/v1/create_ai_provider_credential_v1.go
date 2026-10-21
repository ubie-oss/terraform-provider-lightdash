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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type CreateAiProviderCredentialV1Response struct {
	Results struct {
		UUID string `json:"uuid"`
	} `json:"results"`
	Status string `json:"status"`
}

func CreateAiProviderCredentialV1(c *api.Client, payload models.CreateAiProviderCredentialRequest) (string, error) {
	marshalled, err := json.Marshal(payload) // #nosec G117 -- marshals BYOK apiKey for Lightdash API
	if err != nil {
		return "", fmt.Errorf("failed to marshal create AI provider credential request: %w", err)
	}

	path := fmt.Sprintf("%s/api/v1/ai/provider-credentials", c.HostUrl)
	req, err := http.NewRequest(http.MethodPost, path, bytes.NewReader(marshalled))
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider credential request: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return "", fmt.Errorf("create AI provider credential request failed: %w", err)
	}

	response := CreateAiProviderCredentialV1Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal create AI provider credential response: %w", err)
	}
	if response.Results.UUID == "" {
		return "", fmt.Errorf("credential UUID is missing in the create AI provider credential response")
	}

	return response.Results.UUID, nil
}
