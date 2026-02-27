##########################################################################
# Public and private spaces
##########################################################################
resource "lightdash_space" "test_public" {
  project_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name         = "zzz_test_public_space"
  is_private   = false

  deletion_protection = false
}


resource "lightdash_space" "test_private" {
  project_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name         = "zzz_test_private_space"
  // The visibility is private by default.
  is_private = true

  deletion_protection = false

  access {
    user_uuid  = "xxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxx"
    space_role = "editor"
  }

  access {
    user_uuid  = "xxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxx"
    space_role = "viewer"
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

// Nested spaces inherit visibility and access from the parent by default,
// but can also have their own Restricted Access.
resource "lightdash_space" "test_child_space_inherited" {
  project_uuid        = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  parent_space_uuid   = lightdash_space.test_parent_space.space_uuid
  name                = "zzz_test_child_space_inherited"
  deletion_protection = false
}

// Restricted nested space with its own access control list.
resource "lightdash_space" "test_child_space_restricted" {
  project_uuid        = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  parent_space_uuid   = lightdash_space.test_parent_space.space_uuid
  name                = "zzz_test_child_space_restricted"
  is_private          = true
  deletion_protection = false

  access {
    user_uuid  = "xxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxx"
    space_role = "editor"
  }
}
