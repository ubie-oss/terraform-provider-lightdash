// Copyright 2025 Ubie, inc.
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
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"

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

// NewSpaceController creates a new SpaceController
func NewSpaceController(client *api.Client) *SpaceController {
	return &SpaceController{
		spaceService:               services.NewSpaceService(client),
		organizationMembersService: services.GetOrganizationMembersService(client),
		organizationGroupsService:  services.NewOrganizationGroupsService(client),
		projectService:             services.NewProjectService(client),
	}
}

// CreateSpaceOptions contains all the options for creating a space
type CreateSpaceOptions struct {
	ProjectUUID     string
	SpaceName       string
	IsPrivate       *bool
	ParentSpaceUUID *string
	MemberAccess    []models.SpaceAccessMember
	GroupAccess     []models.SpaceAccessGroup
}

func (o *CreateSpaceOptions) IsNestedSpace() bool {
	return o.ParentSpaceUUID != nil && *o.ParentSpaceUUID != ""
}

// UpdateSpaceOptions contains all the options for updating a space
type UpdateSpaceOptions struct {
	ProjectUUID     string
	SpaceUUID       string
	SpaceName       string
	IsPrivate       *bool
	ParentSpaceUUID *string
	MemberAccess    []models.SpaceAccessMember
	GroupAccess     []models.SpaceAccessGroup
}

