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

type SpaceMemberRole string

// List of ProjectMemberRole
const (
	VIEWER_SpaceMemberRole SpaceMemberRole = "viewer"
	EDITOR_SpaceMemberRole SpaceMemberRole = "editor"
	ADMIN_SpaceMemberRole  SpaceMemberRole = "admin"
)

// convert ProjectMemberRole to string
func (s SpaceMemberRole) String() string {
	return string(s)
}

// Check if a given string is a valid SpaceMemberRole
func IsValidSpaceMemberRole(s string) bool {
	switch SpaceMemberRole(s) {
	case VIEWER_SpaceMemberRole,
		EDITOR_SpaceMemberRole,
		ADMIN_SpaceMemberRole:
		return true
	}
	return false
}
