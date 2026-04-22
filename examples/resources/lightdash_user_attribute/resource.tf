resource "lightdash_user_attribute" "simple" {
  name              = "department"
  description       = "The department the user belongs to"
  attribute_default = "unknown"
}

resource "lightdash_user_attribute" "with_assignments" {
  name              = "region"
  description       = "Sales region assigned to the user"
  attribute_default = "global"

  users = [
    {
      user_uuid = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      value     = "apac"
    },
    {
      user_uuid = "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy"
      value     = "emea"
    },
  ]

  groups = [
    {
      group_uuid = "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz"
      value      = "na"
    },
  ]
}