func (o *UpdateSpaceOptions) IsNestedSpace() bool {
	return o.ParentSpaceUUID != nil && *o.ParentSpaceUUID != ""
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

// CreateSpace creates a new space with the specified properties and access settings.
// Access settings (memberAccess and groupAccess) are only applied to root spaces.
func (c *SpaceController) CreateSpace(
	ctx context.Context,
	options CreateSpaceOptions,
) (*models.SpaceDetails, []error) {
	tflog.Debug(ctx, "(SpaceController.CreateSpace) Creating space", map[string]interface{}{
		"options": options,
	})

	var createdSpaceDetails *models.SpaceDetails
	var errors []error

	if options.IsNestedSpace() {
		createdSpaceDetails, errors = c.createNestedSpace(ctx, options)
	} else {
		createdSpaceDetails, errors = c.createRootSpace(ctx, options)
	}
	if len(errors) > 0 {
		return nil, errors
	}

	// Get the final space details to return the complete state
	actualCreatedSpaceDetails, err := c.GetSpace(ctx, options.ProjectUUID, createdSpaceDetails.SpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get final space details after creation: %w", err)}
	}

	return actualCreatedSpaceDetails, nil
}

// GetSpace retrieves the details of a space by its project and space UUIDs.
// For root spaces, it populates MemberAccess and GroupAccess with all members/groups returned by API.
// Filtering for direct access (for the 'access' attribute in Terraform) is handled in the resource layer.
func (c *SpaceController) GetSpace(ctx context.Context, projectUUID, spaceUUID string) (*models.SpaceDetails, error) {
	tflog.Debug(ctx, "(SpaceController.GetSpace) Getting space", map[string]interface{}{
		"projectUUID": projectUUID,
		"spaceUUID":   spaceUUID,
	})

	// Get space details from the service layer
	space, err := c.spaceService.GetSpace(ctx, projectUUID, spaceUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get space: %w", err)
	}

	tflog.Debug(ctx, "(SpaceController.GetSpace) Space details from service layer", map[string]interface{}{
		"space": space,
	})

	// Convert API SpaceAccessMember to models.SpaceAccessMember
	spaceAccessMembers := []models.SpaceMemberAccess{}
	for _, member := range space.SpaceAccessMembers {
		spaceAccessMembers = append(spaceAccessMembers, models.SpaceMemberAccess{
			UserUUID:        member.UserUUID,
			SpaceRole:       member.SpaceRole,
			HasDirectAccess: &member.HasDirectAccess,
			InheritedRole:   &member.InheritedRole,
			InheritedFrom:   &member.InheritedFrom,
			ProjectRole:     &member.ProjectRole,
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
			AccessList: []models.SpaceMemberAccess{}, // API doesn't provide access list for child spaces
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

	tflog.Debug(ctx, "(SpaceController.GetSpace) Space details", map[string]interface{}{
		"spaceDetails": spaceDetails,
	})

	return spaceDetails, nil
}

// UpdateSpace updates a space based on whether it's a root or nested space and if its parent is changing.
// It orchestrates calls to specific update/move functions.
func (c *SpaceController) UpdateSpace(
	ctx context.Context,
	options UpdateSpaceOptions,
) (*models.SpaceDetails, []error) {
	tflog.Debug(ctx, "(SpaceController.UpdateSpace) Updating space", map[string]interface{}{
		"options": options,
	})

	// Get the current space details to determine if it's a root or nested space.
	currentSpaceDetails, err := c.GetSpace(ctx, options.ProjectUUID, options.SpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get current space details for update: %w", err)}
	}

	// Check if the space is currently a root space (ParentSpaceUUID is nil) and if the plan indicates it should become root
	isCurrentlyRootSpace := !currentSpaceDetails.IsNestedSpace()
	isBecomingRootSpace := models.IsEmptyStringPointer(options.ParentSpaceUUID)
	tflog.Debug(ctx, "(SpaceController.UpdateSpace) options", map[string]interface{}{
		"options":              options,
		"ParentSpaceUUID":      &options.ParentSpaceUUID,
		"isCurrentlyRootSpace": isCurrentlyRootSpace,
		"isBecomingRootSpace":  isBecomingRootSpace,
	})

	var errors []error

	if isCurrentlyRootSpace && isBecomingRootSpace {
		// Scenario 1: Remains a root space - Update properties and access.
		errors = c.updateRootSpace(
			ctx,
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
			ctx,
			options.ProjectUUID,
			options.SpaceUUID,
			options.SpaceName,
			options.ParentSpaceUUID,
		)
	} else if !isCurrentlyRootSpace && isBecomingRootSpace {
		// Scenario 3: Nested space becoming a root space - Move to root and then apply access controls.
		// The space will initially inherit project access, and then direct access can be set.
		errors = c.moveNestedToRootSpace(
			ctx,
			options.ProjectUUID,
			options.SpaceUUID,
			options.SpaceName,
			options.IsPrivate,
			options.MemberAccess,
			options.GroupAccess,
		)
	} else {
		// Scenario 4: Nested space staying nested (either same parent or different parent).
		// Only name and parent space can be updated via the API for nested spaces.
		// Access controls and privacy are inherited and cannot be managed.
		errors = c.updateNestedSpace(
			ctx,
			options.ProjectUUID,
			options.SpaceUUID,
			options.SpaceName,
			options.ParentSpaceUUID,
			currentSpaceDetails,
		)
	}

	// Return the updated space details converted to Terraform state
	if len(errors) > 0 {
		return nil, errors
	}

	// Get the final space details to return the complete state
	actualUpdatedSpaceDetails, err := c.GetSpace(ctx, options.ProjectUUID, options.SpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get final space details after update: %w", err)}
	}

	return actualUpdatedSpaceDetails, nil
}

// DeleteSpace deletes a space if deletion protection is disabled.
func (c *SpaceController) DeleteSpace(ctx context.Context, options DeleteSpaceOptions) error {
	tflog.Debug(ctx, "(SpaceController.DeleteSpace) Deleting space", map[string]interface{}{
		"options": options,
	})

	if options.DeletionProtection {
		return fmt.Errorf("cannot delete space %s: deletion protection is enabled", options.SpaceUUID)
	}

	// Check if the space has any child spaces
	space, err := c.GetSpace(ctx, options.ProjectUUID, options.SpaceUUID)
	if err != nil {
		return fmt.Errorf("failed to get space details: %w", err)
	}
	childSpaces := space.ChildSpaces
	if len(childSpaces) > 0 {
		return fmt.Errorf("cannot delete space %s: it has child spaces. Please delete the child spaces first", options.SpaceUUID)
	}

	// Delete the space via the service layer
	return c.spaceService.DeleteSpace(ctx, options.ProjectUUID, options.SpaceUUID)
}

// ImportSpace imports an existing space by its resource ID.
// It retrieves the space details and access settings.
func (c *SpaceController) ImportSpace(ctx context.Context, options ImportSpaceOptions) (*models.SpaceDetails, error) {
	tflog.Debug(ctx, "(SpaceController.ImportSpace) Importing space", map[string]interface{}{
		"options": options,
	})

	// Extract project and space UUIDs from the resource ID string
	projectUUID, spaceUUID, err := c.spaceService.ExtractSpaceResourceID(options.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract resource ID %s: %w", options.ResourceID, err)
	}

	// Get space details via the service layer
	spaceDetails, err := c.GetSpace(ctx, projectUUID, spaceUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get space details: %w", err)
	}

	return spaceDetails, nil
}

// --- Private Helper Methods ---

// createRootSpace creates a new root-level space and manages its direct access settings.
func (c *SpaceController) createRootSpace(
	ctx context.Context,
	options CreateSpaceOptions,
) (*models.SpaceDetails, []error) {
	tflog.Debug(ctx, "Creating root space", map[string]interface{}{
		"options": options,
	})

	// 1. Validate the input
	validationErrors := c.validateSpaceCreation(ctx, options)
	if len(validationErrors) > 0 {
		return nil, validationErrors
	}

	// 2. Create the space via the service layer
	createdSpace, err := c.spaceService.CreateSpace(
		ctx,
		options.ProjectUUID,
		options.SpaceName,
		options.IsPrivate,
		nil, // nil parentSpaceUUID for root space
	)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to create space: %w", err)}
	}

	// 3. Manage access for the root-level space after creation.
	// Use the specific helper functions, passing empty current access lists.
	var accessErrors []error

	memberErrors := c.manageRootSpaceMemberAccess(
		ctx,
		options.ProjectUUID,
		createdSpace.SpaceUUID,
		options.MemberAccess,
		[]models.SpaceMemberAccess{}, // No existing direct member access on creation
	)
	accessErrors = append(accessErrors, memberErrors...)

	groupErrors := c.manageRootSpaceGroupAccess(
		ctx,
		options.ProjectUUID,
		createdSpace.SpaceUUID,
		options.GroupAccess,
		[]models.SpaceAccessGroup{}, // No existing group access on creation
	)
	accessErrors = append(accessErrors, groupErrors...)

	if len(accessErrors) > 0 {
		// If access management fails, try to clean up by deleting the space
		errDel := c.spaceService.DeleteSpace(ctx, options.ProjectUUID, createdSpace.SpaceUUID)
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
	ctx context.Context,
	options CreateSpaceOptions,
) (*models.SpaceDetails, []error) {
	tflog.Debug(ctx, "Creating nested space", map[string]interface{}{
		"options": options,
	})

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
	createdSpace, err := c.spaceService.CreateSpace(ctx, options.ProjectUUID, options.SpaceName, nil, options.ParentSpaceUUID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to create nested space: %w", err)}
	}

	return createdSpace, nil
}

// validateSpaceCreation validates space creation parameters
func (c *SpaceController) validateSpaceCreation(ctx context.Context, options CreateSpaceOptions) []error {
	var errors []error

	// 1.1 Check if the member can become a space member (must be project member, not org admin)
	projectMembers, err := c.projectService.GetProjectMembers(ctx, options.ProjectUUID)
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
	}

	// 1.2 Check if the groups exist in the organization
	for _, group := range options.GroupAccess {
		_, err := c.organizationGroupsService.GetGroup(ctx, group.GroupUUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("group %s not found: %w", group.GroupUUID, err))
			continue
		}
	}

	return errors
}

// manageRootSpaceMemberAccess handles adding, updating, and removing direct member access for a root space.
func (c *SpaceController) manageRootSpaceMemberAccess(
	ctx context.Context,
	projectUUID string,
	spaceUUID string,
	newMemberAccess []models.SpaceAccessMember,
	currentMemberAccess []models.SpaceMemberAccess,
) []error {
	var errors []error

	tflog.Debug(ctx, "(SpaceController.manageRootSpaceMemberAccess) Managing root space member access", map[string]interface{}{
		"projectUUID":         projectUUID,
		"spaceUUID":           spaceUUID,
		"newMemberAccess":     newMemberAccess,
		"currentMemberAccess": currentMemberAccess,
	})

	currentDirectMemberMap := make(map[string]models.SpaceMemberAccess)
	for _, currentMember := range currentMemberAccess {
		// Only consider members with direct access for management
		if currentMember.HasDirectAccess != nil && *currentMember.HasDirectAccess {
			currentDirectMemberMap[currentMember.UserUUID] = currentMember
		}
	}

	newDirectMemberMap := make(map[string]models.SpaceAccessMember)
	for _, member := range newMemberAccess {
		newDirectMemberMap[member.UserUUID] = member
	}

	if len(errors) > 0 {
		return errors
	}

	// Find members to remove (in current but not in new)
	for userUUID := range currentDirectMemberMap {
		if _, exists := newDirectMemberMap[userUUID]; !exists {
			err := c.spaceService.RemoveUserFromSpace(ctx, projectUUID, spaceUUID, userUUID)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to remove direct access for user %s from space %s: %w", userUUID, spaceUUID, err))
			}
			delete(currentDirectMemberMap, userUUID) // Remove from map to avoid re-processing
		}
	}

	// Find members to add or update (in new)
	for userUUID, newMember := range newDirectMemberMap {
		currentMember, exists := currentDirectMemberMap[userUUID]

		// Add if not in current direct access or update if role has changed
		if !exists || currentMember.SpaceRole != newMember.SpaceRole {
			err := c.spaceService.AddUserToSpace(ctx, projectUUID, spaceUUID, userUUID, newMember.SpaceRole)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to add/update direct access for user %s to space %s: %w", userUUID, spaceUUID, err))
			}
		}
	}

	return errors
}

