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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type MoveSpaceV2Request struct {
	Item struct {
		UUID        string `json:"uuid"`
		Type        string `json:"type"`
		ContentType string `json:"contentType"`
	} `json:"item"`
	Action struct {
		Type            string `json:"type"`
		TargetSpaceUUID string `json:"targetSpaceUuid"`
	} `json:"action"`
}

type MoveSpaceV2Response struct {
	Status string `json:"status"`
}

// MoveSpaceV2 moves a space to a new parent space using the v2 API
func (c *Client) MoveSpaceV2(projectUuid string, spaceUuid string, parentSpaceUuid *string) error {
	data := MoveSpaceV2Request{}
	data.Item.UUID = spaceUuid
	data.Item.Type = "space"
	data.Item.ContentType = "space"
	data.Action.Type = "move"
	data.Action.TargetSpaceUUID = *parentSpaceUuid

	marshalled, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("MoveSpaceV2(projectUuid=%s, spaceUuid=%s, parentSpaceUuid=%v): failed to marshal MoveSpaceV2Request: %w", projectUuid, spaceUuid, parentSpaceUuid, err)
	}
	path := fmt.Sprintf("%s/api/v2/content/%s/move", c.HostUrl, projectUuid)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return fmt.Errorf("MoveSpaceV2(projectUuid=%s, spaceUuid=%s, parentSpaceUuid=%v): failed to create new request: %w", projectUuid, spaceUuid, parentSpaceUuid, err)
	}
	// Do request
	body, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("MoveSpaceV2(projectUuid=%s, spaceUuid=%s, parentSpaceUuid=%v): request failed: %w", projectUuid, spaceUuid, parentSpaceUuid, err)
	}
	// Marshal the response
	response := MoveSpaceV2Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("MoveSpaceV2(projectUuid=%s, spaceUuid=%s, parentSpaceUuid=%v): failed to unmarshal response: %w", projectUuid, spaceUuid, parentSpaceUuid, err)
	}
	return nil
}
