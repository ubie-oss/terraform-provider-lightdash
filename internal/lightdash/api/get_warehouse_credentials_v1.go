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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type GetWarehouseCredentialsV1Response struct {
	Results models.WarehouseCredentials `json:"results,omitempty"`
	Status  string                      `json:"status"`
}

func (c *Client) GetWarehouseCredentialsV1(uuid string) (*models.WarehouseCredentials, error) {
	// Create the request
	path := fmt.Sprintf("%s/api/v1/org/warehouse-credentials/%s", c.HostUrl, uuid)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}
	// Do request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	// Marshal the response
	response := GetWarehouseCredentialsV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v, body: %s", err, string(body))
	}
	return &response.Results, nil
}
