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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// SpaceController orchestrates operations related to Lightdash spaces.
// It provides a consistent API using options structs for all operations:
// - CreateSpace with CreateSpaceOptions
// - UpdateSpace with UpdateSpaceOptions
// - DeleteSpace with DeleteSpaceOptions
// - ImportSpace with ImportSpaceOptions
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

// CreateSpaceOptions contains all the options for creating a space
type CreateSpaceOptions struct {
	ProjectUUID     string
	SpaceName       string
	IsPrivate       bool
	ParentSpaceUUID *string
	MemberAccess    []SpaceAccessMemberRequest
	GroupAccess     []SpaceGroupAccess
}

// UpdateSpaceOptions contains all the options for updating a space
type UpdateSpaceOptions struct {
	ProjectUUID     string
	SpaceUUID       string
	SpaceName       string
	IsPrivate       *bool
	ParentSpaceUUID *string
	MemberAccess    []SpaceAccessMemberRequest
	GroupAccess     []SpaceGroupAccess
}

// DeleteSpaceOptions contains all the options for deleting a space
type DeleteSpaceOptions struct {
	ProjectUUID        string
	SpaceUUID          string
	DeletionProtection bool
}

// ImportSpaceOptions contains all the options for importing a space
type ImportSpaceOptions struct {
	ResourceID string // Format: "projects/{projectUUID}/spaces/{spaceUUID}"
}

// NewSpaceController creates a new SpaceController
func NewSpaceController(client *api.Client) *SpaceController {
	return &SpaceController{
		spaceService:               services.NewSpaceService(client),
		organizationMembersService: services.NewOrganizationMembersService(client),
		organizationGroupsService:  services.NewOrganizationGroupsService(client),
		projectService:             services.NewProjectService(client),
	}
}

// CreateSpace creates a new space with the specified properties and access settings.
// Access settings (memberAccess and groupAccess) are only applied to root spaces.
func (c *SpaceController) CreateSpace(
	options CreateSpaceOptions,
) (*models.SpaceDetails, []error) {
	// Check if this will be a nested space (has parent space UUID)
	isNestedSpace := options.ParentSpaceUUID != nil

	var createdSpaceDetails *models.SpaceDetails
	var errors []error

	if isNestedSpace {
		createdSpaceDetails, errors = c.createNestedSpace(options)
	} else {
		createdSpaceDetails, errors = c.createRootSpace(options)
	}
	if len(errors) > 0 {
		return nil, errors
	}

	// Get the final space details to return the complete state
	actualCreatedSpaceDetails, err := c.GetSpace(options.ProjectUUID, createdSpaceDetails.SpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get final space details after creation: %w", err)}
	}

	return actualCreatedSpaceDetails, nil
}

// GetSpace retrieves the details of a space by its project and space UUIDs.
// For root spaces, it populates MemberAccess and GroupAccess with all members/groups returned by API.
// Filtering for direct access (for the 'access' attribute in Terraform) is handled in the resource layer.
func (c *SpaceController) GetSpace(projectUUID, spaceUUID string) (*models.SpaceDetails, error) {
	// Get space details from the service layer
	space, err := c.spaceService.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get space: %w", err)
	}

	// Convert API SpaceAccessMember to models.SpaceAccessMember
	spaceAccessMembers := []models.SpaceAccessMember{}
	for _, member := range space.SpaceAccessMembers {
		spaceAccessMembers = append(spaceAccessMembers, models.SpaceAccessMember{
			UserUUID:        member.UserUUID,
			SpaceRole:       member.SpaceRole,
			HasDirectAccess: member.HasDirectAccess,
			InheritedRole:   member.InheritedRole,
			InheritedFrom:   member.InheritedFrom,
			ProjectRole:     member.ProjectRole,
		})
	}

	// Convert API SpaceAccessGroup to models.SpaceAccessGroup
	spaceAccessGroups := []models.SpaceAccessGroup{}
	for _, group := range space.SpaceAccessGroups {
		spaceAccessGroups = append(spaceAccessGroups, models.SpaceAccessGroup{
			GroupUUID: group.GroupUUID,
			SpaceRole: group.SpaceRole,
		})
	}

	// Convert API ChildSpace to models.ChildSpace
	childSpaces := []models.ChildSpace{}
	for _, child := range space.ChildSpaces {
		childSpaces = append(childSpaces, models.ChildSpace{
			SpaceUUID:  child.SpaceUUID,
			SpaceName:  child.Name,
			IsPrivate:  child.IsPrivate,
			AccessList: []models.SpaceAccessMember{}, // API doesn't provide access list for child spaces
		})
	}

	// Build result SpaceDetails object
	spaceDetails := &models.SpaceDetails{
		ProjectUUID:        space.ProjectUUID,
		SpaceUUID:          space.SpaceUUID,
		ParentSpaceUUID:    space.ParentSpaceUUID,
		SpaceName:          space.SpaceName,
		IsPrivate:          space.IsPrivate,
		SpaceAccessMembers: spaceAccessMembers,
		SpaceAccessGroups:  spaceAccessGroups,
		ChildSpaces:        childSpaces,
	}

	return spaceDetails, nil
}

