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
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv2 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v2"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type RoleService struct {
	client     *api.Client
	rolesByOrg map[string][]models.Role
}

func NewRoleService(client *api.Client) *RoleService {
	return &RoleService{
		client:     client,
		rolesByOrg: make(map[string][]models.Role),
	}
}

func (s *RoleService) GetRoles(ctx context.Context, orgUUID string) ([]models.Role, error) {
	if roles, ok := s.rolesByOrg[orgUUID]; ok {
		return roles, nil
	}

	roles, err := apiv2.GetOrganizationRolesV2(s.client, orgUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization roles: %w", err)
	}

	s.rolesByOrg[orgUUID] = roles
	return roles, nil
}

func (s *RoleService) ResolveRoleID(ctx context.Context, orgUUID string, roleName string) (string, error) {
	roles, err := s.GetRoles(ctx, orgUUID)
	if err != nil {
		return "", err
	}

	return resolveRoleIDFromRoles(roles, roleName)
}

func (s *RoleService) ResolveRoleName(ctx context.Context, orgUUID string, roleID string) (string, error) {
	roles, err := s.GetRoles(ctx, orgUUID)
	if err != nil {
		return "", err
	}

	return resolveRoleNameFromRoles(roles, roleID)
}

func (s *RoleService) GetOrgUserAssignment(ctx context.Context, orgUUID string, userUUID string) (*models.RoleAssignment, error) {
	assignments, err := apiv2.ListOrganizationRoleAssignmentsV2(s.client, orgUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization role assignments: %w", err)
	}

	for i := range assignments {
		assignment := &assignments[i]
		if assignment.AssigneeType == "user" && assignment.AssigneeID == userUUID {
			return assignment, nil
		}
	}

	return nil, fmt.Errorf("organization role assignment not found for user %s", userUUID)
}

func (s *RoleService) GetProjectUserAssignment(ctx context.Context, projectUUID string, userUUID string) (*models.RoleAssignment, error) {
	assignments, err := apiv2.ListProjectRoleAssignmentsV2(s.client, projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list project role assignments: %w", err)
	}

	for i := range assignments {
		assignment := &assignments[i]
		if assignment.AssigneeType == "user" && assignment.AssigneeID == userUUID {
			return assignment, nil
		}
	}

	return nil, fmt.Errorf("project role assignment not found for user %s", userUUID)
}

func (s *RoleService) ListProjectGroupAssignments(ctx context.Context, projectUUID string) ([]models.RoleAssignment, error) {
	assignments, err := apiv2.ListProjectRoleAssignmentsV2(s.client, projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list project role assignments: %w", err)
	}

	return filterGroupAssignments(assignments), nil
}

func (s *RoleService) GetProjectGroupAssignment(ctx context.Context, projectUUID string, groupUUID string) (*models.RoleAssignment, error) {
	assignments, err := s.ListProjectGroupAssignments(ctx, projectUUID)
	if err != nil {
		return nil, err
	}

	for i := range assignments {
		assignment := &assignments[i]
		if assignment.AssigneeID == groupUUID {
			return assignment, nil
		}
	}

	return nil, fmt.Errorf("project role assignment not found for group %s", groupUUID)
}

func filterGroupAssignments(assignments []models.RoleAssignment) []models.RoleAssignment {
	var groups []models.RoleAssignment
	for _, assignment := range assignments {
		if assignment.AssigneeType == "group" {
			groups = append(groups, assignment)
		}
	}
	return groups
}

func (s *RoleService) AssignOrgUserRole(ctx context.Context, orgUUID string, userUUID string, roleName string) (*models.RoleAssignment, error) {
	roleID, err := s.ResolveRoleID(ctx, orgUUID, roleName)
	if err != nil {
		return nil, err
	}

	assignment, err := apiv2.AssignOrganizationRoleToUserV2(s.client, orgUUID, userUUID, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to assign organization role to user: %w", err)
	}

	return assignment, nil
}

func (s *RoleService) AssignProjectUserRole(ctx context.Context, orgUUID string, projectUUID string, userUUID string, roleName string, sendEmail bool) (*models.RoleAssignment, error) {
	roleID, err := s.ResolveRoleID(ctx, orgUUID, roleName)
	if err != nil {
		return nil, err
	}

	assignment, err := apiv2.AssignProjectRoleToUserV2(s.client, projectUUID, userUUID, roleID, sendEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to assign project role to user: %w", err)
	}

	return assignment, nil
}

