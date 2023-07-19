locals {
  test_projects = tolist(data.lightdash_projects.test.projects)
}

resource "lightdash_space" "test_public" {
  count = (length(local.test_projects) > 0) ? 1 : 0

  project_uuid = local.test_projects[0].project_uuid
  name         = "zzz_test_public_space"
  // The visibility is public by default.
  // is_private   = false

  deletion_protection = false
}

resource "lightdash_space" "test_private" {
  count = (length(local.test_projects) > 0) ? 1 : 0

  project_uuid = local.test_projects[0].project_uuid
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
