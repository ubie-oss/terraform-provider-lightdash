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

type UpdateWarehouseCredentialsV1Request struct {
	Name        string      `json:"name" validate:"required"`
	Credentials interface{} `json:"credentials" validate:"required"`
	Description *string     `json:"description,omitempty"`
}

type UpdateWarehouseCredentialsV1Response struct {
	Results models.WarehouseCredentials `json:"results,omitempty"`
	Status  string                      `json:"status"`
}

func (c *Client) UpdateWarehouseCredentialsV1(uuid string, name string, credentials interface{}, description *string) (*models.WarehouseCredentials, error) {
	// Create the request body
	data := UpdateWarehouseCredentialsV1Request{
		Name:        name,
		Credentials: credentials,
		Description: description,
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	// Create the request
	path := fmt.Sprintf("%s/api/v1/org/warehouse-credentials/%s", c.HostUrl, uuid)
	req, err := http.NewRequest("PATCH", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v, body: %s", err, string(marshalled))
	}
	// Do request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v, body: %s", err, string(marshalled))
	}
	// Marshal the response
	response := UpdateWarehouseCredentialsV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v, body: %s", err, string(body))
	}
	return &response.Results, nil
}
