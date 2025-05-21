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

func TestAccOrganizationMembersByEmailsDataSource(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for data_source_lightdash_organization_members_by_emails")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Test of simple organization members retrieval
	config, err := ReadAccTestResource([]string{"data_sources", "lightdash_organization_members_by_emails", "data", "01_data.tf"})
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + config,
				Check: resource.ComposeTestCheckFunc(
					// The number of members should be the same as the number of members in the all_members data source
					resource.TestCheckResourceAttrPair(
						"data.lightdash_organization_members_by_emails.all_members", "members.#",
						"data.lightdash_organization_members.all_members", "members.#",
					),
					// The number of members should be 1
					resource.TestCheckResourceAttr("data.lightdash_organization_members_by_emails.one_member", "members.#", "1")),
			},
		},
	})
}
