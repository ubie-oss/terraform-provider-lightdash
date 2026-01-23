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
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

type UpdateSpaceV1Request struct {
	Name      string `json:"name"`
	IsPrivate *bool  `json:"isPrivate,omitempty"`
}

type UpdateSpaceV1Results struct {
	OrganizationUUID string  `json:"organizationUuid"`
	ProjectUUID      string  `json:"projectUuid"`
	ParentSpaceUUID  *string `json:"parentSpaceUuid,omitempty"`
	SpaceUUID        string  `json:"uuid"`
	SpaceName        string  `json:"name"`
	IsPrivate        bool    `json:"isPrivate"`
}

type UpdateSpaceV1Response struct {
	Results UpdateSpaceV1Results `json:"results,omitempty"`
	Status  string               `json:"status"`
}

func UpdateSpaceV1(c *api.Client, _ context.Context, projectUuid string, spaceUuid string, spaceName string, isPrivate *bool) (*UpdateSpaceV1Results, error) {
	// Create the request body, including parentSpaceUuid if provided
	data := UpdateSpaceV1Request{
		Name: spaceName,
	}
	if isPrivate != nil {
		data.IsPrivate = isPrivate
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal UpdateSpaceV1Request (data=%#v): %w",
			data, err,
		)
	}
	// Create the request
	path := fmt.Sprintf("%s/api/v1/projects/%s/spaces/%s", c.HostUrl, projectUuid, spaceUuid)
	req, err := http.NewRequest("PATCH", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create new request (data=%#v): %w",
			data, err,
		)
	}

	// Do request
	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf(
			"request failed (data=%#v): %w",
			data, err,
		)
	}

	// Marshal the response
	response := UpdateSpaceV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal response (data=%#v): %w",
			data, err,
		)
	}
	// Make sure if the space UUID is not empty
	if response.Results.SpaceUUID == "" {
		return nil, fmt.Errorf(
			"space UUID is nil (data=%#v)",
			data,
		)
	}
	return &response.Results, nil
}
