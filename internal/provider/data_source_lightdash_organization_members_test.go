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

func TestAccOrganizationMembersDataSource(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for data_source_lightdash_organization_members")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of simple organization members retrieval
	config, err := ReadAccTestResource([]string{"data_source_lightdash_organization_members", "data", "01_data.tf"})
	if err != nil {
		t.Fatalf("Failed to get organizationMembersConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + config,
				Check: resource.ComposeTestCheckFunc(
					// lightdash_project_role_group.resource_lightdash_project_role_group__grant
					resource.TestCheckResourceAttrSet("data.lightdash_organization_members.test", "organization_uuid"),
					// TODO: Add check for members
					//       We need to figure out the expression to specify wildcard for members in the terraform acceptance test
					resource.TestCheckResourceAttrSet("data.lightdash_organization_members.test", "members.0.user_uuid"),
					resource.TestCheckResourceAttrSet("data.lightdash_organization_members.test", "members.0.email"),
					resource.TestCheckResourceAttrSet("data.lightdash_organization_members.test", "members.0.role"),
				),
			},
		},
	})
}
