data "lightdash_organization_member" "test_admin_user" {
  email = var.test_organization_admin_user_email
}

output "lightdash_organization_member_test_admin_user" {
  value = data.lightdash_organization_member.test_admin_user
}


data "lightdash_organization_member" "test_member_user" {
  count = (var.test_organization_member_user_email != null ? 1 : 0)

  email = var.test_organization_member_user_email
}

output "lightdash_organization_member_test_member_user" {
  value = (length(data.lightdash_organization_member.test_member_user) > 0
  ? element(data.lightdash_organization_member.test_member_user, 0) : null)
}
