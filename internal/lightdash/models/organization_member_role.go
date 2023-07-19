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

type OrganizationMemberRole string

// List of OrganizationMemberRole
const (
	MEMBER_OrganizationMemberRole             OrganizationMemberRole = "member"
	VIEWER_OrganizationMemberRole             OrganizationMemberRole = "viewer"
	INTERACTIVE_VIEWER_OrganizationMemberRole OrganizationMemberRole = "interactive_viewer"
	EDITOR_OrganizationMemberRole             OrganizationMemberRole = "editor"
	DEVELOPER_OrganizationMemberRole          OrganizationMemberRole = "developer"
	ADMIN_OrganizationMemberRole              OrganizationMemberRole = "admin"
)

// default string
func (e OrganizationMemberRole) String() string {
	return string(e)
}

// Check if a given string is a valid OrganizationMemberRole
func IsValidOrganizationMemberRole(s string) bool {
	switch OrganizationMemberRole(s) {
	case MEMBER_OrganizationMemberRole,
		VIEWER_OrganizationMemberRole,
		INTERACTIVE_VIEWER_OrganizationMemberRole,
		EDITOR_OrganizationMemberRole,
		DEVELOPER_OrganizationMemberRole,
		ADMIN_OrganizationMemberRole:
		return true
	}
	return false
}
