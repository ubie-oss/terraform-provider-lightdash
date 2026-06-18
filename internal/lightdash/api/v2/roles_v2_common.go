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

package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type roleAssignmentResponse struct {
	Results models.RoleAssignment `json:"results"`
	Status  string                `json:"status"`
}

type roleAssignmentListResponse struct {
	Results []models.RoleAssignment `json:"results"`
	Status  string                  `json:"status"`
}

type upsertRoleAssignmentRequest struct {
	RoleID    string `json:"roleId"`
	SendEmail *bool  `json:"sendEmail,omitempty"`
}

type updateRoleAssignmentRequest struct {
	RoleID string `json:"roleId"`
}

func unmarshalRoleAssignmentResponse(body []byte) (*models.RoleAssignment, error) {
	response := roleAssignmentResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role assignment response: %w", err)
	}
	return &response.Results, nil
}

func unmarshalRoleAssignmentListResponse(body []byte) ([]models.RoleAssignment, error) {
	response := roleAssignmentListResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role assignment list response: %w", err)
	}
	return response.Results, nil
}

func requireNonEmpty(value string, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is empty", fieldName)
	}
	return nil
}

func doJSONRequest(c *api.Client, method string, path string, payload any) ([]byte, error) {
	var req *http.Request
	var err error
	if payload != nil {
		marshalled, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", marshalErr)
		}
		req, err = http.NewRequest(method, path, bytes.NewReader(marshalled))
	} else {
		req, err = http.NewRequest(method, path, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return c.DoRequest(req)
}
