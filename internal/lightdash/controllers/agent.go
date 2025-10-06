// Copyright 2025 Ubie, inc.
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

package controllers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// AgentController orchestrates operations related to Lightdash project agents.
type AgentController struct {
	agentService *services.AgentService
}

// NewAgentController creates a new AgentController.
func NewAgentController(client *api.Client) *AgentController {
	return &AgentController{
		agentService: services.NewAgentService(client),
	}
}

// CreateAgentOptions contains all the options for creating an agent.
type CreateAgentOptions struct {
	ProjectUUID           string
	Name                  string
	Instruction           *string
	ImageURL              *string
	Tags                  []string
	Integrations          []models.AgentIntegration
	GroupAccess           []string
	UserAccess            []string
	EnableDataAccess      bool
	EnableSelfImprovement *bool
}

// UpdateAgentOptions contains all the options for updating an agent.
type UpdateAgentOptions struct {
	ProjectUUID           string
	AgentUUID             string
	Name                  *string
	Instruction           *string
	ImageURL              *string
	Tags                  []string
	Integrations          []models.AgentIntegration
	GroupAccess           []string
	UserAccess            []string
	EnableDataAccess      *bool
	EnableSelfImprovement *bool
}

// DeleteAgentOptions contains all the options for deleting an agent.
type DeleteAgentOptions struct {
	ProjectUUID        string
	AgentUUID          string
	DeletionProtection bool
}

// GetAgent retrieves the details of an agent by its project and agent UUIDs.
func (c *AgentController) GetAgent(ctx context.Context, projectUUID, agentUUID string) (*models.Agent, error) {
	tflog.Debug(ctx, "(AgentController.GetAgent) Getting agent", map[string]interface{}{
		"projectUUID": projectUUID,
		"agentUUID":   agentUUID,
	})
	return c.agentService.GetAgent(ctx, projectUUID, agentUUID)
}

// CreateAgent creates a new agent with the specified properties.
func (c *AgentController) CreateAgent(ctx context.Context, options CreateAgentOptions) (*models.Agent, error) {
	tflog.Debug(ctx, "(AgentController.CreateAgent) Creating agent", map[string]interface{}{
		"options": options,
	})

	agent, err := c.agentService.CreateAgent(
		ctx,
		options.ProjectUUID,
		options.Name,
		options.Instruction,
		options.ImageURL,
		options.Tags,
		options.Integrations,
		options.GroupAccess,
		options.UserAccess,
		options.EnableDataAccess,
		options.EnableSelfImprovement,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
}

// UpdateAgent updates an existing agent.
func (c *AgentController) UpdateAgent(ctx context.Context, options UpdateAgentOptions) (*models.Agent, error) {
	tflog.Debug(ctx, "(AgentController.UpdateAgent) Updating agent", map[string]interface{}{
		"options": options,
	})

	agent, err := c.agentService.UpdateAgent(
		ctx,
		options.ProjectUUID,
		options.AgentUUID,
		options.Name,
		options.Instruction,
		options.ImageURL,
		options.Tags,
		options.Integrations,
		options.GroupAccess,
		options.UserAccess,
		options.EnableDataAccess,
		options.EnableSelfImprovement,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}
	return agent, nil
}

// DeleteAgent deletes an agent if deletion protection is disabled.
func (c *AgentController) DeleteAgent(ctx context.Context, options DeleteAgentOptions) error {
	tflog.Debug(ctx, "(AgentController.DeleteAgent) Deleting agent", map[string]interface{}{
		"options": options,
	})

	if options.DeletionProtection {
		return fmt.Errorf("cannot delete agent %s: deletion protection is enabled", options.AgentUUID)
	}

	return c.agentService.DeleteAgent(ctx, options.ProjectUUID, options.AgentUUID)
}
