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

package models

import (
	"encoding/json"
	"testing"
)

func TestRole_UnmarshalJSON_systemRoleWithScopes(t *testing.T) {
	const fixture = `{
		"roleUuid": "viewer",
		"name": "Viewer",
		"description": "Viewer",
		"ownerType": "system",
		"scopes": ["view:Dashboard", "view:Project"],
		"organizationUuid": null,
		"createdAt": null,
		"updatedAt": null,
		"createdBy": null
	}`

	var role Role
	if err := json.Unmarshal([]byte(fixture), &role); err != nil {
		t.Fatalf("unmarshal Role: %v", err)
	}

	if role.RoleUUID != "viewer" {
		t.Errorf("RoleUUID = %q, want viewer", role.RoleUUID)
	}
	if role.Name != "Viewer" {
		t.Errorf("Name = %q, want Viewer", role.Name)
	}
	if role.OwnerType != "system" {
		t.Errorf("OwnerType = %q, want system", role.OwnerType)
	}
	if len(role.Scopes) != 2 {
		t.Errorf("Scopes len = %d, want 2", len(role.Scopes))
	}
}

func TestRoleAssignment_UnmarshalJSON_orgUserAssignment(t *testing.T) {
	const fixture = `{
		"roleId": "admin",
		"roleName": "admin",
		"assigneeType": "user",
		"ownerType": "system",
		"assigneeId": "b61b8510-dca3-4cab-8080-db170dd9a2ee",
		"organizationId": "089a18c4-667e-41cb-9d10-b088461ac941",
		"createdAt": "2023-08-28T18:27:36.097Z",
		"updatedAt": "2023-08-28T18:27:36.097Z"
	}`

	var assignment RoleAssignment
	if err := json.Unmarshal([]byte(fixture), &assignment); err != nil {
		t.Fatalf("unmarshal RoleAssignment: %v", err)
	}

	if assignment.RoleID != "admin" {
		t.Errorf("RoleID = %q, want admin", assignment.RoleID)
	}
	if assignment.AssigneeType != "user" {
		t.Errorf("AssigneeType = %q, want user", assignment.AssigneeType)
	}
	if assignment.OrganizationID != "089a18c4-667e-41cb-9d10-b088461ac941" {
		t.Errorf("OrganizationID = %q", assignment.OrganizationID)
	}
}

func TestRoleAssignment_UnmarshalJSON_projectUserAssignment(t *testing.T) {
	const fixture = `{
		"roleId": "editor",
		"roleName": "editor",
		"ownerType": "system",
		"assigneeType": "user",
		"assigneeId": "83d3ce52-4e26-4005-96b2-e3e945ac34ca",
		"assigneeName": "Test User",
		"projectId": "9cc0bae8-f552-4ac0-bdcc-44933d7031ae",
		"createdAt": "2026-06-18T09:37:53.439Z",
		"updatedAt": "2026-06-18T09:37:53.439Z"
	}`

	var assignment RoleAssignment
	if err := json.Unmarshal([]byte(fixture), &assignment); err != nil {
		t.Fatalf("unmarshal RoleAssignment: %v", err)
	}

	if assignment.RoleName != "editor" {
		t.Errorf("RoleName = %q, want editor", assignment.RoleName)
	}
	if assignment.ProjectID != "9cc0bae8-f552-4ac0-bdcc-44933d7031ae" {
		t.Errorf("ProjectID = %q", assignment.ProjectID)
	}
	if assignment.AssigneeName != "Test User" {
		t.Errorf("AssigneeName = %q, want Test User", assignment.AssigneeName)
	}
}
