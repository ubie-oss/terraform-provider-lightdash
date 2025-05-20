data "lightdash_organization" "test" {
}

resource "lightdash_group" "test_group" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "test (Acceptance Test - import)"
  members           = []
}
