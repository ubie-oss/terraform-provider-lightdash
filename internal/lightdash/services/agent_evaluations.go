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

	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type AgentEvaluationsService struct {
	client *api.Client
}

func NewAgentEvaluationsService(client *api.Client) *AgentEvaluationsService {
	return &AgentEvaluationsService{
		client: client,
	}
}

func (s *AgentEvaluationsService) GetEvaluations(ctx context.Context, projectUUID string, agentUUID string, evalUUID string) (*models.AgentEvaluations, error) {
	tflog.Debug(ctx, "Getting evaluation", map[string]interface{}{
		"projectUUID": projectUUID,
		"agentUUID":   agentUUID,
		"evalUUID":    evalUUID,
	})

	evaluation, err := apiv1.GetEvaluationsV1(s.client, projectUUID, agentUUID, evalUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get evaluations: %w", err)
	}

	result := &models.AgentEvaluations{
		EvalUUID:    evaluation.EvalUUID,
		AgentUUID:   evaluation.AgentUUID,
		Title:       evaluation.Title,
		Description: evaluation.Description,
		CreatedAt:   evaluation.CreatedAt,
		UpdatedAt:   evaluation.UpdatedAt,
		Prompts:     evaluation.Prompts,
	}

	return result, nil
}

func (s *AgentEvaluationsService) CreateEvaluations(ctx context.Context, projectUUID string, agentUUID string, title string, description *string, prompts []models.EvaluationsPrompt) (*models.AgentEvaluations, error) {
	tflog.Debug(ctx, "Creating evaluation", map[string]interface{}{
		"projectUUID": projectUUID,
		"agentUUID":   agentUUID,
		"title":       title,
	})

	request := apiv1.CreateEvaluationsV1Request{
		Title:       title,
		Description: description,
		Prompts:     prompts,
	}

	evaluation, err := apiv1.CreateEvaluationsV1(s.client, projectUUID, agentUUID, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create evaluations: %w", err)
	}

	// Get the full evaluation details after creation
	return s.GetEvaluations(ctx, projectUUID, agentUUID, evaluation.EvalUUID)
}

func (s *AgentEvaluationsService) UpdateEvaluations(ctx context.Context, projectUUID string, agentUUID string, evalUUID string, title *string, description *string, prompts []models.EvaluationsPrompt) (*models.AgentEvaluations, error) {
	tflog.Debug(ctx, "Updating evaluation", map[string]interface{}{
		"projectUUID": projectUUID,
		"agentUUID":   agentUUID,
		"evalUUID":    evalUUID,
	})

	request := apiv1.UpdateEvaluationsV1Request{
		Title:       title,
		Description: description,
		Prompts:     prompts,
	}

	_, err := apiv1.UpdateEvaluationsV1(s.client, projectUUID, agentUUID, evalUUID, request)
	if err != nil {
		return nil, fmt.Errorf("failed to update evaluations: %w", err)
	}

	// Get the updated evaluation details
	return s.GetEvaluations(ctx, projectUUID, agentUUID, evalUUID)
}

func (s *AgentEvaluationsService) AppendEvaluations(ctx context.Context, projectUUID string, agentUUID string, evalUUID string, prompts []models.EvaluationsPrompt) (*models.AgentEvaluations, error) {
	tflog.Debug(ctx, "Appending to evaluation", map[string]interface{}{
		"projectUUID": projectUUID,
		"agentUUID":   agentUUID,
		"evalUUID":    evalUUID,
	})

	request := apiv1.AppendEvaluationsV1Request{
		Prompts: prompts,
	}

	_, err := apiv1.AppendEvaluationsV1(s.client, projectUUID, agentUUID, evalUUID, request)
	if err != nil {
		return nil, fmt.Errorf("failed to append to evaluations: %w", err)
	}

	// Get the updated evaluation details
	return s.GetEvaluations(ctx, projectUUID, agentUUID, evalUUID)
}

func (s *AgentEvaluationsService) DeleteEvaluations(ctx context.Context, projectUUID string, agentUUID string, evalUUID string) error {
	tflog.Debug(ctx, "Deleting evaluation", map[string]interface{}{
		"projectUUID": projectUUID,
		"agentUUID":   agentUUID,
		"evalUUID":    evalUUID,
	})

	err := apiv1.DeleteEvaluationsV1(s.client, projectUUID, agentUUID, evalUUID)
	if err != nil {
		return fmt.Errorf("failed to delete evaluations: %w", err)
	}

	return nil
}
