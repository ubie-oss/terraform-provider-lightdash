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

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUpdateSpaceV1Request_MarshalJSON(t *testing.T) {
	parentUUID := "parent-uuid-456"
	cases := []struct {
		name     string
		input    UpdateSpaceV1Request
		expected map[string]interface{}
	}{
		{
			name: "All fields set",
			input: UpdateSpaceV1Request{
				Name:            "Updated Space",
				IsPrivate:       true,
				ParentSpaceUUID: &parentUUID,
			},
			expected: map[string]interface{}{
				"name":            "Updated Space",
				"isPrivate":       true,
				"parentSpaceUuid": "parent-uuid-456",
			},
		},
		{
			name: "ParentSpaceUUID nil",
			input: UpdateSpaceV1Request{
				Name:            "No Parent Space",
				IsPrivate:       false,
				ParentSpaceUUID: nil,
			},
			expected: map[string]interface{}{
				"name":      "No Parent Space",
				"isPrivate": false,
			},
		},
		{
			name: "Empty name, isPrivate true, ParentSpaceUUID nil",
			input: UpdateSpaceV1Request{
				Name:            "",
				IsPrivate:       true,
				ParentSpaceUUID: nil,
			},
			expected: map[string]interface{}{
				"name":      "",
				"isPrivate": true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("failed to marshal UpdateSpaceV1Request: %v", err)
			}
			var got map[string]interface{}
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatalf("failed to unmarshal marshaled JSON: %v", err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("unexpected marshaled JSON.\nGot:      %v\nExpected: %v", got, tc.expected)
			}
		})
	}
}

func TestUpdateSpaceV1Request_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected UpdateSpaceV1Request
	}{
		{
			name:  "With ParentSpaceUUID",
			input: `{"name":"Space X","isPrivate":true,"parentSpaceUuid":"parent-uuid-xyz"}`,
			expected: func() UpdateSpaceV1Request {
				uuid := "parent-uuid-xyz"
				return UpdateSpaceV1Request{
					Name:            "Space X",
					IsPrivate:       true,
					ParentSpaceUUID: &uuid,
				}
			}(),
		},
		{
			name:  "Without ParentSpaceUUID",
			input: `{"name":"Space Y","isPrivate":false}`,
			expected: UpdateSpaceV1Request{
				Name:            "Space Y",
				IsPrivate:       false,
				ParentSpaceUUID: nil,
			},
		},
		{
			name:  "Empty name, isPrivate true, no parent",
			input: `{"name":"","isPrivate":true}`,
			expected: UpdateSpaceV1Request{
				Name:            "",
				IsPrivate:       true,
				ParentSpaceUUID: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req UpdateSpaceV1Request
			err := json.Unmarshal([]byte(tc.input), &req)
			if err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if req.Name != tc.expected.Name {
				t.Errorf("expected Name '%s', got '%s'", tc.expected.Name, req.Name)
			}
			if req.IsPrivate != tc.expected.IsPrivate {
				t.Errorf("expected IsPrivate %v, got %v", tc.expected.IsPrivate, req.IsPrivate)
			}
			if tc.expected.ParentSpaceUUID == nil {
				if req.ParentSpaceUUID != nil {
					t.Errorf("expected ParentSpaceUUID nil, got '%v'", req.ParentSpaceUUID)
				}
			} else {
				if req.ParentSpaceUUID == nil || *req.ParentSpaceUUID != *tc.expected.ParentSpaceUUID {
					t.Errorf("expected ParentSpaceUUID '%v', got '%v'", *tc.expected.ParentSpaceUUID, req.ParentSpaceUUID)
				}
			}
		})
	}
}

