# NOTE We don't recommend using this resource in production. It's only for testing purposes,
#      because you as an admin potentailly lose access to the project.

resource "lightdash_organization_role_member" "test" {
  count = (length(data.lightdash_organization_member.test_member_user) > 0 ? 1 : 0)

  organization_uuid = data.lightdash_organization.test.organization_uuid
  user_uuid         = data.lightdash_organization_member.test_member_user[0].user_uuid
  role              = "interactive_viewer"
}

output "lightdash_organization_role_member__test" {
  value = (length(lightdash_organization_role_member.test) > 0 ? lightdash_organization_role_member.test : null)
}
