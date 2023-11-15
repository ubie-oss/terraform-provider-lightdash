resource "lightdash_space_access_member" "test_private" {
  project_uuid = var.test_lightdash_project_uuid
  space_uuid   = lightdash_space.test_private.space_uuid
  user_uuid    = data.lightdash_organization_member.test_admin_user.user_uuid
}

output "lightdash_space_access_member__test" {
  value = lightdash_space_access_member.test_private
}
