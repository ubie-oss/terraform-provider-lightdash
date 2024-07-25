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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type AddSpaceGroupAccessRequest struct {
	SpaceRole string `json:"spaceRole"`
	GroupUUID string `json:"groupUuid"`
}

type AddSpaceGroupAccessResults struct {
	ProjectUUID string `json:"projectUuid"`
	SpaceUUID   string `json:"spaceUuid"`
	GroupUUID   string `json:"groupUuid"`
	SpaceRole   string `json:"spaceRole"`
}

type AddSpaceGroupAccessResponse struct {
	Results AddSpaceGroupAccessResults `json:"results,omitempty"`
	Status  string                     `json:"status"`
}

func (c *Client) AddSpaceGroupAccessV1(
	projectUuid string, spaceUuid string, groupUuid string, role models.SpaceMemberRole) error {
	// Validate the role
	if !models.IsValidSpaceMemberRole(role.String()) {
		return fmt.Errorf("invalid role: %s", role)
	}

	// Create the request body
	data := AddSpaceGroupAccessRequest{
		SpaceRole: role.String(),
		GroupUUID: groupUuid,
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to marshal request data: %s", err)
	}

	// Create the request
	path := fmt.Sprintf("%s/api/v1/projects/%s/spaces/%s/group/share", c.HostUrl, projectUuid, spaceUuid)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return fmt.Errorf("failed to create new HTTP request: %v", err)
	}

	// Do request
	_, err = c.doRequest(req)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request for project %s, space %s, group %s with role %s: %v", projectUuid, spaceUuid, groupUuid, role.String(), err)
	}

	return nil
}
