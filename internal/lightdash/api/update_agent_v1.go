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

type UpdateAgentV1Request struct {
	UUID                  string             `json:"uuid"`
	Name                  *string            `json:"name,omitempty"`
	Instruction           *string            `json:"instruction,omitempty"`
	ImageURL              *string            `json:"imageUrl,omitempty"`
	Tags                  []string           `json:"tags"`
	Integrations          []AgentIntegration `json:"integrations,omitempty"`
	GroupAccess           []string           `json:"groupAccess"`
	UserAccess            []string           `json:"userAccess"`
	EnableDataAccess      *bool              `json:"enableDataAccess,omitempty"`
	EnableSelfImprovement *bool              `json:"enableSelfImprovement,omitempty"`
	Version               int64              `json:"version"`
}

type UpdateAgentV1Results struct {
	UUID                  string             `json:"uuid"`
	OrganizationUUID      string             `json:"organizationUuid"`
	ProjectUUID           string             `json:"projectUuid"`
	Name                  string             `json:"name"`
	Tags                  []string           `json:"tags,omitempty"`
	Integrations          []AgentIntegration `json:"integrations"`
	UpdatedAt             string             `json:"updatedAt"`
	CreatedAt             string             `json:"createdAt"`
	Instruction           *string            `json:"instruction"`
	ImageURL              *string            `json:"imageUrl,omitempty"`
	EnableDataAccess      bool               `json:"enableDataAccess,omitempty"`
	EnableSelfImprovement bool               `json:"enableSelfImprovement,omitempty"`
	GroupAccess           []string           `json:"groupAccess,omitempty"`
	UserAccess            []string           `json:"userAccess,omitempty"`
}

type UpdateAgentV1Response struct {
	Results UpdateAgentV1Results `json:"results,omitempty"`
	Status  string               `json:"status"`
}

func (c *Client) UpdateAgentV1(projectUUID string, agentUUID string, request UpdateAgentV1Request) (*UpdateAgentV1Results, error) {
	marshalled, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal UpdateAgentV1Request: %w", err)
	}

	path := fmt.Sprintf("%s/api/v1/projects/%s/aiAgents/%s", c.HostUrl, projectUUID, agentUUID)
	req, err := http.NewRequest("PATCH", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("error creating PATCH request for agent: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing PATCH request for agent: %w", err)
	}

	response := UpdateAgentV1Response{}
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
