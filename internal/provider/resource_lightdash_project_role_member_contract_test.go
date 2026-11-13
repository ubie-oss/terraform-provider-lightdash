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
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const contractProjectMemberResource = "lightdash_project_role_member.test"

func contractProjectMemberConfig(f *fakeLightdash, role string) string {
	return contractProviderBlock(f) + fmt.Sprintf(`
resource "lightdash_project_role_member" "test" {
  project_uuid = %q
  user_uuid    = %q
  role         = %q
}
`, contractProjectUUID, contractUserUUID, role)
}

// TestContractProjectRoleMemberLifecycle checks create and update against the fake
// API, and that destroy removes the assignment.
func TestContractProjectRoleMemberLifecycle(t *testing.T) {
	skipInAcceptanceMode(t)
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkNoRole(contractProjectUserRef),
		Steps: []resource.TestStep{
			{
				Config: contractProjectMemberConfig(f, "editor"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contractProjectMemberResource, "role", "editor"),
					f.checkRole(contractProjectUserRef, "editor"),
				),
			},
			{
				Config: contractProjectMemberConfig(f, "viewer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contractProjectMemberResource, "role", "viewer"),
					f.checkRole(contractProjectUserRef, "viewer"),
				),
			},
		},
	})
}

// TestContractProjectRoleMemberImport checks that importing by
// "projects/{project}/access/{user}" rebuilds state that matches the managed state.
func TestContractProjectRoleMemberImport(t *testing.T) {
	skipInAcceptanceMode(t)
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkNoRole(contractProjectUserRef),
		Steps: []resource.TestStep{
			{
				Config: contractProjectMemberConfig(f, "editor"),
				Check:  f.checkRole(contractProjectUserRef, "editor"),
			},
			{
				ResourceName:            contractProjectMemberResource,
				ImportState:             true,
				ImportStateId:           "projects/" + contractProjectUUID + "/access/" + contractUserUUID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			{
				// The first plan after import must be empty, or the member is updated for no reason.
				Config:   contractProjectMemberConfig(f, "editor"),
				PlanOnly: true,
			},
		},
	})
}

// TestContractProjectRoleMemberRecreatesRemovedMember checks that a member removed
// outside Terraform is granted again by the next apply, instead of staying in state.
func TestContractProjectRoleMemberRecreatesRemovedMember(t *testing.T) {
	skipInAcceptanceMode(t)
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkNoRole(contractProjectUserRef),
		Steps: []resource.TestStep{
			{
				Config: contractProjectMemberConfig(f, "editor"),
				Check:  f.checkRole(contractProjectUserRef, "editor"),
			},
			{
				PreConfig: func() { f.clear(contractProjectUserRef) },
				Config:    contractProjectMemberConfig(f, "editor"),
				Check:     f.checkRole(contractProjectUserRef, "editor"),
			},
		},
	})
}

// TestContractProjectRoleMemberDetectsDrift checks that a role changed outside
// Terraform shows up as a planned change.
func TestContractProjectRoleMemberDetectsDrift(t *testing.T) {
	skipInAcceptanceMode(t)
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: contractProjectMemberConfig(f, "editor"),
				Check:  f.checkRole(contractProjectUserRef, "editor"),
			},
			{
				PreConfig:          func() { f.set(contractProjectUserRef, "admin") },
				Config:             contractProjectMemberConfig(f, "editor"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestContractProjectRoleMemberRejectsMixedCaseRole checks that a spelling other than
// the lowercase slug fails at validation, before any assignment is made.
func TestContractProjectRoleMemberRejectsMixedCaseRole(t *testing.T) {
	skipInAcceptanceMode(t)
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkNoRole(contractProjectUserRef),
		Steps: []resource.TestStep{
			{
				Config:      contractProjectMemberConfig(f, "Editor"),
				ExpectError: regexp.MustCompile(`must be one of: viewer, interactive_viewer, editor, developer, admin`),
			},
		},
	})
}
