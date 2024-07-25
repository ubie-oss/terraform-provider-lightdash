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

data "lightdash_group" "test1" {
  organization_uuid = data.lightdash_projects.test.organization_uuid
  project_uuid      = var.test_lightdash_project_uuid
  group_uuid        = lightdash_group.test1.group_uuid
}

data "lightdash_group" "test2" {
  organization_uuid = data.lightdash_projects.test.organization_uuid
  project_uuid      = var.test_lightdash_project_uuid
  group_uuid        = lightdash_group.test2.group_uuid
}

output "data_lightdash_group__test1" {
  value = data.lightdash_group.test1
}

output "data_lightdash_group__test2" {
  value = data.lightdash_group.test2
}
