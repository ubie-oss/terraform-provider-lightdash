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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

func TestFindUserAttributesByName(t *testing.T) {
	attributes := []models.UserAttribute{
		{UserAttributeUUID: "attr-1", Name: "region"},
		{UserAttributeUUID: "attr-2", Name: "Region"},
		{UserAttributeUUID: "attr-3", Name: "region"},
	}

	matches := findUserAttributesByName(attributes, "region")
	if len(matches) != 2 || matches[0].UserAttributeUUID != "attr-1" || matches[1].UserAttributeUUID != "attr-3" {
		t.Fatalf("unexpected matches: %+v", matches)
	}
	if matches := findUserAttributesByName(attributes, "team"); len(matches) != 0 {
		t.Fatalf("expected no matches, got %+v", matches)
	}
}

func TestKeepUserValues(t *testing.T) {
	current := models.UserAttribute{
		Users: []models.UserAttributeValue{{UUID: "user-1", Values: []string{"eu"}}},
	}
	desired := models.UserAttribute{
		Name:              "region",
		Description:       "updated",
		AttributeDefaults: []string{"jp"},
		Groups:            []models.UserAttributeValue{{UUID: "group-1", Values: []string{"us"}}},
	}

	got := keepUserValues(desired, current)
	if !reflect.DeepEqual(got.Users, current.Users) {
		t.Fatalf("users were not kept: %+v", got.Users)
	}
	if got.Name != desired.Name || got.Description != desired.Description || !reflect.DeepEqual(got.Groups, desired.Groups) {
		t.Fatalf("desired fields changed: %+v", got)
	}
	if got := keepUserValues(desired, models.UserAttribute{}); got.Users != nil {
		t.Fatalf("expected no users, got %+v", got.Users)
	}
}

func TestCanonicalValues(t *testing.T) {
	if got := canonicalValues([]string{"us", "eu"}, "jp"); !reflect.DeepEqual(got, []string{"us", "eu"}) {
		t.Fatalf("expected the values array, got %+v", got)
	}
	if got := canonicalValues(nil, "jp"); !reflect.DeepEqual(got, []string{"jp"}) {
		t.Fatalf("expected the deprecated value, got %+v", got)
	}
	if got := canonicalValues(nil, " "); len(got) != 0 {
		t.Fatalf("expected no values, got %+v", got)
	}
}

func TestToUserAttributeV1RequestKeepsEmptyArrays(t *testing.T) {
	request := toUserAttributeV1Request(models.UserAttribute{Name: "region"})
	if request.Groups == nil || request.Users == nil {
		t.Fatalf("expected empty arrays, got groups=%#v users=%#v", request.Groups, request.Users)
	}
	if request.AttributeDefaults != nil {
		t.Fatalf("expected nil defaults, got %+v", request.AttributeDefaults)
	}
}

func TestUserAttributesServiceGetByName(t *testing.T) {
	service := newUserAttributeTestService(t, userAttributeTestHandler(t, nil))

	_, err := service.GetByName(context.Background(), "missing")
	if !errors.Is(err, ErrUserAttributeNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}

	_, err = service.GetByName(context.Background(), "region")
	if !errors.Is(err, ErrUserAttributeAmbiguous) {
		t.Fatalf("expected ambiguous, got %v", err)
	}

	got, err := service.GetByName(context.Background(), "team")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if got.UserAttributeUUID != "attr-3" {
		t.Fatalf("unexpected attribute: %+v", got)
	}
}

