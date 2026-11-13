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

const contractProjectGroupResource = "lightdash_project_role_group.test"

func contractProjectGroupConfig(f *fakeLightdash, role string) string {
	return contractProviderBlock(f) + fmt.Sprintf(`
resource "lightdash_project_role_group" "test" {
  project_uuid = %q
  group_uuid   = %q
  role         = %q
}
`, contractProjectUUID, contractGroupUUID, role)
}

// TestContractProjectRoleGroupLifecycle checks create, update, and destroy of a group
// role against the fake API.
func TestContractProjectRoleGroupLifecycle(t *testing.T) {
	skipInAcceptanceMode(t)
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkNoRole(contractProjectGroupRef),
		Steps: []resource.TestStep{
			{
				Config: contractProjectGroupConfig(f, "editor"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contractProjectGroupResource, "role", "editor"),
					f.checkRole(contractProjectGroupRef, "editor"),
				),
			},
			{
				Config: contractProjectGroupConfig(f, "viewer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contractProjectGroupResource, "role", "viewer"),
					f.checkRole(contractProjectGroupRef, "viewer"),
				),
			},
		},
	})
}

// TestContractProjectRoleGroupRecreatesRemovedGroup checks that a group removed outside
// Terraform is granted again by the next apply, instead of staying in state.
func TestContractProjectRoleGroupRecreatesRemovedGroup(t *testing.T) {
	skipInAcceptanceMode(t)
	f := contractFake()
	f.reset()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkNoRole(contractProjectGroupRef),
		Steps: []resource.TestStep{
			{
				Config: contractProjectGroupConfig(f, "editor"),
				Check:  f.checkRole(contractProjectGroupRef, "editor"),
			},
			{
				PreConfig: func() { f.clear(contractProjectGroupRef) },
				Config:    contractProjectGroupConfig(f, "editor"),
				Check:     f.checkRole(contractProjectGroupRef, "editor"),
			},
		},
	})
}
