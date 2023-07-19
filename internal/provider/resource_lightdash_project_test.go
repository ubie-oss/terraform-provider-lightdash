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

func TestAccLightdashProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    []resource.TestStep{
			// // Create and Read testing
			// {
			// 	Config: providerConfig + testAccLightdashProjectResourceConfig("one"),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("lightdash_project.test", "id", "example-id"),
			// 		resource.TestCheckResourceAttr("lightdash_project.test", "name", "one"),
			// 	),
			// },
			// // ImportState testing
			// {
			// 	ResourceName:            "lightdash_project.test",
			// 	ImportState:             true,
			// 	ImportStateVerify:       true,
			// 	ImportStateVerifyIgnore: []string{"name", "defaulted"},
			// },
			// // Update and Read testing
			// {
			// 	Config: providerConfig + testAccLightdashProjectResourceConfig("two"),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("lightdash_project.test", "name", "two"),
			// 	),
			// },
			// // Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLightdashProjectResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
resource "lightdash_project" "test" {
  name = %[1]q
}
`, configurableAttribute)
}
