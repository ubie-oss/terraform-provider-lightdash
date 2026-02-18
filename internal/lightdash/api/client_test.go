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
	"testing"
)

func TestNewClient(t *testing.T) {
	// Test case 1: Both host and token provided
	host := "example.com"
	expectedHost := "https://example.com"
	token := "abc123"
	maxRequests := int64(15)
	client, err := NewClient(&host, &token, &maxRequests)
	if err != nil {
		t.Errorf("Error creating client: %s", err.Error())
	}

	// Assert client properties
	if client.HTTPClient == nil {
		t.Error("Expected non-nil HTTPClient")
	}
	if client.HostUrl != expectedHost {
		t.Errorf("Expected HostUrl: %s, got: %s", expectedHost, client.HostUrl)
	}
	if client.Token != token {
		t.Errorf("Expected Token: %s, got: %s", token, client.Token)
	}
	if cap(client.Semaphore) != 15 {
		t.Errorf("Expected Semaphore capacity: 15, got: %d", cap(client.Semaphore))
	}

	// Test case 2: Only host provided
	host = "example.com"
	expectedHost = "https://example.com"
	token = ""
	client, err = NewClient(&host, nil, nil)
	if err != nil {
		t.Errorf("Error creating client: %s", err.Error())
	}

	// Assert client properties
	if client.HTTPClient == nil {
		t.Error("Expected non-nil HTTPClient")
	}
	if client.HostUrl != expectedHost {
		t.Errorf("Expected HostUrl: %s, got: %s", expectedHost, client.HostUrl)
	}
	if client.Token != "" {
		t.Errorf("Expected empty Token, got: %s", client.Token)
	}
	if cap(client.Semaphore) != 10 {
		t.Errorf("Expected default Semaphore capacity: 10, got: %d", cap(client.Semaphore))
	}

	// Test case 3: Only token provided
	host = ""
	token = "abc123"
	client, err = NewClient(nil, &token, nil)
	if err != nil {
		t.Errorf("Error creating client: %s", err.Error())
	}

	// Assert client properties
	if client.HTTPClient == nil {
		t.Error("Expected non-nil HTTPClient")
	}
	if client.HostUrl != "" {
		t.Errorf("Expected empty HostUrl, got: %s", client.HostUrl)
	}
	if client.Token != token {
		t.Errorf("Expected Token: %s, got: %s", token, client.Token)
	}

	// Test case 4: Neither host nor token provided
	host = ""
	token = ""
	client, err = NewClient(nil, nil, nil)
	if err != nil {
		t.Errorf("Error creating client: %s", err.Error())
	}

	// Assert client properties
	if client.HTTPClient == nil {
		t.Error("Expected non-nil HTTPClient")
	}
	if client.HostUrl != "" {
		t.Errorf("Expected empty HostUrl, got: %s", client.HostUrl)
	}
	if client.Token != "" {
		t.Errorf("Expected empty Token, got: %s", client.Token)
	}
}
