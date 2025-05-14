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

package models

import (
	"testing"
)

func TestSpaceMemberRoleString(t *testing.T) {
	tests := []struct {
		role     SpaceMemberRole
		expected string
	}{
		{SPACE_VIEWER_ROLE, "viewer"},
		{SPACE_EDITOR_ROLE, "editor"},
		{SPACE_ADMIN_ROLE, "admin"},
	}

	for _, test := range tests {
		if test.role.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.role.String())
		}
	}
}

func TestIsValidSpaceMemberRole(t *testing.T) {
	tests := []struct {
		role     string
		expected bool
	}{
		{"viewer", true},
		{"editor", true},
		{"admin", true},
		{"invalid", false},
	}

	for _, test := range tests {
		if SpaceMemberRole(test.role).IsValid() != test.expected {
			t.Errorf("Expected %v for role %s", test.expected, test.role)
		}
	}
}
