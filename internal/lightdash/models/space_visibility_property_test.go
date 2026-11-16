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

// TestInheritParentPermissions_roundTripsIsPrivate checks that a Terraform
// is_private value survives the trip to the API flag and back.
func TestInheritParentPermissions_roundTripsIsPrivate(t *testing.T) {
	if InheritParentPermissionsFromTerraformIsPrivate(nil) != nil {
		t.Fatal("nil is_private must send no inheritParentPermissions flag")
	}
	for _, isPrivate := range []bool{true, false} {
		inherit := InheritParentPermissionsFromTerraformIsPrivate(&isPrivate)
		if inherit == nil {
			t.Fatalf("is_private=%v: got nil inherit flag", isPrivate)
		}
		if got := TerraformIsPrivateFromInheritParentPermissions(*inherit); got != isPrivate {
			t.Errorf("is_private=%v round-trips to %v", isPrivate, got)
		}
	}
}

// TestInheritUpdatePointerIfChanged_alwaysReachesDesiredVisibility enumerates
// every combination of desired is_private and current inherit flag. After
// applying the returned update (or keeping the current flag when nil), the
// effective inherit flag must equal the desired one.
func TestInheritUpdatePointerIfChanged_alwaysReachesDesiredVisibility(t *testing.T) {
	trueVal, falseVal := true, false
	desiredCases := []*bool{nil, &trueVal, &falseVal}
	for _, tfIsPrivate := range desiredCases {
		for _, currentInherit := range []bool{true, false} {
			update := InheritUpdatePointerIfChanged(tfIsPrivate, currentInherit)
			effective := currentInherit
			if update != nil {
				effective = *update
			}
			if tfIsPrivate == nil {
				if update != nil {
					t.Errorf("nil is_private with current=%v returned update %v, want nil", currentInherit, *update)
				}
				continue
			}
			if want := !*tfIsPrivate; effective != want {
				t.Errorf("is_private=%v current=%v: effective inherit=%v, want %v",
					*tfIsPrivate, currentInherit, effective, want)
			}
		}
	}
}

// TestTerraformIsPrivateFromAPIForNestedSpace_invariants enumerates every
// combination of inherit flag, parent visibility, and root status, and checks the rules a
// nested space must follow: a non-inheriting space is always private, and an
// inheriting child on a root space follows the project; an inheriting nested child
// follows its parent.
func TestTerraformIsPrivateFromAPIForNestedSpace_invariants(t *testing.T) {
	for _, inherit := range []bool{true, false} {
		for _, parentPrivate := range []bool{true, false} {
			for _, isRoot := range []bool{true, false} {
				got := TerraformIsPrivateFromAPIForNestedSpace(inherit, parentPrivate, isRoot)
				switch {
				case !inherit && !got:
					t.Errorf("inherit=false parent=%v root=%v: got public, want private", parentPrivate, isRoot)
				case inherit && isRoot && got:
					t.Errorf("inherit=true root=true: got private, want public")
				case inherit && !isRoot && got != parentPrivate:
					t.Errorf("inherit=true nested parent=%v: got %v, want parent's visibility", parentPrivate, got)
				}
			}
		}
	}
}
