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

// SpaceAccessMember represents the core information for a space access member used in requests.
type SpaceAccessMember struct {
	UserUUID  string
	SpaceRole SpaceMemberRole
}

// SpaceGroupAccess represents a group's access to a space
type SpaceGroupAccess struct {
	GroupUUID string
	SpaceRole SpaceMemberRole
}

// SpaceAccessGroup represents a group's access to a space as returned by the API
type SpaceAccessGroup struct {
	GroupUUID string
	SpaceRole SpaceMemberRole
}

// SpaceMemberAccess represents a user's access to a space as returned by the API
type SpaceMemberAccess struct {
	UserUUID  string
	SpaceRole SpaceMemberRole // Assuming SpaceMemberRole is also in models package
	// Fields from API response indicating how access is granted
	HasDirectAccess *bool   `json:"hasDirectAccess,omitempty"`
	InheritedRole   *string `json:"inheritedRole,omitempty"`
	InheritedFrom   *string `json:"inheritedFrom,omitempty"`
	ProjectRole     *string `json:"projectRole,omitempty"`
}

// GetSpaceAccessType returns the type of space access for a member
func (s *SpaceMemberAccess) GetSpaceAccessType() *string {
	// No direct access
	if s.HasDirectAccess == nil || !*s.HasDirectAccess {
		return nil
	}

	// Group space access
	if s.InheritedFrom != nil && *s.InheritedFrom == "group" {
		group := "group"
		return &group
	}
	// Individual space access member
	member := "member"
	return &member
}

// HasDirectSpaceMemberAccess returns true if the member has direct access to the space
func (s *SpaceMemberAccess) HasDirectSpaceMemberAccess() bool {
	return s.HasDirectAccess != nil && *s.HasDirectAccess &&
		(s.InheritedFrom == nil || *s.InheritedFrom != "group")
}

// ChildSpace represents a nested space within a parent space
type ChildSpace struct {
	SpaceUUID  string
	SpaceName  string
	IsPrivate  bool
	AccessList []SpaceMemberAccess
}

// SpaceDetails contains all the details of a space returned by the GetSpace API.
// Note: For nested spaces, MemberAccess and GroupAccess lists will be empty as access is inherited.
type SpaceDetails struct {
	ProjectUUID        string
	SpaceUUID          string
	ParentSpaceUUID    *string
	SpaceName          string
	IsPrivate          bool
	SpaceAccessMembers []SpaceMemberAccess // Full list from API for access_all
	SpaceAccessGroups  []SpaceAccessGroup  // Full list from API for group_access_all
	ChildSpaces        []ChildSpace        // Child spaces, if any
}

// IsNestedSpace returns true if the space is nested (has a parent)
func (s *SpaceDetails) IsNestedSpace() bool {
	return s.ParentSpaceUUID != nil && *s.ParentSpaceUUID != ""
}

// GetDirectMemberAccess returns only the members with direct access to the space
func (s *SpaceDetails) GetDirectMemberAccess() []SpaceMemberAccess {
	var directMembers []SpaceMemberAccess

	for _, member := range s.SpaceAccessMembers {
		if member.HasDirectSpaceMemberAccess() {
			directMembers = append(directMembers, member)
		}
	}

	return directMembers
}

// GetMemberByUUID returns a member by UUID, or nil if not found
func (s *SpaceDetails) GetMemberByUUID(userUUID string) *SpaceMemberAccess {
	for i, member := range s.SpaceAccessMembers {
		if member.UserUUID == userUUID {
			return &s.SpaceAccessMembers[i]
		}
	}
	return nil
}

// GetGroupByUUID returns a group by UUID, or nil if not found
func (s *SpaceDetails) GetGroupByUUID(groupUUID string) *SpaceAccessGroup {
	for i, group := range s.SpaceAccessGroups {
		if group.GroupUUID == groupUUID {
			return &s.SpaceAccessGroups[i]
		}
	}
	return nil
}

// ResourceID returns the formatted resource ID for a space
func (s *SpaceDetails) ResourceID() string {
	return "projects/" + s.ProjectUUID + "/spaces/" + s.SpaceUUID
}