// manageRootSpaceGroupAccess handles adding, updating, and removing group access for a root space.
func (c *SpaceController) manageRootSpaceGroupAccess(
	ctx context.Context,
	projectUUID string,
	spaceUUID string,
	newGroupAccess []models.SpaceAccessGroup,
	currentGroupAccess []models.SpaceAccessGroup,
) []error {
	var errors []error

	tflog.Debug(ctx, "(SpaceController.manageRootSpaceGroupAccess) Managing root space group access", map[string]interface{}{
		"projectUUID":        projectUUID,
		"spaceUUID":          spaceUUID,
		"newGroupAccess":     newGroupAccess,
		"currentGroupAccess": currentGroupAccess,
	})

	currentGroupMap := make(map[string]models.SpaceAccessGroup)
	for _, group := range currentGroupAccess {
		currentGroupMap[group.GroupUUID] = group
	}

	newGroupMap := make(map[string]models.SpaceAccessGroup)
	for _, group := range newGroupAccess {
		newGroupMap[group.GroupUUID] = group
	}

	// Find groups to remove (in current but not in new)
	for groupUUID := range currentGroupMap {
		if _, exists := newGroupMap[groupUUID]; !exists {
			// Check if group exists in Lightdash before attempting to remove access.
			// This prevents errors if the group was deleted outside of Terraform.
			_, err := c.organizationGroupsService.GetGroup(ctx, groupUUID)
			if err != nil {
				// Skip if group no longer exists in Lightdash
				tflog.Debug(ctx, fmt.Sprintf("group %s not found in Lightdash, skipping removal from space %s", groupUUID, spaceUUID))
				continue
			}

			// Remove access via API
			err = c.spaceService.RemoveGroupFromSpace(ctx, projectUUID, spaceUUID, groupUUID)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to remove group %s from space %s: %w", groupUUID, spaceUUID, err))
			}
			delete(currentGroupMap, groupUUID) // Remove from map to avoid re-processing
		}
	}

	// Find groups to add or update (in new)
	for groupUUID, newGroup := range newGroupMap {
		currentGroup, exists := currentGroupMap[groupUUID]

		// Check if group exists in Lightdash before adding/updating
		_, err := c.organizationGroupsService.GetGroup(ctx, groupUUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("group %s not found in Lightdash: %w", groupUUID, err))
			continue
		}

		// Add if not in current or update if role has changed
		if !exists || currentGroup.SpaceRole != newGroup.SpaceRole {
			err := c.spaceService.UpdateGroupAccessInSpace(ctx, projectUUID, spaceUUID, groupUUID, newGroup.SpaceRole)
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
	ctx context.Context,
	projectUUID string,
	spaceUUID string,
	spaceName string,
	isPrivate *bool,
	newMemberAccess []models.SpaceAccessMember,
	newGroupAccess []models.SpaceAccessGroup,
	currentSpaceDetails *models.SpaceDetails,
) []error {
	var errors []error

	tflog.Debug(ctx, "(SpaceController.updateRootSpace) Updating root space", map[string]interface{}{
		"projectUUID":         projectUUID,
		"spaceUUID":           spaceUUID,
		"spaceName":           spaceName,
		"isPrivate":           isPrivate,
		"newMemberAccess":     newMemberAccess,
		"newGroupAccess":      newGroupAccess,
		"currentSpaceDetails": currentSpaceDetails,
	})

	// If isPrivate isn't changed, then it is nil.
	// This is a workaround to avoid the API from returning an error.
	var isPrivateForUpdate *bool
	if isPrivate != nil && *isPrivate != currentSpaceDetails.IsPrivate {
		isPrivateForUpdate = isPrivate
	}

	// 1. Update the space properties via the service layer if they have changed
	if spaceName != currentSpaceDetails.SpaceName || isPrivateForUpdate != nil {
		updatedSpaceDetails, err := c.spaceService.UpdateRootSpace(ctx, projectUUID, spaceUUID, spaceName, isPrivateForUpdate)
		if err != nil {
			return []error{fmt.Errorf("failed to update space properties: %w", err)}
		}
		tflog.Debug(ctx, "(SpaceController.updateRootSpace) Updated space details", map[string]interface{}{
			"projectUUID": updatedSpaceDetails.ProjectUUID,
			"spaceUUID":   updatedSpaceDetails.SpaceUUID,
			"spaceName":   updatedSpaceDetails.SpaceName,
			"isPrivate":   updatedSpaceDetails.IsPrivate,
		})
	}

	// 2. Manage member access (add/update/remove direct access)
	memberErrors := c.manageRootSpaceMemberAccess(
		ctx,
		projectUUID,
		spaceUUID,
		newMemberAccess,
		currentSpaceDetails.SpaceAccessMembers,
	)
	errors = append(errors, memberErrors...)

	// 3. Handle group access updates (add/update/remove groups)
	groupErrors := c.manageRootSpaceGroupAccess(
		ctx,
		projectUUID,
		spaceUUID,
		newGroupAccess,
		currentSpaceDetails.SpaceAccessGroups,
	)
	errors = append(errors, groupErrors...)

	return errors
}

