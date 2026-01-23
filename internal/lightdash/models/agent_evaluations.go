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

package models

type EvaluationsPrompt struct {
	EvalPromptUUID   string `json:"evalPromptUuid,omitempty"`
	CreatedAt        string `json:"createdAt,omitempty"`
	Type             string `json:"type,omitempty"`
	Prompt           string `json:"prompt"`
	ExpectedResponse string `json:"expectedResponse"`
	ThreadUUID       string `json:"threadUuid,omitempty"`
	PromptUUID       string `json:"promptUuid,omitempty"`
}

type AgentEvaluations struct {
	EvalUUID    string
	AgentUUID   string
	Title       string
	Description *string
	CreatedAt   string
	UpdatedAt   string
	Prompts     []EvaluationsPrompt
}
