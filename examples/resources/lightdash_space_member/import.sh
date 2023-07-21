# Space can be impported by specifing the resource identifier.
terraform import lightdash_project_role_member.example "projects/${project_uuid}/spaces/${space_uuid}/access/${user_uuid}"
