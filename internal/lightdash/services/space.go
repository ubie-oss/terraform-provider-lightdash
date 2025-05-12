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
func (s *SpaceService) CreateSpace(projectUuid, spaceName string, isPrivate bool, parentSpaceUuid *string) (*api.CreateSpaceV1Results, error) {
	return s.client.CreateSpaceV1(projectUuid, spaceName, isPrivate, parentSpaceUuid)
}

// GetSpace retrieves a space by UUID
func (s *SpaceService) GetSpace(projectUuid, spaceUuid string) (*api.GetSpaceV1Results, error) {
	return s.client.GetSpaceV1(projectUuid, spaceUuid)
}

// UpdateSpaceProperties updates just the space properties without changing access controls
func (s *SpaceService) UpdateSpaceProperties(projectUuid, spaceUuid, spaceName string, isPrivate *bool, parentSpaceUuid *string) (*api.UpdateSpaceV1Results, error) {
	return s.client.UpdateSpaceV1(projectUuid, spaceUuid, spaceName, isPrivate, parentSpaceUuid)
}

// UpdateSpace updates the space's name, privacy, and optionally its parent space.
// For root spaces, it can update name, privacy, and optionally move to become a nested space.
// For nested spaces, it can only update name and parent space (move to another parent or become a root space).
// parentSpaceUuidPointer == nil means the space should be a root space.
// isPrivate == nil means the privacy setting should not be changed.
func (s *SpaceService) UpdateSpace(projectUuid, spaceUuid, spaceName string, isPrivate *bool, parentSpaceUuidPointer *string) (*api.UpdateSpaceV1Results, error) {
	// First get current space details to determine if it's root or nested
	currentSpace, err := s.GetSpace(projectUuid, spaceUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get current space details: %w", err)
	}

	// Check if this is a nested space
	isCurrentlyNestedSpace := currentSpace.ParentSpaceUUID != nil

	// For nested spaces, only update name if privacy change is requested
	// (because privacy can't be changed for nested spaces)
	if isCurrentlyNestedSpace && isPrivate != nil {
		// Log that we're ignoring privacy change for nested spaces
		// Not returning error to avoid breaking existing configurations
		return s.client.UpdateSpaceV1(projectUuid, spaceUuid, spaceName, nil, parentSpaceUuidPointer)
	}

	// Update space properties based on the space type
	updatedSpace, err := s.client.UpdateSpaceV1(projectUuid, spaceUuid, spaceName, isPrivate, parentSpaceUuidPointer)
	if err != nil {
		return nil, fmt.Errorf("failed to update space properties: %w", err)
	}

	return updatedSpace, nil
}

// RenameSpace updates only the name of a space
func (s *SpaceService) RenameSpace(projectUuid, spaceUuid, newSpaceName string) (*api.UpdateSpaceV1Results, error) {
	return s.client.UpdateSpaceV1(projectUuid, spaceUuid, newSpaceName, nil, nil)
}

// DeleteSpace deletes a space
func (s *SpaceService) DeleteSpace(projectUuid, spaceUuid string) error {
	return s.client.DeleteSpaceV1(projectUuid, spaceUuid)
}

// MoveSpace moves a space to a new parent space
// parentSpaceUuidPointer == nil means the space should become a root space
func (s *SpaceService) MoveSpace(projectUuid, spaceUuid string, parentSpaceUuidPointer *string) error {
	// Get the current space details to check the current parent
	currentSpace, err := s.GetSpace(projectUuid, spaceUuid)
	if err != nil {
		return fmt.Errorf("failed to get current space details: %w", err)
	}

	// Compare the desired parent with the current parent
	isSameParentSpaceUUID := compareTwoStringPointers(parentSpaceUuidPointer, currentSpace.ParentSpaceUUID)

	// Only move the space if the parent is different
	if !isSameParentSpaceUUID {
		return s.client.MoveSpaceV1(projectUuid, spaceUuid, parentSpaceUuidPointer)
	}

	// No move needed, return nil
	return nil
}

