package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLightdashProjectDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    []resource.TestStep{
			// Read testing
			// {
			// 	Config: providerConfig + testAccLightdashProjectDataSourceConfig,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("data.lightdash_project.test", "id", "example-id"),
			// 	),
			// },
		},
	})
}

const testAccLightdashProjectDataSourceConfig = `
data "lightdash_project" "test" {
  project_uuid = "example"
}
`
