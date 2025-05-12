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

// Space Management Methods

// CreateSpace creates a new space with specified properties
func (s *SpaceService) CreateSpace(projectUuid, spaceName string, isPrivate bool, parentSpaceUuid *string) (*models.SpaceDetails, error) {
	createdSpace, err := s.client.CreateSpaceV1(projectUuid, spaceName, isPrivate, parentSpaceUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to create space: %w", err)
	}

	spaceDetails := &models.SpaceDetails{
		ProjectUUID:        projectUuid,
		SpaceUUID:          createdSpace.SpaceUUID,
		SpaceName:          createdSpace.SpaceName,
		IsPrivate:          createdSpace.IsPrivate,
		ParentSpaceUUID:    parentSpaceUuid,
		SpaceAccessMembers: []models.SpaceAccessMember{},
		SpaceAccessGroups:  []models.SpaceAccessGroup{},
		ChildSpaces:        []models.ChildSpace{},
	}

	return spaceDetails, nil
}

// GetSpace retrieves a space by UUID
func (s *SpaceService) GetSpace(projectUuid, spaceUuid string) (*api.GetSpaceV1Results, error) {
	return s.client.GetSpaceV1(projectUuid, spaceUuid)
}

// UpdateRootSpace updates the space properties for a root space
func (s *SpaceService) UpdateRootSpace(projectUuid, spaceUuid, spaceName string, isPrivate *bool) (*api.UpdateSpaceV1Results, error) {
	updatedSpace, err := s.client.UpdateSpaceV1(projectUuid, spaceUuid, spaceName, isPrivate)
	if err != nil {
		return nil, fmt.Errorf("failed to update space properties: %w", err)
	}
	return updatedSpace, nil
}

// UpdateNestedSpace updates the space properties for a nested space
func (s *SpaceService) UpdateNestedSpace(projectUuid, spaceUuid string, spaceName string, parentSpaceUuidPointer *string) (*api.UpdateSpaceV1Results, error) {
	updatedSpace, err := s.client.UpdateSpaceV1(projectUuid, spaceUuid, spaceName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update nested space: %w", err)
	}
	return updatedSpace, nil
}

// DeleteSpace deletes a space
func (s *SpaceService) DeleteSpace(projectUuid, spaceUuid string) error {
	return s.client.DeleteSpaceV1(projectUuid, spaceUuid)
}

// MoveSpace moves a space to a new parent space
// parentSpaceUuidPointer == nil means the space should become a root space
func (s *SpaceService) MoveSpace(projectUuid, spaceUuid string, parentSpaceUuidPointer *string) error {
	err := s.client.MoveSpaceV2(projectUuid, spaceUuid, parentSpaceUuidPointer)
	if err != nil {
		return fmt.Errorf("failed to move space: %w", err)
	}
	return nil
}

// Resource ID Handling Methods

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

// AddUserToSpace grants a user access to a space with the specified role
// NOTE: Should only be called for root spaces
func (s *SpaceService) AddUserToSpace(projectUuid, spaceUuid, userUuid string, role models.SpaceMemberRole) error {
	return s.client.AddSpaceShareToUserV1(projectUuid, spaceUuid, userUuid, role)
}

// RemoveUserFromSpace revokes a user's access to a space
// NOTE: Should only be called for root spaces
func (s *SpaceService) RemoveUserFromSpace(projectUuid, spaceUuid, userUuid string) error {
	return s.client.RevokeSpaceAccessV1(projectUuid, spaceUuid, userUuid)
}

// AddGroupToSpace grants a group access to a space with the specified role
// NOTE: Should only be called for root spaces
func (s *SpaceService) AddGroupToSpace(projectUuid, spaceUuid, groupUuid string, role models.SpaceMemberRole) error {
	return s.client.AddSpaceShareToGroupV1(projectUuid, spaceUuid, groupUuid, role)
}

// UpdateGroupAccessInSpace updates a group's role in a space
// NOTE: Should only be called for root spaces
func (s *SpaceService) UpdateGroupAccessInSpace(projectUuid, spaceUuid, groupUuid string, role models.SpaceMemberRole) error {
	return s.client.AddSpaceGroupAccessV1(projectUuid, spaceUuid, groupUuid, role)
}

// RemoveGroupFromSpace revokes a group's access to a space
// NOTE: Should only be called for root spaces
func (s *SpaceService) RemoveGroupFromSpace(projectUuid, spaceUuid, groupUuid string) error {
	return s.client.RevokeSpaceGroupAccessV1(projectUuid, spaceUuid, groupUuid)
}

// GetChildSpaces returns all child spaces of a space
func (s *SpaceService) GetChildSpaces(projectUuid, spaceUuid string) ([]api.ChildSpace, error) {
	space, err := s.GetSpace(projectUuid, spaceUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get space details: %w", err)
	}

	return space.ChildSpaces, nil
}
