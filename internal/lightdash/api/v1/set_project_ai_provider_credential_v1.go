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
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

func SetProjectAiProviderCredentialV1(c *api.Client, projectUUID string, selection models.ProjectAiCredentialSelection) error {
	if strings.TrimSpace(projectUUID) == "" {
		return fmt.Errorf("project UUID is empty")
	}

	marshalled, err := json.Marshal(selection)
	if err != nil {
		return fmt.Errorf("failed to marshal set project AI provider credential request: %w", err)
	}

	path := fmt.Sprintf("%s/api/v1/projects/%s/ai/provider-credential", c.HostUrl, projectUUID)
	req, err := http.NewRequest(http.MethodPut, path, bytes.NewReader(marshalled))
	if err != nil {
		return fmt.Errorf("failed to create set project AI provider credential request: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return fmt.Errorf("set project AI provider credential request failed: %w", err)
	}

	response := ApiSuccessEmptyV1Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to unmarshal set project AI provider credential response: %w", err)
	}

	return nil
}