// IsNestedSpace checks if a space is nested (has a parent)
func (s *SpaceService) IsNestedSpace(projectUuid, spaceUuid string) (bool, error) {
	space, err := s.GetSpace(projectUuid, spaceUuid)
	if err != nil {
		return false, fmt.Errorf("failed to get space details: %w", err)
	}

	return space.ParentSpaceUUID != nil, nil
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

// Space Access Management Methods

// SpaceAccessManager provides a dedicated interface for managing space access
type SpaceAccessManager struct {
	client       *api.Client
	spaceService *SpaceService
}

// NewSpaceAccessManager creates a new SpaceAccessManager
func NewSpaceAccessManager(client *api.Client) *SpaceAccessManager {
	return &SpaceAccessManager{
		client:       client,
		spaceService: NewSpaceService(client),
	}
}

// EnsureMemberAccess ensures a user has the specified access level to a space
// It will add or update access as needed
func (sam *SpaceAccessManager) EnsureMemberAccess(projectUuid, spaceUuid, userUuid string, role models.SpaceMemberRole) error {
	isNested, err := sam.spaceService.IsNestedSpace(projectUuid, spaceUuid)
	if err != nil {
		return err
	}

	if isNested {
		return fmt.Errorf("cannot manage access for nested space: access is inherited from parent")
	}

	return sam.client.AddSpaceShareToUserV1(projectUuid, spaceUuid, userUuid, role)
}

// RevokeMemberAccess removes a user's access to a space
func (sam *SpaceAccessManager) RevokeMemberAccess(projectUuid, spaceUuid, userUuid string) error {
	isNested, err := sam.spaceService.IsNestedSpace(projectUuid, spaceUuid)
	if err != nil {
		return err
	}

	if isNested {
		return fmt.Errorf("cannot remove access for nested space: access is inherited from parent")
	}

	return sam.client.RevokeSpaceAccessV1(projectUuid, spaceUuid, userUuid)
}

// EnsureGroupAccess ensures a group has the specified access level to a space
// It will add or update access as needed
func (sam *SpaceAccessManager) EnsureGroupAccess(projectUuid, spaceUuid, groupUuid string, role models.SpaceMemberRole) error {
	isNested, err := sam.spaceService.IsNestedSpace(projectUuid, spaceUuid)
	if err != nil {
		return err
	}

	if isNested {
		return fmt.Errorf("cannot manage group access for nested space: access is inherited from parent")
	}

	return sam.client.AddSpaceShareToGroupV1(projectUuid, spaceUuid, groupUuid, role)
}

// RevokeGroupAccess removes a group's access to a space
func (sam *SpaceAccessManager) RevokeGroupAccess(projectUuid, spaceUuid, groupUuid string) error {
	isNested, err := sam.spaceService.IsNestedSpace(projectUuid, spaceUuid)
	if err != nil {
		return err
	}

	if isNested {
		return fmt.Errorf("cannot remove group access for nested space: access is inherited from parent")
	}

	return sam.client.RevokeSpaceGroupAccessV1(projectUuid, spaceUuid, groupUuid)
}

// AddUserToSpace grants a user access to a space with the specified role
// NOTE: Should only be called for root spaces
func (s *SpaceService) AddUserToSpace(projectUuid, spaceUuid, userUuid string, role models.SpaceMemberRole) error {
	// Check if this is a nested space
	space, err := s.GetSpace(projectUuid, spaceUuid)
	if err != nil {
		return fmt.Errorf("failed to get space details: %w", err)
	}

	if space.ParentSpaceUUID != nil {
		return fmt.Errorf("cannot add user to nested space: space access is inherited from parent")
	}

	return s.client.AddSpaceShareToUserV1(projectUuid, spaceUuid, userUuid, role)
}

// RemoveUserFromSpace revokes a user's access to a space
// NOTE: Should only be called for root spaces
func (s *SpaceService) RemoveUserFromSpace(projectUuid, spaceUuid, userUuid string) error {
	// Check if this is a nested space
	space, err := s.GetSpace(projectUuid, spaceUuid)
	if err != nil {
		return fmt.Errorf("failed to get space details: %w", err)
	}

	if space.ParentSpaceUUID != nil {
		return fmt.Errorf("cannot remove user from nested space: space access is inherited from parent")
	}

	return s.client.RevokeSpaceAccessV1(projectUuid, spaceUuid, userUuid)
}

// AddGroupToSpace grants a group access to a space with the specified role
// NOTE: Should only be called for root spaces
func (s *SpaceService) AddGroupToSpace(projectUuid, spaceUuid, groupUuid string, role models.SpaceMemberRole) error {
	// Check if this is a nested space
	space, err := s.GetSpace(projectUuid, spaceUuid)
	if err != nil {
		return fmt.Errorf("failed to get space details: %w", err)
	}

	if space.ParentSpaceUUID != nil {
		return fmt.Errorf("cannot add group to nested space: space access is inherited from parent")
	}

	return s.client.AddSpaceShareToGroupV1(projectUuid, spaceUuid, groupUuid, role)
}

// UpdateGroupAccessInSpace updates a group's role in a space
// NOTE: Should only be called for root spaces
func (s *SpaceService) UpdateGroupAccessInSpace(projectUuid, spaceUuid, groupUuid string, role models.SpaceMemberRole) error {
	// Check if this is a nested space
	space, err := s.GetSpace(projectUuid, spaceUuid)
	if err != nil {
		return fmt.Errorf("failed to get space details: %w", err)
	}

	if space.ParentSpaceUUID != nil {
		return fmt.Errorf("cannot update group access in nested space: space access is inherited from parent")
	}

	return s.client.AddSpaceGroupAccessV1(projectUuid, spaceUuid, groupUuid, role)
}

// RemoveGroupFromSpace revokes a group's access to a space
// NOTE: Should only be called for root spaces
func (s *SpaceService) RemoveGroupFromSpace(projectUuid, spaceUuid, groupUuid string) error {
	// Check if this is a nested space
	space, err := s.GetSpace(projectUuid, spaceUuid)
	if err != nil {
		return fmt.Errorf("failed to get space details: %w", err)
	}

	if space.ParentSpaceUUID != nil {
		return fmt.Errorf("cannot remove group from nested space: space access is inherited from parent")
	}

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
