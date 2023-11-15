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

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

test:
	# TF_ACC mustn't be set, otherwise acceptance tests will run
	unset TF_ACC && cd "internal/" && go test -count=1 -v ./...

build: gen-docs
	go build -v ./

format:
	go fmt ./internal/...

install:
	go build -v ./ && go install .

gen-docs:
	go generate ./...

tidy:
	go mod tidy

setup-dev: setup-pre-commit

setup-pre-commit:
	pre-commit install

run-pre-commit:
	pre-commit run --all-files

plan-integration-tests:
	cd ./integration_tests/ && TF_LOG=1 terraform plan 2>&1

apply-integration-tests:
	cd ./integration_tests/ && TF_LOG=1 terraform apply -auto-approve 2>&1

destroy-integration-tests:
	cd ./integration_tests/ && TF_LOG=1 terraform destroy -auto-approve 2>&1
