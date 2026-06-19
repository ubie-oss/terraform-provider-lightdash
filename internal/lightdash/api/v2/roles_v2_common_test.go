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

package v2

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalRoleAssignmentResponse(t *testing.T) {
	const fixture = `{
		"status": "ok",
		"results": {
			"roleId": "editor",
			"roleName": "editor",
			"ownerType": "system",
			"assigneeType": "user",
			"assigneeId": "83d3ce52-4e26-4005-96b2-e3e945ac34ca",
			"assigneeName": "Test User",
			"projectId": "9cc0bae8-f552-4ac0-bdcc-44933d7031ae",
			"createdAt": "2026-06-18T09:37:53.439Z",
			"updatedAt": "2026-06-18T09:37:53.439Z"
		}
	}`

	assignment, err := unmarshalRoleAssignmentResponse([]byte(fixture))
	if err != nil {
		t.Fatalf("unmarshalRoleAssignmentResponse: %v", err)
	}
	if assignment.RoleID != "editor" {
		t.Errorf("RoleID = %q, want editor", assignment.RoleID)
	}
	if assignment.AssigneeID != "83d3ce52-4e26-4005-96b2-e3e945ac34ca" {
		t.Errorf("AssigneeID = %q", assignment.AssigneeID)
	}
}

func TestUpsertRoleAssignmentRequest_MarshalJSON(t *testing.T) {
	sendEmail := true
	req := upsertRoleAssignmentRequest{
		RoleID:    "editor",
		SendEmail: &sendEmail,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal upsertRoleAssignmentRequest: %v", err)
	}

	const want = `{"roleId":"editor","sendEmail":true}`
	if string(data) != want {
		t.Errorf("got %s, want %s", data, want)
	}
}

func TestUpdateRoleAssignmentRequest_MarshalJSON(t *testing.T) {
	req := updateRoleAssignmentRequest{RoleID: "viewer"}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal updateRoleAssignmentRequest: %v", err)
	}

	const want = `{"roleId":"viewer"}`
	if string(data) != want {
		t.Errorf("got %s, want %s", data, want)
	}
}
