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

package controllers

import (
	"fmt"
	"time"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// SpaceController orchestrates operations related to Lightdash spaces
type SpaceController struct {
	spaceService               *services.SpaceService
	organizationMembersService *services.OrganizationMembersService
	organizationGroupsService  *services.OrganizationGroupsService
}

// SpaceMemberAccess represents a user's access to a space
type SpaceMemberAccess struct {
	UserUUID            string
	SpaceRole           models.SpaceMemberRole
	IsOrganizationAdmin bool
}

// SpaceGroupAccess represents a group's access to a space
type SpaceGroupAccess struct {
	GroupUUID string
	SpaceRole models.SpaceMemberRole
}

// SpaceDetails contains all the details of a space
type SpaceDetails struct {
	ProjectUUID     string
	SpaceUUID       string
	ParentSpaceUUID *string
	SpaceName       string
	IsPrivate       bool
	CreatedAt       time.Time
	MemberAccess    []SpaceMemberAccess
	GroupAccess     []SpaceGroupAccess
}

// NewSpaceController creates a new SpaceController
func NewSpaceController(client *api.Client) *SpaceController {
	return &SpaceController{
		spaceService:               services.NewSpaceService(client),
		organizationMembersService: services.NewOrganizationMembersService(client),
		organizationGroupsService:  services.NewOrganizationGroupsService(client),
	}
}

// CreateSpace creates a new space with the specified properties and access settings
func (c *SpaceController) CreateSpace(
	projectUUID string,
	spaceName string,
	isPrivate bool,
	parentSpaceUUID *string,
	memberAccess []SpaceMemberAccess,
	groupAccess []SpaceGroupAccess,
) (*SpaceDetails, []error) {
	var errors []error

	// Create the space
	createdSpace, err := c.spaceService.CreateSpace(projectUUID, spaceName, isPrivate, parentSpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to create space: %w", err)}
	}

	// Add member access
	memberAccessList := []SpaceMemberAccess{}
	for _, access := range memberAccess {
		// Skip organization admins as they inherently have access to all spaces
		isOrganizationAdmin, err := c.organizationMembersService.IsOrganizationAdmin(access.UserUUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to check if user %s is an organization admin: %w", access.UserUUID, err))
			continue
		}

		if isOrganizationAdmin {
			// Add to result list but don't try to grant access
			memberAccessList = append(memberAccessList, SpaceMemberAccess{
				UserUUID:            access.UserUUID,
				SpaceRole:           access.SpaceRole,
				IsOrganizationAdmin: true,
			})
			continue
		}

		// Add space access
		err = c.spaceService.AddUserToSpace(projectUUID, createdSpace.SpaceUUID, access.UserUUID, access.SpaceRole)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to add user %s to space: %w", access.UserUUID, err))
		} else {
			memberAccessList = append(memberAccessList, SpaceMemberAccess{
				UserUUID:            access.UserUUID,
				SpaceRole:           access.SpaceRole,
				IsOrganizationAdmin: isOrganizationAdmin,
			})
		}
	}

	// Add group access
	groupAccessList := []SpaceGroupAccess{}
	for _, access := range groupAccess {
		// Skip if the group no longer exists in the organization
		_, err := c.getGroupFromOrganization(access.GroupUUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("group %s not found: %w", access.GroupUUID, err))
			continue
		}

		// Add space access
		err = c.spaceService.AddGroupToSpace(projectUUID, createdSpace.SpaceUUID, access.GroupUUID, access.SpaceRole)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to add group %s to space: %w", access.GroupUUID, err))
		} else {
			groupAccessList = append(groupAccessList, SpaceGroupAccess{
				GroupUUID: access.GroupUUID,
				SpaceRole: access.SpaceRole,
			})
		}
	}

	// Build result
	spaceDetails := &SpaceDetails{
		ProjectUUID:     createdSpace.ProjectUUID,
		SpaceUUID:       createdSpace.SpaceUUID,
		ParentSpaceUUID: createdSpace.ParentSpaceUUID,
		SpaceName:       createdSpace.SpaceName,
		IsPrivate:       createdSpace.IsPrivate,
		CreatedAt:       time.Now(),
		MemberAccess:    memberAccessList,
		GroupAccess:     groupAccessList,
	}

	return spaceDetails, errors
}

