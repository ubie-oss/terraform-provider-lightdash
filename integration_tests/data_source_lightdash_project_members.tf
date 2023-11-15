data "lightdash_project_members" "test" {
  project_uuid = var.test_lightdash_project_uuid
}

output "lightdash_project_members_test" {
  value = data.lightdash_project_members.test
}
