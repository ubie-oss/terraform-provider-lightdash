data "lightdash_project" "test" {
  project_uuid = var.test_lightdash_project_uuid
}

output "lightdash_project_test" {
  value = data.lightdash_project.test
}
