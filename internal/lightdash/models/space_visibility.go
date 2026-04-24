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

// InheritParentPermissionsFromTerraformIsPrivate maps the Terraform `is_private`
// attribute to Lightdash `inheritParentPermissions` on CreateSpace / UpdateSpace
// (see OpenAPI `CreateSpace` and `UpdateSpace` in Lightdash swagger).
//
// Semantics used by this provider:
//   - is_private=false (public / inherits project permissions) => inheritParentPermissions=true
//   - is_private=true (restricted / private)                  => inheritParentPermissions=false
//
// When isPrivate is nil, no visibility flag is sent and the server default applies.
func InheritParentPermissionsFromTerraformIsPrivate(isPrivate *bool) *bool {
	if isPrivate == nil {
		return nil
	}
	v := !*isPrivate
	return &v
}

// TerraformIsPrivateFromInheritParentPermissions maps Lightdash `inheritParentPermissions`
// to the Terraform `is_private` attribute for a root space (no parent): inverse of inherit.
//
// For nested spaces, use TerraformIsPrivateFromAPIForNestedSpace instead.
func TerraformIsPrivateFromInheritParentPermissions(inheritParentPermissions bool) bool {
	return !inheritParentPermissions
}

// TerraformIsPrivateFromAPIFieldsRootSemantics returns Terraform is_private using only API
// fields on the row, treating the space as a root for this projection (inverse of effective
// inherit). Use for list summaries and debug logs where parent chain is unavailable; for
// nested resolution in state, use SpaceService.ResolveTerraformIsPrivateFromGetResults.
func TerraformIsPrivateFromAPIFieldsRootSemantics(inherit *bool, legacyIsPrivate bool) bool {
	return TerraformIsPrivateFromInheritParentPermissions(
		EffectiveInheritFromOptional(inherit, legacyIsPrivate),
	)
}

// TerraformIsPrivateFromAPIForNestedSpace maps API visibility to Terraform `is_private`.
//
// For a root space (under the project only), inheritParentPermissions refers to inheriting
// project permissions: true => public (is_private false), false => restricted (is_private true).
//
// For a nested space, inheritParentPermissions refers to inheriting the immediate parent space:
// when true, the child follows the parent's Terraform is_private; when false, the space uses
// its own access list (restricted) and is_private is true.
func TerraformIsPrivateFromAPIForNestedSpace(inheritEffective bool, parentTerraformIsPrivate bool, isRootSpace bool) bool {
	if isRootSpace {
		return !inheritEffective
	}
	if !inheritEffective {
		return true
	}
	return parentTerraformIsPrivate
}

// EffectiveInheritFromOptional resolves the effective inheritParentPermissions flag from API
// JSON when the field may be omitted, inferring from legacy isPrivate (restricted => isPrivate true).
func EffectiveInheritFromOptional(inherit *bool, legacyIsPrivate bool) bool {
	if inherit != nil {
		return *inherit
	}
	return !legacyIsPrivate
}

// InheritUpdatePointerIfChanged returns a non-nil pointer only when tfIsPrivate is set
// and desired inheritParentPermissions (!*tfIsPrivate) differs from currentInherit.
func InheritUpdatePointerIfChanged(tfIsPrivate *bool, currentInherit bool) *bool {
	if tfIsPrivate == nil {
		return nil
	}
	desired := !*tfIsPrivate
	if desired == currentInherit {
		return nil
	}
	return &desired
}
