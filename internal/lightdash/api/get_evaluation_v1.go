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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type GetEvaluationV1Results struct {
	EvalUUID    string                    `json:"evalUuid"`
	AgentUUID   string                    `json:"agentUuid"`
	Title       string                    `json:"title"`
	Description *string                   `json:"description,omitempty"`
	CreatedAt   string                    `json:"createdAt"`
	UpdatedAt   string                    `json:"updatedAt"`
	Prompts     []models.EvaluationPrompt `json:"prompts"`
}

type GetEvaluationV1Response struct {
	Results GetEvaluationV1Results `json:"results,omitempty"`
	Status  string                 `json:"status"`
}

func (c *Client) GetEvaluationV1(projectUUID string, agentUUID string, evalUUID string) (*GetEvaluationV1Results, error) {
	path := fmt.Sprintf("%s/api/v1/projects/%s/aiAgents/%s/evaluations/%s", c.HostUrl, projectUUID, agentUUID, evalUUID)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request for evaluation: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing GET request for evaluation: %w", err)
	}

	response := GetEvaluationV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response for evaluation: %w", err)
	}

	return &response.Results, nil
}
