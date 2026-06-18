data "lightdash_project_group_accesses" "test" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  project_uuid      = data.lightdash_project.test.project_uuid

  depends_on = [lightdash_project_role_group.test_project_role_group]
}