func TestUserAttributesServiceUpdateKeepsUsers(t *testing.T) {
	var putBody apiv1.CreateUserAttributeV1Request
	service := newUserAttributeTestService(t, userAttributeTestHandler(t, &putBody))

	updated, err := service.Update(context.Background(), "org-1", models.UserAttribute{
		OrganizationUUID:  "org-1",
		UserAttributeUUID: "attr-3",
		Name:              "team",
		Description:       "updated",
		Groups:            []models.UserAttributeValue{{UUID: "group-1", Values: []string{"us"}}},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Description != "updated" {
		t.Fatalf("unexpected description: %+v", updated)
	}
	if len(putBody.Users) != 1 || putBody.Users[0].UserUUID != "user-1" || !reflect.DeepEqual(putBody.Users[0].Values, []string{"eu"}) {
		t.Fatalf("update did not send the current users: %+v", putBody.Users)
	}
	if len(putBody.Groups) != 1 || putBody.Groups[0].GroupUUID != "group-1" {
		t.Fatalf("update did not send the groups: %+v", putBody.Groups)
	}
}

func TestUserAttributesServiceDeleteIsIdempotent(t *testing.T) {
	deleted := false
	service := newUserAttributeTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleted = true
			writeUserAttributeJSON(w, map[string]any{"status": "ok", "results": map[string]any{}})
			return
		}
		userAttributeTestHandler(t, nil)(w, r)
	})

	if err := service.Delete(context.Background(), "org-1", "missing"); err != nil {
		t.Fatalf("delete of a missing attribute: %v", err)
	}
	if deleted {
		t.Fatal("delete was called for a missing attribute")
	}
	if err := service.Delete(context.Background(), "org-1", "attr-3"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !deleted {
		t.Fatal("delete was not called")
	}
}

func TestUserAttributesServiceGetByUUIDOrganizationMismatch(t *testing.T) {
	service := newUserAttributeTestService(t, userAttributeTestHandler(t, nil))
	_, err := service.GetByUUID(context.Background(), "other-org", "attr-3")
	if !errors.Is(err, ErrUserAttributeOrganizationMismatch) {
		t.Fatalf("expected organization mismatch, got %v", err)
	}
	if errors.Is(err, ErrUserAttributeNotFound) {
		t.Fatal("a mismatch must not be reported as not found")
	}
}

func newUserAttributeTestService(t *testing.T, handler http.HandlerFunc) *UserAttributesService {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	host := server.URL
	token := "test-token"
	client, err := api.NewClient(&host, &token, nil)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	return NewUserAttributesService(client)
}

func userAttributeTestHandler(t *testing.T, putBody *apiv1.CreateUserAttributeV1Request) http.HandlerFunc {
	t.Helper()
	listed := []map[string]any{
		userAttributeJSON("attr-1", "region", []any{userValueJSON("user-9", []string{"jp"})}),
		userAttributeJSON("attr-2", "region", []any{}),
		userAttributeJSON("attr-3", "team", []any{userValueJSON("user-1", []string{"eu"})}),
	}
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/org/attributes":
			writeUserAttributeJSON(w, map[string]any{"status": "ok", "results": listed})
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/org/attributes/attr-3":
			if putBody != nil {
				if err := json.NewDecoder(r.Body).Decode(putBody); err != nil {
					t.Errorf("decode put body: %v", err)
				}
			} else {
				_, _ = io.Copy(io.Discard, r.Body)
			}
			updated := userAttributeJSON("attr-3", "team", []any{userValueJSON("user-1", []string{"eu"})})
			updated["description"] = "updated"
			writeUserAttributeJSON(w, map[string]any{"status": "ok", "results": updated})
		default:
			http.NotFound(w, r)
		}
	}
}

func userAttributeJSON(uuid string, name string, users []any) map[string]any {
	return map[string]any{
		"uuid":              uuid,
		"organizationUuid":  "org-1",
		"name":              name,
		"description":       "",
		"attributeDefault":  nil,
		"attributeDefaults": nil,
		"users":             users,
		"groups":            []any{},
		"createdAt":         "2024-01-01T00:00:00.000Z",
	}
}

func userValueJSON(userUUID string, values []string) map[string]any {
	return map[string]any{
		"userUuid": userUUID,
		"email":    userUUID + "@example.com",
		"value":    values[0],
		"values":   values,
	}
}

func writeUserAttributeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}
