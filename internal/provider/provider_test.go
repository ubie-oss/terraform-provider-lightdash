// Copyright 2023 Ubie, inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func getTestProviderConfig(api_key *string) string {
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the client is properly configured.
	template := `
provider "lightdash" {
  host = "https://app.lightdash.cloud"
  token = "%[1]%"
}`
	providerConfig := fmt.Sprintf(template, api_key)
	return providerConfig
}

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"lightdash": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.

	// Use the utility function from utils.go to check if in integration test mode
	if !isIntegrationTestMode() {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set to '1'")
	}

	// Use the utility functions from utils.go to check for API Key and Project UUID
	if _, err := getLightdashApiKey(); err != nil {
		t.Fatalf("LIGHTDASH_API_KEY must be set for acceptance tests: %v", err)
	}

	if _, err := getLightdashProjectUuid(); err != nil {
		t.Fatalf("LIGHTDASH_PROJECT must be set for acceptance tests: %v", err)
	}
}
