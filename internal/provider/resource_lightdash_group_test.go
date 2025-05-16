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

func TestAccGroupResource_grant(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_group")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of simple space creation
	createConfig010, err := ReadAccTestResource([]string{"resource_lightdash_group", "create", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get createConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig010,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_group.test_group
					resource.TestCheckResourceAttr("lightdash_group.test_group", "name", "test (Acceptance Test - create)"),
					resource.TestCheckResourceAttr("lightdash_group.test_group", "members.#", "0"),
				),
			},
		},
	})
}

func TestAccGroupResource_import(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_group")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of simple space creation
	importConfig010, err := ReadAccTestResource([]string{"resource_lightdash_group", "import", "010_import.tf"})
	if err != nil {
		t.Fatalf("Failed to get importConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + importConfig010,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_group.test_group
					resource.TestCheckResourceAttr("lightdash_group.test_group", "name", "test (Acceptance Test - import)"),
					resource.TestCheckResourceAttr("lightdash_group.test_group", "members.#", "0"),
				),
			},
			{
				Config:                  providerConfig + importConfig010,
				ResourceName:            "lightdash_group.test_group",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					res, ok := state.RootModule().Resources["lightdash_group.test_group"]
					if !ok {
						return "", fmt.Errorf("resource not found in state for import")
					}
					// Get the organization_uuid from the state
					organization_uuid, ok := res.Primary.Attributes["organization_uuid"]
					if !ok || organization_uuid == "" {
						return "", fmt.Errorf("organization_uuid attribute not present in state")
					}
					// Get the group_uuid from the state
					group_uuid, ok := res.Primary.Attributes["group_uuid"]
					if !ok || group_uuid == "" {
						return "", fmt.Errorf("group_uuid attribute not present in state")
					}
					// Construct the import ID in the form 'organizations/<organization_uuid>/groups/<group_uuid>'
					id := getGroupResourceId(organization_uuid, group_uuid)
					return id, nil
				},
			},
		},
	})
}
