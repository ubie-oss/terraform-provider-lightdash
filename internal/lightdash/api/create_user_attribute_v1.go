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

type CreateUserAttributeV1Response struct {
	Results models.UserAttribute `json:"results"`
	Status  string               `json:"status"`
}

func (c *Client) CreateUserAttributeV1(request *models.CreateUserAttribute) (*models.UserAttribute, error) {
	// Create the request body
	marshalled, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshalling create user attribute request: %v", err)
	}

	// Create the request
	path := fmt.Sprintf("%s/api/v1/org/attributes", c.HostUrl)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("error creating POST request for user attribute: %v", err)
	}

	// Do the request
	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing POST request for user attribute: %v, body: %s", err, string(marshalled))
	}

	// Parse the response
	response := CreateUserAttributeV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response for user attribute: %v, body: %s", err, string(body))
	}

	// Validate that the user attribute UUID is present in the response
	if response.Results.UUID == "" {
		return nil, fmt.Errorf("user attribute UUID is missing in the response")
	}

	return &response.Results, nil
}
