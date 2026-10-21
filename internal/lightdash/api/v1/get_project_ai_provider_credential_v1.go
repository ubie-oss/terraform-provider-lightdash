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
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type GetProjectAiProviderCredentialV1Response struct {
	Results models.ProjectAiCredentialSelection `json:"results"`
	Status  string                              `json:"status"`
}

func GetProjectAiProviderCredentialV1(c *api.Client, projectUUID string) (*models.ProjectAiCredentialSelection, error) {
	if strings.TrimSpace(projectUUID) == "" {
		return nil, fmt.Errorf("project UUID is empty")
	}

	path := fmt.Sprintf("%s/api/v1/projects/%s/ai/provider-credential", c.HostUrl, projectUUID)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get project AI provider credential request: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("get project AI provider credential request failed: %w", err)
	}

	response := GetProjectAiProviderCredentialV1Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal get project AI provider credential response: %w", err)
	}

	return &response.Results, nil
}
