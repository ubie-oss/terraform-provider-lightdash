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

func TestAccSpaceResource_simple(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_space")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of simple space creation
	simpleSpaceConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_space", "create_space", "010_create_space.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	simpleSpaceConfig020, err := ReadAccTestResource([]string{"resources", "lightdash_space", "create_space", "020_create_space.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + simpleSpaceConfig010,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_space.create_space__test_public
					resource.TestCheckResourceAttrSet("lightdash_space.create_space__test_public", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_public", "name", "Public Space (Acceptance Test: create_space)"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_public", "is_private", "false"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_public", "deletion_protection", "true"),
					// data.lightdash_space.test_public
					resource.TestCheckResourceAttrPair(
						"data.lightdash_space.create_space__test_public",
						"space_uuid",
						"lightdash_space.create_space__test_public",
						"space_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_space.create_space__test_public",
						"name",
						"lightdash_space.create_space__test_public",
						"name",
					),

					// lightdash_space.create_space__test_private
					resource.TestCheckResourceAttrSet("lightdash_space.create_space__test_private", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_private", "name", "Private Space (Acceptance Test: create_space)"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_private", "is_private", "true"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_private", "deletion_protection", "true"),
					// data.lightdash_space.test_private
					resource.TestCheckResourceAttrPair(
						"data.lightdash_space.create_space__test_private",
						"space_uuid",
						"lightdash_space.create_space__test_private",
						"space_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_space.create_space__test_private",
						"name",
						"lightdash_space.create_space__test_private",
						"name",
					),
				),
			},
			{
				Config: providerConfig + simpleSpaceConfig020,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_space.create_space__test_public
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_public", "name", "Public Space (Acceptance Test: create_space - 020)"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_public", "is_private", "true"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_public", "deletion_protection", "false"),
					// lightdash_space.create_space__test_private
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_private", "name", "Private Space (Acceptance Test: create_space - 020)"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_private", "is_private", "false"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_private", "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccSpaceResource_nested(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_space - nested")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of nested spaces
	nestedSpaceConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_space", "nested_space", "010_nested_space.tf"})
	if err != nil {
		t.Fatalf("Failed to get nestedSpaceConfig: %v", err)
	}
	nestedSpaceConfig020, err := ReadAccTestResource([]string{"resources", "lightdash_space", "nested_space", "020_nested_space.tf"})
	if err != nil {
		t.Fatalf("Failed to get nestedSpaceConfig: %v", err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + nestedSpaceConfig010,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_space.nested_space_public_root
					resource.TestCheckResourceAttrSet("lightdash_space.nested_space_public_root", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_root", "name", "Public Root Space (Acceptance Test: nested_space)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_root", "is_private", "false"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_root", "deletion_protection", "false"),
					// lightdash_space.nested_space_public_child
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_child", "name", "Public Child Space (Acceptance Test: nested_space)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_child", "deletion_protection", "false"),
					resource.TestCheckResourceAttrPair(
						"lightdash_space.nested_space_public_child",
						"parent_space_uuid",
						"lightdash_space.nested_space_public_root",
						"space_uuid",
					),
					// lightdash_space.nested_space_public_grandchild
					resource.TestCheckResourceAttrSet("lightdash_space.nested_space_public_grandchild", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_grandchild", "name", "Public Grandchild Space (Acceptance Test: nested_space)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_grandchild", "deletion_protection", "false"),
					resource.TestCheckResourceAttrPair(
						"lightdash_space.nested_space_public_grandchild",
						"parent_space_uuid",
						"lightdash_space.nested_space_public_child",
						"space_uuid",
					),

					// lightdash_space.nested_space_private_root
					resource.TestCheckResourceAttrSet("lightdash_space.nested_space_private_root", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_root", "name", "Private Root Space (Acceptance Test: nested_space)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_root", "is_private", "true"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_root", "deletion_protection", "false"),
					// lightdash_space.nested_space_private_child
					resource.TestCheckResourceAttrSet("lightdash_space.nested_space_private_child", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_child", "name", "Private Child Space (Acceptance Test: nested_space)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_child", "deletion_protection", "false"),
					resource.TestCheckResourceAttrPair(
						"lightdash_space.nested_space_private_child",
						"parent_space_uuid",
						"lightdash_space.nested_space_private_root",
						"space_uuid",
					),
					// lightdash_space.nested_space_private_grandchild
					resource.TestCheckResourceAttrSet("lightdash_space.nested_space_private_grandchild", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_grandchild", "name", "Private Grandchild Space (Acceptance Test: nested_space)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_grandchild", "deletion_protection", "false"),
					resource.TestCheckResourceAttrPair(
						"lightdash_space.nested_space_private_grandchild",
						"parent_space_uuid",
						"lightdash_space.nested_space_private_child",
						"space_uuid",
					),
				),
			},
			{
				Config: providerConfig + nestedSpaceConfig020,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_space.nested_space_public_root
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_root", "name", "Public Root Space (Acceptance Test: nested_space - 020)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_root", "is_private", "true"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_root", "deletion_protection", "false"),
					// lightdash_space.nested_space_public_child
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_child", "name", "Public Child Space (Acceptance Test: nested_space - 020)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_child", "is_private", "false"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_child", "deletion_protection", "false"),
					// lightdash_space.nested_space_public_grandchild
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_grandchild", "name", "Public Grandchild Space (Acceptance Test: nested_space - 020)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_grandchild", "is_private", "true"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_public_grandchild", "deletion_protection", "false"),
					resource.TestCheckResourceAttrPair(
						"lightdash_space.nested_space_public_grandchild",
						"parent_space_uuid",
						"lightdash_space.nested_space_public_root",
						"space_uuid",
					),

					// lightdash_space.nested_space_private_root
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_root", "name", "Private Root Space (Acceptance Test: nested_space - 020)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_root", "is_private", "false"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_root", "deletion_protection", "false"),
					// lightdash_space.nested_space_private_child
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_child", "name", "Private Child Space (Acceptance Test: nested_space - 020)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_child", "is_private", "false"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_child", "deletion_protection", "false"),
					// lightdash_space.nested_space_private_grandchild
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_grandchild", "name", "Private Grandchild Space (Acceptance Test: nested_space - 020)"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_grandchild", "is_private", "false"),
					resource.TestCheckResourceAttr("lightdash_space.nested_space_private_grandchild", "deletion_protection", "false"),
					resource.TestCheckResourceAttrPair(
						"lightdash_space.nested_space_private_grandchild",
						"parent_space_uuid",
						"lightdash_space.nested_space_private_root",
						"space_uuid",
					),
				),
			},
		},
	})
}

