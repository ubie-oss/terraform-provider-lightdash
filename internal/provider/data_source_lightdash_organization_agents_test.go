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

func TestAccOrganizationAgentsDataSource(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for data_source_lightdash_organization_agents")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of simple organization agents retrieval
	config, err := ReadAccTestResource([]string{"data_sources", "lightdash_organization_agents", "010_data.tf"})
	if err != nil {
		t.Fatalf("Failed to get organizationAgentsConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + config,

				Check: resource.ComposeTestCheckFunc(
					// Check that organization UUID is set
					resource.TestCheckResourceAttrSet("data.lightdash_organization_agents.test", "organization_uuid"),
					// Check that we have at least one agent
					resource.TestCheckResourceAttrSet("data.lightdash_organization_agents.test", "agents.0.agent_uuid"),
					resource.TestCheckResourceAttrSet("data.lightdash_organization_agents.test", "agents.0.organization_uuid"),
					resource.TestCheckResourceAttrSet("data.lightdash_organization_agents.test", "agents.0.project_uuid"),
					resource.TestCheckResourceAttrSet("data.lightdash_organization_agents.test", "agents.0.name"),
					resource.TestCheckResourceAttrSet("data.lightdash_organization_agents.test", "agents.0.updated_at"),
					resource.TestCheckResourceAttrSet("data.lightdash_organization_agents.test", "agents.0.created_at"),
					// enable_data_access should be boolean and set
					resource.TestCheckResourceAttrSet("data.lightdash_organization_agents.test", "agents.0.enable_data_access"),
				),
			},
		},
	})
}
