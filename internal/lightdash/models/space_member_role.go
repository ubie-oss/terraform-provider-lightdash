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
	SPACE_VIEWER_ROLE SpaceMemberRole = "viewer"
	SPACE_EDITOR_ROLE SpaceMemberRole = "editor"
	SPACE_ADMIN_ROLE  SpaceMemberRole = "admin"
)

// convert ProjectMemberRole to string
func (s SpaceMemberRole) String() string {
	return string(s)
}

// Check if a given string is a valid SpaceMemberRole
func (s SpaceMemberRole) IsValid() bool {
	switch s {
	case SPACE_VIEWER_ROLE,
		SPACE_EDITOR_ROLE,
		SPACE_ADMIN_ROLE:
		return true
	}
	return false
}
