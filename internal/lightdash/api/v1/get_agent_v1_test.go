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

func TestGetAgentV1Results_UnmarshalJSON(t *testing.T) {
	jsonStr := `{
		"uuid": "c389d0b6-418c-4d2b-9784-3c76474ed28d",
		"organizationUuid": "089a18c4-667e-41cb-9d10-b088461ac941",
		"projectUuid": "f58b2903-de95-4bcc-8a11-194a35f31f15",
		"name": "AI Analyst [removed until 2025-07-31]",
		"tags": ["test-lightdash-ai"],
		"integrations": [],
		"updatedAt": "2025-09-30T04:37:54.806Z",
		"createdAt": "2025-07-07T02:46:02.041Z",
		"instruction": "You are an expert AI data analyst...",
		"imageUrl": null,
		"enableDataAccess": false,
		"groupAccess": ["223eef74-67cc-4ffd-a757-0803b823503c"],
		"userAccess": []
	}`

	var results GetAgentV1Results
	err := json.Unmarshal([]byte(jsonStr), &results)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if results.UUID != "c389d0b6-418c-4d2b-9784-3c76474ed28d" {
		t.Errorf("expected UUID to be 'c389d0b6-418c-4d2b-9784-3c76474ed28d', got '%s'", results.UUID)
	}
	if results.OrganizationUUID != "089a18c4-667e-41cb-9d10-b088461ac941" {
		t.Errorf("expected OrganizationUUID to be '089a18c4-667e-41cb-9d10-b088461ac941', got '%s'", results.OrganizationUUID)
	}
	if results.ProjectUUID != "f58b2903-de95-4bcc-8a11-194a35f31f15" {
		t.Errorf("expected ProjectUUID to be 'f58b2903-de95-4bcc-8a11-194a35f31f15', got '%s'", results.ProjectUUID)
	}
	if results.Name != "AI Analyst [removed until 2025-07-31]" {
		t.Errorf("expected Name to be 'AI Analyst [removed until 2025-07-31]', got '%s'", results.Name)
	}
	if !reflect.DeepEqual(results.Tags, []string{"test-lightdash-ai"}) {
		t.Errorf("expected Tags to be ['test-lightdash-ai'], got %v", results.Tags)
	}
	if results.EnableDataAccess != false {
		t.Errorf("expected EnableDataAccess to be false, got %v", results.EnableDataAccess)
	}
	if !reflect.DeepEqual(results.GroupAccess, []string{"223eef74-67cc-4ffd-a757-0803b823503c"}) {
		t.Errorf("expected GroupAccess to be ['223eef74-67cc-4ffd-a757-0803b823503c'], got %v", results.GroupAccess)
	}
	if len(results.UserAccess) != 0 {
		t.Errorf("expected UserAccess to be empty, got %v", results.UserAccess)
	}
}

func TestGetAgentV1Response_UnmarshalJSON(t *testing.T) {
	jsonStr := `{
		"status": "ok",
		"results": {
			"uuid": "c389d0b6-418c-4d2b-9784-3c76474ed28d",
			"organizationUuid": "089a18c4-667e-41cb-9d10-b088461ac941",
			"projectUuid": "f58b2903-de95-4bcc-8a11-194a35f31f15",
			"name": "AI Analyst",
			"tags": [],
			"integrations": [],
			"updatedAt": "2025-09-30T04:37:54.806Z",
			"createdAt": "2025-07-07T02:46:02.041Z",
			"instruction": null,
			"imageUrl": null,
			"enableDataAccess": true,
			"groupAccess": [],
			"userAccess": []
		}
	}`

	var response GetAgentV1Response
	err := json.Unmarshal([]byte(jsonStr), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected Status to be 'ok', got '%s'", response.Status)
	}
	if response.Results.UUID != "c389d0b6-418c-4d2b-9784-3c76474ed28d" {
		t.Errorf("expected Results.UUID to be 'c389d0b6-418c-4d2b-9784-3c76474ed28d', got '%s'", response.Results.UUID)
	}
	if response.Results.EnableDataAccess != true {
		t.Errorf("expected Results.EnableDataAccess to be true, got %v", response.Results.EnableDataAccess)
	}
}