// UpdateSpace updates a space based on whether it's a root or nested space and if its parent is changing.
// It orchestrates calls to specific update/move functions.
func (c *SpaceController) UpdateSpace(
	options UpdateSpaceOptions,
) (*models.SpaceDetails, []error) {
	// Get the current space details to determine if it's a root or nested space.
	currentSpaceDetails, err := c.GetSpace(options.ProjectUUID, options.SpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get current space details for update: %w", err)}
	}

	// Check if the space is currently a root space (ParentSpaceUUID is nil) and if the plan indicates it should become root
	isCurrentlyRootSpace := currentSpaceDetails.ParentSpaceUUID == nil
	isBecomingRootSpace := options.ParentSpaceUUID == nil

	var errors []error

	if isCurrentlyRootSpace && isBecomingRootSpace {
		// Scenario 1: Remains a root space - Update properties and access.
		errors = c.updateRootSpace(
			options.ProjectUUID,
			options.SpaceUUID,
			options.SpaceName,
			options.IsPrivate,
			options.MemberAccess,
			options.GroupAccess,
			currentSpaceDetails,
		)
	} else if isCurrentlyRootSpace && !isBecomingRootSpace {
		// Scenario 2: Root space becoming a nested space - Update name and move.
		// Access controls will be inherited from the new parent and any direct access will be ignored by the API.
		errors = c.moveRootToNestedSpace(
			options.ProjectUUID,
			options.SpaceUUID,
			options.SpaceName,
			options.ParentSpaceUUID,
		)
	} else if !isCurrentlyRootSpace && isBecomingRootSpace {
		// Scenario 3: Nested space becoming a root space - Move to root and then apply access controls.
		// The space will initially inherit project access, and then direct access can be set.
		errors = c.moveNestedToRootSpace(
			options.ProjectUUID,
			options.SpaceUUID,
			options.SpaceName,
			options.IsPrivate,
		)
	} else {
		// Scenario 4: Nested space staying nested (either same parent or different parent).
		// Only name and parent space can be updated via the API for nested spaces.
		// Access controls and privacy are inherited and cannot be managed.
		errors = c.updateNestedSpace(
			options.ProjectUUID,
			options.SpaceUUID,
			options.SpaceName,
			options.ParentSpaceUUID,
		)
	}

	// Return the updated space details converted to Terraform state
	if len(errors) > 0 {
		return nil, errors
	}

	// Get the final space details to return the complete state
	actualUpdatedSpaceDetails, err := c.GetSpace(options.ProjectUUID, options.SpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get final space details after update: %w", err)}
	}

	return actualUpdatedSpaceDetails, nil
}

// DeleteSpace deletes a space if deletion protection is disabled.
func (c *SpaceController) DeleteSpace(options DeleteSpaceOptions) error {
	if options.DeletionProtection {
		return fmt.Errorf("cannot delete space %s: deletion protection is enabled", options.SpaceUUID)
	}

	// Check if the space has any child spaces
	space, err := c.GetSpace(options.ProjectUUID, options.SpaceUUID)
	if err != nil {
		return fmt.Errorf("failed to get space details: %w", err)
	}
	childSpaces := space.ChildSpaces
	if len(childSpaces) > 0 {
		return fmt.Errorf("cannot delete space %s: it has child spaces", options.SpaceUUID)
	}

	// Delete the space via the service layer
	return c.spaceService.DeleteSpace(options.ProjectUUID, options.SpaceUUID)
}

// ImportSpace imports an existing space by its resource ID.
// It retrieves the space details and access settings.
func (c *SpaceController) ImportSpace(options ImportSpaceOptions) (*models.SpaceDetails, error) {
	// Extract project and space UUIDs from the resource ID string
	projectUUID, spaceUUID, err := c.spaceService.ExtractSpaceResourceID(options.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract resource ID %s: %w", options.ResourceID, err)
	}

	// Get space details via the service layer
	spaceDetails, err := c.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get space details: %w", err)
	}

	return spaceDetails, nil
}

