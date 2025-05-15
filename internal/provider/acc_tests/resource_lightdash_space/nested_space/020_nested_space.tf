# Public spaces
resource "lightdash_space" "nested_space_public_root" {
  project_uuid = data.lightdash_project.test.project_uuid
  name         = "Public Root Space (Acceptance Test: nested_space - 020)"
  // Test changing the is_private attribute to true
  is_private          = true
  deletion_protection = false
}

resource "lightdash_space" "nested_space_public_child" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Public Child Space (Acceptance Test: nested_space - 020)"
  deletion_protection = false
  parent_space_uuid   = lightdash_space.nested_space_public_root.space_uuid
}

resource "lightdash_space" "nested_space_public_grandchild" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Public Grandchild Space (Acceptance Test: nested_space - 020)"
  deletion_protection = false
  parent_space_uuid   = lightdash_space.nested_space_public_root.space_uuid
}

# Private spaces
resource "lightdash_space" "nested_space_private_root" {
  project_uuid = data.lightdash_project.test.project_uuid
  name         = "Private Root Space (Acceptance Test: nested_space - 020)"
  // Test changing the is_private attribute to false
  is_private          = false
  deletion_protection = false
}

resource "lightdash_space" "nested_space_private_child" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Private Child Space (Acceptance Test: nested_space - 020)"
  deletion_protection = false
  parent_space_uuid   = lightdash_space.nested_space_private_root.space_uuid
}

resource "lightdash_space" "nested_space_private_grandchild" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Private Grandchild Space (Acceptance Test: nested_space - 020)"
  deletion_protection = false
  parent_space_uuid   = lightdash_space.nested_space_private_root.space_uuid
}
