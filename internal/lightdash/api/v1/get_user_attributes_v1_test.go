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

func TestUserAttributeV1JSON(t *testing.T) {
	raw := `{
		"uuid": "attr-1",
		"organizationUuid": "org-1",
		"name": "region",
		"description": "Sales region",
		"attributeDefault": "jp",
		"attributeDefaults": ["jp", "us"],
		"users": [{"userUuid": "user-1", "email": "a@example.com", "value": "eu", "values": ["eu", "uk"]}],
		"groups": [{"groupUuid": "group-1", "value": "us", "values": ["us"]}],
		"createdAt": "2024-01-01T00:00:00.000Z"
	}`

	var got UserAttributeV1
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.UUID != "attr-1" || got.OrganizationUUID != "org-1" || got.Name != "region" || got.Description != "Sales region" {
		t.Fatalf("unexpected attribute: %+v", got)
	}
	if got.AttributeDefault == nil || *got.AttributeDefault != "jp" {
		t.Fatalf("unexpected attributeDefault: %+v", got.AttributeDefault)
	}
	if !reflect.DeepEqual(got.AttributeDefaults, []string{"jp", "us"}) {
		t.Fatalf("unexpected defaults: %+v", got.AttributeDefaults)
	}
	if len(got.Users) != 1 || got.Users[0].UserUUID != "user-1" || !reflect.DeepEqual(got.Users[0].Values, []string{"eu", "uk"}) {
		t.Fatalf("unexpected users: %+v", got.Users)
	}
	if len(got.Groups) != 1 || got.Groups[0].GroupUUID != "group-1" || !reflect.DeepEqual(got.Groups[0].Values, []string{"us"}) {
		t.Fatalf("unexpected groups: %+v", got.Groups)
	}
}

func TestUserAttributeV1JSONNullDefaults(t *testing.T) {
	raw := `{
		"uuid": "attr-1",
		"organizationUuid": "org-1",
		"name": "region",
		"description": "",
		"attributeDefault": null,
		"attributeDefaults": null,
		"users": [],
		"groups": [],
		"createdAt": "2024-01-01T00:00:00.000Z"
	}`
	var got UserAttributeV1
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.AttributeDefault != nil || got.AttributeDefaults != nil {
		t.Fatalf("expected null defaults, got %+v %+v", got.AttributeDefault, got.AttributeDefaults)
	}
	if got.Users == nil || got.Groups == nil {
		t.Fatalf("expected empty arrays, got users=%v groups=%v", got.Users, got.Groups)
	}
}
