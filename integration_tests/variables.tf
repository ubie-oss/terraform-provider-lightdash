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

variable "lightdash_url" {
  description = "The URL of the Lightdash instance to test."
  type        = string
}

variable "personal_access_token" {
  description = "Lightdash personal access token"
  type        = string
  sensitive   = true
}

variable "test_lightdash_project_uuid" {
  description = "The UUID of the Lightdash project to test."
  type        = string
}

variable "test_organization_admin_user_email" {
  description = "The email of the Lightdash organization admin user to test."
  type        = string
}

variable "test_organization_member_user_emails" {
  description = "The email of the Lightdash organization member user to test in the testing project."
  type        = list(string)
  default     = []
}
