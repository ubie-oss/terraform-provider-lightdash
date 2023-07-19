data "lightdash_projects" "test" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
}

output "lightdash_projects_test" {
  value = tolist(data.lightdash_projects.test.projects)[0]
}
