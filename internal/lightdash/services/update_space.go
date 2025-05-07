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
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

// UpdateSpaceService provides methods to update and move a Lightdash space.
type UpdateSpaceService struct {
	client *api.Client
}

// NewUpdateSpaceService creates a new UpdateSpaceService.
func NewUpdateSpaceService(client *api.Client) *UpdateSpaceService {
	return &UpdateSpaceService{client: client}
}

// UpdateSpace updates the space's name, privacy, and optionally its parent space.
// projectUuid is required for both UpdateSpaceV1 and MoveSpaceV1 API calls.
func (s *UpdateSpaceService) UpdateSpace(projectUuid, spaceUuid, spaceName string, isPrivate bool, parentSpaceUuid *string) error {
	// Update the space's name, privacy, and parent (if provided)
	updatedSpace, err := s.client.UpdateSpaceV1(projectUuid, spaceUuid, spaceName, isPrivate, parentSpaceUuid)
	if err != nil {
		return err
	}

	// If parentSpaceUuid is provided and isn't the same as the updatedSpace's parentSpaceUUID, move the space to the new parent
	isSameParentSpaceUUID := compareTwoStringPointers(parentSpaceUuid, updatedSpace.ParentSpaceUUID)
	if !isSameParentSpaceUUID {
		err := s.client.MoveSpaceV1(projectUuid, spaceUuid, parentSpaceUuid)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO move the function to an appropriate package
func compareTwoStringPointers(a, b *string) bool {
	/**
	 * This function compares two string pointers and returns true if they are the same.
	 * It handles the case where both pointers are nil, one pointer is nil and the other is not,
	 * and where both pointers are not nil and have the same value.
	 */

	// If both are nil, they are the same
	if a == nil && b == nil {
		return true
	}
	// If one is nil and the other is not, they are not the same
	if a != nil && b != nil && *a == *b {
		return true
	}
	return false
}
