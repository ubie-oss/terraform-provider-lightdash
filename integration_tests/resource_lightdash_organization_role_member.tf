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

resource "lightdash_organization_role_member" "test" {
  count = (length(data.lightdash_organization_member.test_member_user) > 0 ? 1 : 0)

  organization_uuid = data.lightdash_organization.test.organization_uuid
  user_uuid         = data.lightdash_organization_member.test_member_user[0].user_uuid
  role              = "member"
}

output "lightdash_organization_role_member__test" {
  value = (length(lightdash_organization_role_member.test) > 0 ? lightdash_organization_role_member.test : null)
}