func (s *RoleService) AssignProjectGroupRole(ctx context.Context, orgUUID string, projectUUID string, groupUUID string, roleName string, sendEmail bool) (*models.RoleAssignment, error) {
	roleID, err := s.ResolveRoleID(ctx, orgUUID, roleName)
	if err != nil {
		return nil, err
	}

	assignment, err := apiv2.AssignProjectRoleToGroupV2(s.client, projectUUID, groupUUID, roleID, sendEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to assign project role to group: %w", err)
	}

	return assignment, nil
}

func (s *RoleService) UpdateProjectGroupRole(ctx context.Context, orgUUID string, projectUUID string, groupUUID string, roleName string) (*models.RoleAssignment, error) {
	roleID, err := s.ResolveRoleID(ctx, orgUUID, roleName)
	if err != nil {
		return nil, err
	}

	assignment, err := apiv2.UpdateProjectGroupRoleV2(s.client, projectUUID, groupUUID, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to update project group role: %w", err)
	}

	return assignment, nil
}

func (s *RoleService) RemoveProjectUserRole(ctx context.Context, projectUUID string, userUUID string) error {
	if err := apiv2.RemoveProjectRoleFromUserV2(s.client, projectUUID, userUUID); err != nil {
		return fmt.Errorf("failed to remove project role from user: %w", err)
	}
	return nil
}

func (s *RoleService) RemoveProjectGroupRole(ctx context.Context, projectUUID string, groupUUID string) error {
	if err := apiv2.RemoveProjectRoleFromGroupV2(s.client, projectUUID, groupUUID); err != nil {
		return fmt.Errorf("failed to remove project role from group: %w", err)
	}
	return nil
}

// TerraformProjectRoleFromAssignment maps a v2 assignment to a Terraform project role string.
func TerraformProjectRoleFromAssignment(assignment *models.RoleAssignment) (models.ProjectMemberRole, error) {
	return terraformRoleFromAssignment(
		assignment,
		func(role string) bool { return models.ProjectMemberRole(role).IsValid() },
		func(role string) models.ProjectMemberRole { return models.ProjectMemberRole(role) },
		"project",
	)
}

// TerraformOrganizationRoleFromAssignment maps a v2 assignment to a Terraform organization role string.
func TerraformOrganizationRoleFromAssignment(assignment *models.RoleAssignment) (models.OrganizationMemberRole, error) {
	return terraformRoleFromAssignment(
		assignment,
		func(role string) bool { return models.OrganizationMemberRole(role).IsValid() },
		func(role string) models.OrganizationMemberRole { return models.OrganizationMemberRole(role) },
		"organization",
	)
}

func terraformRoleFromAssignment[T ~string](
	assignment *models.RoleAssignment,
	isValid func(string) bool,
	toRole func(string) T,
	scope string,
) (T, error) {
	var zero T
	if assignment == nil {
		return zero, fmt.Errorf("%s role assignment is nil", scope)
	}

	for _, candidate := range []string{assignment.RoleName, assignment.RoleID} {
		candidate = strings.ToLower(strings.TrimSpace(candidate))
		if candidate == "" {
			continue
		}
		if isValid(candidate) {
			return toRole(candidate), nil
		}
		normalized := normalizeRoleName(candidate)
		if isValid(normalized) {
			return toRole(normalized), nil
		}
	}

	return zero, fmt.Errorf("unknown %s role from assignment: roleId=%q roleName=%q", scope, assignment.RoleID, assignment.RoleName)
}

func normalizeRoleName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	return strings.ReplaceAll(normalized, " ", "_")
}

func roleMatchesName(role models.Role, roleName string) bool {
	target := strings.ToLower(strings.TrimSpace(roleName))
	if strings.ToLower(role.RoleUUID) == target {
		return true
	}
	if normalizeRoleName(role.Name) == target {
		return true
	}
	return strings.ToLower(role.Name) == target
}

func resolveRoleIDFromRoles(roles []models.Role, roleName string) (string, error) {
	if strings.TrimSpace(roleName) == "" {
		return "", fmt.Errorf("role name is empty")
	}

	var matches []models.Role
	for _, role := range roles {
		if roleMatchesName(role, roleName) {
			matches = append(matches, role)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("role %q not found", roleName)
	case 1:
		return matches[0].RoleUUID, nil
	default:
		return "", fmt.Errorf("ambiguous role name %q: %d roles matched", roleName, len(matches))
	}
}

func resolveRoleNameFromRoles(roles []models.Role, roleID string) (string, error) {
	if strings.TrimSpace(roleID) == "" {
		return "", fmt.Errorf("role ID is empty")
	}

	target := strings.ToLower(strings.TrimSpace(roleID))
	for _, role := range roles {
		if strings.ToLower(role.RoleUUID) == target {
			return role.RoleUUID, nil
		}
	}

	return "", fmt.Errorf("role ID %q not found", roleID)
}
