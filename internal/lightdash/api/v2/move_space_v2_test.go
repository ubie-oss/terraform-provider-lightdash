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
	"reflect"
	"testing"
)

func TestMoveSpaceV2Request_UnmarshalJSON(t *testing.T) {
	parentID := "parent-space-uuid"

	type itemStruct struct {
		UUID        string `json:"uuid"`
		Type        string `json:"type"`
		ContentType string `json:"contentType"`
	}

	type actionStruct struct {
		Type            string  `json:"type"`
		TargetSpaceUUID *string `json:"targetSpaceUuid"`
	}

	tests := []struct {
		name        string
		jsonStr     string
		expected    MoveSpaceV2Request
		expectError bool
	}{
		{
			name: "Valid JSON with targetSpaceUuid",
			jsonStr: `{
				"item": {
					"uuid": "space-uuid",
					"type": "space",
					"contentType": "space"
				},
				"action": {
					"type": "move",
					"targetSpaceUuid": "parent-space-uuid"
				}
			}`,
			expected: MoveSpaceV2Request{
				Item: itemStruct{
					UUID:        "space-uuid",
					Type:        "space",
					ContentType: "space",
				},
				Action: actionStruct{
					Type:            "move",
					TargetSpaceUUID: &parentID,
				},
			},
			expectError: false,
		},
		{
			name: "Valid JSON with targetSpaceUuid as null",
			jsonStr: `{
				"item": {
					"uuid": "space-uuid",
					"type": "space",
					"contentType": "space"
				},
				"action": {
					"type": "move",
					"targetSpaceUuid": null
				}
			}`,
			expected: MoveSpaceV2Request{
				Item: itemStruct{
					UUID:        "space-uuid",
					Type:        "space",
					ContentType: "space",
				},
				Action: actionStruct{
					Type:            "move",
					TargetSpaceUUID: nil,
				},
			},
			expectError: false,
		},
		{
			name: "Valid JSON with targetSpaceUuid omitted",
			jsonStr: `{
				"item": {
					"uuid": "space-uuid",
					"type": "space",
					"contentType": "space"
				},
				"action": {
					"type": "move"
				}
			}`,
			expected: MoveSpaceV2Request{
				Item: itemStruct{
					UUID:        "space-uuid",
					Type:        "space",
					ContentType: "space",
				},
				Action: actionStruct{
					Type:            "move",
					TargetSpaceUUID: nil,
				},
			},
			expectError: false,
		},
		{
			name: "Missing 'item' object",
			jsonStr: `{
				"action": {
					"type": "move",
					"targetSpaceUuid": "parent-space-uuid"
				}
			}`,
			expected: MoveSpaceV2Request{
				Item: itemStruct{}, // Expected zero value itemStruct
				Action: actionStruct{
					Type:            "move",
					TargetSpaceUUID: &parentID,
				},
			},
			expectError: false, // json.Unmarshal does not return error for missing objects
		},
		{
			name: "Missing 'action' object",
			jsonStr: `{
				"item": {
					"uuid": "space-uuid",
					"type": "space",
					"contentType": "space"
				}
			}`,
			expected: MoveSpaceV2Request{
				Item: itemStruct{
					UUID:        "space-uuid",
					Type:        "space",
					ContentType: "space",
				},
				Action: actionStruct{}, // Expected zero value actionStruct
			},
			expectError: false, // json.Unmarshal does not return error for missing objects
		},
		{
			name: "Missing 'uuid' in item",
			jsonStr: `{
				"item": {
					"type": "space",
					"contentType": "space"
				},
				"action": {
					"type": "move"
				}
			}`,
			expected: MoveSpaceV2Request{
				Item: itemStruct{
					UUID:        "", // Zero value for string
					Type:        "space",
					ContentType: "space",
				},
				Action: actionStruct{
					Type:            "move",
					TargetSpaceUUID: nil,
				},
			},
			expectError: false,
		},
		{
			name: "Invalid JSON format",
			jsonStr: `{
				"item": {
					"uuid": "space-uuid",
					"type": "space",
					"contentType": "space"
				},
				"action": {
					"type": "move"
				}
			`, // Missing closing brace
			expected:    MoveSpaceV2Request{}, // Expected zero value on error
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result MoveSpaceV2Request
			err := json.Unmarshal([]byte(test.jsonStr), &result)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}

				// Use reflect.DeepEqual for comprehensive comparison, including nil pointer for TargetSpaceUUID
				if !reflect.DeepEqual(result, test.expected) {
					t.Errorf("Unmarshal mismatch:\nExpected: %+v\nGot:      %+v", test.expected, result)
				}
			}
		})
	}
}
