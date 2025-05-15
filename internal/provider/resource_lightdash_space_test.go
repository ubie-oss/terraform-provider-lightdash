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
	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of simple space creation
	publicSpaceConfig, err := ReadAccTestResource([]string{"resource_lightdash_space", "create_space", "010_create_space.tf"})
	if err != nil {
		t.Fatalf("Failed to get publicSpaceConfig: %v", err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + publicSpaceConfig,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_space.test_public
					resource.TestCheckResourceAttrSet("lightdash_space.test_public", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.test_public", "name", "Public Space (Acceptance Test: create_space)"),
					resource.TestCheckResourceAttr("lightdash_space.test_public", "is_private", "false"),
					resource.TestCheckResourceAttr("lightdash_space.test_public", "deletion_protection", "false"),
					// lightdash_space.test_private
					resource.TestCheckResourceAttrSet("lightdash_space.test_private", "space_uuid"),
					resource.TestCheckResourceAttr("lightdash_space.test_private", "name", "Private Space (Acceptance Test: create_space)"),
					resource.TestCheckResourceAttr("lightdash_space.test_private", "is_private", "true"),
					resource.TestCheckResourceAttr("lightdash_space.test_private", "deletion_protection", "false"),
				),
			},
			// Add more test steps for update, delete, import later if needed
		},
	})
}
