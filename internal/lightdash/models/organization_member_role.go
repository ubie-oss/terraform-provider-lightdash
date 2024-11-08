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
	DEFAULT_ORGANIZATION_MEMBER_ROLE     OrganizationMemberRole = ORGANIZATION_MEMBER_ROLE
	ORGANIZATION_MEMBER_ROLE             OrganizationMemberRole = "member"
	ORGANIZATION_VIEWER_ROLE             OrganizationMemberRole = "viewer"
	ORGANIZATION_INTERACTIVE_VIEWER_ROLE OrganizationMemberRole = "interactive_viewer"
	ORGANIZATION_EDITOR_ROLE             OrganizationMemberRole = "editor"
	ORGANIZATION_DEVELOPER_ROLE          OrganizationMemberRole = "developer"
	ORGANIZATION_ADMIN_ROLE              OrganizationMemberRole = "admin"
)

// default string
func (e OrganizationMemberRole) String() string {
	return string(e)
}

// Check if a given string is a valid OrganizationMemberRole
func IsValidOrganizationMemberRole(s string) bool {
	switch OrganizationMemberRole(s) {
	case ORGANIZATION_MEMBER_ROLE,
		ORGANIZATION_VIEWER_ROLE,
		ORGANIZATION_INTERACTIVE_VIEWER_ROLE,
		ORGANIZATION_EDITOR_ROLE,
		ORGANIZATION_DEVELOPER_ROLE,
		ORGANIZATION_ADMIN_ROLE:
		return true
	}
	return false
}