func TestUpdateSpaceV1Results_MarshalUnmarshalJSON(t *testing.T) {
	parentUUID := "parent-uuid-abc"
	marshalCases := []struct {
		name     string
		input    UpdateSpaceV1Results
		expected map[string]interface{}
	}{
		{
			name: "All fields set",
			input: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-10",
				ProjectUUID:      "proj-uuid-10",
				ParentSpaceUUID:  &parentUUID,
				SpaceUUID:        "space-uuid-10",
				SpaceName:        "Updated Space 10",
				IsPrivate:        false,
			},
			expected: map[string]interface{}{
				"organizationUuid": "org-uuid-10",
				"projectUuid":      "proj-uuid-10",
				"parentSpaceUuid":  "parent-uuid-abc",
				"uuid":             "space-uuid-10",
				"name":             "Updated Space 10",
				"isPrivate":        false,
			},
		},
		{
			name: "No ParentSpaceUUID",
			input: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-11",
				ProjectUUID:      "proj-uuid-11",
				ParentSpaceUUID:  nil,
				SpaceUUID:        "space-uuid-11",
				SpaceName:        "Updated Space 11",
				IsPrivate:        true,
			},
			expected: map[string]interface{}{
				"organizationUuid": "org-uuid-11",
				"projectUuid":      "proj-uuid-11",
				"uuid":             "space-uuid-11",
				"name":             "Updated Space 11",
				"isPrivate":        true,
			},
		},
	}

	for _, tc := range marshalCases {
		t.Run("Marshal/"+tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("failed to marshal UpdateSpaceV1Results: %v", err)
			}
			var got map[string]interface{}
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatalf("failed to unmarshal marshaled JSON: %v", err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("unexpected marshaled JSON.\nGot:      %v\nExpected: %v", got, tc.expected)
			}
		})
	}

	unmarshalCases := []struct {
		name     string
		input    string
		expected UpdateSpaceV1Results
	}{
		{
			name: "All fields set",
			input: `{
				"organizationUuid": "org-uuid-20",
				"projectUuid": "proj-uuid-20",
				"parentSpaceUuid": "parent-uuid-20",
				"uuid": "space-uuid-20",
				"name": "Updated Space 20",
				"isPrivate": true
			}`,
			expected: func() UpdateSpaceV1Results {
				uuid := "parent-uuid-20"
				return UpdateSpaceV1Results{
					OrganizationUUID: "org-uuid-20",
					ProjectUUID:      "proj-uuid-20",
					ParentSpaceUUID:  &uuid,
					SpaceUUID:        "space-uuid-20",
					SpaceName:        "Updated Space 20",
					IsPrivate:        true,
				}
			}(),
		},
		{
			name: "Without ParentSpaceUUID",
			input: `{
				"organizationUuid": "org-uuid-21",
				"projectUuid": "proj-uuid-21",
				"uuid": "space-uuid-21",
				"name": "Updated Space 21",
				"isPrivate": false
			}`,
			expected: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-21",
				ProjectUUID:      "proj-uuid-21",
				ParentSpaceUUID:  nil,
				SpaceUUID:        "space-uuid-21",
				SpaceName:        "Updated Space 21",
				IsPrivate:        false,
			},
		},
	}

	for _, tc := range unmarshalCases {
		t.Run("Unmarshal/"+tc.name, func(t *testing.T) {
			var res UpdateSpaceV1Results
			err := json.Unmarshal([]byte(tc.input), &res)
			if err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if res.OrganizationUUID != tc.expected.OrganizationUUID {
				t.Errorf("expected OrganizationUUID '%s', got '%s'", tc.expected.OrganizationUUID, res.OrganizationUUID)
			}
			if res.ProjectUUID != tc.expected.ProjectUUID {
				t.Errorf("expected ProjectUUID '%s', got '%s'", tc.expected.ProjectUUID, res.ProjectUUID)
			}
			if tc.expected.ParentSpaceUUID == nil {
				if res.ParentSpaceUUID != nil {
					t.Errorf("expected ParentSpaceUUID nil, got '%v'", res.ParentSpaceUUID)
				}
			} else {
				if res.ParentSpaceUUID == nil || *res.ParentSpaceUUID != *tc.expected.ParentSpaceUUID {
					t.Errorf("expected ParentSpaceUUID '%v', got '%v'", *tc.expected.ParentSpaceUUID, res.ParentSpaceUUID)
				}
			}
			if res.SpaceUUID != tc.expected.SpaceUUID {
				t.Errorf("expected SpaceUUID '%s', got '%s'", tc.expected.SpaceUUID, res.SpaceUUID)
			}
			if res.SpaceName != tc.expected.SpaceName {
				t.Errorf("expected SpaceName '%s', got '%s'", tc.expected.SpaceName, res.SpaceName)
			}
			if res.IsPrivate != tc.expected.IsPrivate {
				t.Errorf("expected IsPrivate %v, got %v", tc.expected.IsPrivate, res.IsPrivate)
			}
		})
	}
}

