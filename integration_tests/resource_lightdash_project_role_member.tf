resource "lightdash_project_role_member" "test_admin_user" {
  project_uuid = var.test_lightdash_project_uuid
  user_uuid    = data.lightdash_organization_member.test_admin_user.user_uuid
  role         = "admin"
}

output "lightdash_project_role_member__test_admin_user" {
  value     = lightdash_project_role_member.test_admin_user
  sensitive = true
}

resource "lightdash_project_role_member" "test_member_user" {
  count = (length(data.lightdash_organization_member.test_member_user) > 0 ? 1 : 0)

  project_uuid = var.test_lightdash_project_uuid
  user_uuid    = data.lightdash_organization_member.test_member_user[0].user_uuid
  role         = "editor"
}

output "lightdash_project_role_member__test_member_user" {
  value     = lightdash_project_role_member.test_member_user
  sensitive = true
}
