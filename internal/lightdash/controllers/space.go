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
	projectService             *services.ProjectService
}

// BaseSpaceAccessMember represents the core information for a space access member.
type BaseSpaceAccessMember struct {
	UserUUID  string
	SpaceRole models.SpaceMemberRole
}

// SpaceAccessMemberRequest represents a request to add or update space access for a member.
type SpaceAccessMemberRequest struct {
	BaseSpaceAccessMember
	IsOrganizationAdmin bool // Indicates if the user is an organization admin (cannot be added as a direct space member)
}

// SpaceAccessMemberResponse represents the response details for a space access member.
// This includes how access is granted (direct, inherited, etc.).
type SpaceAccessMemberResponse struct {
	BaseSpaceAccessMember
	HasDirectAccess *bool   // Whether the user has direct access to the space
	InheritedRole   *string // The role inherited from an upper level (org, group)
	InheritedFrom   *string // The source of the inherited role (e.g., "organization", "group")
	ProjectRole     *string // The user's role within the associated project
}

// GetSpaceAccessType returns the type of space access for a member.
// It returns "member" for direct access, "group" for group-inherited access, or nil.
// Note: Organization admin access is not represented by this function.
func (s *SpaceAccessMemberResponse) GetSpaceAccessType() *string {
	// No direct access
	if s.HasDirectAccess == nil || !*s.HasDirectAccess {
		return nil
	}

	// Group space access
	if s.InheritedFrom != nil && *s.InheritedFrom == "group" {
		group := "group"
		return &group
	}
	// Individual space access member
	member := "member"
	return &member
}

// SpaceGroupAccess represents a group's access to a space
type SpaceGroupAccess struct {
	GroupUUID string
	SpaceRole models.SpaceMemberRole // The role the group has in the space
}

// SpaceDetails contains all the details of a space returned by the GetSpace API.
// Note: For nested spaces, MemberAccess and GroupAccess lists will be empty as access is inherited.
type SpaceDetails struct {
	ProjectUUID     string
	SpaceUUID       string
	ParentSpaceUUID *string // UUID of the parent space, nil for root spaces
	SpaceName       string
	IsPrivate       bool                        // Whether the space is private
	CreatedAt       time.Time                   // Timestamp of space creation (Note: This field is managed by Terraform state, not directly returned by GetSpace API)
	MemberAccess    []SpaceAccessMemberResponse // List of members with direct access (only for root spaces)
	GroupAccess     []SpaceGroupAccess          // List of groups with access (only for root spaces)
}

func (s *SpaceDetails) GetSpaceAccessType(userUUID string) (*string, error) {
	for _, member := range s.MemberAccess {
		if member.UserUUID == userUUID {
			return member.GetSpaceAccessType(), nil
		}
	}
	return nil, fmt.Errorf("user %s not found in space %s", userUUID, s.SpaceUUID)
}

// NewSpaceController creates a new SpaceController
func NewSpaceController(client *api.Client) *SpaceController {
	return &SpaceController{
		spaceService: services.NewSpaceService(client),

		organizationMembersService: services.NewOrganizationMembersService(client),
		organizationGroupsService:  services.NewOrganizationGroupsService(client),

		projectService: services.NewProjectService(client),
	}
}

