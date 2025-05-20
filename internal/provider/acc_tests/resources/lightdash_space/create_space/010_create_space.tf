resource "lightdash_space" "create_space__test_public" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Public Space (Acceptance Test: create_space)"
  is_private          = false
  deletion_protection = true
}

resource "lightdash_space" "create_space__test_private" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Private Space (Acceptance Test: create_space)"
  is_private          = true
  deletion_protection = true
}
