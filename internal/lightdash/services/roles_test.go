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
	"strings"
	"testing"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

func testSystemRoles() []models.Role {
	return []models.Role{
		{RoleUUID: "viewer", Name: "Viewer", OwnerType: "system"},
		{RoleUUID: "interactive_viewer", Name: "Interactive Viewer", OwnerType: "system"},
		{RoleUUID: "editor", Name: "Editor", OwnerType: "system"},
		{RoleUUID: "developer", Name: "Developer", OwnerType: "system"},
		{RoleUUID: "admin", Name: "Admin", OwnerType: "system"},
		{RoleUUID: "member", Name: "Member", OwnerType: "system"},
	}
}

func TestResolveRoleIDFromRoles(t *testing.T) {
	roles := testSystemRoles()

	tests := []struct {
		name    string
		role    string
		want    string
		wantErr string
	}{
		{name: "by roleUuid", role: "editor", want: "editor"},
		{name: "by display name", role: "Editor", want: "editor"},
		{name: "interactive viewer slug", role: "interactive_viewer", want: "interactive_viewer"},
		{name: "interactive viewer display name", role: "Interactive Viewer", want: "interactive_viewer"},
		{name: "not found", role: "superadmin", wantErr: "not found"},
		{name: "empty", role: "  ", wantErr: "empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveRoleIDFromRoles(roles, tt.role)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveRoleIDFromRoles_duplicateName(t *testing.T) {
	roles := []models.Role{
		{RoleUUID: "editor", Name: "Editor", OwnerType: "system"},
		{RoleUUID: "custom-editor", Name: "Editor", OwnerType: "user"},
	}

	_, err := resolveRoleIDFromRoles(roles, "Editor")
	if err == nil {
		t.Fatal("expected ambiguous role error")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("error = %q, want ambiguous", err.Error())
	}
}

func TestNormalizeRoleName(t *testing.T) {
	if got := normalizeRoleName("Interactive Viewer"); got != "interactive_viewer" {
		t.Errorf("got %q, want interactive_viewer", got)
	}
}

func TestTerraformProjectRoleFromAssignment(t *testing.T) {
	role, err := TerraformProjectRoleFromAssignment(&models.RoleAssignment{
		RoleID:   "editor",
		RoleName: "editor",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role != models.PROJECT_EDITOR_ROLE {
		t.Errorf("got %q, want editor", role)
	}
}

func TestTerraformOrganizationRoleFromAssignment(t *testing.T) {
	role, err := TerraformOrganizationRoleFromAssignment(&models.RoleAssignment{
		RoleID:   "admin",
		RoleName: "admin",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role != models.ORGANIZATION_ADMIN_ROLE {
		t.Errorf("got %q, want admin", role)
	}
}

func TestFilterAssignmentsByType(t *testing.T) {
	assignments := []models.RoleAssignment{
		{AssigneeType: models.AssigneeTypeUser, AssigneeID: "user-1", RoleID: "viewer"},
		{AssigneeType: models.AssigneeTypeGroup, AssigneeID: "group-1", RoleID: "editor"},
		{AssigneeType: models.AssigneeTypeGroup, AssigneeID: "group-2", RoleID: "admin"},
		{AssigneeType: "other", AssigneeID: "other-1", RoleID: "viewer"},
	}

	got := filterAssignmentsByType(assignments, models.AssigneeTypeGroup)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].AssigneeID != "group-1" || got[1].AssigneeID != "group-2" {
		t.Errorf("got group IDs %q and %q", got[0].AssigneeID, got[1].AssigneeID)
	}
}