// CreateSpace creates a new space with the specified properties and access settings.
// Access settings (memberAccess and groupAccess) are only applied to root spaces.
func (c *SpaceController) CreateSpace(
	projectUUID string,
	spaceName string,
	isPrivate bool,
	parentSpaceUUID *string,
	memberAccess []SpaceAccessMemberRequest,
	groupAccess []SpaceGroupAccess,
) (*SpaceDetails, []error) {
	var errors []error

	// Check if this will be a nested space (has parent space UUID)
	isNestedSpace := parentSpaceUUID != nil

	// 1. Validate the input
	// Only validate member and group access for root spaces since nested spaces inherit access
	if !isNestedSpace {
		// 1.1 Check if the member can become a space member (must be project member, not org admin)
		projectMembers, err := c.projectService.GetProjectMembers(projectUUID)
		if err != nil {
			return nil, []error{fmt.Errorf("failed to get project members: %w", err)}
		}
		for _, member := range memberAccess {
			// Check if the member is a project member
			isProjectMember := false
			for _, projectMember := range projectMembers {
				if projectMember.UserUUID == member.UserUUID {
					isProjectMember = true
					break
				}
			}
			if !isProjectMember {
				errors = append(errors, fmt.Errorf("user %s is not a project member", member.UserUUID))
				continue
			}

			// Check if the member is an organization admin (org admins have implicit access and cannot be added explicitly)
			if member.IsOrganizationAdmin {
				errors = append(errors, fmt.Errorf("user %s is an organization admin, so they shouldn't be added as a space member", member.UserUUID))
				continue
			}
		}
		// 1.2 Check if the groups exist in the organization
		for _, group := range groupAccess {
			_, err := c.organizationGroupsService.GetGroup(group.GroupUUID)
			if err != nil {
				errors = append(errors, fmt.Errorf("group %s not found: %w", group.GroupUUID, err))
				continue
			}
		}
		if len(errors) > 0 {
			return nil, errors
		}
	}

	// 2. Create the space via the service layer
	createdSpace, err := c.spaceService.CreateSpace(projectUUID, spaceName, isPrivate, parentSpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to create space: %w", err)}
	}

	// 3. Only manage access for root-level spaces after creation.
	// For nested spaces, access is inherited from the parent space and cannot be set here.
	if !isNestedSpace {
		// 3.1 Grant space access to groups
		for _, group := range groupAccess {
			err = c.spaceService.AddGroupToSpace(projectUUID, createdSpace.SpaceUUID, group.GroupUUID, group.SpaceRole)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to add group %s to space: %w", group.GroupUUID, err))
			}
		}

		// 3.2 Grant space access to members
		// NOTE: We need to get the actual space details *after* creation to check if the member already has space access
		// (e.g., via group membership or other inherited means) before attempting to add them explicitly.
		actualSpace, err := c.GetSpace(projectUUID, createdSpace.SpaceUUID)
		if err != nil {
			// If we fail to get the space after creation, we should attempt to delete it.
			errDel := c.spaceService.DeleteSpace(projectUUID, createdSpace.SpaceUUID)
			if errDel != nil {
				errors = append(errors, fmt.Errorf("failed to delete partially created space %s after GetSpace failure: %w", createdSpace.SpaceUUID, errDel))
			}
			errors = append(errors, fmt.Errorf("failed to get space after creation: %w", err))
			return nil, errors
		}
		for _, member := range memberAccess {
			// Check if the member already has space access through any means (direct, group, etc.)
			memberHasAccess := false
			for _, actualSpaceAccess := range actualSpace.MemberAccess {
				if actualSpaceAccess.UserUUID == member.UserUUID {
					memberHasAccess = true
					// TODO: Consider logging a warning if access already exists but role differs from plan?
					break
				}
			}

			if !memberHasAccess {
				// Grant space access to the member via the service layer
				err = c.spaceService.AddUserToSpace(projectUUID, createdSpace.SpaceUUID, member.UserUUID, member.SpaceRole)
				if err != nil {
					errors = append(errors, fmt.Errorf("failed to add user %s to space: %w", member.UserUUID, err))
				}
			}
		}
	}

	// 4. (Optional) Delete the space if failing at any after creation step
	if len(errors) > 0 {
		err = c.spaceService.DeleteSpace(projectUUID, createdSpace.SpaceUUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to delete space %s after access management failure: %w", createdSpace.SpaceUUID, err))
		}
		errors = append(errors, fmt.Errorf("failed to create space with specified access: %w", errors[0])) // Add a general creation failure error
		return nil, errors
	}

	// Get the final space details to return the complete state
	// Note: The GetSpace method already filters members to only include direct access for root spaces.
	finalSpace, err := c.GetSpace(projectUUID, createdSpace.SpaceUUID)
	if err != nil {
		// If we successfully created but fail to get the final state, it's an issue.
		errors = append(errors, fmt.Errorf("failed to get final space details after creation and access management: %w", err))
		// Consider attempting deletion here too? Maybe too risky. Just return the error.
		return nil, errors
	}

	return finalSpace, nil
}

