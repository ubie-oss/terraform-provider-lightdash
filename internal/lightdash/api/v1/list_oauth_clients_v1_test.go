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
	"reflect"
	"testing"
)

func TestListOAuthClientsV1Response_UnmarshalJSON(t *testing.T) {
	jsonStr := `{
		"status": "ok",
		"results": [
			{
				"clientId": "oauth-V1StGXR8_Z5jdHi6",
				"clientName": "My MCP Server",
				"redirectUris": [
					"https://myapp.example.com/oauth/callback",
					"cursor://anysphere.cursor-mcp/oauth/callback"
				],
				"scopes": [],
				"createdAt": "2026-06-15T10:30:00.000Z",
				"createdByUserUuid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
			}
		]
	}`

	var response ListOAuthClientsV1Response
	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected Status to be 'ok', got '%s'", response.Status)
	}
	if len(response.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(response.Results))
	}

	client := response.Results[0]
	if client.ClientID != "oauth-V1StGXR8_Z5jdHi6" {
		t.Errorf("expected ClientID oauth-V1StGXR8_Z5jdHi6, got %s", client.ClientID)
	}
	if client.ClientName != "My MCP Server" {
		t.Errorf("expected ClientName My MCP Server, got %s", client.ClientName)
	}
	expectedURIs := []string{
		"https://myapp.example.com/oauth/callback",
		"cursor://anysphere.cursor-mcp/oauth/callback",
	}
	if !reflect.DeepEqual(client.RedirectURIs, expectedURIs) {
		t.Errorf("expected RedirectURIs %v, got %v", expectedURIs, client.RedirectURIs)
	}
	if client.CreatedByUserUUID == nil || *client.CreatedByUserUUID != "a1b2c3d4-e5f6-7890-abcd-ef1234567890" {
		t.Errorf("unexpected CreatedByUserUUID: %v", client.CreatedByUserUUID)
	}
}

func TestCreateOAuthClientV1Response_UnmarshalJSON(t *testing.T) {
	testClientSecret := "fixture-" + "oauth-client-secret-for-unmarshal-test"

	payload := map[string]any{
		"status": "ok",
		"results": map[string]any{
			"clientId":          "oauth-V1StGXR8_Z5jdHi6",
			"clientName":        "My MCP Server",
			"redirectUris":      []string{"https://myapp.example.com/oauth/callback"},
			"scopes":            []string{},
			"createdAt":         "2026-06-15T10:30:00.000Z",
			"createdByUserUuid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			"clientSecret":      testClientSecret,
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal test payload: %v", err)
	}

	var response CreateOAuthClientV1Response
	if err := json.Unmarshal(jsonBytes, &response); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if response.Results.ClientSecret != testClientSecret {
		t.Errorf("expected clientSecret to be set, got %q", response.Results.ClientSecret)
	}
}
