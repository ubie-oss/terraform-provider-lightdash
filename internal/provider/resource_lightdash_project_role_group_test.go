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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	grantConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_project_role_group", "grant", "010_grant.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	grantConfig020, err := ReadAccTestResource([]string{"resources", "lightdash_project_role_group", "grant", "020_grant.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	grantConfig030, err := ReadAccTestResource([]string{"resources", "lightdash_project_role_group", "grant", "030_grant.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	grantConfig040, err := ReadAccTestResource([]string{"resources", "lightdash_project_role_group", "grant", "040_grant.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	grantConfig050, err := ReadAccTestResource([]string{"resources", "lightdash_project_role_group", "grant", "050_grant.tf"})
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

func TestAccProjectRoleGroupResource_import(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project_role_group")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of simple space creation
	importConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_project_role_group", "import", "010_import.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + importConfig010,
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
				Config:            providerConfig + importConfig010,
				ResourceName:      "lightdash_project_role_group.test_project_role_group",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"last_updated",
				},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					res, ok := state.RootModule().Resources["lightdash_project_role_group.test_project_role_group"]
					if !ok {
						return "", fmt.Errorf("resource not found in state for import")
					}
					// Get the project_uuid from the state
					project_uuid, ok := res.Primary.Attributes["project_uuid"]
					if !ok || project_uuid == "" {
						return "", fmt.Errorf("project_uuid attribute not present in state")
					}
					// Get the group_uuid from the state
					group_uuid, ok := res.Primary.Attributes["group_uuid"]
					if !ok || group_uuid == "" {
						return "", fmt.Errorf("group_uuid attribute not present in state")
					}
					// Construct the import ID in the form 'projects/<project_uuid>/access-groups/<group_uuid>'
					id := getProjectRoleGroupResourceId(project_uuid, group_uuid)
					return id, nil
				},
			},
		},
	})
}
