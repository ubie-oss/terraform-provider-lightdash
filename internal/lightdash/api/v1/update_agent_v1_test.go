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
	"strings"
	"testing"
)

func TestUpdateAgentV1Request_MarshalJSON_includesEmptySpaceAccess(t *testing.T) {
	req := UpdateAgentV1Request{
		UUID:        "11111111-1111-1111-1111-111111111111",
		Version:     2,
		Tags:        []string{},
		GroupAccess: []string{},
		UserAccess:  []string{},
		SpaceAccess: []string{},
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, `"spaceAccess":[]`) {
		t.Fatalf("expected JSON to contain explicit empty spaceAccess array, got: %s", s)
	}
}