func TestAccSpaceResource_nestedRestricted(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_space - nested restricted")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of nested restricted spaces
	nestedRestrictedConfig030, err := ReadAccTestResource([]string{"resources", "lightdash_space", "nested_space", "030_nested_restricted_space.tf"})
	if err != nil {
		t.Fatalf("Failed to get nestedRestrictedConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + nestedRestrictedConfig030,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_space.nested_restricted_root
					resource.TestCheckResourceAttrSet("lightdash_space.nested_restricted_root", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.nested_restricted_root", "is_private", "true"),
					// lightdash_space.nested_restricted_child
					resource.TestCheckResourceAttrSet("lightdash_space.nested_restricted_child", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.nested_restricted_child", "is_private", "true"),
					resource.TestCheckResourceAttrPair(
						"lightdash_space.nested_restricted_child",
						"parent_space_uuid",
						"lightdash_space.nested_restricted_root",
						"space_uuid",
					),
					resource.TestCheckResourceAttr("lightdash_space.nested_restricted_child", "group_access.#", "1"),
				),
			},
		},
	})
}

func TestAccSpaceResource_nestedRestrictedComplex(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_space - nested restricted complex")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of nested restricted spaces complex scenarios
	nestedRestrictedConfig040, err := ReadAccTestResource([]string{"resources", "lightdash_space", "nested_space", "040_nested_restricted_complex.tf"})
	if err != nil {
		t.Fatalf("Failed to get nestedRestrictedConfig040: %v", err)
	}
	nestedRestrictedConfig050, err := ReadAccTestResource([]string{"resources", "lightdash_space", "nested_space", "050_nested_restricted_complex_update.tf"})
	if err != nil {
		t.Fatalf("Failed to get nestedRestrictedConfig050: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + nestedRestrictedConfig040,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_space.test_space", "name", "Test Space (Acceptance Test)"),
					resource.TestCheckResourceAttr("lightdash_space.test_space", "is_private", "true"),
					resource.TestCheckResourceAttrPair("lightdash_space.test_space", "parent_space_uuid", "lightdash_space.parent1", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.test_space", "group_access.#", "1"),
				),
			},
			{
				Config: providerConfig + nestedRestrictedConfig050,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_space.test_space", "name", "Test Space Updated (Acceptance Test)"),
					resource.TestCheckResourceAttr("lightdash_space.test_space", "is_private", "true"),
					resource.TestCheckResourceAttrPair("lightdash_space.test_space", "parent_space_uuid", "lightdash_space.parent2", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.test_space", "group_access.#", "0"),
				),
			},
		},
	})
}

