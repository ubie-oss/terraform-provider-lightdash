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

resource "lightdash_space" "test_public" {
  project_uuid = var.test_lightdash_project_uuid
  name         = "zzz_test_public_space"
  // The visibility is private by default.
  is_private = false

  deletion_protection = false
}

resource "lightdash_space" "test_private" {
  project_uuid = var.test_lightdash_project_uuid
  name         = "zzz_test_private_space"
  is_private   = true

  deletion_protection = false
}

output "lightdash_space__test_public_space" {
  value = lightdash_space.test_public
}

output "lightdash_space__test_private_space" {
  value = lightdash_space.test_private
}