func TestUpdateSpaceV1Results_FieldAssignment(t *testing.T) {
	parentUUID := "parent-uuid-abc"
	cases := []struct {
		name     string
		input    UpdateSpaceV1Results
		expected UpdateSpaceV1Results
	}{
		{
			name: "All fields set",
			input: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-30",
				ProjectUUID:      "proj-uuid-30",
				ParentSpaceUUID:  &parentUUID,
				SpaceUUID:        "space-uuid-30",
				SpaceName:        "Updated Space 30",
				IsPrivate:        false,
			},
			expected: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-30",
				ProjectUUID:      "proj-uuid-30",
				ParentSpaceUUID:  &parentUUID,
				SpaceUUID:        "space-uuid-30",
				SpaceName:        "Updated Space 30",
				IsPrivate:        false,
			},
		},
		{
			name: "No ParentSpaceUUID",
			input: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-31",
				ProjectUUID:      "proj-uuid-31",
				ParentSpaceUUID:  nil,
				SpaceUUID:        "space-uuid-31",
				SpaceName:        "Updated Space 31",
				IsPrivate:        true,
			},
			expected: UpdateSpaceV1Results{
				OrganizationUUID: "org-uuid-31",
				ProjectUUID:      "proj-uuid-31",
				ParentSpaceUUID:  nil,
				SpaceUUID:        "space-uuid-31",
				SpaceName:        "Updated Space 31",
				IsPrivate:        true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			results := tc.input
			if results.OrganizationUUID != tc.expected.OrganizationUUID {
				t.Errorf("expected OrganizationUUID '%s', got '%s'", tc.expected.OrganizationUUID, results.OrganizationUUID)
			}
			if results.ProjectUUID != tc.expected.ProjectUUID {
				t.Errorf("expected ProjectUUID '%s', got '%s'", tc.expected.ProjectUUID, results.ProjectUUID)
			}
			if tc.expected.ParentSpaceUUID == nil {
				if results.ParentSpaceUUID != nil {
					t.Errorf("expected ParentSpaceUUID nil, got '%v'", results.ParentSpaceUUID)
				}
			} else {
				if results.ParentSpaceUUID == nil || *results.ParentSpaceUUID != *tc.expected.ParentSpaceUUID {
					t.Errorf("expected ParentSpaceUUID '%v', got '%v'", *tc.expected.ParentSpaceUUID, results.ParentSpaceUUID)
				}
			}
			if results.SpaceUUID != tc.expected.SpaceUUID {
				t.Errorf("expected SpaceUUID '%s', got '%s'", tc.expected.SpaceUUID, results.SpaceUUID)
			}
			if results.SpaceName != tc.expected.SpaceName {
				t.Errorf("expected SpaceName '%s', got '%s'", tc.expected.SpaceName, results.SpaceName)
			}
			if results.IsPrivate != tc.expected.IsPrivate {
				t.Errorf("expected IsPrivate %v, got %v", tc.expected.IsPrivate, results.IsPrivate)
			}
		})
	}
}

func TestClient_UpdateSpaceV1(t *testing.T) {
	// Define a dummy response body for the mock server
	dummyResponseBody := `{"status": "ok", "results": {"organizationUuid": "org-uuid", "projectUuid": "proj-uuid", "uuid": "space-uuid", "name": "Updated Space", "isPrivate": false}}`

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		expectedPath := "/api/v1/projects/test-project-uuid/spaces/test-space-uuid"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected request path %s, got %s", expectedPath, r.URL.Path)
		}

		// Check the request body
		var requestBody UpdateSpaceV1Request
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		expectedRequestBody := UpdateSpaceV1Request{
			Name:      "Updated Space",
			IsPrivate: false,
		}

		// Compare the decoded request body with the expected body
		if !reflect.DeepEqual(requestBody, expectedRequestBody) {
			t.Errorf("Request body mismatch:\n%s", cmp.Diff(expectedRequestBody, requestBody))
		}

		// Write the dummy response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, dummyResponseBody)
	}))
	defer server.Close()

	// Create a new client with the mock server URL
	client := NewClient(server.URL, "dummy-token")

	// Call the API function
	spaceName := "Updated Space"
	isPrivate := false
	var isPrivatePtr *bool
	isPrivatePtr = &isPrivate

	// The UpdateSpaceV1 function should not take parentSpaceUuid as an argument
	// It only updates name and isPrivate
	resp, err := client.UpdateSpaceV1("test-project-uuid", "test-space-uuid", spaceName, isPrivatePtr)

	// Check for errors
	if err != nil {
		t.Fatalf("UpdateSpaceV1 failed: %v", err)
	}

	// Define the expected results
	expectedResults := &UpdateSpaceV1Results{
		OrganizationUUID: "org-uuid",
		ProjectUUID:      "proj-uuid",
		SpaceUUID:        "space-uuid",
		SpaceName:        "Updated Space",
		IsPrivate:        false,
		ParentSpaceUUID:  nil, // The API response should return the *actual* parent UUID
	}

	// Compare the received results with the expected results
	if !reflect.DeepEqual(resp, expectedResults) {
		t.Errorf("Response results mismatch:\n%s", cmp.Diff(expectedResults, resp))
	}
}

