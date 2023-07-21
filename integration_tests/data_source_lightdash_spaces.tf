data "lightdash_spaces" "test" {
  for_each = { for project in data.lightdash_projects.test.projects : project.project_uuid => project }

  organization_uuid = data.lightdash_projects.test.organization_uuid
  project_uuid      = each.value.project_uuid
}

output "lightdash_spaces__test" {
  value = data.lightdash_spaces.test
}
