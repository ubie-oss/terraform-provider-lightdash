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

// TestNormalizeMembers_matchesHighestRoleSpec enumerates every assignment of the
// three members {a, b, c} to the five role lists (2^15 cases). Each member must
// appear in exactly one output list: the highest role it was given. Lists are sorted.
func TestNormalizeMembers_matchesHighestRoleSpec(t *testing.T) {
	const (
		roleCount   = 5 // 0 viewer, 1 interactive viewer, 2 editor, 3 developer, 4 admin
		memberCount = 3
	)
	members := []string{"a", "b", "c"}
	f := &NormalizeProjectMembersFunction{}

	for code := 0; code < 1<<(memberCount*roleCount); code++ {
		var masks [roleCount]int
		var inputs [roleCount][]string
		for k := 0; k < roleCount; k++ {
			masks[k] = (code >> (memberCount * k)) & (1<<memberCount - 1)
			for j, m := range members {
				if masks[k]&(1<<j) != 0 {
					inputs[k] = append(inputs[k], m)
				}
			}
		}

		// Oracle: each member goes to the highest role that lists it.
		want := [roleCount][]string{{}, {}, {}, {}, {}}
		for j, m := range members {
			for k := roleCount - 1; k >= 0; k-- {
				if masks[k]&(1<<j) != 0 {
					want[k] = append(want[k], m)
					break
				}
			}
		}

		admins, developers, editors, interactiveViewers, viewers := f.normalizeMembers(
			inputs[4], inputs[3], inputs[2], inputs[1], inputs[0],
		)
		got := [roleCount][]string{viewers, interactiveViewers, editors, developers, admins}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("inputs by role (viewer..admin) %v: got %v, want %v", inputs, got, want)
		}
	}
}

// TestNormalizeMembers_duplicateInputsCollapse checks that a member repeated
// within one input list appears once in the output.
func TestNormalizeMembers_duplicateInputsCollapse(t *testing.T) {
	f := &NormalizeProjectMembersFunction{}
	admins, _, _, _, viewers := f.normalizeMembers(
		[]string{"a", "a"}, nil, nil, nil, []string{"a", "b", "b"},
	)
	if !reflect.DeepEqual(admins, []string{"a"}) {
		t.Errorf("admins = %v, want [a]", admins)
	}
	if !reflect.DeepEqual(viewers, []string{"b"}) {
		t.Errorf("viewers = %v, want [b]", viewers)
	}
}
