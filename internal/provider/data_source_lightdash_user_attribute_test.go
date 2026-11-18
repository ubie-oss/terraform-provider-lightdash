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
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The successful lookup is covered by TestAccUserAttributeResource, which reads the attribute
// it creates back through this data source.
func TestAccUserAttributeDataSourceNotFound(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for data_source_lightdash_user_attribute")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}
	config, err := ReadAccTestResource([]string{"data_sources", "lightdash_user_attribute", "error", "010_not_found.tf"})
	if err != nil {
		t.Fatalf("Failed to get userAttributeNotFoundConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig + config,
				ExpectError: regexp.MustCompile(`user attribute not found`),
			},
		},
	})
}
