##########################################################################
# Public and private spaces
##########################################################################
resource "lightdash_space" "test_public" {
  project_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name         = "zzz_test_private_space"
  // The visibility is private by default.
  is_private = true

  deletion_protection = false
}


resource "lightdash_space" "test_privacte" {
  project_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name         = "zzz_test_private_space"
  // The visibility is private by default.
  is_private = true

  deletion_protection = false

  access {
    user_uuid = "xxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxx"
  }

  access {
    user_uuid = "xxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxx"
  }
}

##########################################################################
# Nested spaces
##########################################################################
resource "lightdash_space" "test_parent_space" {
  project_uuid        = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name                = "zzz_test_parent_space"
  is_private          = true
  deletion_protection = false
}

resource "lightdash_space" "test_child_space" {
  project_uuid        = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  parent_space_uuid   = lightdash_space.test_parent_space.space_uuid
  name                = "zzz_test_child_space"
  is_private          = true
  deletion_protection = false
}

resource "lightdash_space" "test_grandchild_space" {
  project_uuid        = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  parent_space_uuid   = lightdash_space.test_child_space.space_uuid
  name                = "zzz_test_grandchild_space"
  is_private          = true
  deletion_protection = false
}
