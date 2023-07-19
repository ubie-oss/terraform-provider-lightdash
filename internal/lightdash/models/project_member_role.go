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

type ProjectMemberRole string

// List of ProjectMemberRole
const (
	VIEWER_ProjectMemberRole             ProjectMemberRole = "viewer"
	INTERACTIVE_VIEWER_ProjectMemberRole ProjectMemberRole = "interactive_viewer"
	EDITOR_ProjectMemberRole             ProjectMemberRole = "editor"
	DEVELOPER_ProjectMemberRole          ProjectMemberRole = "developer"
	ADMIN_ProjectMemberRole              ProjectMemberRole = "admin"
)

// convert ProjectMemberRole to string
func (s ProjectMemberRole) String() string {
	return string(s)
}

// Check if a given string is a valid ProjectMemberRole
func IsValidProjectMemberRole(s string) bool {
	switch ProjectMemberRole(s) {
	case VIEWER_ProjectMemberRole,
		INTERACTIVE_VIEWER_ProjectMemberRole,
		EDITOR_ProjectMemberRole,
		DEVELOPER_ProjectMemberRole,
		ADMIN_ProjectMemberRole:
		return true
	}
	return false
}
