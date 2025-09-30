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

package services

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type AgentService struct {
	client *api.Client
}

func NewAgentService(client *api.Client) *AgentService {
	return &AgentService{
		client: client,
	}
}

func (s *AgentService) GetAllAgents(ctx context.Context) ([]models.Agent, error) {
	tflog.Debug(ctx, "Getting all agents")

	agents, err := s.client.GetAllAgentsV1()
	if err != nil {
		return nil, fmt.Errorf("failed to get all agents: %w", err)
	}

	results := []models.Agent{}
	for _, agent := range agents {
		// Convert API integrations to model integrations
		integrations := []models.AgentIntegration{}
		for _, integration := range agent.Integrations {
			integrations = append(integrations, models.AgentIntegration{
				Type:      integration.Type,
				ChannelID: integration.ChannelID,
			})
		}

		results = append(results, models.Agent{
			AgentUUID:        agent.UUID,
			OrganizationUUID: agent.OrganizationUUID,
			ProjectUUID:      agent.ProjectUUID,
			Name:             agent.Name,
			Tags:             agent.Tags,
			Integrations:     integrations,
			UpdatedAt:        agent.UpdatedAt,
			CreatedAt:        agent.CreatedAt,
			Instruction:      agent.Instruction,
			ImageURL:         agent.ImageURL,
			EnableDataAccess: agent.EnableDataAccess,
			GroupAccess:      agent.GroupAccess,
			UserAccess:       agent.UserAccess,
		})
	}

	return results, nil
}