// GetSpace retrieves the details of a space by its project and space UUIDs.
// For root spaces, it populates MemberAccess and GroupAccess with directly managed access.
// For nested spaces, these lists will be empty.
func (c *SpaceController) GetSpace(projectUUID, spaceUUID string) (*SpaceDetails, error) {
	// Get space details from the service layer
	space, err := c.spaceService.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get space: %w", err)
	}

	// Determine if this is a nested space based on the parent UUID
	isNestedSpace := space.ParentSpaceUUID != nil

	// Process members
	memberAccessList := []SpaceAccessMemberResponse{}

	// Only process and filter member access for root spaces. Nested spaces inherit access.
	if !isNestedSpace {
		for _, member := range space.SpaceAccessMembers {
			// Convert to SpaceAccessMemberResponse from models.SpaceAccessMember
			spaceAccessMemberResponse := SpaceAccessMemberResponse{
				BaseSpaceAccessMember: BaseSpaceAccessMember{
					UserUUID:  member.UserUUID,
					SpaceRole: member.SpaceRole,
				},
				HasDirectAccess: &member.HasDirectAccess,
				InheritedRole:   &member.InheritedRole,
				InheritedFrom:   &member.InheritedFrom,
				ProjectRole:     &member.ProjectRole,
			}

			// Filter for members who have direct access based on the API response.
			// We only want to represent directly managed access in Terraform state.
			accessType := spaceAccessMemberResponse.GetSpaceAccessType()
			if accessType != nil && *accessType == "member" {
				memberAccessList = append(memberAccessList, spaceAccessMemberResponse)
			}
		}
	}

	// Process group access
	groupAccessList := []SpaceGroupAccess{}

	// Only process group access for root spaces. Nested spaces inherit access.
	if !isNestedSpace {
		for _, group := range space.SpaceAccessGroups {
			groupAccessList = append(groupAccessList, SpaceGroupAccess{
				GroupUUID: group.GroupUUID,
				SpaceRole: group.SpaceRole,
			})
		}
	}

	// Build result SpaceDetails object
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

// UpdateSpace updates a space based on whether it's a root or nested space and if its parent is changing.
// It orchestrates calls to specific update/move functions.
func (c *SpaceController) UpdateSpace(
	projectUUID string,
	spaceUUID string,
	spaceName string,
	isPrivate *bool,
	parentSpaceUUID *string,
	newMemberAccess []SpaceAccessMemberRequest,
	newGroupAccess []SpaceGroupAccess,
) (*SpaceDetails, []error) {
	// Get the current space details to determine if it's a root or nested space.
	currentSpaceDetails, err := c.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get current space details for update: %w", err)}
	}

	// Check if the space is currently a root space (ParentSpaceUUID is nil) and if the plan indicates it should become root
	isCurrentlyRootSpace := currentSpaceDetails.ParentSpaceUUID == nil
	isBecomingRootSpace := parentSpaceUUID == nil

	if isCurrentlyRootSpace && isBecomingRootSpace {
		// Scenario 1: Remains a root space - Update properties and access.
		return c.UpdateRootSpace(
			projectUUID,
			spaceUUID,
			spaceName,
			isPrivate,
			nil, // explicitly nil as it remains a root space
			newMemberAccess,
			newGroupAccess,
			currentSpaceDetails,
		)
	} else if isCurrentlyRootSpace && !isBecomingRootSpace {
		// Scenario 2: Root space becoming a nested space - Update name and move.
		// Access controls will be inherited from the new parent and any direct access will be ignored by the API.
		return c.MoveRootToNestedSpace(
			projectUUID,
			spaceUUID,
			spaceName,
			parentSpaceUUID,
			currentSpaceDetails,
		)
	} else if !isCurrentlyRootSpace && isBecomingRootSpace {
		// Scenario 3: Nested space becoming a root space - Move to root and then apply access controls.
		// The space will initially inherit project access, and then direct access can be set.
		return c.MoveNestedToRootSpace(
			projectUUID,
			spaceUUID,
			spaceName,
			isPrivate,
			newMemberAccess,
			newGroupAccess,
			currentSpaceDetails,
		)
	} else {
		// Scenario 4: Nested space staying nested (either same parent or different parent).
		// Only name and parent space can be updated via the API for nested spaces.
		// Access controls and privacy are inherited and cannot be managed.
		return c.UpdateNestedSpace(
			projectUUID,
			spaceUUID,
			spaceName,
			parentSpaceUUID,
		)
	}
}

