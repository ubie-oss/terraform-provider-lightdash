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
	"context"
	"fmt"

	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"
	apiv2 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v2"

	"github.com/hashicorp/terraform-plugin-log/tflog"

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

const maxSpaceParentChainDepth = 64

// resolveTerraformIsPrivateByUUID loads a space and resolves its Terraform is_private,
// recursing through parents when the space inherits parent permissions.
func (s *SpaceService) resolveTerraformIsPrivateByUUID(ctx context.Context, projectUuid, spaceUuid string, depth int) (bool, error) {
	if depth > maxSpaceParentChainDepth {
		return false, fmt.Errorf("space parent chain exceeded max depth %d", maxSpaceParentChainDepth)
	}
	sp, err := apiv1.GetSpaceV1(s.client, projectUuid, spaceUuid)
	if err != nil {
		return false, fmt.Errorf("get space for visibility: %w", err)
	}
	return s.ResolveTerraformIsPrivateFromGetResults(ctx, projectUuid, sp, depth)
}

// ResolveTerraformIsPrivateFromGetResults maps GetSpace API results to Terraform `is_private`,
// following parent inheritance semantics for nested spaces.
func (s *SpaceService) ResolveTerraformIsPrivateFromGetResults(ctx context.Context, projectUuid string, sp *apiv1.GetSpaceV1Results, depth int) (bool, error) {
	inherit := models.EffectiveInheritFromOptional(sp.InheritParentPermissions, sp.IsPrivate)
	if sp.ParentSpaceUUID == nil || *sp.ParentSpaceUUID == "" {
		return models.TerraformIsPrivateFromAPIForNestedSpace(inherit, false, true), nil
	}
	parentPriv, err := s.resolveTerraformIsPrivateByUUID(ctx, projectUuid, *sp.ParentSpaceUUID, depth+1)
	if err != nil {
		return false, err
	}
	return models.TerraformIsPrivateFromAPIForNestedSpace(inherit, parentPriv, false), nil
}

// Space Management Methods

// CreateSpace creates a new space with specified properties
func (s *SpaceService) CreateSpace(ctx context.Context, projectUuid, spaceName string, isPrivate *bool, parentSpaceUuid *string) (*models.SpaceDetails, error) {
	createdSpace, err := apiv1.CreateSpaceV1(s.client, projectUuid, spaceName, isPrivate, parentSpaceUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to create space: %w", err)
	}

	inherit := models.EffectiveInheritFromOptional(createdSpace.InheritParentPermissions, createdSpace.IsPrivate)
	parentUUID := createdSpace.ParentSpaceUUID
	if parentUUID == nil || *parentUUID == "" {
		parentUUID = parentSpaceUuid
	}
	var isPrivateResolved bool
	if parentUUID == nil || *parentUUID == "" {
		isPrivateResolved = models.TerraformIsPrivateFromAPIForNestedSpace(inherit, false, true)
	} else {
		parentPriv, err := s.resolveTerraformIsPrivateByUUID(ctx, projectUuid, *parentUUID, 0)
		if err != nil {
			return nil, fmt.Errorf("resolve parent visibility: %w", err)
		}
		isPrivateResolved = models.TerraformIsPrivateFromAPIForNestedSpace(inherit, parentPriv, false)
	}
	spaceDetails := &models.SpaceDetails{
		ProjectUUID:              projectUuid,
		SpaceUUID:                createdSpace.SpaceUUID,
		SpaceName:                createdSpace.SpaceName,
		IsPrivate:                isPrivateResolved,
		InheritParentPermissions: inherit,
		ParentSpaceUUID:          parentSpaceUuid,
		SpaceAccessMembers:       []models.SpaceMemberAccess{},
		SpaceAccessGroups:        []models.SpaceAccessGroup{},
		ChildSpaces:              []models.ChildSpace{},
	}

	return spaceDetails, nil
}

// GetSpace retrieves a space by UUID
func (s *SpaceService) GetSpace(ctx context.Context, projectUuid, spaceUuid string) (*apiv1.GetSpaceV1Results, error) {
	return apiv1.GetSpaceV1(s.client, projectUuid, spaceUuid)
}

