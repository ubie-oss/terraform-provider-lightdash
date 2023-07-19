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

package provider

import (
	"reflect"
	"testing"
)

func TestExtractSpaceResourceId(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
		wantErr  bool
	}{
		{
			input:    "projects/xyz-567/spaces/abc-123",
			expected: []string{"xyz-567", "abc-123"},
			wantErr:  false,
		},
		{
			input:    "projects/123/spaces/456",
			expected: []string{"123", "456"},
			wantErr:  false,
		},
		{
			input:    "projects/xyz/spaces/",
			expected: nil,
			wantErr:  true,
		},
		{
			input:    "projects/xyz/spaces/abc/def",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		info, err := extractSpaceResourceId(test.input)

		if (err != nil) != test.wantErr {
			t.Errorf("Input: %s, Expected error: %v, Got error: %v", test.input, test.wantErr, err)
		}

		if !reflect.DeepEqual(info, test.expected) {
			t.Errorf("Input: %s, Expected: %v, Got: %v", test.input, test.expected, info)
		}
	}
}
