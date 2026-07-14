data "lightdash_organization" "test" {
}

data "lightdash_project_upstream" "test" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  project_uuid      = data.lightdash_project.test.project_uuid
}
