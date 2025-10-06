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
			AgentUUID:             agent.UUID,
			OrganizationUUID:      agent.OrganizationUUID,
			ProjectUUID:           agent.ProjectUUID,
			Name:                  agent.Name,
			Tags:                  agent.Tags,
			Integrations:          integrations,
			UpdatedAt:             agent.UpdatedAt,
			CreatedAt:             agent.CreatedAt,
			Instruction:           agent.Instruction,
			ImageURL:              agent.ImageURL,
			EnableDataAccess:      agent.EnableDataAccess,
			GroupAccess:           agent.GroupAccess,
			UserAccess:            agent.UserAccess,
			EnableSelfImprovement: agent.EnableSelfImprovement,
		})
	}

	return results, nil
}

func (s *AgentService) GetAgent(ctx context.Context, projectUuid string, agentUuid string) (*models.Agent, error) {
	tflog.Debug(ctx, "Getting single agent", map[string]interface{}{
		"projectUuid": projectUuid,
		"agentUuid":   agentUuid,
	})

	agent, err := s.client.GetAgentV1(projectUuid, agentUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Convert API integrations to model integrations
	integrations := []models.AgentIntegration{}
	for _, integration := range agent.Integrations {
		integrations = append(integrations, models.AgentIntegration{
			Type:      integration.Type,
			ChannelID: integration.ChannelID,
		})
	}

	result := &models.Agent{
		AgentUUID:             agent.UUID,
		OrganizationUUID:      agent.OrganizationUUID,
		ProjectUUID:           agent.ProjectUUID,
		Name:                  agent.Name,
		Tags:                  agent.Tags,
		Integrations:          integrations,
		UpdatedAt:             agent.UpdatedAt,
		CreatedAt:             agent.CreatedAt,
		Instruction:           agent.Instruction,
		ImageURL:              agent.ImageURL,
		EnableDataAccess:      agent.EnableDataAccess,
		GroupAccess:           agent.GroupAccess,
		UserAccess:            agent.UserAccess,
		EnableSelfImprovement: agent.EnableSelfImprovement,
	}

	return result, nil
}

func (s *AgentService) CreateAgent(ctx context.Context, projectUuid string, name string, instruction *string, imageUrl *string, tags []string, integrations []models.AgentIntegration, groupAccess []string, userAccess []string, enableDataAccess bool, enableSelfImprovement *bool) (*models.Agent, error) {
	tflog.Debug(ctx, "Creating agent", map[string]interface{}{
		"projectUuid":      projectUuid,
		"name":             name,
		"enableDataAccess": enableDataAccess,
	})

	// Convert model integrations to API integrations
	apiIntegrations := []api.AgentIntegration{}
	for _, integration := range integrations {
		apiIntegrations = append(apiIntegrations, api.AgentIntegration{
			Type:      integration.Type,
			ChannelID: integration.ChannelID,
		})
	}

	request := api.CreateAgentV1Request{
		Name:                  name,
		Instruction:           instruction,
		ImageURL:              imageUrl,
		Tags:                  tags,
		Integrations:          apiIntegrations,
		GroupAccess:           groupAccess,
		UserAccess:            userAccess,
		EnableDataAccess:      enableDataAccess,
		EnableSelfImprovement: enableSelfImprovement,
	}

	agent, err := s.client.CreateAgentV1(projectUuid, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Convert API integrations back to model integrations
	modelIntegrations := []models.AgentIntegration{}
	for _, integration := range agent.Integrations {
		modelIntegrations = append(modelIntegrations, models.AgentIntegration{
			Type:      integration.Type,
			ChannelID: integration.ChannelID,
		})
	}

	result := &models.Agent{
		AgentUUID:             agent.UUID,
		OrganizationUUID:      agent.OrganizationUUID,
		ProjectUUID:           agent.ProjectUUID,
		Name:                  agent.Name,
		Tags:                  agent.Tags,
		Integrations:          modelIntegrations,
		UpdatedAt:             agent.UpdatedAt,
		CreatedAt:             agent.CreatedAt,
		Instruction:           agent.Instruction,
		ImageURL:              agent.ImageURL,
		EnableDataAccess:      agent.EnableDataAccess,
		GroupAccess:           agent.GroupAccess,
		UserAccess:            agent.UserAccess,
		EnableSelfImprovement: agent.EnableSelfImprovement,
	}

	return result, nil
}

func (s *AgentService) DeleteAgent(ctx context.Context, projectUuid string, agentUuid string) error {
	tflog.Debug(ctx, "Deleting agent", map[string]interface{}{
		"projectUuid": projectUuid,
		"agentUuid":   agentUuid,
	})

	err := s.client.DeleteAgentV1(projectUuid, agentUuid)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	return nil
}

func (s *AgentService) UpdateAgent(ctx context.Context, projectUuid string, agentUuid string, name *string, instruction *string, imageUrl *string, tags []string, integrations []models.AgentIntegration, groupAccess []string, userAccess []string, enableDataAccess *bool, enableSelfImprovement *bool) (*models.Agent, error) {
	tflog.Debug(ctx, "Updating agent", map[string]interface{}{
		"projectUuid": projectUuid,
		"agentUuid":   agentUuid,
	})

	// Convert model integrations to API integrations
	apiIntegrations := []api.AgentIntegration{}
	for _, integration := range integrations {
		apiIntegrations = append(apiIntegrations, api.AgentIntegration{
			Type:      integration.Type,
			ChannelID: integration.ChannelID,
		})
	}

	request := api.UpdateAgentV1Request{
		UUID:                  agentUuid,
		Name:                  name,
		Instruction:           instruction,
		ImageURL:              imageUrl,
		Tags:                  tags,
		Integrations:          apiIntegrations,
		GroupAccess:           groupAccess,
		UserAccess:            userAccess,
		EnableDataAccess:      enableDataAccess,
		EnableSelfImprovement: enableSelfImprovement,
	}

	agent, err := s.client.UpdateAgentV1(projectUuid, agentUuid, request)
	if err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	// Convert API integrations back to model integrations
	modelIntegrations := []models.AgentIntegration{}
	for _, integration := range agent.Integrations {
		modelIntegrations = append(modelIntegrations, models.AgentIntegration{
			Type:      integration.Type,
			ChannelID: integration.ChannelID,
		})
	}

	result := &models.Agent{
		AgentUUID:             agent.UUID,
		OrganizationUUID:      agent.OrganizationUUID,
		ProjectUUID:           agent.ProjectUUID,
		Name:                  agent.Name,
		Tags:                  agent.Tags,
		Integrations:          modelIntegrations,
		UpdatedAt:             agent.UpdatedAt,
		CreatedAt:             agent.CreatedAt,
		Instruction:           agent.Instruction,
		ImageURL:              agent.ImageURL,
		EnableDataAccess:      agent.EnableDataAccess,
		GroupAccess:           agent.GroupAccess,
		UserAccess:            agent.UserAccess,
		EnableSelfImprovement: agent.EnableSelfImprovement,
	}

	return result, nil
}
