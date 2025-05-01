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
)

type ListSpacesInProjectV1Results struct {
	OrganizationUUID string  `json:"organizationUuid"`
	ProjectUUID      string  `json:"projectUuid"`
	ParentSpaceUUID  *string `json:"parentSpaceUuid,omitempty"`
	SpaceUUID        string  `json:"uuid"`
	SpaceName        string  `json:"name"`
	IsPrivate        bool    `json:"isPrivate"` // nolint: govet
}

type ListSpacesInProjectV1Response struct {
	Results []ListSpacesInProjectV1Results `json:"results,omitempty"`
	Status  string                         `json:"status"`
}

func (c *Client) ListSpacesInProjectV1(projectUuid string) ([]ListSpacesInProjectV1Results, error) {
	path := fmt.Sprintf("%s/api/v1/projects/%s/spaces", c.HostUrl, projectUuid)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request for spaces: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request for spaces: %w", err)
	}

	response := ListSpacesInProjectV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling spaces response: %w", err)
	}

	return response.Results, nil
}
