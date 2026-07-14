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

# NOTE Destroying this resource clears upstreamProjectUuid on the source project.
# NOTE Test the import manually.
#  export ORGANIZATION_UUID="..."
#  export PROJECT_UUID="..."
#  terraform import -var-file testing.tfvars \
#    "lightdash_project_upstream.test[0]" \
#    "organizations/${ORGANIZATION_UUID}/projects/${PROJECT_UUID}/upstream"
#
# Set test_lightdash_upstream_project_uuid in testing.tfvars to another project
# UUID in the same organization before applying. Leave it unset/null to skip.
resource "lightdash_project_upstream" "test" {
  count = var.test_lightdash_upstream_project_uuid != null && var.test_lightdash_upstream_project_uuid != "" ? 1 : 0

  organization_uuid     = data.lightdash_organization.test.organization_uuid
  project_uuid          = data.lightdash_project.test.project_uuid
  upstream_project_uuid = var.test_lightdash_upstream_project_uuid
}
