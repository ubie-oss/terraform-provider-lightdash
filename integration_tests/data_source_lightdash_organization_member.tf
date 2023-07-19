data "lightdash_organization_member" "test" {
  for_each = { for member in data.lightdash_organization_members.test.members : member.user_uuid => member }

  email = each.value.email
}

output "lightdash_organization_member_test" {
  value = data.lightdash_organization_member.test
}
