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

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Offline contract tests for lightdash_project_role_member. The real provider
// runs in-process under Terraform against fakeLightdash, an in-memory stand-in
// for the Lightdash endpoints the resource calls. No credentials or live
// instance are needed, and no real access is changed.

const (
	contractOrgUUID     = "11111111-1111-4111-8111-111111111111"
	contractProjectUUID = "22222222-2222-4222-8222-222222222222"
	contractUserUUID    = "33333333-3333-4333-8333-333333333333"
	contractMemberEmail = "contract.user@example.com"
	contractResourceRef = "lightdash_project_role_member.test"
)

var contractSystemRoles = []string{"viewer", "interactive_viewer", "editor", "developer", "admin"}

// fakeLightdash holds project role assignments in memory. It is the only mutable
// state the contract tests rely on, so tests reset it before running.
type fakeLightdash struct {
	mu          sync.Mutex
	url         string
	assignments map[string]string // "projectUUID/userUUID" -> roleID
}

// contractFake starts one shared fake server for the test process. The provider's
// role catalog is cached in a process-wide singleton bound to the first server
// it sees, so every test must reuse this one.
var contractFake = sync.OnceValue(func() *fakeLightdash {
	f := &fakeLightdash{assignments: map[string]string{}}
	f.url = httptest.NewServer(f).URL
	return f
})

func (f *fakeLightdash) reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.assignments = map[string]string{}
}

func (f *fakeLightdash) assign(projectUUID, userUUID, roleID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.assignments[projectUUID+"/"+userUUID] = roleID
}

func (f *fakeLightdash) roleOf(projectUUID, userUUID string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	role, ok := f.assignments[projectUUID+"/"+userUUID]
	return role, ok
}

func (f *fakeLightdash) checkAssigned(projectUUID, userUUID, roleID string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		got, ok := f.roleOf(projectUUID, userUUID)
		if !ok || got != roleID {
			return fmt.Errorf("API role for user %s = %q (assigned=%v), want %q", userUUID, got, ok, roleID)
		}
		return nil
	}
}

func (f *fakeLightdash) checkNotAssigned(projectUUID, userUUID string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if got, ok := f.roleOf(projectUUID, userUUID); ok {
			return fmt.Errorf("user %s still has API role %q after destroy", userUUID, got)
		}
		return nil
	}
}

