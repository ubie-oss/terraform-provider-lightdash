resource "lightdash_space" "test_public" {
  project_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name         = "zzz_test_private_space"
  // The visibility is public by default.
  // is_private   = false

  deletion_protection = false
}


resource "lightdash_space" "test_public" {
  project_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name         = "zzz_test_private_space"
  is_private   = true

  deletion_protection = false
}
