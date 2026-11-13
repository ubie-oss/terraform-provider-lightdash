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
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv2 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v2"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

var (
	roleServiceInstance *RoleService
	roleServiceOnce     sync.Once
)

// ErrRoleAssignmentNotFound is returned, wrapped, when a role assignment is not in the list.
// Callers use errors.Is to tell "removed outside Terraform" from a failed API call.
var ErrRoleAssignmentNotFound = errors.New("role assignment not found")

type RoleService struct {
	client     *api.Client
	rolesByOrg map[string][]models.Role
	orgUUID    string
}

func NewRoleService(client *api.Client) *RoleService {
	return &RoleService{
		client:     client,
		rolesByOrg: make(map[string][]models.Role),
	}
}

// GetRoleService returns a shared RoleService instance with a cached org role catalog.
func GetRoleService(client *api.Client) *RoleService {
	roleServiceOnce.Do(func() {
		roleServiceInstance = NewRoleService(client)
	})
	return roleServiceInstance
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

func (s *RoleService) OrganizationUUID(ctx context.Context) (string, error) {
	if s.orgUUID != "" {
		return s.orgUUID, nil
	}
	orgUUID, err := GetOrganizationUUID(ctx, s.client)
	if err != nil {
		return "", err
	}
	s.orgUUID = orgUUID
	return orgUUID, nil
}

func (s *RoleService) GetOrgUserAssignment(ctx context.Context, orgUUID string, userUUID string) (*models.RoleAssignment, error) {
	assignments, err := apiv2.ListOrganizationRoleAssignmentsV2(s.client, orgUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization role assignments: %w", err)
	}

	return findAssignment(assignments, models.AssigneeTypeUser, userUUID, "organization")
}

func (s *RoleService) GetProjectUserAssignment(ctx context.Context, projectUUID string, userUUID string) (*models.RoleAssignment, error) {
	assignments, err := apiv2.ListProjectRoleAssignmentsV2(s.client, projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list project role assignments: %w", err)
	}

	return findAssignment(assignments, models.AssigneeTypeUser, userUUID, "project")
}

func (s *RoleService) ListProjectGroupAssignments(ctx context.Context, projectUUID string) ([]models.RoleAssignment, error) {
	assignments, err := apiv2.ListProjectRoleAssignmentsV2(s.client, projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list project role assignments: %w", err)
	}

	return filterAssignmentsByType(assignments, models.AssigneeTypeGroup), nil
}

func (s *RoleService) GetProjectGroupAssignment(ctx context.Context, projectUUID string, groupUUID string) (*models.RoleAssignment, error) {
	assignments, err := s.ListProjectGroupAssignments(ctx, projectUUID)
	if err != nil {
		return nil, err
	}

	return findAssignment(assignments, models.AssigneeTypeGroup, groupUUID, "project")
}

func findAssignment(assignments []models.RoleAssignment, assigneeType string, assigneeID string, scope string) (*models.RoleAssignment, error) {
	for i := range assignments {
		assignment := &assignments[i]
		if assignment.AssigneeType == assigneeType && assignment.AssigneeID == assigneeID {
			return assignment, nil
		}
	}
	return nil, fmt.Errorf("%s role assignment for %s %s: %w", scope, assigneeType, assigneeID, ErrRoleAssignmentNotFound)
}

func filterAssignmentsByType(assignments []models.RoleAssignment, assigneeType string) []models.RoleAssignment {
	var filtered []models.RoleAssignment
	for _, assignment := range assignments {
		if assignment.AssigneeType == assigneeType {
			filtered = append(filtered, assignment)
		}
	}
	return filtered
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
	return terraformRoleFromAssignment(assignment, models.ProjectMemberRole.IsValid, "project")
}

// TerraformOrganizationRoleFromAssignment maps a v2 assignment to a Terraform organization role string.
func TerraformOrganizationRoleFromAssignment(assignment *models.RoleAssignment) (models.OrganizationMemberRole, error) {
	return terraformRoleFromAssignment(assignment, models.OrganizationMemberRole.IsValid, "organization")
}

// terraformRoleFromAssignment returns the first of the assignment's role name or ID that is a valid role
// once normalized. Valid role names contain no spaces, so normalizing first accepts the same inputs.
func terraformRoleFromAssignment[T ~string](assignment *models.RoleAssignment, isValid func(T) bool, scope string) (T, error) {
	if assignment == nil {
		return "", fmt.Errorf("%s role assignment is nil", scope)
	}
	for _, raw := range []string{assignment.RoleName, assignment.RoleID} {
		if candidate := T(normalizeRoleName(raw)); isValid(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("unknown %s role from assignment: roleId=%q roleName=%q", scope, assignment.RoleID, assignment.RoleName)
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
