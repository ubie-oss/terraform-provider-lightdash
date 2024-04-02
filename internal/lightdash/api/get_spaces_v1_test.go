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
	tests := []struct {
		name     string
		jsonStr  string
		expected ListSpacesInProjectV1Results
	}{
		{
			name: "Test with valid data",
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
				SpaceUUID:        "space-uuid",
				SpaceName:        "Test Space",
				IsPrivate:        true,
			},
		},
		{
			name: "Test with practical empty data",
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
				SpaceUUID:        "space-empty-uuid",
				SpaceName:        "Empty Space",
				IsPrivate:        false,
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
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("TestListSpacesInProjectV1Results failed for %s, expected %+v, got %+v", test.name, test.expected, result)
			}
		})
	}
}
