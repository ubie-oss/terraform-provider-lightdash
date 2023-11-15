resource "lightdash_space" "test_public" {
  project_uuid = var.test_lightdash_project_uuid
  name         = "zzz_test_public_space"
  // The visibility is private by default.
  is_private = false

  deletion_protection = false
}

resource "lightdash_space" "test_private" {
  project_uuid = var.test_lightdash_project_uuid
  name         = "zzz_test_private_space"
  is_private   = true

  deletion_protection = false
}

output "lightdash_space__test_public_space" {
  value = lightdash_space.test_public
}

output "lightdash_space__test_private_space" {
  value = lightdash_space.test_private
}