// UpdateRootSpace updates the space properties for a root space
func (s *SpaceService) UpdateRootSpace(ctx context.Context, projectUuid, spaceUuid, spaceName string, inheritParentPermissions *bool) (*apiv1.UpdateSpaceV1Results, error) {
	tflog.Debug(ctx, "(SpaceService.UpdateRootSpace) Updating root space", map[string]interface{}{
		"projectUuid":              projectUuid,
		"spaceUuid":                spaceUuid,
		"spaceName":                spaceName,
		"inheritParentPermissions": inheritParentPermissions,
	})
	updatedSpace, err := apiv1.UpdateSpaceV1(s.client, ctx, projectUuid, spaceUuid, spaceName, inheritParentPermissions)
	if err != nil {
		return nil, fmt.Errorf("failed to update space properties: %w", err)
	}
	eff := models.EffectiveInheritFromOptional(updatedSpace.InheritParentPermissions, updatedSpace.IsPrivate)
	tflog.Debug(ctx, "(SpaceService.UpdateRootSpace) Updated space properties", map[string]interface{}{
		"projectUuid":              updatedSpace.ProjectUUID,
		"spaceUuid":                updatedSpace.SpaceUUID,
		"spaceName":                updatedSpace.SpaceName,
		"inheritParentPermissions": eff,
		"isPrivate": models.TerraformIsPrivateFromAPIFieldsRootSemantics(
			updatedSpace.InheritParentPermissions, updatedSpace.IsPrivate),
	})
	return updatedSpace, nil
}

// UpdateNestedSpace updates the space properties for a nested space
func (s *SpaceService) UpdateNestedSpace(ctx context.Context, projectUuid, spaceUuid string, spaceName string, inheritParentPermissions *bool) (*apiv1.UpdateSpaceV1Results, error) {
	updatedSpace, err := apiv1.UpdateSpaceV1(s.client, ctx, projectUuid, spaceUuid, spaceName, inheritParentPermissions)
	if err != nil {
		return nil, fmt.Errorf("failed to update nested space: %w", err)
	}
	return updatedSpace, nil
}

// DeleteSpace deletes a space
func (s *SpaceService) DeleteSpace(ctx context.Context, projectUuid, spaceUuid string) error {
	return apiv1.DeleteSpaceV1(s.client, projectUuid, spaceUuid)
}

// MoveSpace moves a space to a new parent space
// parentSpaceUuidPointer == nil means the space should become a root space
func (s *SpaceService) MoveSpace(ctx context.Context, projectUuid, spaceUuid string, parentSpaceUuidPointer *string) error {
	err := apiv2.MoveSpaceV2(s.client, projectUuid, spaceUuid, parentSpaceUuidPointer)
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
func (s *SpaceService) AddUserToSpace(ctx context.Context, projectUuid, spaceUuid, userUuid string, role models.SpaceMemberRole) error {
	return apiv1.AddSpaceShareToUserV1(s.client, projectUuid, spaceUuid, userUuid, role)
}

// RemoveUserFromSpace revokes a user's access to a space
// NOTE: Should only be called for root spaces
func (s *SpaceService) RemoveUserFromSpace(ctx context.Context, projectUuid, spaceUuid, userUuid string) error {
	return apiv1.RevokeSpaceAccessV1(s.client, projectUuid, spaceUuid, userUuid)
}

// AddGroupToSpace grants a group access to a space with the specified role
// NOTE: Should only be called for root spaces
func (s *SpaceService) AddGroupToSpace(ctx context.Context, projectUuid, spaceUuid, groupUuid string, role models.SpaceMemberRole) error {
	return apiv1.AddSpaceGroupAccessV1(s.client, projectUuid, spaceUuid, groupUuid, role)
}

// UpdateGroupAccessInSpace updates a group's role in a space
// NOTE: Should only be called for root spaces
func (s *SpaceService) UpdateGroupAccessInSpace(ctx context.Context, projectUuid, spaceUuid, groupUuid string, role models.SpaceMemberRole) error {
	return apiv1.AddSpaceGroupAccessV1(s.client, projectUuid, spaceUuid, groupUuid, role)
}

// RemoveGroupFromSpace revokes a group's access to a space
// NOTE: Should only be called for root spaces
func (s *SpaceService) RemoveGroupFromSpace(ctx context.Context, projectUuid, spaceUuid, groupUuid string) error {
	return apiv1.RevokeSpaceGroupAccessV1(s.client, projectUuid, spaceUuid, groupUuid)
}

// GetChildSpaces returns all child spaces of a space
func (s *SpaceService) GetChildSpaces(ctx context.Context, projectUuid, spaceUuid string) ([]apiv1.ChildSpace, error) {
	space, err := s.GetSpace(ctx, projectUuid, spaceUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get space details: %w", err)
	}

	return space.ChildSpaces, nil
}
