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
)

func TestAccProvider_concurrencyConfig(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for provider concurrency configuration")
	}

	// Get the Lightdash URL and API key
	lightdashUrl, err := getLightdashUrl()
	if err != nil {
		t.Fatalf("Failed to get Lightdash URL: %v", err)
	}
	lightdashApiKey, err := getLightdashApiKey()
	if err != nil {
		t.Fatalf("Failed to get Lightdash API key: %v", err)
	}
	lightdashProjectUuid, err := getLightdashProjectUuid()
	if err != nil {
		t.Fatalf("Failed to get Lightdash project UUID: %v", err)
	}

	// Create provider config with max_concurrent_requests set to 5
	providerConfig := fmt.Sprintf(`
provider "lightdash" {
	host                  = "%s"
	token                 = "%s"
	max_concurrent_requests = 5
}

data "lightdash_project" "test" {
	project_uuid = "%s"
}
`, *lightdashUrl, *lightdashApiKey, *lightdashProjectUuid)

	// Simple test config that uses the organization data source
	testConfig := `
data "lightdash_organization" "test" {
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testConfig,
				Check: resource.ComposeTestCheckFunc(
					// Verify the organization data source works with the custom concurrency setting
					resource.TestCheckResourceAttrSet("data.lightdash_organization.test", "organization_uuid"),
				),
			},
		},
	})
}

func TestAccProvider_concurrencyConfigDefault(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for provider default concurrency configuration")
	}

	// Get the provider config (without max_concurrent_requests, should use default)
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Simple test config that uses the organization data source
	testConfig := `
data "lightdash_organization" "test" {
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testConfig,
				Check: resource.ComposeTestCheckFunc(
					// Verify the organization data source works with the default concurrency setting
					resource.TestCheckResourceAttrSet("data.lightdash_organization.test", "organization_uuid"),
				),
			},
		},
	})
}
