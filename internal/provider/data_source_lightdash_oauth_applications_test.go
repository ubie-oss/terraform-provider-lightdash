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

// Requires org-admin LIGHTDASH_API_KEY.
func TestAccOAuthApplicationDataSources(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for lightdash_oauth_application data sources")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	dataConfig, err := ReadAccTestResource([]string{"data_sources", "lightdash_oauth_applications", "data", "010_data.tf"})
	if err != nil {
		t.Fatalf("Failed to get data source config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + dataConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_oauth_application.acc_test", "client_name", "test (Acceptance Test - oauth data sources)"),
					resource.TestCheckResourceAttr("data.lightdash_oauth_application.by_id", "client_name", "test (Acceptance Test - oauth data sources)"),
					resource.TestCheckResourceAttrSet("data.lightdash_oauth_application.by_id", "organization_uuid"),
					resource.TestCheckResourceAttrSet("data.lightdash_oauth_applications.all", "applications.0.client_id"),
				),
			},
		},
	})
}