// --- Private Helper Methods ---

// createRootSpace creates a new root-level space and manages its direct access settings.
func (c *SpaceController) createRootSpace(
	options CreateSpaceOptions,
) (*models.SpaceDetails, []error) {
	// 1. Validate the input
	validationErrors := c.validateSpaceCreation(options)
	if len(validationErrors) > 0 {
		return nil, validationErrors
	}

	// 2. Create the space via the service layer
	createdSpace, err := c.spaceService.CreateSpace(
		options.ProjectUUID,
		options.SpaceName,
		options.IsPrivate,
		nil, // nil parentSpaceUUID for root space
	)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to create space: %w", err)}
	}

	// 3. Manage access for the root-level space after creation.
	accessErrors := c.manageSpaceAccess(
		options.ProjectUUID,
		createdSpace.SpaceUUID,
		options.MemberAccess,
		options.GroupAccess,
	)

	if len(accessErrors) > 0 {
		// If access management fails, try to clean up by deleting the space
		errDel := c.spaceService.DeleteSpace(options.ProjectUUID, createdSpace.SpaceUUID)
		if errDel != nil {
			accessErrors = append(accessErrors, fmt.Errorf("failed to delete space after access management failure: %w", errDel))
		}
		return nil, accessErrors
	}

	return createdSpace, nil
}

// createNestedSpace creates a new nested space. Access controls are inherited from the parent.
// isPrivate will be ignored by Lightdash for nested spaces as privacy is inherited.
// memberAccess and groupAccess will be ignored by Lightdash for nested spaces as access is inherited.
func (c *SpaceController) createNestedSpace(
	options CreateSpaceOptions,
) (*models.SpaceDetails, []error) {
	// 1. Validate inputs - ensure no space access is specified for nested spaces
	var errors []error
	if len(options.MemberAccess) > 0 {
		errors = append(errors, fmt.Errorf("cannot manage member access for nested space %s: access is inherited from parent", options.SpaceName))
	}
	if len(options.GroupAccess) > 0 {
		errors = append(errors, fmt.Errorf("cannot manage group access for nested space %s: access is inherited from parent", options.SpaceName))
	}
	if len(errors) > 0 {
		return nil, errors
	}

	// 2. Create the space via the service layer. Note that isPrivate, memberAccess, and groupAccess
	// are effectively ignored by Lightdash for nested spaces as they inherit these from the parent.
	createdSpace, err := c.spaceService.CreateSpace(options.ProjectUUID, options.SpaceName, options.IsPrivate, options.ParentSpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to create nested space: %w", err)}
	}

	return createdSpace, nil
}

// validateSpaceCreation validates space creation parameters
func (c *SpaceController) validateSpaceCreation(options CreateSpaceOptions) []error {
	var errors []error

	// 1.1 Check if the member can become a space member (must be project member, not org admin)
	projectMembers, err := c.projectService.GetProjectMembers(options.ProjectUUID)
	if err != nil {
		return []error{fmt.Errorf("failed to get project members: %w", err)}
	}

	for _, member := range options.MemberAccess {
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
	for _, group := range options.GroupAccess {
		_, err := c.organizationGroupsService.GetGroup(group.GroupUUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("group %s not found: %w", group.GroupUUID, err))
			continue
		}
	}

	return errors
}

// manageSpaceAccess handles adding members and groups to a space
func (c *SpaceController) manageSpaceAccess(
	projectUUID string,
	spaceUUID string,
	memberAccess []SpaceAccessMemberRequest,
	groupAccess []SpaceGroupAccess,
) []error {
	var errors []error

	// 1. Get the actual space details to check existing access
	actualSpace, err := c.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		return []error{fmt.Errorf("failed to get space details: %w", err)}
	}
	// If the space is a nested space, we cannot manage access via the API
	// because it is inherited from the root space
	if actualSpace.ParentSpaceUUID != nil {
		return []error{fmt.Errorf("space %s is a nested space and cannot have direct access", spaceUUID)}
	}

	// 2. Add groups to the space
	for _, group := range groupAccess {
		err = c.spaceService.AddGroupToSpace(
			projectUUID,
			spaceUUID,
			group.GroupUUID,
			group.SpaceRole,
		)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to add group %s to space: %w", group.GroupUUID, err))
		}
	}

	// 3. Add members to the space if they don't already have access
	for _, member := range memberAccess {
		// Check if the member already has space access through any means (direct, group, etc.)
		existingMember := actualSpace.GetMemberByUUID(member.UserUUID)

		// If member doesn't have access or has different access level, add/update it
		if existingMember == nil || string(existingMember.SpaceRole) != string(member.SpaceRole) {
			err := c.spaceService.AddUserToSpace(
				projectUUID,
				spaceUUID,
				member.UserUUID,
				member.SpaceRole,
			)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to add user %s to space: %w", member.UserUUID, err))
			}
		}
	}

	return errors
}

