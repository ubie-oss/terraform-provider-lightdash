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
)

func DeleteAiProviderCredentialV1(c *api.Client, credentialUUID string) error {
	if strings.TrimSpace(credentialUUID) == "" {
		return fmt.Errorf("credential UUID is empty")
	}

	path := fmt.Sprintf("%s/api/v1/ai/provider-credentials/%s", c.HostUrl, credentialUUID)
	req, err := http.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete AI provider credential request: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		if strings.Contains(err.Error(), "status code: 404") {
			return nil
		}
		return fmt.Errorf("delete AI provider credential request failed: %w", err)
	}

	response := ApiSuccessEmptyV1Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to unmarshal delete AI provider credential response: %w", err)
	}

	return nil
}
