// Copyright 2023 Ubie, inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import "testing"

func TestTerraformIsPrivateFromAPIFieldsRootSemantics(t *testing.T) {
	t.Parallel()
	inheritTrue := true
	if TerraformIsPrivateFromAPIFieldsRootSemantics(&inheritTrue, false) {
		t.Fatal("public row should be is_private false")
	}
	inheritFalse := false
	if !TerraformIsPrivateFromAPIFieldsRootSemantics(&inheritFalse, true) {
		t.Fatal("restricted row should be is_private true")
	}
}

func TestTerraformIsPrivateFromAPIForNestedSpace(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		inherit  bool
		parent   bool
		isRoot   bool
		expected bool
	}{
		{"root_public", true, false, true, false},
		{"root_restricted", false, false, true, true},
		{"nested_own_acl", false, false, false, true},
		{"nested_inherit_public_parent", true, false, false, false},
		{"nested_inherit_private_parent", true, true, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := TerraformIsPrivateFromAPIForNestedSpace(tc.inherit, tc.parent, tc.isRoot)
			if got != tc.expected {
				t.Fatalf("TerraformIsPrivateFromAPIForNestedSpace(%v,%v,%v) = %v, want %v",
					tc.inherit, tc.parent, tc.isRoot, got, tc.expected)
			}
		})
	}
}
