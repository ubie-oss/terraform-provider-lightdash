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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const contractOrgMemberResource = "lightdash_organization_role_member.test"

func contractOrgMemberConfig(f *fakeLightdash, role string) string {
	return contractProviderBlock(f) + fmt.Sprintf(`
resource "lightdash_organization_role_member" "test" {
  organization_uuid = %q
  user_uuid         = %q
  role              = %q
}
`, contractOrgUUID, contractUserUUID, role)
}

// TestContractOrganizationRoleMemberLifecycle checks create and update against the fake
// API. Destroy does not remove the member: the resource grants the default "member" role.
func TestContractOrganizationRoleMemberLifecycle(t *testing.T) {
	skipInAcceptanceMode(t)
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkRole(contractOrgUserRef, "member"),
		Steps: []resource.TestStep{
			{
				Config: contractOrgMemberConfig(f, "editor"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contractOrgMemberResource, "role", "editor"),
					f.checkRole(contractOrgUserRef, "editor"),
				),
			},
			{
				Config: contractOrgMemberConfig(f, "viewer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contractOrgMemberResource, "role", "viewer"),
					f.checkRole(contractOrgUserRef, "viewer"),
				),
			},
		},
	})
}

// TestContractOrganizationRoleMemberRecreatesRemovedMember checks that a member removed
// outside Terraform is granted again by the next apply, instead of staying in state.
func TestContractOrganizationRoleMemberRecreatesRemovedMember(t *testing.T) {
	skipInAcceptanceMode(t)
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkRole(contractOrgUserRef, "member"),
		Steps: []resource.TestStep{
			{
				Config: contractOrgMemberConfig(f, "editor"),
				Check:  f.checkRole(contractOrgUserRef, "editor"),
			},
			{
				PreConfig: func() { f.clear(contractOrgUserRef) },
				Config:    contractOrgMemberConfig(f, "editor"),
				Check:     f.checkRole(contractOrgUserRef, "editor"),
			},
		},
	})
}
