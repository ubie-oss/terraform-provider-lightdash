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

// SpaceMemberAccess represents a user's access to a space
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

// SpaceGroupAccess represents a group's access to a space
type SpaceGroupAccess struct {
	GroupUUID string
	SpaceRole SpaceMemberRole // Assuming SpaceMemberRole is also in models package
}

// SpaceDetails contains all the details of a space
type SpaceDetails struct {
	ProjectUUID     string
	SpaceUUID       string
	ParentSpaceUUID *string
	SpaceName       string
	IsPrivate       bool
	MemberAccess    []SpaceMemberAccess
	GroupAccess     []SpaceGroupAccess
}
