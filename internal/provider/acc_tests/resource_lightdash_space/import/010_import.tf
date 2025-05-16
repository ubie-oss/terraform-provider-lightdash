# Public spaces
resource "lightdash_space" "import__public_root_space" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Public Root Space (Acceptance Test: import)"
  is_private          = false
  deletion_protection = false
}

resource "lightdash_space" "import__public_child_space" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Public Child Space (Acceptance Test: import)"
  deletion_protection = false
  parent_space_uuid   = lightdash_space.import__public_root_space.space_uuid
}

# Private spaces
resource "lightdash_space" "import__private_root_space" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Private Root Space (Acceptance Test: import)"
  is_private          = true
  deletion_protection = false
}

resource "lightdash_space" "import__private_child_space" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Private Child Space (Acceptance Test: import)"
  deletion_protection = false
  parent_space_uuid   = lightdash_space.import__private_root_space.space_uuid
}
