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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type AppendEvaluationsV1Request struct {
	Prompts []models.EvaluationsPrompt `json:"prompts"`
}

type AppendEvaluationsV1Results struct {
	EvalUUID    string                     `json:"evalUuid"`
	AgentUUID   string                     `json:"agentUuid"`
	Title       string                     `json:"title"`
	Description *string                    `json:"description,omitempty"`
	CreatedAt   string                     `json:"createdAt"`
	UpdatedAt   string                     `json:"updatedAt"`
	Prompts     []models.EvaluationsPrompt `json:"prompts"`
}

type AppendEvaluationsV1Response struct {
	Results AppendEvaluationsV1Results `json:"results,omitempty"`
	Status  string                     `json:"status"`
}

func AppendEvaluationsV1(c *api.Client, projectUUID string, agentUUID string, evalUUID string, request AppendEvaluationsV1Request) (*AppendEvaluationsV1Results, error) {
	marshalled, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AppendEvaluationsV1Request: %w", err)
	}

	path := fmt.Sprintf("%s/api/v1/projects/%s/aiAgents/%s/evaluations/%s/append", c.HostUrl, projectUUID, agentUUID, evalUUID)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("error creating POST request for append evaluations: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing POST request for append evaluations: %w", err)
	}

	response := AppendEvaluationsV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response for append evaluations: %w", err)
	}

	return &response.Results, nil
}
