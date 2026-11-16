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

// spellingVariants returns spellings the resolver must treat as the same
// identifier: casing, surrounding whitespace, and space/underscore separators.
func spellingVariants(s string) []string {
	spaced := strings.ReplaceAll(s, "_", " ")
	return []string{
		s,
		strings.ToUpper(s),
		"  " + s + "\t",
		spaced,
		strings.ToUpper(spaced),
		" " + spaced + " ",
	}
}

// singleEdits returns every string one deletion, insertion, or substitution away from s.
func singleEdits(s string, alphabet []rune) []string {
	runes := []rune(s)
	var out []string
	for i := 0; i <= len(runes); i++ {
		for _, r := range alphabet {
			out = append(out, string(runes[:i])+string(r)+string(runes[i:]))
		}
	}
	for i := range runes {
		out = append(out, string(runes[:i])+string(runes[i+1:]))
		for _, r := range alphabet {
			edited := append([]rune{}, runes...)
			edited[i] = r
			out = append(out, string(edited))
		}
	}
	return out
}

// allStringsUpTo returns every string of length 0..maxLen over alphabet.
func allStringsUpTo(alphabet []rune, maxLen int) []string {
	out := []string{""}
	frontier := []string{""}
	for n := 0; n < maxLen; n++ {
		var next []string
		for _, s := range frontier {
			for _, r := range alphabet {
				next = append(next, s+string(r))
			}
		}
		out = append(out, next...)
		frontier = next
	}
	return out
}

// TestResolveRoleIDFromRoles_everySpellingResolvesToItsRole checks that every
// role in the catalog resolves from its UUID and display name in every spelling variant.
func TestResolveRoleIDFromRoles_everySpellingResolvesToItsRole(t *testing.T) {
	roles := testSystemRoles()
	for _, role := range roles {
		candidates := append(spellingVariants(role.RoleUUID), spellingVariants(role.Name)...)
		for _, candidate := range candidates {
			got, err := resolveRoleIDFromRoles(roles, candidate)
			if err != nil {
				t.Errorf("resolve(%q): unexpected error: %v", candidate, err)
				continue
			}
			if got != role.RoleUUID {
				t.Errorf("resolve(%q) = %q, want %q", candidate, got, role.RoleUUID)
			}
		}
	}
}

// matchesRole is the resolution rule written out independently of the code under test:
// the trimmed, lowercased input equals the role UUID, or the name with spaces
// replaced by underscores, or the name itself.
func matchesRole(role models.Role, candidate string) bool {
	key := strings.ToLower(strings.TrimSpace(candidate))
	return key == strings.ToLower(role.RoleUUID) ||
		key == strings.ReplaceAll(strings.ToLower(role.Name), " ", "_") ||
		key == strings.ToLower(role.Name)
}

// TestResolveRoleIDFromRoles_succeedsOnlyOnUniqueMatch checks, over a neighborhood
// of every catalog identifier plus all short strings, that resolution succeeds only
// when exactly one role matches, and then returns that role. The catalog includes a
// duplicate display name so the ambiguous path is exercised too.
func TestResolveRoleIDFromRoles_succeedsOnlyOnUniqueMatch(t *testing.T) {
	catalogs := map[string][]models.Role{
		"system roles": testSystemRoles(),
		"with duplicate name": append(testSystemRoles(),
			models.Role{RoleUUID: "custom-editor", Name: "Editor", OwnerType: "user"},
		),
	}
	alphabet := []rune("vieweradmn_ ")

	for catalogName, roles := range catalogs {
		candidates := allStringsUpTo(alphabet, 2)
		for _, role := range roles {
			for _, id := range append(spellingVariants(role.RoleUUID), spellingVariants(role.Name)...) {
				candidates = append(candidates, singleEdits(id, alphabet)...)
			}
		}

		successes, ambiguous := 0, 0
		for _, candidate := range candidates {
			var wantMatches []models.Role
			for _, role := range roles {
				if matchesRole(role, candidate) {
					wantMatches = append(wantMatches, role)
				}
			}

			got, err := resolveRoleIDFromRoles(roles, candidate)
			switch len(wantMatches) {
			case 1:
				successes++
				if err != nil {
					t.Errorf("[%s] resolve(%q): unexpected error: %v", catalogName, candidate, err)
				} else if got != wantMatches[0].RoleUUID {
					t.Errorf("[%s] resolve(%q) = %q, want %q", catalogName, candidate, got, wantMatches[0].RoleUUID)
				}
			default:
				if len(wantMatches) > 1 {
					ambiguous++
				}
				if err == nil {
					t.Errorf("[%s] resolve(%q) = %q, want error (%d matches)", catalogName, candidate, got, len(wantMatches))
				}
			}
		}
		if successes == 0 {
			t.Fatalf("[%s] no candidate resolved; the enumeration is not exercising the success path", catalogName)
		}
		if catalogName == "with duplicate name" && ambiguous == 0 {
			t.Fatal("duplicate-name catalog never produced an ambiguous candidate")
		}
	}
}

// TestTerraformProjectRoleFromAssignment_mapsEveryValidRoleSpelling checks that
// every valid project role maps back from each spelling of its name or ID.
func TestTerraformProjectRoleFromAssignment_mapsEveryValidRoleSpelling(t *testing.T) {
	validRoles := []models.ProjectMemberRole{
		models.PROJECT_VIEWER_ROLE,
		models.PROJECT_INTERACTIVE_VIEWER_ROLE,
		models.PROJECT_EDITOR_ROLE,
		models.PROJECT_DEVELOPER_ROLE,
		models.PROJECT_ADMIN_ROLE,
	}
	for _, want := range validRoles {
		for _, spelling := range spellingVariants(want.String()) {
			for _, assignment := range []models.RoleAssignment{
				{RoleID: spelling, RoleName: spelling},
				{RoleID: "not_a_role", RoleName: spelling},
				{RoleID: spelling, RoleName: "not_a_role"},
			} {
				got, err := TerraformProjectRoleFromAssignment(&assignment)
				if err != nil {
					t.Errorf("assignment %+v: unexpected error: %v", assignment, err)
					continue
				}
				if got != want {
					t.Errorf("assignment %+v = %q, want %q", assignment, got, want)
				}
			}
		}
	}
}

// TestTerraformProjectRoleFromAssignment_rejectsNonProjectRoles checks that roles
// valid elsewhere, or unknown ones, never map to a project role.
func TestTerraformProjectRoleFromAssignment_rejectsNonProjectRoles(t *testing.T) {
	for _, bad := range []string{"", "member", "superadmin", "viewer2", "  "} {
		got, err := TerraformProjectRoleFromAssignment(&models.RoleAssignment{RoleID: bad, RoleName: bad})
		if err == nil {
			t.Errorf("role %q mapped to %q, want error", bad, got)
		}
	}
}

// FuzzTerraformProjectRoleFromAssignment checks the core safety property: a
// successful mapping always yields a valid project role.
func FuzzTerraformProjectRoleFromAssignment(f *testing.F) {
	for _, seed := range []string{"editor", "Interactive Viewer", "admin", "member", "", "x_y"} {
		f.Add(seed, seed)
	}
	f.Fuzz(func(t *testing.T, roleID string, roleName string) {
		got, err := TerraformProjectRoleFromAssignment(&models.RoleAssignment{RoleID: roleID, RoleName: roleName})
		if err == nil && !got.IsValid() {
			t.Errorf("roleID=%q roleName=%q mapped to invalid role %q", roleID, roleName, got)
		}
	})
}
