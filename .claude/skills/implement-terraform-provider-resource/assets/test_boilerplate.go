// Copyright 2023 Ubie, inc.
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExampleResource(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test")
	}

	providerConfig, _ := getProviderConfig()
	testConfig, _ := ReadAccTestResource([]string{"resources", "example", "010_create.tf"})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_example.test", "name", "expected-name"),
				),
			},
			{
				ResourceName:      "lightdash_example.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
