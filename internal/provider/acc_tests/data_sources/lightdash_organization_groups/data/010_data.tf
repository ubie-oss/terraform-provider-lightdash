data "lightdash_organization" "test" {
}

// Guarantee that the group exists before the data source is read
resource "lightdash_group" "test_group" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "(Acceptance Test - data_source_lightdash_organization_groups)"
  members           = []
}

data "lightdash_organization_groups" "test" {
  organization_uuid = data.lightdash_organization.test.organization_uuid

  depends_on = [
    lightdash_group.test_group,
  ]
}
