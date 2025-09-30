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

func TestAccProjectAgentEvaluationsResource_create(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project_agent_evaluations")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of project agent evaluations creation with different configurations
	createConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_project_agent_evaluations", "create_evaluation", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config 010: %v", err)
	}
	createConfig020, err := ReadAccTestResource([]string{"resources", "lightdash_project_agent_evaluations", "create_evaluation", "020_update.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config 020: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// Skip CheckDestroy for create test as it can be unreliable with dependent resources
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + createConfig010,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "id"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "evaluation_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "created_at"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "updated_at"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "title", "Data Analysis Evaluation"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "description", "Evaluating the AI assistant's ability to analyze data and provide insights"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "prompts.#", "3"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "prompts.0.prompt", "Show me the top 5 customers by sales."),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "prompts.1.prompt", "What are the most popular products?"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "prompts.2.prompt", "Can you create a chart showing sales trends over time?"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "deletion_protection", "false"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + createConfig020,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "id"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "evaluation_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "created_at"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "updated_at"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "title", "Updated Data Analysis Evaluation"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "description", "Comprehensive evaluation of data analysis capabilities with additional test cases"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "prompts.#", "5"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "prompts.0.prompt", "Show me the top 5 customers by sales."),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.0.eval_prompt_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.0.type"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.0.created_at"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "prompts.1.prompt", "What are the most popular products?"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.1.eval_prompt_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.1.type"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.1.created_at"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "prompts.2.prompt", "Can you create a chart showing sales trends over time?"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.2.eval_prompt_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.2.type"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.2.created_at"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "prompts.3.prompt", "What are the sales trends over time?"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.3.eval_prompt_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.3.type"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.3.created_at"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "prompts.4.prompt", "Show me customer segmentation analysis."),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.4.eval_prompt_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.4.type"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test", "prompts.4.created_at"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test", "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccProjectAgentEvaluationsResource_import(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project_agent_evaluations")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of project agent evaluations import
	importConfig, err := ReadAccTestResource([]string{"resources", "lightdash_project_agent_evaluations", "import", "010_import.tf"})
	if err != nil {
		t.Fatalf("Failed to get import config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// Skip CheckDestroy for import test as it can be unreliable with dependent resources
		Steps: []resource.TestStep{
			// Create testing
			{
				Config: providerConfig + importConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test_evaluation", "id"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test_evaluation", "evaluation_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test_evaluation", "created_at"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test_evaluation", "updated_at"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test_evaluation", "title", "Test Evaluation for Import"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test_evaluation", "description", "This evaluation is created for import testing"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test_evaluation", "prompts.#", "2"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test_evaluation", "prompts.0.prompt", "Show me the top 5 customers by sales."),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test_evaluation", "prompts.0.eval_prompt_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test_evaluation", "prompts.0.type"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test_evaluation", "prompts.0.created_at"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test_evaluation", "prompts.1.prompt", "What are the most popular products?"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test_evaluation", "prompts.1.eval_prompt_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test_evaluation", "prompts.1.type"),
					resource.TestCheckResourceAttrSet("lightdash_project_agent_evaluations.test_evaluation", "prompts.1.created_at"),
					resource.TestCheckResourceAttr("lightdash_project_agent_evaluations.test_evaluation", "deletion_protection", "false"),
				),
			},
			// Import testing
			{
				ResourceName:      "lightdash_project_agent_evaluations.test_evaluation",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
