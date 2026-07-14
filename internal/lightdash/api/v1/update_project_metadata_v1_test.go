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

func TestUpdateProjectMetadataV1Request_JSON_setUpstream(t *testing.T) {
	upstream := "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"
	req := UpdateProjectMetadataV1Request{
		UpstreamProjectUUID: &upstream,
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"upstreamProjectUuid":"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"}`
	if string(b) != want {
		t.Fatalf("json mismatch\n got:  %s\n want: %s", string(b), want)
	}
}

func TestUpdateProjectMetadataV1Request_JSON_clearUpstream(t *testing.T) {
	req := UpdateProjectMetadataV1Request{
		UpstreamProjectUUID: nil,
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"upstreamProjectUuid":null}`
	if string(b) != want {
		t.Fatalf("json mismatch\n got:  %s\n want: %s", string(b), want)
	}
}

func TestUpdateProjectMetadataV1Response_JSON_ok(t *testing.T) {
	const payload = `{"status":"ok"}`
	var response UpdateProjectMetadataV1Response
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if response.Status != "ok" {
		t.Fatalf("status = %q, want ok", response.Status)
	}
}
