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
	"log"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type GrantProjectAccessV1Request struct {
	Email       string                   `json:"email"`
	ProjectRole models.ProjectMemberRole `json:"role"`
	SendEmail   bool                     `json:"sendEmail"`
}

func (c *Client) GrantProjectAccessToUserV1(
	projectUuid string, email string, role models.ProjectMemberRole, sendEmail bool) error {
	// Create the request body
	data := GrantProjectAccessV1Request{
		Email:       email,
		ProjectRole: role,
		SendEmail:   sendEmail,
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("impossible to marshall data: %s", err)
	}
	// Create the request
	path := fmt.Sprintf("%s/api/v1/projects/%s/access", c.HostUrl, projectUuid)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return fmt.Errorf("failed to create new request for project access: %w", err)
	}
	// Do request
	_, err = c.doRequest(req)
	if err != nil {
		return fmt.Errorf("request to grant project access failed: %w", err)
	}

	return nil
}
