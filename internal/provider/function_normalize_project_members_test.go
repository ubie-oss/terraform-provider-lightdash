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
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestNormalizeProjectMembersRun_simple(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project_role_group")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Get the function config
	functionConfig010, err := ReadAccTestResource([]string{"function_normalize_project_members", "simple", "010_simple.tf"})
	if err != nil {
		t.Fatalf("Failed to get functionConfig: %v", err)
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + functionConfig010,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("normalized_project_members_admins_json", "[\"admin1\"]"),
					resource.TestCheckOutput("normalized_project_members_developers_json", "[\"dev1\"]"),
					resource.TestCheckOutput("normalized_project_members_editors_json", "[\"editor1\"]"),
					resource.TestCheckOutput("normalized_project_members_interactive_viewers_json", "[\"interactive_viewer1\"]"),
					resource.TestCheckOutput("normalized_project_members_viewers_json", "[\"viewer1\"]"),
				),
			},
		},
	})
}

func TestNormalizeProjectMembersRun_complicated(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project_role_group")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Get the function config
	functionConfig010, err := ReadAccTestResource([]string{"function_normalize_project_members", "complicated", "010_complicated.tf"})
	if err != nil {
		t.Fatalf("Failed to get functionConfig: %v", err)
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + functionConfig010,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("normalized_project_members_admins_json", "[\"admin1\",\"admin2\"]"),
					resource.TestCheckOutput("normalized_project_members_developers_json", "[\"dev1\",\"dev2\"]"),
					resource.TestCheckOutput("normalized_project_members_editors_json", "[\"editor1\",\"editor2\"]"),
					resource.TestCheckOutput("normalized_project_members_interactive_viewers_json", "[\"interactive_viewer1\",\"interactive_viewer2\"]"),
					resource.TestCheckOutput("normalized_project_members_viewers_json", "[\"viewer1\",\"viewer2\"]"),
				),
			},
		},
	})
}
