data "lightdash_organization" "test" {
}

resource "lightdash_space" "create_space__test_public" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Public Space (Acceptance Test: create_space)"
  is_private          = false
  deletion_protection = true
}


data "lightdash_space" "create_space__test_public" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  project_uuid      = data.lightdash_project.test.project_uuid
  space_uuid        = lightdash_space.create_space__test_public.space_uuid

  depends_on = [
    lightdash_space.create_space__test_public
  ]
}

resource "lightdash_space" "create_space__test_private" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Private Space (Acceptance Test: create_space)"
  is_private          = true
  deletion_protection = true
}


data "lightdash_space" "create_space__test_private" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  project_uuid      = data.lightdash_project.test.project_uuid
  space_uuid        = lightdash_space.create_space__test_private.space_uuid

  depends_on = [
    lightdash_space.create_space__test_private
  ]
}
