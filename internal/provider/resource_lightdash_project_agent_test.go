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

func TestAccProjectAgentResource_create(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project_agent")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of project agent creation with different configurations
	createConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_project_agent", "create_agent", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config 010: %v", err)
	}
	createConfig020, err := ReadAccTestResource([]string{"resources", "lightdash_project_agent", "create_agent", "020_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config 030: %v", err)
	}
	createConfig030, err := ReadAccTestResource([]string{"resources", "lightdash_project_agent", "create_agent", "030_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config 030: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: Basic agent creation (010_create.tf)
				// No tags, group_access, or user_access specified (should default to empty lists)
				Config: providerConfig + createConfig010,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "organization_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "project_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "agent_uuid"),
					resource.TestCheckOutput("is_agent_uuid_set", "true"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "name", "Test Agent"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "instruction", "You are a helpful AI assistant for data analysis."),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "enable_data_access", "false"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "deletion_protection", "true"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "tags.#", "0"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "group_access.#", "0"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "user_access.#", "0"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "updated_at"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "created_at"),
					// Check that agent references match data sources
					resource.TestCheckResourceAttrPair(
						"lightdash_project_agent.test",
						"organization_uuid",
						"data.lightdash_organization.test",
						"organization_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"lightdash_project_agent.test",
						"project_uuid",
						"data.lightdash_project.test",
						"project_uuid",
					),
					// Check data source attributes
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "id"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "organization_uuid"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "name"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "instruction"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "updated_at"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "created_at"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "enable_data_access", "false"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "tags.#", "0"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "group_access.#", "0"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "user_access.#", "0"),
					// Check that data source matches resource
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"organization_uuid",
						"lightdash_project_agent.test",
						"organization_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"project_uuid",
						"lightdash_project_agent.test",
						"project_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"agent_uuid",
						"lightdash_project_agent.test",
						"agent_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"name",
						"lightdash_project_agent.test",
						"name",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"instruction",
						"lightdash_project_agent.test",
						"instruction",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"enable_data_access",
						"lightdash_project_agent.test",
						"enable_data_access",
					),
				),
			},
			{
				// Step 2: Agent update with tags and explicit empty lists (020_create.tf)
				Config: providerConfig + createConfig020,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "organization_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "project_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "agent_uuid"),
					resource.TestCheckOutput("is_agent_uuid_set", "true"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "name", "Test Agent Updated"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "instruction", "You are an updated helpful AI assistant for data analysis and insights."),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "tags.0", "terraform"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "tags.1", "updated"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "enable_data_access", "true"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "group_access.#", "0"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "user_access.#", "0"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "deletion_protection", "false"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "updated_at"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "created_at"),
					// Check that agent references match data sources
					resource.TestCheckResourceAttrPair(
						"lightdash_project_agent.test",
						"organization_uuid",
						"data.lightdash_organization.test",
						"organization_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"lightdash_project_agent.test",
						"project_uuid",
						"data.lightdash_project.test",
						"project_uuid",
					),
					// Check data source attributes for updated agent
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "id"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "organization_uuid"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "name"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "instruction"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "updated_at"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "created_at"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "name", "Test Agent Updated"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "instruction", "You are an updated helpful AI assistant for data analysis and insights."),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "tags.0", "terraform"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "tags.1", "updated"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "enable_data_access", "true"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "group_access.#", "0"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "user_access.#", "0"),
					// Check that data source matches resource
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"organization_uuid",
						"lightdash_project_agent.test",
						"organization_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"project_uuid",
						"lightdash_project_agent.test",
						"project_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"agent_uuid",
						"lightdash_project_agent.test",
						"agent_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"name",
						"lightdash_project_agent.test",
						"name",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"instruction",
						"lightdash_project_agent.test",
						"instruction",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"tags.#",
						"lightdash_project_agent.test",
						"tags.#",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"enable_data_access",
						"lightdash_project_agent.test",
						"enable_data_access",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"group_access.#",
						"lightdash_project_agent.test",
						"group_access.#",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"user_access.#",
						"lightdash_project_agent.test",
						"user_access.#",
					),
				),
			},
			{
				// Step 3: Agent update without tags (030_create.tf)
				// Tags omitted, should default back to empty list
				Config: providerConfig + createConfig030,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "organization_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "project_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "agent_uuid"),
					resource.TestCheckOutput("is_agent_uuid_set", "true"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "name", "Test Agent Updated"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "instruction", "You are an updated helpful AI assistant for data analysis and insights."),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "enable_data_access", "true"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "tags.#", "0"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "group_access.#", "0"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test", "user_access.#", "0"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "updated_at"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test", "created_at"),
					// Check that agent references match data sources
					resource.TestCheckResourceAttrPair(
						"lightdash_project_agent.test",
						"organization_uuid",
						"data.lightdash_organization.test",
						"organization_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"lightdash_project_agent.test",
						"project_uuid",
						"data.lightdash_project.test",
						"project_uuid",
					),
					// Check data source attributes
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "id"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "organization_uuid"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "name"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "instruction"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "updated_at"),
					resource.TestCheckResourceAttrSet("data.lightdash_project_agent.test", "created_at"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "name", "Test Agent Updated"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "instruction", "You are an updated helpful AI assistant for data analysis and insights."),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "enable_data_access", "true"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "tags.#", "0"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "group_access.#", "0"),
					resource.TestCheckResourceAttr("data.lightdash_project_agent.test", "user_access.#", "0"),
					// Check that data source matches resource
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"organization_uuid",
						"lightdash_project_agent.test",
						"organization_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"project_uuid",
						"lightdash_project_agent.test",
						"project_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"agent_uuid",
						"lightdash_project_agent.test",
						"agent_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"name",
						"lightdash_project_agent.test",
						"name",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"instruction",
						"lightdash_project_agent.test",
						"instruction",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"enable_data_access",
						"lightdash_project_agent.test",
						"enable_data_access",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"tags.#",
						"lightdash_project_agent.test",
						"tags.#",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"group_access.#",
						"lightdash_project_agent.test",
						"group_access.#",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_agent.test",
						"user_access.#",
						"lightdash_project_agent.test",
						"user_access.#",
					),
				),
			},
		},
	})
}

