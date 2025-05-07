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

package api

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestUpdateSpaceV1Request_MarshalJSON(t *testing.T) {
	parentUUID := "parent-uuid-456"
	cases := []struct {
		name     string
		input    UpdateSpaceV1Request
		expected map[string]interface{}
	}{
		{
			name: "All fields set",
			input: UpdateSpaceV1Request{
				Name:            "Updated Space",
				IsPrivate:       true,
				ParentSpaceUUID: &parentUUID,
			},
			expected: map[string]interface{}{
				"name":            "Updated Space",
				"isPrivate":       true,
				"parentSpaceUuid": "parent-uuid-456",
			},
		},
		{
			name: "ParentSpaceUUID nil",
			input: UpdateSpaceV1Request{
				Name:            "No Parent Space",
				IsPrivate:       false,
				ParentSpaceUUID: nil,
			},
			expected: map[string]interface{}{
				"name":      "No Parent Space",
				"isPrivate": false,
			},
		},
		{
			name: "Empty name, isPrivate true, ParentSpaceUUID nil",
			input: UpdateSpaceV1Request{
				Name:            "",
				IsPrivate:       true,
				ParentSpaceUUID: nil,
			},
			expected: map[string]interface{}{
				"name":      "",
				"isPrivate": true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("failed to marshal UpdateSpaceV1Request: %v", err)
			}
			var got map[string]interface{}
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatalf("failed to unmarshal marshaled JSON: %v", err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("unexpected marshaled JSON.\nGot:      %v\nExpected: %v", got, tc.expected)
			}
		})
	}
}

func TestUpdateSpaceV1Request_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected UpdateSpaceV1Request
	}{
		{
			name:  "With ParentSpaceUUID",
			input: `{"name":"Space X","isPrivate":true,"parentSpaceUuid":"parent-uuid-xyz"}`,
			expected: func() UpdateSpaceV1Request {
				uuid := "parent-uuid-xyz"
				return UpdateSpaceV1Request{
					Name:            "Space X",
					IsPrivate:       true,
					ParentSpaceUUID: &uuid,
				}
			}(),
		},
		{
			name:  "Without ParentSpaceUUID",
			input: `{"name":"Space Y","isPrivate":false}`,
			expected: UpdateSpaceV1Request{
				Name:            "Space Y",
				IsPrivate:       false,
				ParentSpaceUUID: nil,
			},
		},
		{
			name:  "Empty name, isPrivate true, no parent",
			input: `{"name":"","isPrivate":true}`,
			expected: UpdateSpaceV1Request{
				Name:            "",
				IsPrivate:       true,
				ParentSpaceUUID: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req UpdateSpaceV1Request
			err := json.Unmarshal([]byte(tc.input), &req)
			if err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if req.Name != tc.expected.Name {
				t.Errorf("expected Name '%s', got '%s'", tc.expected.Name, req.Name)
			}
			if req.IsPrivate != tc.expected.IsPrivate {
				t.Errorf("expected IsPrivate %v, got %v", tc.expected.IsPrivate, req.IsPrivate)
			}
			if tc.expected.ParentSpaceUUID == nil {
				if req.ParentSpaceUUID != nil {
					t.Errorf("expected ParentSpaceUUID nil, got '%v'", req.ParentSpaceUUID)
				}
			} else {
				if req.ParentSpaceUUID == nil || *req.ParentSpaceUUID != *tc.expected.ParentSpaceUUID {
					t.Errorf("expected ParentSpaceUUID '%v', got '%v'", *tc.expected.ParentSpaceUUID, req.ParentSpaceUUID)
				}
			}
		})
	}
}

