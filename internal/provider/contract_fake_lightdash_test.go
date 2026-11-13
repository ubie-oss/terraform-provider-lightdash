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

// Offline contract tests for the role resources. The provider runs in-process under
// Terraform against fakeLightdash, an in-memory stand-in for the Lightdash endpoints
// the resources call. No credentials or live instance are needed, and no real access
// is changed.
//
// services.GetRoleService is a process-wide singleton bound to the first client it
// sees, so every contract test must reach the same server. contractFake provides it.

const (
	contractOrgUUID     = "11111111-1111-4111-8111-111111111111"
	contractProjectUUID = "22222222-2222-4222-8222-222222222222"
	contractUserUUID    = "33333333-3333-4333-8333-333333333333"
	contractGroupUUID   = "44444444-4444-4444-8444-444444444444"
	contractMemberEmail = "contract.user@example.com"
)

// contractRoles maps each role slug to its display name. "member" is the org default role.
var contractRoles = map[string]string{
	"member":             "Member",
	"viewer":             "Viewer",
	"interactive_viewer": "Interactive Viewer",
	"editor":             "Editor",
	"developer":          "Developer",
	"admin":              "Admin",
}

// contractRef identifies one role assignment: its scope (org or project UUID),
// assignee type, and assignee UUID.
type contractRef struct {
	scope, assigneeType, assigneeID string
}

var (
	contractOrgUserRef      = contractRef{contractOrgUUID, "user", contractUserUUID}
	contractProjectUserRef  = contractRef{contractProjectUUID, "user", contractUserUUID}
	contractProjectGroupRef = contractRef{contractProjectUUID, "group", contractGroupUUID}
)

// fakeLightdash holds role assignments in memory. Tests reset it before running.
type fakeLightdash struct {
	mu    sync.Mutex
	url   string
	roles map[contractRef]string // assignment -> roleID
}

// contractFake starts the one shared fake server for the test process.
var contractFake = sync.OnceValue(func() *fakeLightdash {
	f := &fakeLightdash{roles: map[contractRef]string{}}
	f.url = httptest.NewServer(f.routes()).URL
	return f
})

// contractProviderBlock configures the provider to talk to the fake server.
func contractProviderBlock(f *fakeLightdash) string {
	return fmt.Sprintf(`
provider "lightdash" {
  host  = %q
  token = "contract-test-token"
}
`, f.url)
}

// skipInAcceptanceMode keeps these tests out of the TF_ACC run. Acceptance tests in
// the same process bind the shared role service to a live instance, so these tests
// would otherwise send their requests to real Lightdash.
func skipInAcceptanceMode(t *testing.T) {
	t.Helper()
	if isIntegrationTestMode() {
		t.Skip("offline contract test; runs only without TF_ACC")
	}
}

func (f *fakeLightdash) reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.roles = map[contractRef]string{}
}

func (f *fakeLightdash) set(ref contractRef, roleID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.roles[ref] = roleID
}

// clear removes the assignment, as if it was changed outside Terraform.
func (f *fakeLightdash) clear(ref contractRef) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.roles, ref)
}

// remove deletes the assignment and reports whether it existed.
func (f *fakeLightdash) remove(ref contractRef) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.roles[ref]
	delete(f.roles, ref)
	return ok
}

func (f *fakeLightdash) get(ref contractRef) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.roles[ref]
}

func (f *fakeLightdash) checkRole(ref contractRef, roleID string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if got := f.get(ref); got != roleID {
			return fmt.Errorf("API role for %s %s = %q, want %q", ref.assigneeType, ref.assigneeID, got, roleID)
		}
		return nil
	}
}

func (f *fakeLightdash) checkNoRole(ref contractRef) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if got := f.get(ref); got != "" {
			return fmt.Errorf("%s %s still has API role %q", ref.assigneeType, ref.assigneeID, got)
		}
		return nil
	}
}

