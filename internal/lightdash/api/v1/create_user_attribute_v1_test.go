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
	"testing"
)

func TestCreateUserAttributeV1RequestJSON(t *testing.T) {
	defaults := []string{"jp"}
	request := CreateUserAttributeV1Request{
		Name:              "region",
		Description:       "Sales region",
		AttributeDefaults: &defaults,
		Groups: []CreateUserAttributeGroupV1{
			{GroupUUID: "group-1", Values: []string{"us"}},
		},
		Users: []CreateUserAttributeUserV1{},
	}

	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"name", "description", "attributeDefaults", "groups", "users"} {
		if _, ok := body[key]; !ok {
			t.Errorf("missing %q in %s", key, raw)
		}
	}
	users, ok := body["users"].([]any)
	if !ok || len(users) != 0 {
		t.Fatalf("expected an empty users array, got %#v", body["users"])
	}
}

func TestCreateUserAttributeV1RequestJSONNullDefaults(t *testing.T) {
	request := CreateUserAttributeV1Request{
		Name:   "region",
		Groups: []CreateUserAttributeGroupV1{},
		Users:  []CreateUserAttributeUserV1{},
	}
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["attributeDefaults"] != nil {
		t.Fatalf("expected null attributeDefaults, got %#v", body["attributeDefaults"])
	}
}
