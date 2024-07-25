# Copyright 2023 Ubie, inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

locals {
  test_private_spaces_access_members = (length(data.lightdash_organization_member.test_member_user) > 0
    ? concat(
      [data.lightdash_organization_member.test_admin_user.user_uuid],
      [for user in data.lightdash_organization_member.test_member_user : user.user_uuid],
    )
    : [
      data.lightdash_organization_member.test_admin_user.user_uuid,
  ])
}

resource "lightdash_space" "test_public" {
  project_uuid = var.test_lightdash_project_uuid
  name         = "zzz_test_public_space"
  // The visibility is private by default.
  is_private = false

  deletion_protection = false
}

# A command to import the resource
# terraform import lightdash_space.test_private "projects/<project_uuid>/spaces/<space_uuid>>"
resource "lightdash_space" "test_private" {
  project_uuid = var.test_lightdash_project_uuid
  name         = "zzz_test_private_space"
  is_private   = true

  deletion_protection = false

  dynamic "access" {
    for_each = toset(local.test_private_spaces_access_members)
    content {
      user_uuid  = access.key
      space_role = "editor"
    }
  }

  group_access {
    group_uuid = lightdash_group.test1.group_uuid
    space_role = "admin"
  }

  group_access {
    group_uuid = lightdash_group.test2.group_uuid
    space_role = "admin"
  }

  depends_on = [
    lightdash_project_role_member.test_admin_user,
    lightdash_project_role_member.test_member_user,
  ]
}

output "lightdash_space__test_public_space" {
  value = lightdash_space.test_public
}

output "lightdash_space__test_private_space" {
  value = lightdash_space.test_private
}
