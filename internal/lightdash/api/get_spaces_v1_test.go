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

func TestListSpacesInProjectV1Results(t *testing.T) {
	parentID := "parent-space-uuid"
	tests := []struct {
		name     string
		jsonStr  string
		expected ListSpacesInProjectV1Results
	}{
		{
			name: "Test with valid data (no parentSpaceUuid)",
			jsonStr: `{
				"organizationUuid": "org-uuid",
				"projectUuid": "proj-uuid",
				"uuid": "space-uuid",
				"name": "Test Space",
				"isPrivate": true
			}`,
			expected: ListSpacesInProjectV1Results{
				OrganizationUUID: "org-uuid",
				ProjectUUID:      "proj-uuid",
				ParentSpaceUUID:  nil,
				SpaceUUID:        "space-uuid",
				SpaceName:        "Test Space",
				IsPrivate:        true,
			},
		},
		{
			name: "Test with practical empty data (no parentSpaceUuid)",
			jsonStr: `{
				"organizationUuid": "org-empty-uuid",
				"projectUuid": "proj-empty-uuid",
				"uuid": "space-empty-uuid",
				"name": "Empty Space",
				"isPrivate": false
			}`,
			expected: ListSpacesInProjectV1Results{
				OrganizationUUID: "org-empty-uuid",
				ProjectUUID:      "proj-empty-uuid",
				ParentSpaceUUID:  nil,
				SpaceUUID:        "space-empty-uuid",
				SpaceName:        "Empty Space",
				IsPrivate:        false,
			},
		},
		{
			name: "Test with parentSpaceUuid present",
			jsonStr: `{
				"organizationUuid": "org-parent-uuid",
				"projectUuid": "proj-parent-uuid",
				"parentSpaceUuid": "parent-space-uuid",
				"uuid": "space-child-uuid",
				"name": "Child Space",
				"isPrivate": false
			}`,
			expected: ListSpacesInProjectV1Results{
				OrganizationUUID: "org-parent-uuid",
				ProjectUUID:      "proj-parent-uuid",
				ParentSpaceUUID:  &parentID,
				SpaceUUID:        "space-child-uuid",
				SpaceName:        "Child Space",
				IsPrivate:        false,
			},
		},
		{
			name: "Test with parentSpaceUuid explicitly null",
			jsonStr: `{
				"organizationUuid": "org-null-uuid",
				"projectUuid": "proj-null-uuid",
				"parentSpaceUuid": null,
				"uuid": "space-null-uuid",
				"name": "Null Parent Space",
				"isPrivate": true
			}`,
			expected: ListSpacesInProjectV1Results{
				OrganizationUUID: "org-null-uuid",
				ProjectUUID:      "proj-null-uuid",
				ParentSpaceUUID:  nil,
				SpaceUUID:        "space-null-uuid",
				SpaceName:        "Null Parent Space",
				IsPrivate:        true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result ListSpacesInProjectV1Results
			err := json.Unmarshal([]byte(test.jsonStr), &result)
			if err != nil {
				t.Errorf("json.Unmarshal failed for %s with error: %v", test.name, err)
			}
			if test.expected.ParentSpaceUUID != nil && result.ParentSpaceUUID != nil {
				// Compare the values pointed to by ParentSpaceUUID
				if *test.expected.ParentSpaceUUID != *result.ParentSpaceUUID {
					t.Errorf("ParentSpaceUUID mismatch for %s: expected %v, got %v", test.name, *test.expected.ParentSpaceUUID, *result.ParentSpaceUUID)
				}
				// Set to nil for DeepEqual comparison of the rest
				result.ParentSpaceUUID = nil
				test.expected.ParentSpaceUUID = nil
			} else if test.expected.ParentSpaceUUID != result.ParentSpaceUUID {
				t.Errorf("ParentSpaceUUID nil mismatch for %s: expected %v, got %v", test.name, test.expected.ParentSpaceUUID, result.ParentSpaceUUID)
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("TestListSpacesInProjectV1Results failed for %s, expected %+v, got %+v", test.name, test.expected, result)
			}
		})
	}
}
