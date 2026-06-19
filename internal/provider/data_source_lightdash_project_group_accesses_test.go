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

func TestAccProjectGroupAccessesDataSource(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for data_source_lightdash_project_group_accesses")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	grantConfig, err := ReadAccTestResource([]string{"resources", "lightdash_project_role_group", "grant", "010_grant.tf"})
	if err != nil {
		t.Fatalf("Failed to get grant config: %v", err)
	}

	dataSourceConfig, err := ReadAccTestResource([]string{"data_sources", "lightdash_project_group_accesses", "data", "01_data.tf"})
	if err != nil {
		t.Fatalf("Failed to get data source config: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + grantConfig + dataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.lightdash_project_group_accesses.test", "organization_uuid"),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_group_accesses.test",
						"project_uuid",
						"data.lightdash_project.test",
						"project_uuid",
					),
					resource.TestCheckResourceAttrPair(
						"data.lightdash_project_group_accesses.test",
						"groups.0.group_uuid",
						"lightdash_group.test_group",
						"group_uuid",
					),
					resource.TestCheckResourceAttr("data.lightdash_project_group_accesses.test", "groups.0.role", "viewer"),
				),
			},
		},
	})
}