// Add a test case for updating only the name
func TestClient_UpdateSpaceV1_OnlyName(t *testing.T) {
	// Define a dummy response body for the mock server
	dummyResponseBody := `{"status": "ok", "results": {"organizationUuid": "org-uuid", "projectUuid": "proj-uuid", "uuid": "space-uuid", "name": "Name Only Update", "isPrivate": true}}`

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		expectedPath := "/api/v1/projects/test-project-uuid/spaces/test-space-uuid"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected request path %s, got %s", expectedPath, r.URL.Path)
		}

		// Check the request body
		var requestBody UpdateSpaceV1Request
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		expectedRequestBody := UpdateSpaceV1Request{
			Name: "Name Only Update",
			// IsPrivate is omitted or set to default in the request if not provided
		}

		// Compare the decoded request body with the expected body (ignore default/omitted fields)
		// We only check the fields explicitly set in the request body in this test case
		if requestBody.Name != expectedRequestBody.Name {
			t.Errorf("Request body name mismatch: Expected %s, got %s", expectedRequestBody.Name, requestBody.Name)
		}

		// Write the dummy response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, dummyResponseBody)
	}))
	defer server.Close()

	// Create a new client with the mock server URL
	client := NewClient(server.URL, "dummy-token")

	// Call the API function with only the name provided
	spaceName := "Name Only Update"
	resp, err := client.UpdateSpaceV1("test-project-uuid", "test-space-uuid", spaceName, nil)

	// Check for errors
	if err != nil {
		t.Fatalf("UpdateSpaceV1 failed: %v", err)
	}

	// Define the expected results
	expectedResults := &UpdateSpaceV1Results{
		OrganizationUUID: "org-uuid",
		ProjectUUID:      "proj-uuid",
		SpaceUUID:        "space-uuid",
		SpaceName:        "Name Only Update", // The API response should return the actual privacy setting
		IsPrivate:        true,
		ParentSpaceUUID:  nil, // The API response should return the actual parent UUID
	}

	// Compare the received results with the expected results
	// Use cmp.Diff for a detailed diff
	if !cmp.Equal(resp, expectedResults) {
		t.Errorf("Response results mismatch:\n%s", cmp.Diff(expectedResults, resp))
	}
}