func (f *fakeLightdash) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	assignPath := "/api/v2/projects/" + contractProjectUUID + "/roles/assignments/user/" + contractUserUUID
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/org":
		writeContractOK(w, map[string]any{"organizationUuid": contractOrgUUID, "name": "contract-org"})

	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/org/users/"+contractUserUUID:
		writeContractOK(w, map[string]any{
			"organizationUuid": contractOrgUUID,
			"userUuid":         contractUserUUID,
			"email":            contractMemberEmail,
			"firstName":        "Contract",
			"lastName":         "User",
			"role":             "viewer",
			"isActive":         true,
			"isInviteExpired":  false,
		})

	case r.Method == http.MethodGet && r.URL.Path == "/api/v2/orgs/"+contractOrgUUID+"/roles":
		roles := make([]map[string]any, 0, len(contractSystemRoles))
		for _, slug := range contractSystemRoles {
			roles = append(roles, map[string]any{"roleUuid": slug, "name": contractRoleDisplayName(slug), "ownerType": "system"})
		}
		writeContractOK(w, roles)

	case r.Method == http.MethodGet && r.URL.Path == "/api/v2/projects/"+contractProjectUUID+"/roles/assignments":
		assignments := []map[string]any{}
		if roleID, ok := f.assignments[contractProjectUUID+"/"+contractUserUUID]; ok {
			assignments = append(assignments, contractAssignment(roleID))
		}
		writeContractOK(w, assignments)

	case r.Method == http.MethodPost && r.URL.Path == assignPath:
		var body struct {
			RoleID string `json:"roleId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !isContractRole(body.RoleID) {
			writeContractError(w, http.StatusBadRequest, "unknown role")
			return
		}
		f.assignments[contractProjectUUID+"/"+contractUserUUID] = body.RoleID
		writeContractOK(w, contractAssignment(body.RoleID))

	case r.Method == http.MethodDelete && r.URL.Path == assignPath:
		key := contractProjectUUID + "/" + contractUserUUID
		if _, ok := f.assignments[key]; !ok {
			writeContractError(w, http.StatusNotFound, "assignment not found")
			return
		}
		delete(f.assignments, key)
		writeContractOK(w, map[string]any{})

	default:
		writeContractError(w, http.StatusNotFound, "API endpoint not found")
	}
}

func contractRoleDisplayName(slug string) string {
	names := map[string]string{
		"viewer":             "Viewer",
		"interactive_viewer": "Interactive Viewer",
		"editor":             "Editor",
		"developer":          "Developer",
		"admin":              "Admin",
	}
	return names[slug]
}

func isContractRole(roleID string) bool {
	for _, slug := range contractSystemRoles {
		if slug == roleID {
			return true
		}
	}
	return false
}

func contractAssignment(roleID string) map[string]any {
	return map[string]any{
		"roleId":       roleID,
		"roleName":     roleID,
		"assigneeType": "user",
		"assigneeId":   contractUserUUID,
		"ownerType":    "system",
		"projectId":    contractProjectUUID,
		"createdAt":    "2026-01-01T00:00:00Z",
		"updatedAt":    "2026-01-01T00:00:00Z",
	}
}

func writeContractOK(w http.ResponseWriter, results any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "results": results})
}

func writeContractError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "error",
		"error":  map[string]any{"statusCode": code, "message": message},
	})
}

func contractConfig(f *fakeLightdash, role string) string {
	return fmt.Sprintf(`
provider "lightdash" {
  host  = %q
  token = "contract-test-token"
}

resource "lightdash_project_role_member" "test" {
  project_uuid = %q
  user_uuid    = %q
  role         = %q
}
`, f.url, contractProjectUUID, contractUserUUID, role)
}

// TestContractProjectRoleMemberLifecycle checks create and update against the fake
// API, and that destroy removes the assignment.
func TestContractProjectRoleMemberLifecycle(t *testing.T) {
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkNotAssigned(contractProjectUUID, contractUserUUID),
		Steps: []resource.TestStep{
			{
				Config: contractConfig(f, "editor"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contractResourceRef, "role", "editor"),
					f.checkAssigned(contractProjectUUID, contractUserUUID, "editor"),
				),
			},
			{
				Config: contractConfig(f, "viewer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contractResourceRef, "role", "viewer"),
					f.checkAssigned(contractProjectUUID, contractUserUUID, "viewer"),
				),
			},
		},
	})
}

// TestContractProjectRoleMemberImport checks that importing by
// "projects/{project}/access/{user}" rebuilds state that matches the managed state.
func TestContractProjectRoleMemberImport(t *testing.T) {
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkNotAssigned(contractProjectUUID, contractUserUUID),
		Steps: []resource.TestStep{
			{
				Config: contractConfig(f, "editor"),
				Check:  f.checkAssigned(contractProjectUUID, contractUserUUID, "editor"),
			},
			{
				ResourceName:            contractResourceRef,
				ImportState:             true,
				ImportStateId:           "projects/" + contractProjectUUID + "/access/" + contractUserUUID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"send_email", "last_updated"},
			},
		},
	})
}

// TestContractProjectRoleMemberDetectsDrift checks that a role changed outside
// Terraform shows up as a planned change.
func TestContractProjectRoleMemberDetectsDrift(t *testing.T) {
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: contractConfig(f, "editor"),
				Check:  f.checkAssigned(contractProjectUUID, contractUserUUID, "editor"),
			},
			{
				PreConfig:          func() { f.assign(contractProjectUUID, contractUserUUID, "admin") },
				Config:             contractConfig(f, "editor"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
