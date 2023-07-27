resource "lightdash_space_access_member" "example" {
  project_uuid = "xxxxx-xxxxx-xxxx"
  space_uuid   = "yyyy-yyyy-yyy"
  user_uuid    = data.lightdash_organization_member.example.user_uuid
}