func (f *fakeLightdash) routes() *http.ServeMux {
	orgRoles := "/api/v2/orgs/" + contractOrgUUID + "/roles"
	orgAssignments := orgRoles + "/assignments"
	projectAssignments := "/api/v2/projects/" + contractProjectUUID + "/roles/assignments"

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/org", fixedResult(map[string]any{"organizationUuid": contractOrgUUID, "name": "contract-org"}))
	mux.HandleFunc("GET /api/v1/org/users/"+contractUserUUID, fixedResult(map[string]any{
		"organizationUuid": contractOrgUUID,
		"userUuid":         contractUserUUID,
		"email":            contractMemberEmail,
		"firstName":        "Contract",
		"lastName":         "User",
		"role":             "viewer",
		"isActive":         true,
		"isInviteExpired":  false,
	}))
	mux.HandleFunc("GET /api/v1/groups/"+contractGroupUUID, fixedResult(map[string]any{
		"organizationUuid": contractOrgUUID,
		"uuid":             contractGroupUUID,
		"name":             "contract-group",
		"createdAt":        "2026-01-01T00:00:00Z",
	}))

	mux.HandleFunc("GET "+orgRoles, f.listRoles)
	mux.HandleFunc("GET "+orgAssignments, f.listAssignments(contractOrgUUID))
	mux.HandleFunc("POST "+orgAssignments+"/user/"+contractUserUUID, f.assignRole(contractOrgUserRef))

	mux.HandleFunc("GET "+projectAssignments, f.listAssignments(contractProjectUUID))
	mux.HandleFunc("POST "+projectAssignments+"/user/"+contractUserUUID, f.assignRole(contractProjectUserRef))
	mux.HandleFunc("DELETE "+projectAssignments+"/user/"+contractUserUUID, f.removeRole(contractProjectUserRef))
	mux.HandleFunc("POST "+projectAssignments+"/group/"+contractGroupUUID, f.assignRole(contractProjectGroupRef))
	mux.HandleFunc("PATCH "+projectAssignments+"/group/"+contractGroupUUID, f.assignRole(contractProjectGroupRef))
	mux.HandleFunc("DELETE "+projectAssignments+"/group/"+contractGroupUUID, f.removeRole(contractProjectGroupRef))
	return mux
}

func fixedResult(results any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeContractOK(w, results)
	}
}

func (f *fakeLightdash) listRoles(w http.ResponseWriter, r *http.Request) {
	roles := make([]map[string]any, 0, len(contractRoles))
	for slug, name := range contractRoles {
		roles = append(roles, map[string]any{"roleUuid": slug, "name": name, "ownerType": "system"})
	}
	writeContractOK(w, roles)
}

// listAssignments returns every assignment whose scope is the given org or project UUID.
func (f *fakeLightdash) listAssignments(scope string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		assignments := []map[string]any{}
		for ref, roleID := range f.roles {
			if ref.scope == scope {
				assignments = append(assignments, contractAssignment(ref, roleID))
			}
		}
		writeContractOK(w, assignments)
	}
}

// assignRole creates or updates the assignment. Unknown roles are rejected.
func (f *fakeLightdash) assignRole(ref contractRef) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			RoleID string `json:"roleId"`
		}
		err := json.NewDecoder(r.Body).Decode(&body)
		_, known := contractRoles[body.RoleID]
		if err != nil || !known {
			writeContractError(w, http.StatusBadRequest, "unknown role")
			return
		}
		f.set(ref, body.RoleID)
		writeContractOK(w, contractAssignment(ref, body.RoleID))
	}
}

func (f *fakeLightdash) removeRole(ref contractRef) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !f.remove(ref) {
			writeContractError(w, http.StatusNotFound, "assignment not found")
			return
		}
		writeContractOK(w, map[string]any{})
	}
}

func contractAssignment(ref contractRef, roleID string) map[string]any {
	scopeKey := "projectId"
	if ref.scope == contractOrgUUID {
		scopeKey = "organizationId"
	}
	return map[string]any{
		"roleId":       roleID,
		"roleName":     roleID,
		"assigneeType": ref.assigneeType,
		"assigneeId":   ref.assigneeID,
		"ownerType":    "system",
		scopeKey:       ref.scope,
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
