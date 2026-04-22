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

type ListUserAttributesV1Response struct {
	Results []models.UserAttribute `json:"results"`
	Status  string                 `json:"status"`
}

func (c *Client) ListUserAttributesV1() ([]models.UserAttribute, error) {
	// Create the request
	path := fmt.Sprintf("%s/api/v1/org/attributes", c.HostUrl)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request for user attributes: %v", err)
	}

	// Do the request
	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing GET request for user attributes: %v", err)
	}

	// Parse the response
	response := ListUserAttributesV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response for user attributes: %v, body: %s", err, string(body))
	}

	return response.Results, nil
}

// GetUserAttributeV1 retrieves a single user attribute by UUID by listing all
// and filtering client-side. The Lightdash API only exposes a list endpoint.
// Returns nil if not found.
func (c *Client) GetUserAttributeV1(userAttributeUuid string) (*models.UserAttribute, error) {
	attributes, err := c.ListUserAttributesV1()
	if err != nil {
		return nil, err
	}
	for i := range attributes {
		if attributes[i].UUID == userAttributeUuid {
			return &attributes[i], nil
		}
	}
	return nil, nil
}
