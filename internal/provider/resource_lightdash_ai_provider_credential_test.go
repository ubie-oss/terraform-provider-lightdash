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
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestExtractProjectAiProviderCredentialResourceID(t *testing.T) {
	t.Parallel()

	id := "organizations/org-uuid/projects/project-uuid/ai_provider_credential"
	got, err := extractProjectAiProviderCredentialResourceID(id)
	if err != nil {
		t.Fatalf("extractProjectAiProviderCredentialResourceID: %v", err)
	}
	if got[0] != "org-uuid" || got[1] != "project-uuid" {
		t.Fatalf("unexpected parts: %v", got)
	}
}

func TestValidateAiProviderCredentialProvider(t *testing.T) {
	t.Parallel()

	if diags := validateAiProviderCredentialProvider("bedrock"); diags.HasError() {
		t.Fatalf("bedrock should be valid: %v", diags)
	}
	if diags := validateAiProviderCredentialProvider("openai"); !diags.HasError() {
		t.Fatal("openai should be invalid")
	}
}

func TestValidateInferenceGeography(t *testing.T) {
	t.Parallel()

	if diags := validateInferenceGeography(types.StringValue("us")); diags.HasError() {
		t.Fatalf("us should be valid: %v", diags)
	}
	if diags := validateInferenceGeography(types.StringValue("invalid")); !diags.HasError() {
		t.Fatal("invalid geography should fail")
	}
}

func TestAiProviderCredentialDeleteErrorMessage(t *testing.T) {
	t.Parallel()

	summary, detail := aiProviderCredentialDeleteErrorMessage("uuid-1", fmt.Errorf("cannot delete default credential"))
	if summary != "Cannot Delete Default AI Provider Credential" {
		t.Fatalf("unexpected summary: %q", summary)
	}
	if !strings.Contains(detail, "uuid-1") {
		t.Fatalf("detail should mention credential uuid: %q", detail)
	}

	summary, detail = aiProviderCredentialDeleteErrorMessage("uuid-2", fmt.Errorf("credential in use by project"))
	if summary != "Cannot Delete AI Provider Credential In Use" {
		t.Fatalf("unexpected summary: %q", summary)
	}
	if detail == "" {
		t.Fatal("expected detail")
	}

	summary, _ = aiProviderCredentialDeleteErrorMessage("uuid-3", fmt.Errorf("unexpected status code: 500"))
	if summary != "Error deleting AI provider credential" {
		t.Fatalf("unexpected summary: %q", summary)
	}
}

// Requires org-admin LIGHTDASH_API_KEY. Uses a dummy api_key in acc config (not validated against Bedrock).
func TestAccAiProviderCredentialResource_lifecycle(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_ai_provider_credential")
	}
	registerAccAiCredentialPurge(t)

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	createConfig, err := ReadAccTestResource([]string{"resources", "lightdash_ai_provider_credential", "lifecycle", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config: %v", err)
	}
	updateConfig, err := ReadAccTestResource([]string{"resources", "lightdash_ai_provider_credential", "lifecycle", "020_update.tf"})
	if err != nil {
		t.Fatalf("Failed to get update config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAiProviderCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_ai_provider_credential.test", "ai_provider", "bedrock"),
					resource.TestCheckResourceAttrSet("lightdash_ai_provider_credential.test", "credential_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_ai_provider_credential.test", "api_key_hint"),
				),
			},
			{
				Config: providerConfig + updateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_ai_provider_credential.test", "label", "test (Acceptance Test - ai credential updated)"),
				),
			},
		},
	})
}

func TestAccAiProviderCredentialResource_import(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_ai_provider_credential")
	}
	registerAccAiCredentialPurge(t)

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	importConfig, err := ReadAccTestResource([]string{"resources", "lightdash_ai_provider_credential", "import", "010_import.tf"})
	if err != nil {
		t.Fatalf("Failed to get import config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAiProviderCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + importConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("lightdash_ai_provider_credential.test", "credential_uuid"),
				),
			},
			{
				Config:            providerConfig + importConfig,
				ResourceName:      "lightdash_ai_provider_credential.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"api_key",
					"deletion_protection",
				},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					res, ok := state.RootModule().Resources["lightdash_ai_provider_credential.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state for import")
					}
					uuid, ok := res.Primary.Attributes["credential_uuid"]
					if !ok || uuid == "" {
						return "", fmt.Errorf("credential_uuid attribute not present in state")
					}
					return uuid, nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_ai_provider_credential.test", "deletion_protection", "true"),
				),
			},
			{
				Config: providerConfig + importConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_ai_provider_credential.test", "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccAiProviderCredentialResource_deletionProtection(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_ai_provider_credential")
	}
	registerAccAiCredentialPurge(t)

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	createConfig, err := ReadAccTestResource([]string{"resources", "lightdash_ai_provider_credential", "deletion_protection", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config: %v", err)
	}
	allowDestroyConfig, err := ReadAccTestResource([]string{"resources", "lightdash_ai_provider_credential", "deletion_protection", "020_allow_destroy.tf"})
	if err != nil {
		t.Fatalf("Failed to get allow destroy config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAiProviderCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_ai_provider_credential.test", "deletion_protection", "true"),
				),
			},
			{
				Config:      providerConfig + createConfig,
				Destroy:     true,
				ExpectError: regexp.MustCompile("Deletion Protection Enabled"),
			},
			{
				Config: providerConfig + allowDestroyConfig,
			},
		},
	})
}
