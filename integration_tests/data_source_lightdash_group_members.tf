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
  groups = {
    for group in data.lightdash_organization_groups.test.groups :
    group.group_uuid => group...
  }
}

data "lightdash_group_members" "test" {
  for_each = (
    data.lightdash_organization_groups.test.groups != null
    ? local.groups : {}
  )

  organization_uuid = data.lightdash_projects.test.organization_uuid
  project_uuid      = var.test_lightdash_project_uuid
  group_uuid        = each.key

  depends_on = [
    lightdash_group.test1,
    lightdash_group.test2,
  ]
}

output "lightdash_group_members__test" {
  value = data.lightdash_group_members.test
}
