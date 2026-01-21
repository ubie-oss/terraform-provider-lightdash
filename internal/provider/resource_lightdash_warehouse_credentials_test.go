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

func TestAccWarehouseCredentialsResource_create(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_warehouse_credentials")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	createConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_warehouse_credentials", "create", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get createConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig010,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_warehouse_credentials.test_bigquery", "name", "BigQuery Test Credentials"),
					resource.TestCheckResourceAttr("lightdash_warehouse_credentials.test_bigquery", "warehouse_type", "bigquery"),
					resource.TestCheckResourceAttrSet("lightdash_warehouse_credentials.test_bigquery", "organization_warehouse_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_warehouse_credentials.test_bigquery", "project"),
				),
			},
		},
	})
}

func TestAccWarehouseCredentialsResource_import(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_warehouse_credentials")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	importConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_warehouse_credentials", "import", "010_import.tf"})
	if err != nil {
		t.Fatalf("Failed to get importConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + importConfig010,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_warehouse_credentials.test_bigquery", "name", "BigQuery Import Test"),
					resource.TestCheckResourceAttr("lightdash_warehouse_credentials.test_bigquery", "warehouse_type", "bigquery"),
				),
			},
			{
				Config:            providerConfig + importConfig010,
				ResourceName:      "lightdash_warehouse_credentials.test_bigquery",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"keyfile_contents", // Sensitive field not returned by API
					"project",          // Credential details not returned by API
					"dataset",
					"location",
					"timeout_seconds",
					"maximum_bytes_billed",
					"priority",
					"retries",
					"start_of_week",
				},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					res, ok := state.RootModule().Resources["lightdash_warehouse_credentials.test_bigquery"]
					if !ok {
						return "", fmt.Errorf("resource not found in state for import")
					}
					organizationUUID := res.Primary.Attributes["organization_uuid"]
					warehouseUUID := res.Primary.Attributes["organization_warehouse_uuid"]
					return fmt.Sprintf("organizations/%s/warehouse-credentials/%s", organizationUUID, warehouseUUID), nil
				},
			},
		},
	})
}