// Add a test case for updating only isPrivate
func TestClient_UpdateSpaceV1_OnlyIsPrivate(t *testing.T) {
	// Define a dummy response body for the mock server
	dummyResponseBody := `{"status": "ok", "results": {"organizationUuid": "org-uuid", "projectUuid": "proj-uuid", "uuid": "space-uuid", "name": "Original Name", "isPrivate": false}}`

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		expectedPath := "/api/v1/projects/test-project-uuid/spaces/test-space-uuid"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected request path %s, got %s", expectedPath, r.URL.Path)
		}

		// Check the request body
		var requestBody UpdateSpaceV1Request
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		expectedRequestBody := UpdateSpaceV1Request{
			// Name is omitted or set to default in the request if not provided
			IsPrivate: false,
		}

		// Compare the decoded request body with the expected body (ignore default/omitted fields)
		// We only check the fields explicitly set in the request body in this test case
		if requestBody.IsPrivate != expectedRequestBody.IsPrivate {
			t.Errorf("Request body isPrivate mismatch: Expected %t, got %t", expectedRequestBody.IsPrivate, requestBody.IsPrivate)
		}

		// Write the dummy response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, dummyResponseBody)
	}))
	defer server.Close()

	// Create a new client with the mock server URL
	client := NewClient(server.URL, "dummy-token")

	// Call the API function with only isPrivate provided
	isPrivate := false
	var isPrivatePtr *bool
	isPrivatePtr = &isPrivate
	resp, err := client.UpdateSpaceV1("test-project-uuid", "test-space-uuid", "Original Name", isPrivatePtr)

	// Check for errors
	if err != nil {
		t.Fatalf("UpdateSpaceV1 failed: %v", err)
	}

	// Define the expected results
	expectedResults := &UpdateSpaceV1Results{
		OrganizationUUID: "org-uuid",
		ProjectUUID:      "proj-uuid",
		SpaceUUID:        "space-uuid",
		SpaceName:        "Original Name", // The API response should return the actual name
		IsPrivate:        false,
		ParentSpaceUUID:  nil, // The API response should return the actual parent UUID
	}

	// Compare the received results with the expected results
	// Use cmp.Diff for a detailed diff
	if !cmp.Equal(resp, expectedResults) {
		t.Errorf("Response results mismatch:\n%s", cmp.Diff(expectedResults, resp))
	}
}

// Add a test case for when the API returns an error
func TestClient_UpdateSpaceV1_APIError(t *testing.T) {
	// Create a mock HTTP server that returns an error status code
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{"status": "error", "error": {"message": "Internal server error"}}`)
	}))
	defer server.Close()

	// Create a new client with the mock server URL
	client := NewClient(server.URL, "dummy-token")

	// Call the API function
	spaceName := "Should Not Matter"
	resp, err := client.UpdateSpaceV1("test-project-uuid", "test-space-uuid", spaceName, nil)

	// Check for errors
	if err == nil {
		t.Fatalf("Expected an error, but got none. Response: %+v", resp)
	}

	// Check if the error message is as expected (or contains the expected substring)
	expectedErrorSubstring := "request failed"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error message to contain '%s', but got '%s'", expectedErrorSubstring, err.Error())
	}
}

// Add a test case for when the response body is invalid JSON
func TestClient_UpdateSpaceV1_InvalidJSON(t *testing.T) {
	// Create a mock HTTP server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `invalid json}`)
	}))
	defer server.Close()

	// Create a new client with the mock server URL
	client := NewClient(server.URL, "dummy-token")

	// Call the API function
	spaceName := "Should Not Matter"
	resp, err := client.UpdateSpaceV1("test-project-uuid", "test-space-uuid", spaceName, nil)

	// Check for errors
	if err == nil {
		t.Fatalf("Expected an error, but got none. Response: %+v", resp)
	}

	// Check if the error message is related to unmarshalling JSON
	expectedErrorSubstring := "failed to unmarshal response"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error message to contain '%s', but got '%s'", expectedErrorSubstring, err.Error())
	}
}

// Add a test case for when the API response results are missing required fields
func TestClient_UpdateSpaceV1_MissingRequiredFields(t *testing.T) {
	// Define a dummy response body with missing required fields (e.g., SpaceUUID)
	dummyResponseBody := `{"status": "ok", "results": {"organizationUuid": "org-uuid", "projectUuid": "proj-uuid", "name": "Updated Space", "isPrivate": false}}`

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, dummyResponseBody)
	}))
	defer server.Close()

	// Create a new client with the mock server URL
	client := NewClient(server.URL, "dummy-token")

	// Call the API function
	spaceName := "Updated Space"
	isPrivate := false
	var isPrivatePtr *bool
	isPrivatePtr = &isPrivate
	resp, err := client.UpdateSpaceV1("test-project-uuid", "test-space-uuid", spaceName, isPrivatePtr)

	// Check for errors
	if err == nil {
		t.Fatalf("Expected an error, but got none. Response: %+v", resp)
	}

	// Check if the error message is related to missing UUID
	expectedErrorSubstring := "space UUID is nil"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error message to contain '%s', but got '%s'", expectedErrorSubstring, err.Error())
	}
}