// UpdateRootSpace updates the properties and access settings for a root-level space.
// This is used when a space remains a root space during an update.
func (c *SpaceController) UpdateRootSpace(
	projectUUID string,
	spaceUUID string,
	spaceName string,
	isPrivate *bool,
	parentSpaceUUID *string, // Should be nil for root spaces - explicitly set to nil in UpdateSpace calls this function.
	newMemberAccess []SpaceAccessMemberRequest,
	newGroupAccess []SpaceGroupAccess,
	currentSpaceDetails *SpaceDetails,
) (*SpaceDetails, []error) {
	var errors []error

	// 1. Update the space properties via the service layer
	_, err := c.spaceService.UpdateSpace(projectUUID, spaceUUID, spaceName, isPrivate, parentSpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to update space properties: %w", err)}
	}

	// 2. Manage member access (add/update/remove direct access)
	// Create maps of *directly* managed member access from current state and new plan
	currentDirectMemberMap := make(map[string]SpaceAccessMemberResponse)
	for _, member := range currentSpaceDetails.MemberAccess {
		// Only consider members that currently have direct 'member' access (as identified by GetSpaceAccessType)
		accessType := member.GetSpaceAccessType()
		if member.HasDirectAccess != nil && *member.HasDirectAccess && accessType != nil && *accessType == "member" {
			currentDirectMemberMap[member.UserUUID] = member
		}
	}

	newDirectMemberMap := make(map[string]SpaceAccessMemberRequest)
	for _, member := range newMemberAccess {
		// Assume presence in the newMemberAccess list implies the intention to manage this member directly.
		newDirectMemberMap[member.UserUUID] = member
	}

	// 2.1. Process members to remove direct access
	// If a user is in the current direct access map but not in the new direct access map, remove their direct access.
	for userUUID := range currentDirectMemberMap {
		if _, exists := newDirectMemberMap[userUUID]; !exists {
			// Remove direct access via API
			err = c.spaceService.RemoveUserFromSpace(projectUUID, spaceUUID, userUUID)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to remove direct access for user %s from space %s: %w", userUUID, spaceUUID, err))
			}
		}
	}

	// 2.2. Process members to add or update direct access
	// If a user is in the new direct access map: If they are not in the current direct access map, add their access.
	// If they are in the current direct access map and the role has changed, update their access.
	for userUUID, newMember := range newDirectMemberMap {
		currentMember, exists := currentDirectMemberMap[userUUID]

		// If user doesn't have direct access currently, or their role has changed, add or update their direct access.
		// We compare the SpaceRole from the current direct access (SpaceAccessMemberResponse) with the role from the new plan (SpaceAccessMemberRequest).
		if !exists || currentMember.SpaceRole != newMember.SpaceRole {
			err = c.spaceService.AddUserToSpace(projectUUID, spaceUUID, userUUID, newMember.SpaceRole)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to add/update direct access for user %s to space %s: %w", userUUID, spaceUUID, err))
			}
		}
	}

	// 3. Handle group access updates (add/update/remove groups)
	currentGroupMap := make(map[string]SpaceGroupAccess)
	for _, group := range currentSpaceDetails.GroupAccess {
		currentGroupMap[group.GroupUUID] = group
	}

	newGroupMap := make(map[string]SpaceGroupAccess)
	for _, group := range newGroupAccess {
		newGroupMap[group.GroupUUID] = group
	}

	// 3.1. Process groups to remove
	for groupUUID := range currentGroupMap {
		if _, exists := newGroupMap[groupUUID]; !exists {
			// Check if group exists in Lightdash before attempting to remove access.
			// This prevents errors if the group was deleted outside of Terraform.
			_, err := c.organizationGroupsService.GetGroup(groupUUID)
			if err != nil {
				// Skip if group no longer exists in Lightdash
				// TODO: Consider logging a warning here?
				continue
			}

			// Remove access via API
			err = c.spaceService.RemoveGroupFromSpace(projectUUID, spaceUUID, groupUUID)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to remove group %s from space %s: %w", groupUUID, spaceUUID, err))
			}
		}
	}

	// 3.2. Process groups to add or update
	for groupUUID, group := range newGroupMap {
		// Check if group exists in Lightdash (important for adding/updating)
		_, err := c.organizationGroupsService.GetGroup(groupUUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("group %s not found in Lightdash: %w", groupUUID, err))
			continue
		}

		// Add or update access via API
		currentGroup, exists := currentGroupMap[groupUUID]

		// If role has changed or group doesn't exist in current state, add/update access
		if !exists || currentGroup.SpaceRole != group.SpaceRole {
			err = c.spaceService.UpdateGroupAccessInSpace(projectUUID, spaceUUID, groupUUID, group.SpaceRole)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to update group %s access in space %s: %w", groupUUID, spaceUUID, err))
			}
		}
	}

	// 4. After updating properties and managing direct access, fetch the complete space details from the API.
	// This is crucial to get the updated state, including all inherited access, although only direct access is returned for root spaces by GetSpace.
	finalSpaceDetails, err := c.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to retrieve space details after root space update: %w", err))
		// If fetching the final state fails, we should still return any previous errors
		return nil, errors
	}

	return finalSpaceDetails, errors
}