// manageRootSpaceMemberAccess handles adding, updating, and removing direct member access for a root space.
func (c *SpaceController) manageRootSpaceMemberAccess(
	projectUUID string,
	spaceUUID string,
	newMemberAccess []SpaceAccessMemberRequest,
	currentMemberAccess []SpaceAccessMemberResponse,
) []error {
	var errors []error

	currentDirectMemberMap := make(map[string]SpaceAccessMemberResponse)
	for _, currentMember := range currentMemberAccess {
		accessType := currentMember.GetSpaceAccessType()
		if currentMember.HasDirectAccess != nil && *currentMember.HasDirectAccess && accessType != nil && *accessType == "member" {
			currentDirectMemberMap[currentMember.UserUUID] = currentMember
		}
	}

	newDirectMemberMap := make(map[string]SpaceAccessMemberRequest)
	for _, member := range newMemberAccess {
		// Assume presence in the newMemberAccess list implies the intention to manage this member directly.
		newDirectMemberMap[member.UserUUID] = member
	}

	// Process members to remove direct access
	for userUUID := range currentDirectMemberMap {
		if _, exists := newDirectMemberMap[userUUID]; !exists {
			err := c.spaceService.RemoveUserFromSpace(projectUUID, spaceUUID, userUUID)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to remove direct access for user %s from space %s: %w", userUUID, spaceUUID, err))
			}
		}
	}

	// Process members to add or update direct access
	for userUUID, newMember := range newDirectMemberMap {
		currentMember, exists := currentDirectMemberMap[userUUID]

		// If user doesn't have direct access currently, or their role has changed, add or update their direct access.
		// We compare the SpaceRole from the current direct access (SpaceAccessMemberResponse) with the role from the new plan (SpaceAccessMemberRequest).
		if !exists || currentMember.SpaceRole != newMember.SpaceRole {
			err := c.spaceService.AddUserToSpace(projectUUID, spaceUUID, userUUID, newMember.SpaceRole)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to add/update direct access for user %s to space %s: %w", userUUID, spaceUUID, err))
			}
		}
	}

	return errors
}

// manageRootSpaceGroupAccess handles adding, updating, and removing group access for a root space.
func (c *SpaceController) manageRootSpaceGroupAccess(
	projectUUID string,
	spaceUUID string,
	newGroupAccess []SpaceGroupAccess,
	currentGroupAccess []SpaceGroupAccess,
) []error {
	var errors []error

	currentGroupMap := make(map[string]SpaceGroupAccess)
	for _, group := range currentGroupAccess {
		currentGroupMap[group.GroupUUID] = group
	}

	newGroupMap := make(map[string]SpaceGroupAccess)
	for _, group := range newGroupAccess {
		newGroupMap[group.GroupUUID] = group
	}

	// Process groups to remove
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

	// Process groups to add or update
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
			err := c.spaceService.UpdateGroupAccessInSpace(projectUUID, spaceUUID, groupUUID, group.SpaceRole)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to update group %s access in space %s: %w", groupUUID, spaceUUID, err))
			}
		}
	}

	return errors
}

