resource "lightdash_organization_role_member" "test" {
  for_each = {
    for member in data.lightdash_organization_members.test.members
    : member.user_uuid => member
    if member.role != "admin"
  }

  organization_uuid = data.lightdash_organization.test.organization_uuid
  user_uuid         = each.value.user_uuid
  role              = each.value.role
}

output "lightdash_organization_role_member__test" {
  value = lightdash_organization_role_member.test
}