func TestAccSpaceResource_access(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_space - access")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of space access
	spaceAccessConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_space", "space_access", "010_space_access.tf"})
	if err != nil {
		t.Fatalf("Failed to get spaceAccessConfig: %v", err)
	}
	spaceAccessConfig020, err := ReadAccTestResource([]string{"resources", "lightdash_space", "space_access", "020_space_access.tf"})
	if err != nil {
		t.Fatalf("Failed to get spaceAccessConfig: %v", err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + spaceAccessConfig010,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_space.space_access__test_space_1
					resource.TestCheckResourceAttr("lightdash_space.space_access__test_space_1", "name", "Space 1 (Acceptance Test: space_access)"),
					resource.TestCheckResourceAttr("lightdash_space.space_access__test_space_1", "is_private", "false"),
					resource.TestCheckResourceAttr("lightdash_space.space_access__test_space_1", "group_access.#", "0"),
					// lightdash_space.space_access__test_space_2
					resource.TestCheckResourceAttr("lightdash_space.space_access__test_space_2", "name", "Space 2 (Acceptance Test: space_access)"),
					resource.TestCheckResourceAttr("lightdash_space.space_access__test_space_2", "is_private", "true"),
					resource.TestCheckResourceAttr("lightdash_space.space_access__test_space_2", "group_access.#", "3"),
				),
			},
			{
				Config: providerConfig + spaceAccessConfig020,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_space.space_access__test_space_1
					resource.TestCheckResourceAttr("lightdash_space.space_access__test_space_1", "is_private", "true"),
					resource.TestCheckResourceAttr("lightdash_space.space_access__test_space_1", "group_access.#", "3"),
					// lightdash_space.space_access__test_space_2
					resource.TestCheckResourceAttr("lightdash_space.space_access__test_space_2", "is_private", "false"),
					resource.TestCheckResourceAttr("lightdash_space.space_access__test_space_2", "group_access.#", "0"),
				),
			},
		},
	})
}

func TestAccSpaceResource_import(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_space - import")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of space access
	importConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_space", "import", "010_import.tf"})
	if err != nil {
		t.Fatalf("Failed to get importConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the states
			{
				Config: providerConfig + importConfig010,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_space.import__public_root_space
					resource.TestCheckResourceAttr("lightdash_space.import__public_root_space", "name", "Public Root Space (Acceptance Test: import)"),
					resource.TestCheckResourceAttr("lightdash_space.import__public_root_space", "is_private", "false"),
					resource.TestCheckResourceAttr("lightdash_space.import__public_root_space", "deletion_protection", "false"),
					// lightdash_space.import__public_child_space
					resource.TestCheckResourceAttr("lightdash_space.import__public_child_space", "name", "Public Child Space (Acceptance Test: import)"),
					resource.TestCheckResourceAttr("lightdash_space.import__public_child_space", "deletion_protection", "false"),
					resource.TestCheckResourceAttrPair(
						"lightdash_space.import__public_child_space",
						"parent_space_uuid",
						"lightdash_space.import__public_root_space",
						"space_uuid",
					),
					// lightdash_space.import__private_root_space
					resource.TestCheckResourceAttr("lightdash_space.import__private_root_space", "name", "Private Root Space (Acceptance Test: import)"),
					resource.TestCheckResourceAttr("lightdash_space.import__private_root_space", "is_private", "true"),
					resource.TestCheckResourceAttr("lightdash_space.import__private_root_space", "deletion_protection", "false"),
					// lightdash_space.import__private_child_space
					resource.TestCheckResourceAttr("lightdash_space.import__private_child_space", "name", "Private Child Space (Acceptance Test: import)"),
					resource.TestCheckResourceAttr("lightdash_space.import__private_child_space", "deletion_protection", "false"),
					resource.TestCheckResourceAttrPair(
						"lightdash_space.import__private_child_space",
						"parent_space_uuid",
						"lightdash_space.import__private_root_space",
						"space_uuid",
					),
				),
			},
			// Remove the states
			// {
			// 	Config: providerConfig + importConfig020,
			// 	Check:  resource.ComposeTestCheckFunc(),
			// },
			// Import the states
			{
				Config:            providerConfig + importConfig010,
				ResourceName:      "lightdash_space.import__public_root_space",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"access",
					"created_at",
					"last_updated",
					// NOTE tentatively ignore the deletion_protection attribute
					"deletion_protection",
				},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					res, ok := state.RootModule().Resources["lightdash_space.import__public_root_space"]
					if !ok {
						return "", fmt.Errorf("resource not found in state for import")
					}
					// Get the project_uuid from the state
					project_uuid, ok := res.Primary.Attributes["project_uuid"]
					if !ok || project_uuid == "" {
						return "", fmt.Errorf("project_uuid attribute not present in state")
					}
					// Get the space_uuid from the state
					space_uuid, ok := res.Primary.Attributes["space_uuid"]
					if !ok || space_uuid == "" {
						return "", fmt.Errorf("space_uuid attribute not present in state")
					}
					// Construct the import ID in the form 'projects/<project_uuid>/spaces/<space_uuid>'
					id := fmt.Sprintf("projects/%s/spaces/%s", project_uuid, space_uuid)
					return id, nil
				},
			},
		},
	})
}
