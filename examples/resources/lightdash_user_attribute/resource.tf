resource "lightdash_group" "sales_us" {
  organization_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name              = "sales_us"
  members           = []
}

resource "lightdash_group" "sales_emea" {
  organization_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name              = "sales_emea"
  members           = []
}

resource "lightdash_user_attribute" "region" {
  name        = "region"
  description = "The sales regions that a user can see"

  attribute_defaults = ["none"]

  groups = [
    {
      group_uuid = lightdash_group.sales_us.group_uuid
      values     = ["us"]
    },
    {
      group_uuid = lightdash_group.sales_emea.group_uuid
      values     = ["uk", "de", "fr"]
    },
  ]
}

resource "lightdash_user_attribute" "is_admin" {
  name = "is_admin"
}
