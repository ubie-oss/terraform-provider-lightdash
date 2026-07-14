// Copyright 2023 Ubie, inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"encoding/json"
	"testing"
)

func TestGetProjectV1Results_JSON_withUpstream(t *testing.T) {
	const payload = `{
		"organizationUuid": "org-1",
		"projectUuid": "proj-1",
		"name": "Dev",
		"type": "DEFAULT",
		"schedulerTimezone": "UTC",
		"upstreamProjectUuid": "proj-upstream"
	}`
	var results GetProjectV1Results
	if err := json.Unmarshal([]byte(payload), &results); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if results.UpstreamProjectUUID == nil || *results.UpstreamProjectUUID != "proj-upstream" {
		t.Fatalf("upstreamProjectUuid = %v, want proj-upstream", results.UpstreamProjectUUID)
	}
}

func TestGetProjectV1Results_JSON_withoutUpstream(t *testing.T) {
	const payload = `{
		"organizationUuid": "org-1",
		"projectUuid": "proj-1",
		"name": "Prod",
		"type": "DEFAULT",
		"schedulerTimezone": "UTC"
	}`
	var results GetProjectV1Results
	if err := json.Unmarshal([]byte(payload), &results); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if results.UpstreamProjectUUID != nil {
		t.Fatalf("upstreamProjectUuid = %v, want nil", results.UpstreamProjectUUID)
	}
}

func TestGetProjectV1Results_JSON_nullUpstream(t *testing.T) {
	const payload = `{
		"organizationUuid": "org-1",
		"projectUuid": "proj-1",
		"name": "Dev",
		"type": "DEFAULT",
		"schedulerTimezone": "UTC",
		"upstreamProjectUuid": null
	}`
	var results GetProjectV1Results
	if err := json.Unmarshal([]byte(payload), &results); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if results.UpstreamProjectUUID != nil {
		t.Fatalf("upstreamProjectUuid = %v, want nil", results.UpstreamProjectUUID)
	}
}
