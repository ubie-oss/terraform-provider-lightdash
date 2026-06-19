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

func TestUnmarshalRoleAssignmentListResponse_orgAssignments(t *testing.T) {
	const fixture = `{
		"status": "ok",
		"results": [
			{
				"roleId": "admin",
				"roleName": "admin",
				"assigneeType": "user",
				"ownerType": "system",
				"assigneeId": "b61b8510-dca3-4cab-8080-db170dd9a2ee",
				"organizationId": "089a18c4-667e-41cb-9d10-b088461ac941",
				"createdAt": "2023-08-28T18:27:36.097Z",
				"updatedAt": "2023-08-28T18:27:36.097Z"
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
	if assignments[0].RoleID != "admin" {
		t.Errorf("RoleID = %q, want admin", assignments[0].RoleID)
	}
	if assignments[0].OrganizationID != "089a18c4-667e-41cb-9d10-b088461ac941" {
		t.Errorf("OrganizationID = %q", assignments[0].OrganizationID)
	}
}
