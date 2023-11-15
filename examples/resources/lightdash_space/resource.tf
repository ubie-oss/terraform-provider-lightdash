resource "lightdash_space" "test_public" {
  project_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name         = "zzz_test_private_space"
  // The visibility is public by default.
  // is_private   = false

  deletion_protection = false
}


locals {
  test_private_accessors = [
    {
      user_uuid = "xxxx-xxxx-xxx"
    },
    {
      user_uuid = "xxxx-xxxx-xxx"
    }
  ]
}

resource "lightdash_space" "test_privacte" {
  project_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name         = "zzz_test_private_space"
  is_private   = true

  deletion_protection = false

  dynamic "access" {
    for_each = local.test_private_accessors
    content {
      user_uuid = access.value.user_uuid
    }
  }
}
