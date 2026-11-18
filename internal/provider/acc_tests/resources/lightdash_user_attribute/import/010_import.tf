data "lightdash_organization" "test" {
}

resource "lightdash_group" "test_group" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "test user attribute (Acceptance Test - import)"
  members           = []
}

resource "lightdash_user_attribute" "test" {
  name               = "zzz_acc_test_user_attribute_import"
  description        = "Acceptance Test - import"
  attribute_defaults = ["default"]

  groups = [
    {
      group_uuid = lightdash_group.test_group.group_uuid
      values     = ["us", "eu"]
    },
  ]
}
