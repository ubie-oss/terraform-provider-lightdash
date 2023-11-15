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
}
