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
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// UpdateProjectGroupRoleV2 updates a group's project role assignment.
func UpdateProjectGroupRoleV2(c *api.Client, projectUUID string, groupUUID string, roleID string) (*models.RoleAssignment, error) {
	if err := validateUUID(projectUUID, "project UUID"); err != nil {
		return nil, err
	}
	if err := validateUUID(groupUUID, "group UUID"); err != nil {
		return nil, err
	}
	if err := validateUUID(roleID, "role ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/api/v2/projects/%s/roles/assignments/group/%s", c.HostUrl, projectUUID, groupUUID)
	body, err := doJSONRequest(c, "PATCH", path, updateRoleAssignmentRequest{RoleID: roleID})
	if err != nil {
		return nil, fmt.Errorf("request to update project group role failed: %w", err)
	}

	return unmarshalRoleAssignmentResponse(body)
}
