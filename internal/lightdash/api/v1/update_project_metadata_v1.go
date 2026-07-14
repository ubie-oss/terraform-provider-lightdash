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

// UpstreamProjectUUID is a pointer so JSON null clears the upstream link.
type UpdateProjectMetadataV1Request struct {
	UpstreamProjectUUID *string `json:"upstreamProjectUuid"`
}

type UpdateProjectMetadataV1Response struct {
	Status string `json:"status"`
}

func UpdateProjectMetadataV1(c *api.Client, projectUuid string, upstreamProjectUuid *string) (*UpdateProjectMetadataV1Response, error) {
	if len(strings.TrimSpace(projectUuid)) == 0 {
		return nil, fmt.Errorf("projectUuid is empty")
	}

	data := UpdateProjectMetadataV1Request{
		UpstreamProjectUUID: upstreamProjectUuid,
	}

	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("impossible to marshal project metadata: %w", err)
	}

	path := fmt.Sprintf("%s/api/v1/projects/%s/metadata", c.HostUrl, projectUuid)
	req, err := http.NewRequest("PATCH", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request for updating project metadata: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request for updating project metadata in project (%s): %w", projectUuid, err)
	}

	response := UpdateProjectMetadataV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body for updating project metadata: %w", err)
	}

	if response.Status != "ok" {
		return nil, fmt.Errorf("unexpected response status: %s", response.Status)
	}

	return &response, nil
}
