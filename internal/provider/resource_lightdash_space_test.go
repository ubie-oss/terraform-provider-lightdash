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

func TestAccSpaceResource(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_space")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of simple space creation
	simpleSpaceConfig010, err := ReadAccTestResource([]string{"resource_lightdash_space", "create_space", "010_create_space.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	simpleSpaceConfig020, err := ReadAccTestResource([]string{"resource_lightdash_space", "create_space", "020_create_space.tf"})
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
					// lightdash_space.create_space__test_private
					resource.TestCheckResourceAttrSet("lightdash_space.create_space__test_private", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_private", "name", "Private Space (Acceptance Test: create_space)"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_private", "is_private", "true"),
					resource.TestCheckResourceAttr("lightdash_space.create_space__test_private", "deletion_protection", "true"),
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

	// Test of nested spaces
	nestedSpaceConfig010, err := ReadAccTestResource([]string{"resource_lightdash_space", "nested_space", "010_nested_space.tf"})
	if err != nil {
		t.Fatalf("Failed to get nestedSpaceConfig: %v", err)
	}
	nestedSpaceConfig020, err := ReadAccTestResource([]string{"resource_lightdash_space", "nested_space", "020_nested_space.tf"})
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

	// Test of space access
	spaceAccessConfig010, err := ReadAccTestResource([]string{"resource_lightdash_space", "space_access", "010_space_access.tf"})
	if err != nil {
		t.Fatalf("Failed to get spaceAccessConfig: %v", err)
	}
	spaceAccessConfig020, err := ReadAccTestResource([]string{"resource_lightdash_space", "space_access", "020_space_access.tf"})
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
