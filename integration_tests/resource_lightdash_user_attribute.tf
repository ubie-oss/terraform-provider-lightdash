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

# Values are assigned only to the groups created for the tests.
resource "lightdash_user_attribute" "test1" {
  name               = "zzz_test_user_attribute_01"
  description        = "Integration test"
  attribute_defaults = ["default"]

  groups = [
    {
      group_uuid = lightdash_group.test1.group_uuid
      values     = ["us", "eu"]
    },
  ]
}

resource "lightdash_user_attribute" "test2" {
  name = "zzz_test_user_attribute_02"
}

output "lightdash_user_attribute__test1" {
  value = lightdash_user_attribute.test1
}

output "lightdash_user_attribute__test2" {
  value = lightdash_user_attribute.test2
}