func TestAccProjectAgentResource_import(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project_agent")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of project agent import
	importConfig, err := ReadAccTestResource([]string{"resources", "lightdash_project_agent", "import", "010_import.tf"})
	if err != nil {
		t.Fatalf("Failed to get import config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// First create an agent
				Config: providerConfig + importConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test_agent", "organization_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test_agent", "project_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent.test_agent", "agent_uuid"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test_agent", "name", "Test Agent for Import"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test_agent", "instruction", "You are a helpful AI assistant for data analysis and imports."),
					resource.TestCheckResourceAttr("lightdash_project_agent.test_agent", "tags.#", "2"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test_agent", "tags.0", "import"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test_agent", "tags.1", "test"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test_agent", "enable_data_access", "true"),
					resource.TestCheckResourceAttr("lightdash_project_agent.test_agent", "deletion_protection", "false"),
				),
			},
			{
				// Then test importing the agent
				Config:                  providerConfig + importConfig,
				ResourceName:            "lightdash_project_agent.test_agent",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					res, ok := state.RootModule().Resources["lightdash_project_agent.test_agent"]
					if !ok {
						return "", fmt.Errorf("resource not found in state for import")
					}
					// Get the required attributes from the state
					organization_uuid, ok := res.Primary.Attributes["organization_uuid"]
					if !ok || organization_uuid == "" {
						return "", fmt.Errorf("organization_uuid attribute not present in state")
					}
					project_uuid, ok := res.Primary.Attributes["project_uuid"]
					if !ok || project_uuid == "" {
						return "", fmt.Errorf("project_uuid attribute not present in state")
					}
					agent_uuid, ok := res.Primary.Attributes["agent_uuid"]
					if !ok || agent_uuid == "" {
						return "", fmt.Errorf("agent_uuid attribute not present in state")
					}
					// Construct the import ID in the form 'organizations/<organization_uuid>/projects/<project_uuid>/agents/<agent_uuid>'
					id := getProjectAgentResourceId(organization_uuid, project_uuid, agent_uuid)
					return id, nil
				},
			},
		},
	})
}