// GetSpace retrieves the details of a space
func (c *SpaceController) GetSpace(projectUUID, spaceUUID string) (*SpaceDetails, error) {
	// Get space details
	space, err := c.spaceService.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get space: %w", err)
	}

	// Process member access
	memberAccessList := []SpaceMemberAccess{}
	for _, member := range space.SpaceAccessMembers {
		isOrganizationAdmin, err := c.organizationMembersService.IsOrganizationAdmin(member.UserUUID)
		if err != nil {
			// Log error but continue
			continue
		}

		memberAccessList = append(memberAccessList, SpaceMemberAccess{
			UserUUID:            member.UserUUID,
			SpaceRole:           member.SpaceRole,
			IsOrganizationAdmin: isOrganizationAdmin,
		})
	}

	// Process group access
	groupAccessList := []SpaceGroupAccess{}
	for _, group := range space.SpaceAccessGroups {
		groupAccessList = append(groupAccessList, SpaceGroupAccess{
			GroupUUID: group.GroupUUID,
			SpaceRole: group.SpaceRole,
		})
	}

	// Build result
	spaceDetails := &SpaceDetails{
		ProjectUUID:     space.ProjectUUID,
		SpaceUUID:       space.SpaceUUID,
		ParentSpaceUUID: space.ParentSpaceUUID,
		SpaceName:       space.SpaceName,
		IsPrivate:       space.IsPrivate,
		MemberAccess:    memberAccessList,
		GroupAccess:     groupAccessList,
	}

	return spaceDetails, nil
}

// UpdateSpace updates a space with the specified properties and access settings
func (c *SpaceController) UpdateSpace(
	projectUUID string,
	spaceUUID string,
	spaceName string,
	isPrivate *bool,
	parentSpaceUUID *string,
	newMemberAccess []SpaceMemberAccess,
	newGroupAccess []SpaceGroupAccess,
) (*SpaceDetails, []error) {
	var errors []error

	// Get the current space
	currentSpace, err := c.spaceService.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get current space: %w", err)}
	}

	// Update the space properties
	updatedSpace, err := c.spaceService.UpdateSpace(projectUUID, spaceUUID, spaceName, isPrivate, parentSpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to update space: %w", err)}
	}

	// Handle member access updates
	currentMemberUUIDs := make(map[string]models.SpaceMemberRole)
	for _, member := range currentSpace.SpaceAccessMembers {
		currentMemberUUIDs[member.UserUUID] = member.SpaceRole
	}

	newMemberUUIDs := make(map[string]models.SpaceMemberRole)
	for _, member := range newMemberAccess {
		newMemberUUIDs[member.UserUUID] = member.SpaceRole
	}

	// Process members to remove
	for userUUID := range currentMemberUUIDs {
		if _, exists := newMemberUUIDs[userUUID]; !exists {
			// Check if member exists in organization before removing
			_, err := c.organizationMembersService.GetOrganizationMemberByUserUuid(userUUID)
			if err != nil {
				// Skip if member no longer exists in organization
				continue
			}

			// Skip organization admins
			isOrganizationAdmin, err := c.organizationMembersService.IsOrganizationAdmin(userUUID)
			if err == nil && isOrganizationAdmin {
				continue
			}

			// Remove access
			err = c.spaceService.RemoveUserFromSpace(projectUUID, spaceUUID, userUUID)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to remove user %s from space: %w", userUUID, err))
			}
		}
	}

	// Process members to add or update
	updatedMemberAccess := []SpaceMemberAccess{}
	for _, member := range newMemberAccess {
		// Check if user exists in organization
		_, err := c.organizationMembersService.GetOrganizationMemberByUserUuid(member.UserUUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("user %s not found in organization: %w", member.UserUUID, err))
			continue
		}

		// Check organization admin status
		isOrganizationAdmin, err := c.organizationMembersService.IsOrganizationAdmin(member.UserUUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to check if user %s is an organization admin: %w", member.UserUUID, err))
			continue
		}

		// Skip granting access to organization admins
		if isOrganizationAdmin {
			updatedMemberAccess = append(updatedMemberAccess, SpaceMemberAccess{
				UserUUID:            member.UserUUID,
				SpaceRole:           member.SpaceRole,
				IsOrganizationAdmin: true,
			})
			continue
		}

		// Add or update access
		currentRole, exists := currentMemberUUIDs[member.UserUUID]

		// If role has changed or user doesn't exist, update access
		if !exists || currentRole != member.SpaceRole {
			err = c.spaceService.AddUserToSpace(projectUUID, spaceUUID, member.UserUUID, member.SpaceRole)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to update user %s access: %w", member.UserUUID, err))
				continue
			}
		}

		updatedMemberAccess = append(updatedMemberAccess, SpaceMemberAccess{
			UserUUID:            member.UserUUID,
			SpaceRole:           member.SpaceRole,
			IsOrganizationAdmin: isOrganizationAdmin,
		})
	}

	// Handle group access updates
	currentGroupUUIDs := make(map[string]models.SpaceMemberRole)
	for _, group := range currentSpace.SpaceAccessGroups {
		currentGroupUUIDs[group.GroupUUID] = group.SpaceRole
	}

	newGroupUUIDs := make(map[string]models.SpaceMemberRole)
	for _, group := range newGroupAccess {
		newGroupUUIDs[group.GroupUUID] = group.SpaceRole
	}

	// Process groups to remove
	for groupUUID := range currentGroupUUIDs {
		if _, exists := newGroupUUIDs[groupUUID]; !exists {
			// Check if group exists before removing
			_, err := c.getGroupFromOrganization(groupUUID)
			if err != nil {
				// Skip if group no longer exists
				continue
			}

			// Remove access
			err = c.spaceService.RemoveGroupFromSpace(projectUUID, spaceUUID, groupUUID)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to remove group %s from space: %w", groupUUID, err))
			}
		}
	}

	// Process groups to add or update
	updatedGroupAccess := []SpaceGroupAccess{}
	for _, group := range newGroupAccess {
		// Check if group exists
		_, err := c.getGroupFromOrganization(group.GroupUUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("group %s not found: %w", group.GroupUUID, err))
			continue
		}

		// Add or update access
		currentRole, exists := currentGroupUUIDs[group.GroupUUID]

		// If role has changed or group doesn't exist, update access
		if !exists || currentRole != group.SpaceRole {
			err = c.spaceService.UpdateGroupAccessInSpace(projectUUID, spaceUUID, group.GroupUUID, group.SpaceRole)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to update group %s access: %w", group.GroupUUID, err))
				continue
			}
		}

		updatedGroupAccess = append(updatedGroupAccess, SpaceGroupAccess{
			GroupUUID: group.GroupUUID,
			SpaceRole: group.SpaceRole,
		})
	}

	// Build result
	spaceDetails := &SpaceDetails{
		ProjectUUID:     updatedSpace.ProjectUUID,
		SpaceUUID:       updatedSpace.SpaceUUID,
		ParentSpaceUUID: updatedSpace.ParentSpaceUUID,
		SpaceName:       updatedSpace.SpaceName,
		IsPrivate:       updatedSpace.IsPrivate,
		MemberAccess:    updatedMemberAccess,
		GroupAccess:     updatedGroupAccess,
	}

	return spaceDetails, errors
}

