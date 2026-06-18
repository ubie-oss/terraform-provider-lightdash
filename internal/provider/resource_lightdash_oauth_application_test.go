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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestExtractOAuthApplicationResourceID(t *testing.T) {
	t.Parallel()

	id := "organizations/org-uuid/oauth_applications/oauth-abc123"
	got, err := extractOAuthApplicationResourceID(id)
	if err != nil {
		t.Fatalf("extractOAuthApplicationResourceID: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(got))
	}
	if got[0] != "org-uuid" {
		t.Errorf("organization UUID: got %q", got[0])
	}
	if got[1] != "oauth-abc123" {
		t.Errorf("client ID: got %q", got[1])
	}
}

func TestGetOAuthApplicationResourceID(t *testing.T) {
	t.Parallel()

	got := getOAuthApplicationResourceID("org-uuid", "oauth-abc123")
	want := "organizations/org-uuid/oauth_applications/oauth-abc123"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractOAuthApplicationResourceID_invalid(t *testing.T) {
	t.Parallel()

	_, err := extractOAuthApplicationResourceID("invalid/id")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

// Requires org-admin LIGHTDASH_API_KEY.
func TestAccOAuthApplicationResource_lifecycle(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_oauth_application")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	createConfig, err := ReadAccTestResource([]string{"resources", "lightdash_oauth_application", "lifecycle", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config: %v", err)
	}
	updateConfig, err := ReadAccTestResource([]string{"resources", "lightdash_oauth_application", "lifecycle", "020_update.tf"})
	if err != nil {
		t.Fatalf("Failed to get update config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_oauth_application.test", "client_name", "test (Acceptance Test - oauth lifecycle)"),
					resource.TestCheckResourceAttrSet("lightdash_oauth_application.test", "client_id"),
					resource.TestCheckResourceAttrSet("lightdash_oauth_application.test", "client_secret"),
					resource.TestCheckResourceAttr("lightdash_oauth_application.test", "redirect_uris.#", "1"),
				),
			},
			{
				Config: providerConfig + updateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_oauth_application.test", "client_name", "test (Acceptance Test - oauth lifecycle updated)"),
					resource.TestCheckResourceAttr("lightdash_oauth_application.test", "redirect_uris.#", "2"),
				),
			},
		},
	})
}

// Requires org-admin LIGHTDASH_API_KEY.
func TestAccOAuthApplicationResource_import(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_oauth_application")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	importConfig, err := ReadAccTestResource([]string{"resources", "lightdash_oauth_application", "import", "010_import.tf"})
	if err != nil {
		t.Fatalf("Failed to get import config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + importConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_oauth_application.test", "client_name", "test (Acceptance Test - oauth import)"),
					resource.TestCheckResourceAttrSet("lightdash_oauth_application.test", "client_id"),
				),
			},
			{
				Config:            providerConfig + importConfig,
				ResourceName:      "lightdash_oauth_application.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"client_secret",
					"deletion_protection",
					"created_by_user_uuid",
				},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					res, ok := state.RootModule().Resources["lightdash_oauth_application.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state for import")
					}
					organizationUUID, ok := res.Primary.Attributes["organization_uuid"]
					if !ok || organizationUUID == "" {
						return "", fmt.Errorf("organization_uuid attribute not present in state")
					}
					clientID, ok := res.Primary.Attributes["client_id"]
					if !ok || clientID == "" {
						return "", fmt.Errorf("client_id attribute not present in state")
					}
					return getOAuthApplicationResourceID(organizationUUID, clientID), nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_oauth_application.test", "deletion_protection", "true"),
				),
			},
		},
	})
}

// Requires org-admin LIGHTDASH_API_KEY.
func TestAccOAuthApplicationResource_deletionProtection(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_oauth_application")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	createConfig, err := ReadAccTestResource([]string{"resources", "lightdash_oauth_application", "deletion_protection", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config: %v", err)
	}
	allowDestroyConfig, err := ReadAccTestResource([]string{"resources", "lightdash_oauth_application", "deletion_protection", "020_allow_destroy.tf"})
	if err != nil {
		t.Fatalf("Failed to get allow destroy config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_oauth_application.test", "deletion_protection", "true"),
					resource.TestCheckResourceAttrSet("lightdash_oauth_application.test", "client_id"),
				),
			},
			{
				Config:      providerConfig + createConfig,
				Destroy:     true,
				ExpectError: regexp.MustCompile("Deletion Protection Enabled"),
			},
			{
				Config: providerConfig + allowDestroyConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_oauth_application.test", "deletion_protection", "false"),
				),
			},
		},
	})
}