// updateNestedSpace updates the properties for a nested space.
// Only the name and parent space UUID can be changed for nested spaces via the API.
// Access controls and privacy are inherited and cannot be managed by this function.
func (c *SpaceController) updateNestedSpace(
	ctx context.Context,
	projectUUID string,
	spaceUUID string,
	spaceName string,
	parentSpaceUUID *string,
	currentSpaceDetails *models.SpaceDetails,
) []error {
	tflog.Debug(ctx, "(SpaceController.updateNestedSpace) Updating nested space", map[string]interface{}{
		"projectUUID":     projectUUID,
		"spaceUUID":       spaceUUID,
		"spaceName":       spaceName,
		"parentSpaceUUID": parentSpaceUUID,
	})

	// If the new parent Space UUID isn't the same as the current one, then move the space to the new parent
	if !services.CompareParentSpaceUUID(currentSpaceDetails.ParentSpaceUUID, parentSpaceUUID) {
		err := c.spaceService.MoveSpace(ctx, projectUUID, spaceUUID, parentSpaceUUID)
		if err != nil {
			return []error{fmt.Errorf("failed to move space to new parent: %w", err)}
		}
	}

	// Update only the name and parent space UUID for nested spaces via the service layer
	// isPrivate is passed as nil because it cannot be updated for nested spaces.
	// Pass the parentSpaceUUID to the service layer to handle moves between nested spaces
	_, err := c.spaceService.UpdateNestedSpace(ctx, projectUUID, spaceUUID, spaceName, nil)
	if err != nil {
		return []error{fmt.Errorf("failed to update nested space %s: %w", spaceUUID, err)}
	}

	return nil
}

