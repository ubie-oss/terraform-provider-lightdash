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

package v2

import (
	"encoding/json"
	"testing"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

func TestGetOrganizationRolesV2Response_UnmarshalJSON(t *testing.T) {
	const fixture = `{
		"status": "ok",
		"results": [
			{
				"roleUuid": "viewer",
				"name": "Viewer",
				"description": "Viewer",
				"ownerType": "system",
				"scopes": ["view:Dashboard", "view:Project"],
				"organizationUuid": null,
				"createdAt": null,
				"updatedAt": null,
				"createdBy": null
			},
			{
				"roleUuid": "editor",
				"name": "Editor",
				"description": "Editor",
				"ownerType": "system",
				"organizationUuid": null,
				"createdAt": null,
				"updatedAt": null,
				"createdBy": null
			}
		]
	}`

	var response getOrganizationRolesV2Response
	if err := json.Unmarshal([]byte(fixture), &response); err != nil {
		t.Fatalf("unmarshal getOrganizationRolesV2Response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Status = %q, want ok", response.Status)
	}
	if len(response.Results) != 2 {
		t.Fatalf("Results len = %d, want 2", len(response.Results))
	}
	if response.Results[0].RoleUUID != "viewer" {
		t.Errorf("Results[0].RoleUUID = %q, want viewer", response.Results[0].RoleUUID)
	}
	if response.Results[1].Name != "Editor" {
		t.Errorf("Results[1].Name = %q, want Editor", response.Results[1].Name)
	}
}

func TestGetOrganizationRolesV2Response_UnmarshalJSON_invalid(t *testing.T) {
	var response getOrganizationRolesV2Response
	if err := json.Unmarshal([]byte(`{`), &response); err == nil {
		t.Fatal("expected unmarshal error for invalid JSON")
	}
}

// Ensure models.Role is used in response type.
var _ = models.Role{}
