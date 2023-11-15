data "lightdash_spaces" "test" {
  organization_uuid = data.lightdash_projects.test.organization_uuid
  project_uuid      = var.test_lightdash_project_uuid
}

output "lightdash_spaces__test" {
  value = data.lightdash_spaces.test
}
