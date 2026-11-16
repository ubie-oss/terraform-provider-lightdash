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
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"
)

func TestProjectReplacementBlocked(t *testing.T) {
	t.Parallel()

	protected := projectResourceModel{
		Name:             types.StringValue("old"),
		ProjectType:      types.StringValue("DEFAULT"),
		DbtVersion:       types.StringValue("latest"),
		DeleteProtection: types.BoolValue(true),
	}
	if projectReplacementBlocked(protected, protected) {
		t.Fatal("same values should not block")
	}

	renamed := protected
	renamed.Name = types.StringValue("new")
	if !projectReplacementBlocked(protected, renamed) {
		t.Fatal("rename while protected should block")
	}

	unprotected := protected
	unprotected.DeleteProtection = types.BoolValue(false)
	renamed.DeleteProtection = types.BoolValue(false)
	if projectReplacementBlocked(unprotected, renamed) {
		t.Fatal("rename while unprotected should not block")
	}
}

func TestDeleteDecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		protected bool
		wantBlock bool
		wantTitle string
	}{
		{
			name:      "protected",
			protected: true,
			wantBlock: true,
			wantTitle: "Deletion Protection Enabled",
		},
		{
			name:      "unprotected",
			protected: false,
			wantBlock: false,
			wantTitle: "Lightdash project was not deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := deleteDecision(tt.protected)
			if got.Block != tt.wantBlock {
				t.Fatalf("Block = %v, want %v", got.Block, tt.wantBlock)
			}
			if got.Summary != tt.wantTitle {
				t.Fatalf("Summary = %q, want %q", got.Summary, tt.wantTitle)
			}
			if got.Detail == "" {
				t.Fatal("Detail is empty")
			}
		})
	}
}

const (
	lightdashTestSandboxURLEnvVar      = "LIGHTDASH_TEST_SANDBOX_URL"
	lightdashTestAllowProductionEnvVar = "LIGHTDASH_TEST_ALLOW_PRODUCTION"
)

// projectAcceptanceAllowed reports whether these acceptance tests may run against the configured host.
// They create a real project that destroy does not delete.
func projectAcceptanceAllowed(url, sandboxURL, allowProduction string) bool {
	if strings.TrimSpace(allowProduction) == "1" {
		return true
	}
	url = strings.TrimSpace(url)
	sandboxURL = strings.TrimSpace(sandboxURL)
	return url != "" && sandboxURL != "" && url == sandboxURL
}

func TestProjectAcceptanceAllowed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		url             string
		sandboxURL      string
		allowProduction string
		want            bool
	}{
		{name: "sandbox match", url: "https://sandbox.example", sandboxURL: "https://sandbox.example", want: true},
		{name: "sandbox match with space", url: " https://sandbox.example ", sandboxURL: "https://sandbox.example", want: true},
		{name: "different host", url: "https://prod.example", sandboxURL: "https://sandbox.example", want: false},
		{name: "sandbox unset", url: "https://prod.example", want: false},
		{name: "both empty", want: false},
		{name: "production opt-in", url: "https://prod.example", allowProduction: "1", want: true},
		{name: "production flag not exactly 1", url: "https://prod.example", allowProduction: "true", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := projectAcceptanceAllowed(tt.url, tt.sandboxURL, tt.allowProduction); got != tt.want {
				t.Fatalf("projectAcceptanceAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func skipUnlessProjectAcceptance(t *testing.T) {
	t.Helper()
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for lightdash_project")
	}
	if projectAcceptanceAllowed(
		os.Getenv(lightdashUrlEnvVar),
		os.Getenv(lightdashTestSandboxURLEnvVar),
		os.Getenv(lightdashTestAllowProductionEnvVar),
	) {
		return
	}
	t.Skipf(
		"Skipping lightdash_project acceptance test: set %s to the same value as %s, or set %s=1 to allow the configured instance",
		lightdashTestSandboxURLEnvVar,
		lightdashUrlEnvVar,
		lightdashTestAllowProductionEnvVar,
	)
}

func TestAccProjectResource_lifecycle(t *testing.T) {
	skipUnlessProjectAcceptance(t)

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}
	createConfig, err := ReadAccTestResource([]string{"resources", "lightdash_project", "lifecycle", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_project.test", "name", "tf-acc-lightdash-project"),
					resource.TestCheckResourceAttr("lightdash_project.test", "type", "DEFAULT"),
					resource.TestCheckResourceAttr("lightdash_project.test", "dbt_version", "latest"),
					resource.TestCheckResourceAttr("lightdash_project.test", "deletion_protection", "false"),
					resource.TestCheckResourceAttrSet("lightdash_project.test", "project_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_project.test", "slug"),
				),
			},
			{
				Config:            providerConfig + createConfig,
				ResourceName:      "lightdash_project.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deletion_protection",
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_project.test", "deletion_protection", "true"),
				),
			},
			// Import sets deletion_protection to true. Apply the config again so test cleanup
			// can remove state. The Lightdash project is left behind either way.
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_project.test", "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccProjectResource_deletionProtection(t *testing.T) {
	skipUnlessProjectAcceptance(t)

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}
	createConfig, err := ReadAccTestResource([]string{"resources", "lightdash_project", "deletion_protection", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get create config: %v", err)
	}
	allowDestroyConfig, err := ReadAccTestResource([]string{"resources", "lightdash_project", "deletion_protection", "020_allow_destroy.tf"})
	if err != nil {
		t.Fatalf("Failed to get allow destroy config: %v", err)
	}

	var projectUUID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_project.test", "deletion_protection", "true"),
					func(state *terraform.State) error {
						rs, ok := state.RootModule().Resources["lightdash_project.test"]
						if !ok {
							return fmt.Errorf("lightdash_project.test not in state")
						}
						projectUUID = rs.Primary.Attributes["project_uuid"]
						if projectUUID == "" {
							return fmt.Errorf("project_uuid is empty")
						}
						return nil
					},
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
					resource.TestCheckResourceAttr("lightdash_project.test", "deletion_protection", "false"),
				),
			},
			{
				Config:  providerConfig + allowDestroyConfig,
				Destroy: true,
				Check: func(state *terraform.State) error {
					if _, ok := state.RootModule().Resources["lightdash_project.test"]; ok {
						return fmt.Errorf("lightdash_project.test is still in state")
					}
					return projectStillExists(projectUUID)
				},
			},
		},
	})
}

// projectStillExists confirms destroy removed state only. The project must remain in Lightdash.
func projectStillExists(projectUUID string) error {
	host, err := getLightdashUrl()
	if err != nil {
		return err
	}
	token, err := getLightdashApiKey()
	if err != nil {
		return err
	}
	client, err := api.NewClient(host, token, nil)
	if err != nil {
		return fmt.Errorf("create API client: %w", err)
	}
	project, err := apiv1.GetProjectV1(client, projectUUID)
	if err != nil {
		return fmt.Errorf("expected project %s to remain in Lightdash after destroy: %w", projectUUID, err)
	}
	if project.ProjectUUID != projectUUID {
		return fmt.Errorf("read project %s, want %s", project.ProjectUUID, projectUUID)
	}
	return nil
}