func TestUpdateSpaceV1Results_MarshalUnmarshalJSON(t *testing.T) {
	parentUUID := "parent-uuid-abc"
	marshalCases := []struct {
		name     string
		input    UpdateSpaceV1Results
		expected map[string]interface{}
	}{
		{
			name: "All fields set",
			input: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-10",
				ProjectUUID:      "proj-uuid-10",
				ParentSpaceUUID:  &parentUUID,
				SpaceUUID:        "space-uuid-10",
				SpaceName:        "Updated Space 10",
				IsPrivate:        false,
			},
			expected: map[string]interface{}{
				"organizationUuid": "org-uuid-10",
				"projectUuid":      "proj-uuid-10",
				"parentSpaceUuid":  "parent-uuid-abc",
				"uuid":             "space-uuid-10",
				"name":             "Updated Space 10",
				"isPrivate":        false,
			},
		},
		{
			name: "No ParentSpaceUUID",
			input: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-11",
				ProjectUUID:      "proj-uuid-11",
				ParentSpaceUUID:  nil,
				SpaceUUID:        "space-uuid-11",
				SpaceName:        "Updated Space 11",
				IsPrivate:        true,
			},
			expected: map[string]interface{}{
				"organizationUuid": "org-uuid-11",
				"projectUuid":      "proj-uuid-11",
				"uuid":             "space-uuid-11",
				"name":             "Updated Space 11",
				"isPrivate":        true,
			},
		},
	}

	for _, tc := range marshalCases {
		t.Run("Marshal/"+tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("failed to marshal UpdateSpaceV1Results: %v", err)
			}
			var got map[string]interface{}
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatalf("failed to unmarshal marshaled JSON: %v", err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("unexpected marshaled JSON.\nGot:      %v\nExpected: %v", got, tc.expected)
			}
		})
	}

	unmarshalCases := []struct {
		name     string
		input    string
		expected UpdateSpaceV1Results
	}{
		{
			name: "All fields set",
			input: `{
				"organizationUuid": "org-uuid-20",
				"projectUuid": "proj-uuid-20",
				"parentSpaceUuid": "parent-uuid-20",
				"uuid": "space-uuid-20",
				"name": "Updated Space 20",
				"isPrivate": true
			}`,
			expected: func() UpdateSpaceV1Results {
				uuid := "parent-uuid-20"
				return UpdateSpaceV1Results{
					OrganizationUUID: "org-uuid-20",
					ProjectUUID:      "proj-uuid-20",
					ParentSpaceUUID:  &uuid,
					SpaceUUID:        "space-uuid-20",
					SpaceName:        "Updated Space 20",
					IsPrivate:        true,
				}
			}(),
		},
		{
			name: "Without ParentSpaceUUID",
			input: `{
				"organizationUuid": "org-uuid-21",
				"projectUuid": "proj-uuid-21",
				"uuid": "space-uuid-21",
				"name": "Updated Space 21",
				"isPrivate": false
			}`,
			expected: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-21",
				ProjectUUID:      "proj-uuid-21",
				ParentSpaceUUID:  nil,
				SpaceUUID:        "space-uuid-21",
				SpaceName:        "Updated Space 21",
				IsPrivate:        false,
			},
		},
	}

	for _, tc := range unmarshalCases {
		t.Run("Unmarshal/"+tc.name, func(t *testing.T) {
			var res UpdateSpaceV1Results
			err := json.Unmarshal([]byte(tc.input), &res)
			if err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if res.OrganizationUUID != tc.expected.OrganizationUUID {
				t.Errorf("expected OrganizationUUID '%s', got '%s'", tc.expected.OrganizationUUID, res.OrganizationUUID)
			}
			if res.ProjectUUID != tc.expected.ProjectUUID {
				t.Errorf("expected ProjectUUID '%s', got '%s'", tc.expected.ProjectUUID, res.ProjectUUID)
			}
			if tc.expected.ParentSpaceUUID == nil {
				if res.ParentSpaceUUID != nil {
					t.Errorf("expected ParentSpaceUUID nil, got '%v'", res.ParentSpaceUUID)
				}
			} else {
				if res.ParentSpaceUUID == nil || *res.ParentSpaceUUID != *tc.expected.ParentSpaceUUID {
					t.Errorf("expected ParentSpaceUUID '%v', got '%v'", *tc.expected.ParentSpaceUUID, res.ParentSpaceUUID)
				}
			}
			if res.SpaceUUID != tc.expected.SpaceUUID {
				t.Errorf("expected SpaceUUID '%s', got '%s'", tc.expected.SpaceUUID, res.SpaceUUID)
			}
			if res.SpaceName != tc.expected.SpaceName {
				t.Errorf("expected SpaceName '%s', got '%s'", tc.expected.SpaceName, res.SpaceName)
			}
			if res.IsPrivate != tc.expected.IsPrivate {
				t.Errorf("expected IsPrivate %v, got %v", tc.expected.IsPrivate, res.IsPrivate)
			}
		})
	}
}

func TestUpdateSpaceV1Results_FieldAssignment(t *testing.T) {
	parentUUID := "parent-uuid-abc"
	cases := []struct {
		name     string
		input    UpdateSpaceV1Results
		expected UpdateSpaceV1Results
	}{
		{
			name: "All fields set",
			input: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-30",
				ProjectUUID:      "proj-uuid-30",
				ParentSpaceUUID:  &parentUUID,
				SpaceUUID:        "space-uuid-30",
				SpaceName:        "Updated Space 30",
				IsPrivate:        false,
			},
			expected: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-30",
				ProjectUUID:      "proj-uuid-30",
				ParentSpaceUUID:  &parentUUID,
				SpaceUUID:        "space-uuid-30",
				SpaceName:        "Updated Space 30",
				IsPrivate:        false,
			},
		},
		{
			name: "No ParentSpaceUUID",
			input: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-31",
				ProjectUUID:      "proj-uuid-31",
				ParentSpaceUUID:  nil,
				SpaceUUID:        "space-uuid-31",
				SpaceName:        "Updated Space 31",
				IsPrivate:        true,
			},
			expected: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-31",
				ProjectUUID:      "proj-uuid-31",
				ParentSpaceUUID:  nil,
				SpaceUUID:        "space-uuid-31",
				SpaceName:        "Updated Space 31",
				IsPrivate:        true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			results := tc.input
			if results.OrganizationUUID != tc.expected.OrganizationUUID {
				t.Errorf("expected OrganizationUUID '%s', got '%s'", tc.expected.OrganizationUUID, results.OrganizationUUID)
			}
			if results.ProjectUUID != tc.expected.ProjectUUID {
				t.Errorf("expected ProjectUUID '%s', got '%s'", tc.expected.ProjectUUID, results.ProjectUUID)
			}
			if tc.expected.ParentSpaceUUID == nil {
				if results.ParentSpaceUUID != nil {
					t.Errorf("expected ParentSpaceUUID nil, got '%v'", results.ParentSpaceUUID)
				}
			} else {
				if results.ParentSpaceUUID == nil || *results.ParentSpaceUUID != *tc.expected.ParentSpaceUUID {
					t.Errorf("expected ParentSpaceUUID '%v', got '%v'", *tc.expected.ParentSpaceUUID, results.ParentSpaceUUID)
				}
			}
			if results.SpaceUUID != tc.expected.SpaceUUID {
				t.Errorf("expected SpaceUUID '%s', got '%s'", tc.expected.SpaceUUID, results.SpaceUUID)
			}
			if results.SpaceName != tc.expected.SpaceName {
				t.Errorf("expected SpaceName '%s', got '%s'", tc.expected.SpaceName, results.SpaceName)
			}
			if results.IsPrivate != tc.expected.IsPrivate {
				t.Errorf("expected IsPrivate %v, got %v", tc.expected.IsPrivate, results.IsPrivate)
			}
		})
	}
}
