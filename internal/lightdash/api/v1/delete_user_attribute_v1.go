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

func DeleteUserAttributeV1(c *api.Client, userAttributeUUID string) error {
	if strings.TrimSpace(userAttributeUUID) == "" {
		return fmt.Errorf("user attribute UUID is empty")
	}

	path := fmt.Sprintf("%s/api/v1/org/attributes/%s", c.HostUrl, userAttributeUUID)
	req, err := http.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request to delete user attribute: %w", err)
	}
	if _, err := c.DoRequest(req); err != nil {
		return fmt.Errorf("request to delete user attribute %s failed: %w", userAttributeUUID, err)
	}
	return nil
}
