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

.PHONY: terraform-init
terraform-init:
	terraform init

.PHONY: terraform-apply
terraform-apply:
	TF_LOG=DEBUG terraform apply -var-file="testing.tfvars"

.PHONY: terraform-destroy
terraform-destroy:
	TF_LOG=DEBUG terraform destroy -var-file="testing.tfvars"

.PHONY: terraform-show
terraform-show:
	terraform show
