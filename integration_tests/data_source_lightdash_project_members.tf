locals {
  test_project = [for project in data.lightdash_projects.test.projects : project][0]
}


data "lightdash_project_members" "test" {
  project_uuid = local.test_project.project_uuid
}

output "lightdash_project_members_test" {
  value = data.lightdash_project_members.test
}