// MoveRootToNestedSpace handles moving a root space to become a nested space.
// This involves updating the parent space UUID.
// Access controls will be inherited from the new parent.
func (c *SpaceController) MoveRootToNestedSpace(
	projectUUID string,
	spaceUUID string,
	spaceName string,
	parentSpaceUUID *string,
	currentSpaceDetails *SpaceDetails, // nolint: govet
) (*SpaceDetails, []error) {
	// Update name and move to parent via the service layer
	_, err := c.spaceService.UpdateSpace(projectUUID, spaceUUID, spaceName, nil, parentSpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to move root space %s to nested space under parent %s: %w", spaceUUID, *parentSpaceUUID, err)}
	}

	// After moving to nested, get final space details to return the updated state.
	// Note: Access lists will be empty as per GetSpace for nested spaces.
	finalSpaceDetails, err := c.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to retrieve space details %s after moving to nested: %w", spaceUUID, err)}
	}

	return finalSpaceDetails, nil
}

// MoveNestedToRootSpace handles moving a nested space to become a root space.
// This involves updating the parent space UUID to nil and then applying explicit access controls.
func (c *SpaceController) MoveNestedToRootSpace(
	projectUUID string,
	spaceUUID string,
	spaceName string,
	isPrivate *bool,
	newMemberAccess []SpaceAccessMemberRequest,
	newGroupAccess []SpaceGroupAccess,
	currentSpaceDetails *SpaceDetails,
) (*SpaceDetails, []error) {
	// 1. First, update the space to make it a root space (no parent) via the service layer
	_, err := c.spaceService.UpdateSpace(projectUUID, spaceUUID, spaceName, isPrivate, nil)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to move nested space %s to root: %w", spaceUUID, err)}
	}

	// 2. Now, set up the access permissions since it's becoming a root space.
	// Get the updated space details first to pass to UpdateRootSpace.
	updatedSpaceDetails, err := c.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get space details %s after moving to root: %w", spaceUUID, err)}
	}

	// Use UpdateRootSpace to manage member and group access.
	return c.UpdateRootSpace(
		projectUUID,
		spaceUUID,
		spaceName,
		isPrivate,
		nil, // explicitly nil as it's now a root space
		newMemberAccess,
		newGroupAccess,
		updatedSpaceDetails,
	)
}

// UpdateNestedSpace updates the properties for a nested space.
// Only the name and parent space UUID can be changed for nested spaces via the API.
// Access controls and privacy are inherited and cannot be managed by this function.
func (c *SpaceController) UpdateNestedSpace(
	projectUUID string,
	spaceUUID string,
	spaceName string,
	parentSpaceUUID *string,
) (*SpaceDetails, []error) {
	// Update only the name and parent space UUID for nested spaces via the service layer
	// isPrivate is passed as nil because it cannot be updated for nested spaces.
	_, err := c.spaceService.UpdateSpace(projectUUID, spaceUUID, spaceName, nil, parentSpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to update nested space %s: %w", spaceUUID, err)}
	}

	// Retrieve the updated space details to return the final state.
	// Note: Access lists will be empty as per GetSpace for nested spaces.
	updatedSpaceDetails, err := c.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to retrieve updated nested space details %s: %w", spaceUUID, err)}
	}

	return updatedSpaceDetails, nil
}

// DeleteSpace deletes a space if deletion protection is disabled.
func (c *SpaceController) DeleteSpace(projectUUID string, spaceUUID string, deletionProtection bool) error {
	if deletionProtection {
		return fmt.Errorf("cannot delete space %s: deletion protection is enabled", spaceUUID)
	}

	// Delete the space via the service layer
	return c.spaceService.DeleteSpace(projectUUID, spaceUUID)
}

// ImportSpace imports an existing space by its resource ID.
// It retrieves the space details and access settings.
func (c *SpaceController) ImportSpace(resourceID string) (*SpaceDetails, error) {
	// Extract project and space UUIDs from the resource ID string
	projectUUID, spaceUUID, err := c.spaceService.ExtractSpaceResourceID(resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract resource ID %s: %w", resourceID, err)
	}

	// Get space details via the service layer
	return c.GetSpace(projectUUID, spaceUUID)
}

// GrantSpaceAccessToMember grants direct access to a space for a member.
// This operation is only supported for root spaces.
func (c *SpaceController) GrantSpaceAccessToMember(projectUUID, spaceUUID, userUUID string, role models.SpaceMemberRole) error {
	// Check if the space is a nested space by getting its details
	spaceDetails, err := c.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		return fmt.Errorf("failed to get space details for access grant: %w", err)
	}

	// Cannot manage access for nested spaces - return an error if parent_space_uuid is set.
	if spaceDetails.ParentSpaceUUID != nil {
		return fmt.Errorf("cannot manage access for nested spaces: space %s is a nested space", spaceUUID)
	}

	// Grant access via the service layer
	return c.spaceService.AddUserToSpace(projectUUID, spaceUUID, userUUID, role)
}