// DeleteSpace deletes a space if deletion protection is disabled
func (c *SpaceController) DeleteSpace(projectUUID string, spaceUUID string, deletionProtection bool) error {
	if deletionProtection {
		return fmt.Errorf("cannot delete space: deletion protection is enabled")
	}

	return c.spaceService.DeleteSpace(projectUUID, spaceUUID)
}

// ImportSpace imports an existing space with its access settings
func (c *SpaceController) ImportSpace(resourceID string) (*SpaceDetails, error) {
	// Extract project and space UUIDs from resource ID
	projectUUID, spaceUUID, err := c.spaceService.ExtractSpaceResourceID(resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract resource ID: %w", err)
	}

	// Get space details
	return c.GetSpace(projectUUID, spaceUUID)
}

// GrantSpaceAccessToMember grants access to a space for a member
func (c *SpaceController) GrantSpaceAccessToMember(projectUUID, spaceUUID, userUUID string, role models.SpaceMemberRole) error {
	// Check if user is an organization admin
	isOrganizationAdmin, err := c.organizationMembersService.IsOrganizationAdmin(userUUID)
	if err != nil {
		return fmt.Errorf("failed to check if user is an organization admin: %w", err)
	}

	// Skip organization admins
	if isOrganizationAdmin {
		return nil
	}

	return c.spaceService.AddUserToSpace(projectUUID, spaceUUID, userUUID, role)
}

// Helper method to retrieve group from organization
func (c *SpaceController) getGroupFromOrganization(groupUUID string) (*models.OrganizationGroup, error) {
	return c.organizationGroupsService.GetGroup(groupUUID)
}
