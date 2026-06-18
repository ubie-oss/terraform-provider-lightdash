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

package v2

import (
	"fmt"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

// RemoveProjectRoleFromUserV2 removes a user's project role assignment.
func RemoveProjectRoleFromUserV2(c *api.Client, projectUUID string, userUUID string) error {
	if err := validateUUID(projectUUID, "project UUID"); err != nil {
		return err
	}
	if err := validateUUID(userUUID, "user UUID"); err != nil {
		return err
	}

	path := fmt.Sprintf("%s/api/v2/projects/%s/roles/assignments/user/%s", c.HostUrl, projectUUID, userUUID)
	_, err := doJSONRequest(c, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("request to remove project role from user failed: %w", err)
	}

	return nil
}
