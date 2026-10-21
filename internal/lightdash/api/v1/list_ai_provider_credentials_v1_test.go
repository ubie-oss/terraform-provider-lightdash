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

package v1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestListAiProviderCredentialsV1Response_Unmarshal_fixture(t *testing.T) {
	t.Parallel()

	inner, err := os.ReadFile(filepath.Join("testdata", "list_ai_provider_credentials.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	payload := []byte(`{"status":"ok","results":`)
	payload = append(payload, inner...)
	payload = append(payload, '}')

	var response ListAiProviderCredentialsV1Response
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if response.Status != "ok" {
		t.Fatalf("status = %q", response.Status)
	}
	if len(response.Results.Credentials) != 1 {
		t.Fatalf("credentials len = %d", len(response.Results.Credentials))
	}
}

func TestCreateAiProviderCredentialV1Response_Unmarshal(t *testing.T) {
	t.Parallel()

	const payload = `{"status":"ok","results":{"uuid":"12345678-1234-4123-8123-123456789abc"}}`
	var response CreateAiProviderCredentialV1Response
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if response.Results.UUID != "12345678-1234-4123-8123-123456789abc" {
		t.Fatalf("uuid = %q", response.Results.UUID)
	}
}

func TestProjectAiCredentialSelection_JSON_nullCredential(t *testing.T) {
	t.Parallel()

	const payload = `{"credentialUuid":null}`
	var response GetProjectAiProviderCredentialV1Response
	if err := json.Unmarshal([]byte(`{"status":"ok","results":`+payload+`}`), &response); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if response.Results.CredentialUUID != nil {
		t.Fatalf("expected null credentialUuid, got %v", *response.Results.CredentialUUID)
	}
}
