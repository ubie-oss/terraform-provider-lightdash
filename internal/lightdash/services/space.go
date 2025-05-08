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
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// SpaceService provides methods for managing Lightdash spaces.
type SpaceService struct {
	client *api.Client
}

// NewSpaceService creates a new SpaceService.
func NewSpaceService(client *api.Client) *SpaceService {
	return &SpaceService{client: client}
}

// CreateSpace creates a new space with specified properties
func (s *SpaceService) CreateSpace(projectUuid, spaceName string, isPrivate bool, parentSpaceUuid *string) (*api.CreateSpaceV1Results, error) {
	return s.client.CreateSpaceV1(projectUuid, spaceName, isPrivate, parentSpaceUuid)
}

// GetSpace retrieves a space by UUID
func (s *SpaceService) GetSpace(projectUuid, spaceUuid string) (*api.GetSpaceV1Results, error) {
	return s.client.GetSpaceV1(projectUuid, spaceUuid)
}

// DeleteSpace deletes a space
func (s *SpaceService) DeleteSpace(projectUuid, spaceUuid string) error {
	return s.client.DeleteSpaceV1(projectUuid, spaceUuid)
}

// UpdateSpace updates the space's name, privacy, and optionally its parent space.
// projectUuid is required for both UpdateSpaceV1 and MoveSpaceV1 API calls.
func (s *SpaceService) UpdateSpace(projectUuid, spaceUuid, spaceName string, isPrivate *bool, parentSpaceUuidPointer *string) (*api.UpdateSpaceV1Results, error) {
	// Update the space's name, privacy, and parent (if provided)
	updatedSpace, err := s.client.UpdateSpaceV1(projectUuid, spaceUuid, spaceName, isPrivate, parentSpaceUuidPointer)
	if err != nil {
		return nil, err
	}

	// If parentSpaceUuid is provided and isn't the same as the updatedSpace's parentSpaceUUID, move the space to the new parent
	isSameParentSpaceUUID := compareTwoStringPointers(parentSpaceUuidPointer, updatedSpace.ParentSpaceUUID)
	if !isSameParentSpaceUUID {
		err := s.client.MoveSpaceV1(projectUuid, spaceUuid, parentSpaceUuidPointer)
		if err != nil {
			return nil, err
		}
		// Override the parentSpaceUUID in the updatedSpace to the new parentSpaceUUID
		updatedSpace.ParentSpaceUUID = parentSpaceUuidPointer
	}

	return updatedSpace, nil
}

// RenameSpace updates only the name of a space
func (s *SpaceService) RenameSpace(projectUuid, spaceUuid, newSpaceName string) (*api.UpdateSpaceV1Results, error) {
	return s.client.UpdateSpaceV1(projectUuid, spaceUuid, newSpaceName, nil, nil)
}

// AddUserToSpace grants a user access to a space with the specified role
func (s *SpaceService) AddUserToSpace(projectUuid, spaceUuid, userUuid string, role models.SpaceMemberRole) error {
	return s.client.AddSpaceShareToUserV1(projectUuid, spaceUuid, userUuid, role)
}

// RemoveUserFromSpace revokes a user's access to a space
func (s *SpaceService) RemoveUserFromSpace(projectUuid, spaceUuid, userUuid string) error {
	return s.client.RevokeSpaceAccessV1(projectUuid, spaceUuid, userUuid)
}

// AddGroupToSpace grants a group access to a space with the specified role
func (s *SpaceService) AddGroupToSpace(projectUuid, spaceUuid, groupUuid string, role models.SpaceMemberRole) error {
	return s.client.AddSpaceShareToGroupV1(projectUuid, spaceUuid, groupUuid, role)
}

// UpdateGroupAccessInSpace updates a group's role in a space
func (s *SpaceService) UpdateGroupAccessInSpace(projectUuid, spaceUuid, groupUuid string, role models.SpaceMemberRole) error {
	return s.client.AddSpaceGroupAccessV1(projectUuid, spaceUuid, groupUuid, role)
}

// RemoveGroupFromSpace revokes a group's access to a space
func (s *SpaceService) RemoveGroupFromSpace(projectUuid, spaceUuid, groupUuid string) error {
	return s.client.RevokeSpaceGroupAccessV1(projectUuid, spaceUuid, groupUuid)
}

// GetSpaceResourceID returns the formatted resource ID for a space
func (s *SpaceService) GetSpaceResourceID(projectUuid, spaceUuid string) string {
	return fmt.Sprintf("projects/%s/spaces/%s", projectUuid, spaceUuid)
}

// ExtractSpaceResourceID extracts project and space UUIDs from a resource ID
func (s *SpaceService) ExtractSpaceResourceID(resourceID string) (projectUuid string, spaceUuid string, err error) {
	pattern := `^projects/([^/]+)/spaces/([^/]+)$`
	groups, err := ExtractStringsByPattern(resourceID, pattern)
	if err != nil {
		return "", "", fmt.Errorf("could not extract resource ID: %w", err)
	}

	return groups[0], groups[1], nil
}

// Helper functions

// compareTwoStringPointers compares two string pointers and returns true if they are the same
func compareTwoStringPointers(a, b *string) bool {
	// If both are nil, they are the same
	if a == nil && b == nil {
		return true
	}
	// If both are not nil and have the same value, they are the same
	if a != nil && b != nil && *a == *b {
		return true
	}
	return false
}
