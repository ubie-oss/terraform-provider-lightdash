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
	"fmt"
	"net/http"
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

func DeleteOAuthClientV1(c *api.Client, clientID string) error {
	if strings.TrimSpace(clientID) == "" {
		return fmt.Errorf("client ID is empty")
	}

	path := fmt.Sprintf("%s/api/v1/oauth/clients/%s", c.HostUrl, clientID)
	req, err := http.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete OAuth client request: %w", err)
	}

	_, err = c.DoRequest(req)
	if err != nil {
		return fmt.Errorf("delete OAuth client request failed: %w", err)
	}

	return nil
}
