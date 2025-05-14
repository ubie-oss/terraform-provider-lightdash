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
	PROJECT_VIEWER_ROLE             ProjectMemberRole = "viewer"
	PROJECT_INTERACTIVE_VIEWER_ROLE ProjectMemberRole = "interactive_viewer"
	PROJECT_EDITOR_ROLE             ProjectMemberRole = "editor"
	PROJECT_DEVELOPER_ROLE          ProjectMemberRole = "developer"
	PROJECT_ADMIN_ROLE              ProjectMemberRole = "admin"
)

// convert ProjectMemberRole to string
func (s ProjectMemberRole) String() string {
	return string(s)
}

// Check if a given string is a valid ProjectMemberRole
func (s ProjectMemberRole) IsValid() bool {
	switch s {
	case PROJECT_VIEWER_ROLE,
		PROJECT_INTERACTIVE_VIEWER_ROLE,
		PROJECT_EDITOR_ROLE,
		PROJECT_DEVELOPER_ROLE,
		PROJECT_ADMIN_ROLE:
		return true
	}
	return false
}
