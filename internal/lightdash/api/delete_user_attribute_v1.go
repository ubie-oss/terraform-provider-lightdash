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

package api

import (
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) DeleteUserAttributeV1(userAttributeUuid string) error {
	// Validate the arguments
	if strings.TrimSpace(userAttributeUuid) == "" {
		return fmt.Errorf("user attribute UUID is empty")
	}

	// Create the request
	path := fmt.Sprintf("%s/api/v1/org/attributes/%s", c.HostUrl, userAttributeUuid)
	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("error creating DELETE request for user attribute: %v", err)
	}

	// Do the request
	_, err = c.DoRequest(req)
	if err != nil {
		return fmt.Errorf("error performing DELETE request for user attribute UUID '%s': %v", userAttributeUuid, err)
	}

	return nil
}
