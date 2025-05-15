resource "lightdash_space" "create_space__test_public" {
  project_uuid = data.lightdash_project.test.project_uuid
  name         = "Public Space (Acceptance Test: create_space - 020)"
  // Change the visibility
  is_private = true
  // Change the deletion protection
  deletion_protection = false
}

resource "lightdash_space" "create_space__test_private" {
  project_uuid = data.lightdash_project.test.project_uuid
  name         = "Private Space (Acceptance Test: create_space - 020)"
  // Change the visibility
  is_private = false
  // Change the deletion protection
  deletion_protection = false
}
