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

func TestExtractStrings(t *testing.T) {
	tests := []struct {
		input    string
		pattern  string
		expected []string
		wantErr  bool
	}{
		{
			input:    "projects/abc-123/spaces/xyz-234",
			pattern:  `^projects/([^/]+)/spaces/([^/]+)$`,
			expected: []string{"abc-123", "xyz-234"},
			wantErr:  false,
		},
		{
			input:    "projects/asdfad-234234",
			pattern:  `^projects/([^/]+)$`,
			expected: []string{"asdfad-234234"},
			wantErr:  false,
		},
		{
			input:    "projects/kdjfa-zfadf/users/werw-xvx",
			pattern:  `^projects/([^/]+)/users/([^/]+)$`,
			expected: []string{"kdjfa-zfadf", "werw-xvx"},
			wantErr:  false,
		},
		{
			input:    "projects/invalid_input",
			pattern:  `^projects/([^/]+)/spaces/([^/]+)$`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		output, err := extractStrings(test.input, test.pattern)

		if (err != nil) != test.wantErr {
			t.Errorf("Input: %s, Pattern: %s, Expected error: %v, Got error: %v", test.input, test.pattern, test.wantErr, err)
		}

		if !reflect.DeepEqual(output, test.expected) {
			t.Errorf("Input: %s, Pattern: %s, Expected: %v, Got: %v", test.input, test.pattern, test.expected, output)
		}
	}
}
