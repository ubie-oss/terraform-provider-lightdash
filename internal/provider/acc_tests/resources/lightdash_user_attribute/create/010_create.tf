data "lightdash_organization" "test" {
}

resource "lightdash_group" "test_group1" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "test user attribute 1 (Acceptance Test - create)"
  members           = []
}

resource "lightdash_user_attribute" "test" {
  name               = "zzz_acc_test_user_attribute_create"
  description        = "Acceptance Test - create"
  attribute_defaults = ["default"]

  groups = [
    {
      group_uuid = lightdash_group.test_group1.group_uuid
      values     = ["us"]
    },
  ]
}

data "lightdash_user_attribute" "check" {
  name = lightdash_user_attribute.test.name

  depends_on = [lightdash_user_attribute.test]
}
