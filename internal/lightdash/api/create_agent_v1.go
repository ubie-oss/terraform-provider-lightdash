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
)

type CreateAgentV1Request struct {
	ProjectUUID      string             `json:"projectUuid"`
	Name             string             `json:"name"`
	Instruction      *string            `json:"instruction"`
	ImageURL         *string            `json:"imageUrl"`
	Tags             []string           `json:"tags"`
	Integrations     []AgentIntegration `json:"integrations"`
	GroupAccess      []string           `json:"groupAccess"`
	UserAccess       []string           `json:"userAccess"`
	EnableDataAccess bool               `json:"enableDataAccess"`
}

type CreateAgentV1Results struct {
	UUID             string             `json:"uuid"`
	OrganizationUUID string             `json:"organizationUuid"`
	ProjectUUID      string             `json:"projectUuid"`
	Name             string             `json:"name"`
	Tags             []string           `json:"tags,omitempty"`
	Integrations     []AgentIntegration `json:"integrations,omitempty"`
	UpdatedAt        string             `json:"updatedAt"`
	CreatedAt        string             `json:"createdAt"`
	Instruction      *string            `json:"instruction"`
	ImageURL         *string            `json:"imageUrl,omitempty"`
	EnableDataAccess bool               `json:"enableDataAccess,omitempty"`
	GroupAccess      []string           `json:"groupAccess,omitempty"`
	UserAccess       []string           `json:"userAccess,omitempty"`
}

type CreateAgentV1Response struct {
	Results CreateAgentV1Results `json:"results,omitempty"`
	Status  string               `json:"status"`
}

func (c *Client) CreateAgentV1(projectUUID string, request CreateAgentV1Request) (*CreateAgentV1Results, error) {
	// Set the project UUID in the request
	request.ProjectUUID = projectUUID

	marshalled, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CreateAgentV1Request: %w", err)
	}

	path := fmt.Sprintf("%s/api/v1/projects/%s/aiAgents", c.HostUrl, projectUUID)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("error creating POST request for agent: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing POST request for agent: %w", err)
	}

	response := CreateAgentV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response for agent: %w", err)
	}

	// Validate that the agent UUID is present in the response
	if response.Results.UUID == "" {
		return nil, fmt.Errorf("agent UUID is missing in the response")
	}

	return &response.Results, nil
}
