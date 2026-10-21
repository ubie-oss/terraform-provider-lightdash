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

func TestGetProjectAiProviderCredentialResourceID(t *testing.T) {
	t.Parallel()

	got := getProjectAiProviderCredentialResourceID("org", "proj")
	want := "organizations/org/projects/proj/ai_provider_credential"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Requires org-admin LIGHTDASH_API_KEY. Uses a dummy api_key in acc config (not validated against Bedrock).
func TestAccProjectAiProviderCredentialResource_lifecycle(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project_ai_provider_credential")
	}
	registerAccAiCredentialPurge(t)

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	createConfig, err := ReadAccTestResource([]string{"resources", "lightdash_project_ai_provider_credential", "lifecycle", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAiProviderCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("lightdash_project_ai_provider_credential.test", "credential_uuid"),
				),
			},
			{
				Config:            providerConfig + createConfig,
				ResourceName:      "lightdash_project_ai_provider_credential.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deletion_protection",
				},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					res, ok := state.RootModule().Resources["lightdash_project_ai_provider_credential.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state for import")
					}
					org, ok := res.Primary.Attributes["organization_uuid"]
					if !ok || org == "" {
						return "", fmt.Errorf("organization_uuid missing")
					}
					proj, ok := res.Primary.Attributes["project_uuid"]
					if !ok || proj == "" {
						return "", fmt.Errorf("project_uuid missing")
					}
					return getProjectAiProviderCredentialResourceID(org, proj), nil
				},
			},
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_project_ai_provider_credential.test", "deletion_protection", "false"),
				),
			},
		},
	})
}
