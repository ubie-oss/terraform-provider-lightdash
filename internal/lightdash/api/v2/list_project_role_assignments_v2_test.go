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
	"testing"
)

func TestUnmarshalRoleAssignmentListResponse_projectAssignments(t *testing.T) {
	const fixture = `{
		"status": "ok",
		"results": [
			{
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
		]
	}`

	assignments, err := unmarshalRoleAssignmentListResponse([]byte(fixture))
	if err != nil {
		t.Fatalf("unmarshalRoleAssignmentListResponse: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("len = %d, want 1", len(assignments))
	}
	if assignments[0].AssigneeName != "Test User" {
		t.Errorf("AssigneeName = %q, want Test User", assignments[0].AssigneeName)
	}
	if assignments[0].ProjectID != "9cc0bae8-f552-4ac0-bdcc-44933d7031ae" {
		t.Errorf("ProjectID = %q", assignments[0].ProjectID)
	}
}
