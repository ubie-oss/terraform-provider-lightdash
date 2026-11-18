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

// UpdateUserAttributeV1 replaces the user attribute. The API requires `groups` and `users` in the body.
func UpdateUserAttributeV1(c *api.Client, userAttributeUUID string, request CreateUserAttributeV1Request) (*UserAttributeV1, error) {
	if strings.TrimSpace(userAttributeUUID) == "" {
		return nil, fmt.Errorf("user attribute UUID is empty")
	}
	if strings.TrimSpace(request.Name) == "" {
		return nil, fmt.Errorf("user attribute name is empty")
	}
	path := fmt.Sprintf("%s/api/v1/org/attributes/%s", c.HostUrl, userAttributeUUID)
	return writeUserAttributeV1(c, http.MethodPut, path, request)
}
