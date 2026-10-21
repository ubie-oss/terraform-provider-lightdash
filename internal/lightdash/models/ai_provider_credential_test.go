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

package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAiProviderCredentialsList_Unmarshal_fixture(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "api", "v1", "testdata", "list_ai_provider_credentials.json")
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var list AiProviderCredentialsList
	if err := json.Unmarshal(payload, &list); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(list.Credentials) != 1 {
		t.Fatalf("credentials len = %d, want 1", len(list.Credentials))
	}
	cred := list.Credentials[0]
	if cred.UUID != "12345678-1234-4123-8123-123456789abc" {
		t.Fatalf("uuid = %q", cred.UUID)
	}
	if cred.Provider != AiProviderBedrock {
		t.Fatalf("provider = %q", cred.Provider)
	}
	if !cred.IsDefault {
		t.Fatal("expected isDefault true")
	}
	if cred.InferenceGeography != BedrockInferenceGeographyUS {
		t.Fatalf("inferenceGeography = %q", cred.InferenceGeography)
	}
	if list.LegacyBedrock == nil || list.LegacyBedrock.Region != "us-west-2" {
		t.Fatal("expected legacyBedrock region us-west-2")
	}
}
