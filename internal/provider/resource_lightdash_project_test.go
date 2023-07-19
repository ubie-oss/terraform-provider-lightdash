// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