// moveRootToNestedSpace handles moving a root space to become a nested space.
// This involves updating the parent space UUID.
// Access controls will be inherited from the new parent.
func (c *SpaceController) moveRootToNestedSpace(
	ctx context.Context,
	projectUUID string,
	spaceUUID string,
	spaceName string,
	parentSpaceUUID *string,
) []error {
	tflog.Debug(ctx, "(SpaceController.moveRootToNestedSpace) Moving root space to nested space", map[string]interface{}{
		"projectUUID":     projectUUID,
		"spaceUUID":       spaceUUID,
		"spaceName":       spaceName,
		"parentSpaceUUID": parentSpaceUUID,
	})

	// 1. Move the space to the new parent via the service layer
	err := c.spaceService.MoveSpace(ctx, projectUUID, spaceUUID, parentSpaceUUID)
	if err != nil {
		return []error{fmt.Errorf("failed to move root space %s to nested space under parent %s: %w", spaceUUID, *parentSpaceUUID, err)}
	}

	// 2. Get the space details to check if it is private
	spaceDetails, err := c.spaceService.GetSpace(ctx, projectUUID, spaceUUID)
	if err != nil {
		return []error{fmt.Errorf("failed to get space details: %w", err)}
	}

	// 3. Update the space to make it a nested space (no parent) via the service layer
	_, err = c.spaceService.UpdateNestedSpace(ctx, projectUUID, spaceUUID, spaceName, &spaceDetails.IsPrivate)
	if err != nil {
		return []error{fmt.Errorf("failed to move root space %s to nested space under parent %s: %w", spaceUUID, *parentSpaceUUID, err)}
	}

	return nil
}

