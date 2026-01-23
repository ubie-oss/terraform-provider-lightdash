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

type CreateEvaluationsV1Request struct {
	Title       string                     `json:"title"`
	Description *string                    `json:"description,omitempty"`
	Prompts     []models.EvaluationsPrompt `json:"prompts"`
}

type CreateEvaluationsV1Results struct {
	EvalUUID string `json:"evalUuid"`
}

type CreateEvaluationsV1Response struct {
	Results CreateEvaluationsV1Results `json:"results,omitempty"`
	Status  string                     `json:"status"`
}

func CreateEvaluationsV1(c *api.Client, projectUUID string, agentUUID string, request CreateEvaluationsV1Request) (*CreateEvaluationsV1Results, error) {
	marshalled, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CreateEvaluationsV1Request: %w", err)
	}

	path := fmt.Sprintf("%s/api/v1/projects/%s/aiAgents/%s/evaluations", c.HostUrl, projectUUID, agentUUID)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("error creating POST request for evaluations: %w", err)
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing POST request for evaluations: %w", err)
	}

	response := CreateEvaluationsV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response for evaluations: %w", err)
	}

	// Validate that the evaluation UUID is present in the response
	if response.Results.EvalUUID == "" {
		return nil, fmt.Errorf("evaluation UUID is missing in the response")
	}

	return &response.Results, nil
}
