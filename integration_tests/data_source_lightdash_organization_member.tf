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

data "lightdash_organization_member" "test_admin_user" {
  email = var.test_organization_admin_user_email
}

output "lightdash_organization_member_test_admin_user" {
  value = data.lightdash_organization_member.test_admin_user
}


data "lightdash_organization_member" "test_member_user" {
  count = (var.test_organization_member_user_email != null ? 1 : 0)

  email = var.test_organization_member_user_email
}

output "lightdash_organization_member_test_member_user" {
  value = (length(data.lightdash_organization_member.test_member_user) > 0
  ? element(data.lightdash_organization_member.test_member_user, 0) : null)
}