// updateRootSpace updates the properties and access settings for a root-level space.
// This is used when a space remains a root space during an update.
func (c *SpaceController) updateRootSpace(
	projectUUID string,
	spaceUUID string,
	spaceName string,
	isPrivate *bool,
	newMemberAccess []SpaceAccessMemberRequest,
	newGroupAccess []SpaceGroupAccess,
	currentSpaceDetails *models.SpaceDetails,
) []error {
	var errors []error

	// 1. Update the space properties via the service layer
	_, err := c.spaceService.UpdateRootSpace(projectUUID, spaceUUID, spaceName, isPrivate)
	if err != nil {
		return []error{fmt.Errorf("failed to update space properties: %w", err)}
	}

	// Convert models.SpaceAccessMember to SpaceAccessMemberResponse for member access management
	currentProcessedMemberAccess := []SpaceAccessMemberResponse{}
	for _, member := range currentSpaceDetails.SpaceAccessMembers {
		hasDirectAccess := member.HasDirectAccess
		inheritedRole := member.InheritedRole
		inheritedFrom := member.InheritedFrom
		projectRole := member.ProjectRole

		currentProcessedMemberAccess = append(currentProcessedMemberAccess, SpaceAccessMemberResponse{
			BaseSpaceAccessMember: BaseSpaceAccessMember{
				UserUUID:  member.UserUUID,
				SpaceRole: member.SpaceRole,
			},
			HasDirectAccess: &hasDirectAccess,
			InheritedRole:   &inheritedRole,
			InheritedFrom:   &inheritedFrom,
			ProjectRole:     &projectRole,
		})
	}

	// Convert models.SpaceAccessGroup to SpaceGroupAccess for group access management
	currentProcessedGroupAccess := []SpaceGroupAccess{}
	for _, group := range currentSpaceDetails.SpaceAccessGroups {
		currentProcessedGroupAccess = append(currentProcessedGroupAccess, SpaceGroupAccess{
			GroupUUID: group.GroupUUID,
			SpaceRole: group.SpaceRole,
		})
	}

	// 2. Manage member access (add/update/remove direct access)
	memberErrors := c.manageRootSpaceMemberAccess(
		projectUUID,
		spaceUUID,
		newMemberAccess,
		currentProcessedMemberAccess,
	)
	errors = append(errors, memberErrors...)

	// 3. Handle group access updates (add/update/remove groups)
	groupErrors := c.manageRootSpaceGroupAccess(
		projectUUID,
		spaceUUID,
		newGroupAccess,
		currentProcessedGroupAccess,
	)
	errors = append(errors, groupErrors...)

	return errors
}

// updateNestedSpace updates the properties for a nested space.
// Only the name and parent space UUID can be changed for nested spaces via the API.
// Access controls and privacy are inherited and cannot be managed by this function.
func (c *SpaceController) updateNestedSpace(
	projectUUID string,
	spaceUUID string,
	spaceName string,
	parentSpaceUUID *string,
) []error {
	// Update only the name and parent space UUID for nested spaces via the service layer
	// isPrivate is passed as nil because it cannot be updated for nested spaces.
	// Pass the parentSpaceUUID to the service layer to handle moves between nested spaces
	_, err := c.spaceService.UpdateNestedSpace(projectUUID, spaceUUID, spaceName, parentSpaceUUID)
	if err != nil {
		return []error{fmt.Errorf("failed to update nested space %s: %w", spaceUUID, err)}
	}

	// Move the space to the new parent space
	err = c.spaceService.MoveSpace(projectUUID, spaceUUID, parentSpaceUUID)
	if err != nil {
		return []error{fmt.Errorf("failed to move nested space %s to parent %s: %w", spaceUUID, *parentSpaceUUID, err)}
	}

	return nil
}

// moveRootToNestedSpace handles moving a root space to become a nested space.
// This involves updating the parent space UUID.
// Access controls will be inherited from the new parent.
func (c *SpaceController) moveRootToNestedSpace(
	projectUUID string,
	spaceUUID string,
	spaceName string,
	parentSpaceUUID *string,
) []error {
	// Update name and move to parent via the service layer
	_, err := c.spaceService.UpdateRootSpace(projectUUID, spaceUUID, spaceName, nil)
	if err != nil {
		return []error{fmt.Errorf("failed to move root space %s to nested space under parent %s: %w", spaceUUID, *parentSpaceUUID, err)}
	}

	return nil
}

// moveNestedToRootSpace handles moving a nested space to become a root space.
// This involves updating the parent space UUID to nil and then applying explicit access controls.
func (c *SpaceController) moveNestedToRootSpace(
	projectUUID string,
	spaceUUID string,
	spaceName string,
	isPrivate *bool,
) []error {
	// 1. Move the space to the root space via the service layer
	err1 := c.spaceService.MoveSpace(projectUUID, spaceUUID, nil)
	if err1 != nil {
		return []error{fmt.Errorf("failed to move nested space %s to root: %w", spaceUUID, err1)}
	}

	// 2. Update the space to make it a root space (no parent) via the service layer
	_, err2 := c.spaceService.UpdateRootSpace(projectUUID, spaceUUID, spaceName, isPrivate)
	if err2 != nil {
		return []error{fmt.Errorf("failed to move nested space %s to root: %w", spaceUUID, err2)}
	}

	return nil
}
