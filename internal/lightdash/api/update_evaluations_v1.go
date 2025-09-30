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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type UpdateEvaluationsV1Request struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Prompts     []string `json:"prompts,omitempty"`
}

type UpdateEvaluationsV1Results struct {
	EvalUUID    string                     `json:"evalUuid"`
	AgentUUID   string                     `json:"agentUuid"`
	Title       string                     `json:"title"`
	Description *string                    `json:"description,omitempty"`
	CreatedAt   string                     `json:"createdAt"`
	UpdatedAt   string                     `json:"updatedAt"`
	Prompts     []models.EvaluationsPrompt `json:"prompts"`
}

type UpdateEvaluationsV1Response struct {
	Results UpdateEvaluationsV1Results `json:"results,omitempty"`
	Status  string                     `json:"status"`
}

func (c *Client) UpdateEvaluationsV1(projectUUID string, agentUUID string, evalUUID string, request UpdateEvaluationsV1Request) (*UpdateEvaluationsV1Results, error) {
	marshalled, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal UpdateEvaluationsV1Request: %w", err)
	}

	path := fmt.Sprintf("%s/api/v1/projects/%s/aiAgents/%s/evaluations/%s", c.HostUrl, projectUUID, agentUUID, evalUUID)
	req, err := http.NewRequest("PATCH", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("error creating PATCH request for evaluations: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing PATCH request for evaluations: %w", err)
	}

	response := UpdateEvaluationsV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response for evaluations: %w", err)
	}

	return &response.Results, nil
}
