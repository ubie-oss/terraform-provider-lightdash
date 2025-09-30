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
	"context"
	"os"
	"reflect"
	"testing"
)

func TestIsIntegrationTestMode(t *testing.T) {
	original_value := os.Getenv(integrationTestModeEnvVar)

	t.Setenv(integrationTestModeEnvVar, "1")
	if !isIntegrationTestMode() {
		t.Errorf("Expected: true, Got: %t", isIntegrationTestMode())
	}
	t.Setenv(integrationTestModeEnvVar, "0")
	if isIntegrationTestMode() {
		t.Errorf("Expected: false, Got: %t", isIntegrationTestMode())
	}

	t.Setenv(integrationTestModeEnvVar, original_value)
}

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

func TestSubtractStringList(t *testing.T) {
	tests := []struct {
		list1 []string
		list2 []string

		expected []string
	}{
		{
			list1:    []string{"a", "b", "c"},
			list2:    []string{"b", "c"},
			expected: []string{"a"},
		},
		{
			list1:    []string{"a", "b", "c"},
			list2:    []string{"d", "e"},
			expected: []string{"a", "b", "c"},
		},
		{
			list1:    []string{"a", "b", "c"},
			list2:    []string{"a", "b", "c"},
			expected: []string{},
		},
		{
			list1:    []string{},
			list2:    []string{"a", "b"},
			expected: []string{},
		},
		{
			list1:    []string{"a", "b"},
			list2:    []string{},
			expected: []string{"a", "b"},
		},
		{
			list1:    []string{},
			list2:    []string{},
			expected: []string{},
		},
		{
			list1:    []string{"a", "a", "b", "c"},
			list2:    []string{"a", "c"},
			expected: []string{"a", "b"},
		},
		{
			list1:    []string{"a", "b", "c"},
			list2:    []string{"b", "b", "c"},
			expected: []string{"a"},
		},
		{
			list1:    []string{"a", "a", "b", "b", "c", "c"},
			list2:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			list1:    []string{"c", "a", "b"},
			list2:    []string{"b", "c"},
			expected: []string{"a"},
		},
	}

	for _, test := range tests {
		result := subtractStringList(test.list1, test.list2)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected: %v, Got: %v for list1=%v, list2=%v", test.expected, result, test.list1, test.list2)
		}
	}
}

func TestReadMarkdownDescriptionEmbedded(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "resource with internal/provider prefix",
			filename: "internal/provider/docs/resources/resource_lightdash_group.md",
			wantErr:  false,
		},
		{
			name:     "resource with docs prefix",
			filename: "docs/resources/resource_space.md",
			wantErr:  false,
		},
		{
			name:     "data source",
			filename: "internal/provider/docs/data_sources/data_source_authenticated_user.md",
			wantErr:  false,
		},
		{
			name:     "function",
			filename: "internal/provider/docs/functions/function_normalize_project_members.md",
			wantErr:  false,
		},
		{
			name:     "non-existent file",
			filename: "internal/provider/docs/resources/nonexistent.md",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := readMarkdownDescription(ctx, tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("readMarkdownDescription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(content) == 0 {
				t.Errorf("readMarkdownDescription() returned empty content for %s", tt.filename)
			}
		})
	}
}
