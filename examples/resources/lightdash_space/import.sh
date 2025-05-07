# Spaces can be imported by specifying the resource identifier.
terraform import lightdash_space.example "projects/${project-uuid}/spaces/${space_uuid}"


terraform import \
  --var-file=testing.tfvars \
  lightdash_space.test_grandchild_space \
  "projects/9cc0bae8-f552-4ac0-bdcc-44933d7031ae/spaces/0ca1503b-e5c9-4698-b3db-bd7998974555"
