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

package v1

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

func TestSpaceAccessMemberInstantiation(t *testing.T) {
	member := SpaceAccessMember{
		UserUUID:  "user-uuid-123",
		SpaceRole: models.SpaceMemberRole("editor"),
	}

	if member.UserUUID != "user-uuid-123" {
		t.Errorf("expected UserUUID to be 'user-uuid-123', got '%s'", member.UserUUID)
	}
	if member.SpaceRole != "editor" {
		t.Errorf("expected SpaceRole to be 'editor', got '%s'", member.SpaceRole)
	}
}

func TestSpaceAccessGroupInstantiation(t *testing.T) {
	group := SpaceAccessGroup{
		GroupUUID: "group-uuid-456",
		GroupName: "Test Group",
		SpaceRole: models.SpaceMemberRole("viewer"),
	}

	if group.GroupUUID != "group-uuid-456" {
		t.Errorf("expected GroupUUID to be 'group-uuid-456', got '%s'", group.GroupUUID)
	}
	if group.GroupName != "Test Group" {
		t.Errorf("expected GroupName to be 'Test Group', got '%s'", group.GroupName)
	}
	if group.SpaceRole != "viewer" {
		t.Errorf("expected SpaceRole to be 'viewer', got '%s'", group.SpaceRole)
	}
}

func TestGetSpaceV1ResultsInstantiation(t *testing.T) {
	parentUUID := "parent-uuid"
	members := []SpaceAccessMember{
		{UserUUID: "user1", SpaceRole: models.SpaceMemberRole("editor")},
	}
	groups := []SpaceAccessGroup{
		{GroupUUID: "group1", GroupName: "Group One", SpaceRole: models.SpaceMemberRole("viewer")},
	}
	results := GetSpaceV1Results{
		ProjectUUID:        "proj-uuid",
		ParentSpaceUUID:    &parentUUID,
		SpaceUUID:          "space-uuid",
		SpaceName:          "Space Name",
		IsPrivate:          true,
		SpaceAccessMembers: members,
		SpaceAccessGroups:  groups,
	}

	if results.ProjectUUID != "proj-uuid" {
		t.Errorf("expected ProjectUUID to be 'proj-uuid', got '%s'", results.ProjectUUID)
	}
	if results.ParentSpaceUUID == nil || *results.ParentSpaceUUID != "parent-uuid" {
		t.Errorf("expected ParentSpaceUUID to be 'parent-uuid', got '%v'", results.ParentSpaceUUID)
	}
	if results.SpaceUUID != "space-uuid" {
		t.Errorf("expected SpaceUUID to be 'space-uuid', got '%s'", results.SpaceUUID)
	}
	if results.SpaceName != "Space Name" {
		t.Errorf("expected SpaceName to be 'Space Name', got '%s'", results.SpaceName)
	}
	if !results.IsPrivate {
		t.Errorf("expected IsPrivate to be true")
	}
	if !reflect.DeepEqual(results.SpaceAccessMembers, members) {
		t.Errorf("expected SpaceAccessMembers to be '%v', got '%v'", members, results.SpaceAccessMembers)
	}
	if !reflect.DeepEqual(results.SpaceAccessGroups, groups) {
		t.Errorf("expected SpaceAccessGroups to be '%v', got '%v'", groups, results.SpaceAccessGroups)
	}
}

func TestGetSpaceV1ResultsJSONUnmarshal(t *testing.T) {
	jsonStr := `{
		"projectUuid": "proj-uuid",
		"parentSpaceUuid": "parent-uuid",
		"uuid": "space-uuid",
		"name": "Space Name",
		"isPrivate": false,
		"access": [
			{"userUuid": "user1", "role": "editor"}
		],
		"groupsAccess": [
			{"groupUuid": "group1", "groupName": "Group One", "role": "viewer"}
		]
	}`

	var results GetSpaceV1Results
	err := json.Unmarshal([]byte(jsonStr), &results)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if results.ProjectUUID != "proj-uuid" {
		t.Errorf("expected ProjectUUID to be 'proj-uuid', got '%s'", results.ProjectUUID)
	}
	if results.ParentSpaceUUID == nil || *results.ParentSpaceUUID != "parent-uuid" {
		t.Errorf("expected ParentSpaceUUID to be 'parent-uuid', got '%v'", results.ParentSpaceUUID)
	}
	if results.SpaceUUID != "space-uuid" {
		t.Errorf("expected SpaceUUID to be 'space-uuid', got '%s'", results.SpaceUUID)
	}
	if results.SpaceName != "Space Name" {
		t.Errorf("expected SpaceName to be 'Space Name', got '%s'", results.SpaceName)
	}
	if results.IsPrivate {
		t.Errorf("expected IsPrivate to be false")
	}
	if len(results.SpaceAccessMembers) != 1 || results.SpaceAccessMembers[0].UserUUID != "user1" {
		t.Errorf("expected SpaceAccessMembers[0].UserUUID to be 'user1', got '%v'", results.SpaceAccessMembers)
	}
	if len(results.SpaceAccessGroups) != 1 || results.SpaceAccessGroups[0].GroupUUID != "group1" {
		t.Errorf("expected SpaceAccessGroups[0].GroupUUID to be 'group1', got '%v'", results.SpaceAccessGroups)
	}
}

func TestGetSpaceV1ResponseInstantiation(t *testing.T) {
	results := GetSpaceV1Results{
		ProjectUUID: "proj-uuid",
		SpaceUUID:   "space-uuid",
		SpaceName:   "Space Name",
		IsPrivate:   false,
	}
	resp := GetSpaceV1Response{
		Results: results,
		Status:  "ok",
	}

	if !reflect.DeepEqual(resp.Results, results) {
		t.Errorf("expected Results to be '%v', got '%v'", results, resp.Results)
	}
	if resp.Status != "ok" {
		t.Errorf("expected Status to be 'ok', got '%s'", resp.Status)
	}
}
