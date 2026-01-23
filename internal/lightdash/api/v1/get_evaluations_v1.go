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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type GetEvaluationsV1Results struct {
	EvalUUID    string                     `json:"evalUuid"`
	AgentUUID   string                     `json:"agentUuid"`
	Title       string                     `json:"title"`
	Description *string                    `json:"description,omitempty"`
	CreatedAt   string                     `json:"createdAt"`
	UpdatedAt   string                     `json:"updatedAt"`
	Prompts     []models.EvaluationsPrompt `json:"prompts"`
}

type GetEvaluationsV1Response struct {
	Results GetEvaluationsV1Results `json:"results,omitempty"`
	Status  string                  `json:"status"`
}

func GetEvaluationsV1(c *api.Client, projectUUID string, agentUUID string, evalUUID string) (*GetEvaluationsV1Results, error) {
	path := fmt.Sprintf("%s/api/v1/projects/%s/aiAgents/%s/evaluations/%s", c.HostUrl, projectUUID, agentUUID, evalUUID)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request for evaluations: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing GET request for evaluations: %w", err)
	}

	response := GetEvaluationsV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response for evaluations: %w", err)
	}

	return &response.Results, nil
}