// moveNestedToRootSpace handles moving a nested space to become a root space.
// This involves updating the parent space UUID to nil and then applying explicit access controls.
func (c *SpaceController) moveNestedToRootSpace(
	ctx context.Context,
	projectUUID string,
	spaceUUID string,
	spaceName string,
	isPrivate *bool,
	memberAccess []models.SpaceAccessMember,
	groupAccess []models.SpaceAccessGroup,
) []error {
	tflog.Debug(ctx, "(SpaceController.moveNestedToRootSpace) Moving nested space to root space", map[string]interface{}{
		"projectUUID":  projectUUID,
		"spaceUUID":    spaceUUID,
		"spaceName":    spaceName,
		"isPrivate":    isPrivate,
		"memberAccess": memberAccess,
		"groupAccess":  groupAccess,
	})

	// 1. Move the space to the root space via the service layer
	err1 := c.spaceService.MoveSpace(ctx, projectUUID, spaceUUID, nil)
	if err1 != nil {
		return []error{fmt.Errorf("failed to move nested space %s to root: %w", spaceUUID, err1)}
	}

	// 2. Update the space to make it a root space (no parent) via the service layer
	_, err2 := c.spaceService.UpdateRootSpace(ctx, projectUUID, spaceUUID, spaceName, isPrivate)
	if err2 != nil {
		return []error{fmt.Errorf("failed to update space properties after moving to root: %w", err2)}
	}

	// 3. Manage access for the newly root-level space
	accessErrors := c.manageRootSpaceMemberAccess(
		ctx,
		projectUUID,
		spaceUUID,
		memberAccess,
		[]models.SpaceMemberAccess{}, // No existing direct member access when becoming root
	)

	groupErrors := c.manageRootSpaceGroupAccess(
		ctx,
		projectUUID,
		spaceUUID,
		groupAccess,
		[]models.SpaceAccessGroup{}, // No existing group access when becoming root
	)
	accessErrors = append(accessErrors, groupErrors...)

	return accessErrors
}
