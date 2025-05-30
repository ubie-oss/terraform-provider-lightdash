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

resource "lightdash_group" "test1" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "zzz_test_group_01"
  members           = []
}

resource "lightdash_group" "test2" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "zzz_test_group_02"

  members = [
    { user_uuid = data.lightdash_authenticated_user.test.user_uuid },
  ]
}

output "lightdash_group__test1" {
  value = lightdash_group.test1
}

output "lightdash_group__test2" {
  value = lightdash_group.test2
}
