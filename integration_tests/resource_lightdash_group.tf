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
  test_group_members = {for user in data.lightdash_organization_member.test_member_user : user.user_uuid => user}
}

resource "lightdash_group" "test1" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name         = "zzz_test_group_01"
}

resource "lightdash_group" "test2" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name         = "zzz_test_group_02"

  dynamic "member" {
    for_each = local.test_group_members
    content {
      user_uuid = member.key
    }
  }
}

output "lightdash_group__test1" {
  value = lightdash_group.test1
}

output "lightdash_group__test2" {
  value = lightdash_group.test2
}
