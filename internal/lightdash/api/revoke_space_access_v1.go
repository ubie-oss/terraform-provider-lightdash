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
)

func (c *Client) RevokeSpaceAccessV1(projectUuid string, spaceUuid string, userUuid string) error {
	// Create the request
	path := fmt.Sprintf("%s/api/v1/projects/%s/spaces/%s/share/%s",
		c.HostUrl, projectUuid, spaceUuid, userUuid)
	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to create new request for revoking space access: %w", err)
	}
	// Do request
	_, err = c.doRequest(req)
	if err != nil {
		return fmt.Errorf("request to revoke space access failed: %w", err)
	}

	return nil
}
