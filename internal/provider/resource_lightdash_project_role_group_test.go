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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Using the shared testAccPreCheck and testAccProtoV6ProviderFactories from provider_acc_test.go

func TestAccProjectRoleGroupResource_grant(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project_role_group")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of simple space creation
	grantConfig010, err := ReadAccTestResource([]string{"resource_lightdash_project_role_group", "grant", "010_grant.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	grantConfig020, err := ReadAccTestResource([]string{"resource_lightdash_project_role_group", "grant", "020_grant.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	grantConfig030, err := ReadAccTestResource([]string{"resource_lightdash_project_role_group", "grant", "030_grant.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	grantConfig040, err := ReadAccTestResource([]string{"resource_lightdash_project_role_group", "grant", "040_grant.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	grantConfig050, err := ReadAccTestResource([]string{"resource_lightdash_project_role_group", "grant", "050_grant.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + grantConfig010,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_project_role_group.resource_lightdash_project_role_group__grant
					resource.TestCheckResourceAttrSet("lightdash_project_role_group.test_project_role_group", "project_uuid"),
					resource.TestCheckResourceAttr("lightdash_project_role_group.test_project_role_group", "role", "viewer"),
					resource.TestCheckResourceAttrPair(
						"lightdash_project_role_group.test_project_role_group",
						"group_uuid",
						"lightdash_group.test_group",
						"group_uuid",
					),
				),
			},
			{
				Config: providerConfig + grantConfig020,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_project_role_group.resource_lightdash_project_role_group__grant
					resource.TestCheckResourceAttr("lightdash_project_role_group.test_project_role_group", "role", "interactive_viewer"),
				),
			},
			{
				Config: providerConfig + grantConfig030,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_project_role_group.resource_lightdash_project_role_group__grant
					resource.TestCheckResourceAttr("lightdash_project_role_group.test_project_role_group", "role", "editor"),
				),
			},
			{
				Config: providerConfig + grantConfig040,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_project_role_group.resource_lightdash_project_role_group__grant
					resource.TestCheckResourceAttr("lightdash_project_role_group.test_project_role_group", "role", "developer"),
				),
			},
			{
				Config: providerConfig + grantConfig050,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_project_role_group.resource_lightdash_project_role_group__grant
					resource.TestCheckResourceAttr("lightdash_project_role_group.test_project_role_group", "role", "admin"),
				),
			},
		},
	})
}
