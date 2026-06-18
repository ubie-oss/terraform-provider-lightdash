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

// ListOrganizationRoleAssignmentsV2 returns organization-level role assignments.
func ListOrganizationRoleAssignmentsV2(c *api.Client, orgUUID string) ([]models.RoleAssignment, error) {
	if err := validateUUID(orgUUID, "organization UUID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/api/v2/orgs/%s/roles/assignments", c.HostUrl, orgUUID)
	body, err := doJSONRequest(c, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("request to list organization role assignments failed: %w", err)
	}

	return unmarshalRoleAssignmentListResponse(body)
}
