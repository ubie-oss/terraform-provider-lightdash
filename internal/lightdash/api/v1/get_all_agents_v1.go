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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

type AgentIntegration struct {
	Type      string `json:"type"`
	ChannelID string `json:"channelId,omitempty"`
}

type GetAllAgentsV1Result struct {
	UUID                  string             `json:"uuid"`
	OrganizationUUID      string             `json:"organizationUuid"`
	ProjectUUID           string             `json:"projectUuid"`
	Name                  string             `json:"name"`
	Tags                  []string           `json:"tags,omitempty"`
	Integrations          []AgentIntegration `json:"integrations,omitempty"`
	UpdatedAt             string             `json:"updatedAt"`
	CreatedAt             string             `json:"createdAt"`
	Instruction           *string            `json:"instruction"`
	ImageURL              *string            `json:"imageUrl,omitempty"`
	EnableDataAccess      bool               `json:"enableDataAccess,omitempty"`
	EnableSelfImprovement bool               `json:"enableSelfImprovement,omitempty"`
	GroupAccess           []string           `json:"groupAccess,omitempty"`
	UserAccess            []string           `json:"userAccess,omitempty"`
	Description           string             `json:"description"`
	SpaceAccess           []string           `json:"spaceAccess"`
	EnableReasoning       bool               `json:"enableReasoning"`
}

type GetAllAgentsV1Response struct {
	Results []GetAllAgentsV1Result `json:"results,omitempty"`
	Status  string                 `json:"status"`
}

func GetAllAgentsV1(c *api.Client) ([]GetAllAgentsV1Result, error) {
	path := fmt.Sprintf("%s/api/v1/aiAgents/admin/agents", c.HostUrl)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request for all agents: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing GET request for all agents: %w", err)
	}

	response := GetAllAgentsV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response for all agents: %w", err)
	}

	return response.Results, nil
}
