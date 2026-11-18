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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

// CreateUserAttributeV1Request is the body of POST and PUT /api/v1/org/attributes.
// `groups` and `users` are required by the API, including when they are empty.
// A nil AttributeDefaults is sent as null, which means the attribute has no default.
type CreateUserAttributeV1Request struct {
	Name              string                       `json:"name"`
	Description       string                       `json:"description"`
	AttributeDefaults *[]string                    `json:"attributeDefaults"`
	Groups            []CreateUserAttributeGroupV1 `json:"groups"`
	Users             []CreateUserAttributeUserV1  `json:"users"`
}

type CreateUserAttributeGroupV1 struct {
	GroupUUID string   `json:"groupUuid"`
	Values    []string `json:"values"`
}

type CreateUserAttributeUserV1 struct {
	UserUUID string   `json:"userUuid"`
	Values   []string `json:"values"`
}

type userAttributeWriteV1Response struct {
	Results UserAttributeV1 `json:"results"`
	Status  string          `json:"status"`
}

func CreateUserAttributeV1(c *api.Client, request CreateUserAttributeV1Request) (*UserAttributeV1, error) {
	if strings.TrimSpace(request.Name) == "" {
		return nil, fmt.Errorf("user attribute name is empty")
	}
	return writeUserAttributeV1(c, http.MethodPost, fmt.Sprintf("%s/api/v1/org/attributes", c.HostUrl), request)
}

func writeUserAttributeV1(c *api.Client, method string, path string, request CreateUserAttributeV1Request) (*UserAttributeV1, error) {
	if request.Groups == nil {
		request.Groups = []CreateUserAttributeGroupV1{}
	}
	if request.Users == nil {
		request.Users = []CreateUserAttributeUserV1{}
	}

	marshalled, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user attribute request: %w", err)
	}

	req, err := http.NewRequest(method, path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create user attribute request: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("user attribute request failed: %w", err)
	}

	var response userAttributeWriteV1Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user attribute response: %w", err)
	}
	if response.Status != "ok" {
		return nil, fmt.Errorf("user attribute response status %q", response.Status)
	}
	if response.Results.UUID == "" {
		return nil, fmt.Errorf("user attribute UUID is empty in the response")
	}
	return &response.Results, nil
}
