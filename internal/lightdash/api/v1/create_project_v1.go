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

// CreateProjectV1Request is the body for POST /api/v1/org/projects.
// dbtConnection is always type none. Warehouse credentials are not sent.
type CreateProjectV1Request struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	DbtVersion    string                 `json:"dbtVersion"`
	DbtConnection ProjectDbtConnectionV1 `json:"dbtConnection"`
}

type CreateProjectV1Results struct {
	HasContentCopy bool                `json:"hasContentCopy"`
	Project        GetProjectV1Results `json:"project"`
}

type CreateProjectV1Response struct {
	Results CreateProjectV1Results `json:"results"`
	Status  string                 `json:"status"`
}

// CreateProjectV1 creates a project with a dbt connection of type none.
// The Lightdash API has no project delete endpoint; callers must not assume this can be undone.
func CreateProjectV1(c *api.Client, name, projectType, dbtVersion string) (*GetProjectV1Results, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("project name is empty")
	}
	if strings.TrimSpace(projectType) == "" {
		return nil, fmt.Errorf("project type is empty")
	}
	if strings.TrimSpace(dbtVersion) == "" {
		return nil, fmt.Errorf("dbt version is empty")
	}

	data := CreateProjectV1Request{
		Name:       name,
		Type:       projectType,
		DbtVersion: dbtVersion,
		DbtConnection: ProjectDbtConnectionV1{
			Type: DbtConnectionTypeNone,
		},
	}

	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create project request: %w", err)
	}

	path := fmt.Sprintf("%s/api/v1/org/projects", c.HostUrl)
	req, err := http.NewRequest(http.MethodPost, path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request for creating project: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request for creating project: %w", err)
	}

	response := CreateProjectV1Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create project response: %w", err)
	}
	if response.Status != "ok" {
		return nil, fmt.Errorf("unexpected response status: %s", response.Status)
	}
	if response.Results.Project.ProjectUUID == "" {
		return nil, fmt.Errorf("project UUID is empty")
	}

	return &response.Results.Project, nil
}
