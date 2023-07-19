resource "lightdash_space_access_member" "example" {
  project_uuid = lightdash_space.example_public.project_uuid
  space_uuid   = lightdash_space.example_public.uuid
  user_uuid    = data.lightdash_organization_member.example.user_uuid
  role         = "developer"
}
