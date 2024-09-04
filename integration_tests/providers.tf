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

terraform {
  required_version = "1.1.0"

  # Please reconfigure the backend for your own environment
  # TODO set the backend configuration with a command line option of Terraform
  backend "gcs" {
    bucket = "ubie-oss-data-engineering-terraform-provider-lightdash"
    prefix = "state"
    # Service account impersonation is required to access the GCS bucket
    impersonate_service_account = "terraform-provider-lightdash@ubie-oss-data-engineering.iam.gserviceaccount.com"
  }

  required_providers {
    # tflint-ignore: terraform_required_providers
    lightdash = {
      source = "github.com/ubie-oss/lightdash"
    }
  }
}

provider "lightdash" {
  host  = var.lightdash_url
  token = var.personal_access_token
}
