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

package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

// listedBothAssignments is the project role list when the user and the group still have a role.
const listedBothAssignments = `{"status":"ok","results":[` +
	`{"roleId":"r1","roleName":"editor","assigneeType":"user","assigneeId":"user-uuid"},` +
	`{"roleId":"r1","roleName":"editor","assigneeType":"group","assigneeId":"group-uuid"}]}`

// newRemoveTestService returns a RoleService whose DELETE answers deleteStatus and whose
// project role list answers listStatus with listBody.
func newRemoveTestService(t *testing.T, deleteStatus int, listStatus int, listBody string) *RoleService {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(deleteStatus)
			return
		}
		w.WriteHeader(listStatus)
		_, _ = w.Write([]byte(listBody))
	}))
	t.Cleanup(server.Close)

	host, token := server.URL, "test-token"
	client, err := api.NewClient(&host, &token, nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return NewRoleService(client)
}

// A 404 on remove counts as removed when the list no longer has the assignment.
func TestRemoveProjectRole_notFoundAndNotListedIsRemoved(t *testing.T) {
	s := newRemoveTestService(t, http.StatusNotFound, http.StatusOK, `{"status":"ok","results":[]}`)
	ctx := context.Background()

	if err := s.RemoveProjectUserRole(ctx, "project-uuid", "user-uuid"); err != nil {
		t.Errorf("RemoveProjectUserRole when not listed: want nil, got %v", err)
	}
	if err := s.RemoveProjectGroupRole(ctx, "project-uuid", "group-uuid"); err != nil {
		t.Errorf("RemoveProjectGroupRole when not listed: want nil, got %v", err)
	}
}

// A 404 on remove still fails when the assignment is listed, so a wrong route or UUID is not
// reported as a successful removal.
func TestRemoveProjectRole_notFoundButStillListedFails(t *testing.T) {
	s := newRemoveTestService(t, http.StatusNotFound, http.StatusOK, listedBothAssignments)
	ctx := context.Background()

	if err := s.RemoveProjectUserRole(ctx, "project-uuid", "user-uuid"); err == nil {
		t.Error("RemoveProjectUserRole when still listed: want an error")
	}
	if err := s.RemoveProjectGroupRole(ctx, "project-uuid", "group-uuid"); err == nil {
		t.Error("RemoveProjectGroupRole when still listed: want an error")
	}
}

// A 404 on remove also fails when the list itself fails, for example for an unknown project UUID.
func TestRemoveProjectRole_notFoundAndListFailsFails(t *testing.T) {
	s := newRemoveTestService(t, http.StatusNotFound, http.StatusNotFound, `{"status":"error"}`)
	ctx := context.Background()

	if err := s.RemoveProjectUserRole(ctx, "project-uuid", "user-uuid"); err == nil {
		t.Error("RemoveProjectUserRole with an unknown project: want an error")
	}
	if err := s.RemoveProjectGroupRole(ctx, "project-uuid", "group-uuid"); err == nil {
		t.Error("RemoveProjectGroupRole with an unknown project: want an error")
	}
}

// Other remove failures surface without a list call.
func TestRemoveProjectRole_serverErrorFails(t *testing.T) {
	s := newRemoveTestService(t, http.StatusInternalServerError, http.StatusOK, `{"status":"ok","results":[]}`)
	ctx := context.Background()

	if err := s.RemoveProjectUserRole(ctx, "project-uuid", "user-uuid"); err == nil {
		t.Error("RemoveProjectUserRole on a server error: want an error")
	}
	if err := s.RemoveProjectGroupRole(ctx, "project-uuid", "group-uuid"); err == nil {
		t.Error("RemoveProjectGroupRole on a server error: want an error")
	}
}
