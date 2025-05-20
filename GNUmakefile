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
include .env
testacc:
	echo "${LIGHTDASH_PROJECT}"
	TF_ACC=1 \
		LIGHTDASH_URL="${LIGHTDASH_URL}" \
		LIGHTDASH_API_KEY="${LIGHTDASH_API_KEY}" \
		LIGHTDASH_PROJECT="${LIGHTDASH_PROJECT}" \
		TF_LOG=DEBUG \
		go test ./internal/provider/... -v $(TESTARGS) -timeout 120m

test:
	# TF_ACC mustn't be set, otherwise acceptance tests will run
	unset TF_ACC && cd "internal/" && go test -count=1 -v ./...

clean:
	go clean -cache -modcache -i -r

build: gen-docs go-tidy gosec deadcode
	go build -v ./

gosec:
	gosec ./internal/...

deadcode:
	deadcode -test ./...

upgrade-go-mod:
	# Upgrade dependencies
	go get -u ./...
	# Upgrade go.mod
	go mod tidy
	# Upgrade go.sum
	go mod vendor

lint: run-trunk-check run-pre-commit

run-trunk-check:
	trunk check --all

format: format-go format-trunk

format-trunk:
	trunk fmt --all

format-go:
	go fmt ./internal/...

install:
	go build -v ./ && go install .

gen-docs:
	go generate ./...

go-tidy:
	go mod tidy

# Set up the development environment
setup-dev: unset-git-hooks setup-pre-commit setup-trunk

unset-git-hooks:
	git config --unset-all core.hooksPath || true

setup-trunk:
	trunk git-hooks sync

setup-pre-commit:
	pre-commit install

update: update-pre-commit update-trunk

update-trunk:
	trunk upgrade

update-pre-commit:
	pre-commit autoupdate

run-pre-commit:
	pre-commit run --all-files

###################################################################
# Integration Tests
###################################################################
integration-tests: build terraform-apply

terraform-init:
	@make -C ./integration_tests/ terraform-init

terraform-apply:
	@make -C ./integration_tests/ terraform-apply

terraform-destroy:
	@make -C ./integration_tests/ terraform-destroy

terraform-show:
	@make -C ./integration_tests/ terraform-show
