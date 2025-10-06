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

type AgentIntegration struct {
	Type      string `json:"type"`
	ChannelID string `json:"channelId"`
}

type Agent struct {
	AgentUUID             string             `json:"uuid"`
	OrganizationUUID      string             `json:"organizationUuid"`
	ProjectUUID           string             `json:"projectUuid"`
	Name                  string             `json:"name"`
	Tags                  []string           `json:"tags"`
	Integrations          []AgentIntegration `json:"integrations,omitempty"`
	UpdatedAt             string             `json:"updatedAt"`
	CreatedAt             string             `json:"createdAt"`
	Instruction           *string            `json:"instruction"`
	ImageURL              *string            `json:"imageUrl,omitempty"`
	EnableDataAccess      bool               `json:"enableDataAccess"`
	GroupAccess           []string           `json:"groupAccess,omitempty"`
	UserAccess            []string           `json:"userAccess,omitempty"`
	EnableSelfImprovement *bool              `json:"enableSelfImprovement,omitempty"`
}
