data "lightdash_organization" "test" {
}

resource "lightdash_group" "test_group1" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "test user attribute 1 (Acceptance Test - create)"
  members           = []
}

resource "lightdash_group" "test_group2" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "test user attribute 2 (Acceptance Test - create)"
  members           = []
}

resource "lightdash_user_attribute" "test" {
  name               = "zzz_acc_test_user_attribute_update"
  description        = "Acceptance Test - update"
  attribute_defaults = ["default", "other"]

  groups = [
    {
      group_uuid = lightdash_group.test_group1.group_uuid
      values     = ["us", "eu"]
    },
    {
      group_uuid = lightdash_group.test_group2.group_uuid
      values     = ["jp"]
    },
  ]
}

data "lightdash_user_attribute" "check" {
  name = lightdash_user_attribute.test.name

  depends_on = [lightdash_user_attribute.test]
}
