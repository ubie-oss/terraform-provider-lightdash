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

func TestAccOrganizationGroupsDataSource(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for data_source_lightdash_organization_groups")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of simple organization members retrieval
	config, err := ReadAccTestResource([]string{"data_source_lightdash_organization_groups", "data", "010_data.tf"})
	if err != nil {
		t.Fatalf("Failed to get organizationGroupsConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + config,

				Check: resource.ComposeTestCheckFunc(
					// lightdash_project_role_group.resource_lightdash_project_role_group__grant
					resource.TestCheckResourceAttrSet("data.lightdash_organization_groups.test", "organization_uuid"),
					resource.TestCheckResourceAttrSet("data.lightdash_organization_groups.test", "groups.0.group_uuid"),
					resource.TestCheckResourceAttrSet("data.lightdash_organization_groups.test", "groups.0.name"),
					resource.TestCheckResourceAttrSet("data.lightdash_organization_groups.test", "groups.0.created_at"),
				),
			},
		},
	})
}
