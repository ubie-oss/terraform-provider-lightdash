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

package services

import (
	"fmt"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

func GrantSpaceAccess(client *api.Client, project_uuid string, space_uuid string, user_uuid string) error {
	// Check if the member is a member of the project.
	_, err := client.GetProjectMemberByUuidV1(project_uuid, user_uuid)
	if err != nil {
		return fmt.Errorf("failed to get project member by UUID: %w", err)
	}

	// Add space access
	err = client.AddSpaceShareToUserV1(project_uuid, space_uuid, user_uuid)
	if err != nil {
		return fmt.Errorf("failed to add space share to user: %w", err)
	}

	return nil
}
