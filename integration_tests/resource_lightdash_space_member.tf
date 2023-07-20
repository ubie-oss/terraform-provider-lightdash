resource "lightdash_space_member" "test" {
  for_each = {
    for member in data.lightdash_project_members.test.members
    : member.user_uuid => member
    if member.role != "admin"
  }

  project_uuid = lightdash_space.test_private[0].project_uuid
  space_uuid   = lightdash_space.test_private[0].space_uuid
  user_uuid    = each.value.user_uuid

  depends_on = [
    # Members must be project members a head.
    lightdash_project_role_member.test,
  ]
}

output "lightdash_space_member__test" {
  value = lightdash_space_member.test
}
