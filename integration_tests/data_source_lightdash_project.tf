data "lightdash_project" "test" {
  for_each = { for project in data.lightdash_projects.test.projects : project.project_uuid => project }

  project_uuid = each.key
}

output "lightdash_project_test" {
  value = data.lightdash_project.test
}
