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

const (
	AiProviderBedrock = "bedrock"
)

const (
	BedrockInferenceGeographyJP     = "jp"
	BedrockInferenceGeographyUS     = "us"
	BedrockInferenceGeographyEU     = "eu"
	BedrockInferenceGeographyAPAC   = "apac"
	BedrockInferenceGeographyGlobal = "global"
)

var ValidBedrockInferenceGeographies = []string{
	BedrockInferenceGeographyJP,
	BedrockInferenceGeographyUS,
	BedrockInferenceGeographyEU,
	BedrockInferenceGeographyAPAC,
	BedrockInferenceGeographyGlobal,
}

// AiProviderCredential is a named org credential from GET /api/v1/ai/provider-credentials.
type AiProviderCredential struct {
	UUID               string   `json:"uuid"`
	Provider           string   `json:"provider"`
	Label              string   `json:"label"`
	Region             string   `json:"region"`
	AllowedModels      []string `json:"allowedModels"`
	APIKeyHint         string   `json:"apiKeyHint"`
	IsDefault          bool     `json:"isDefault"`
	InferenceGeography string   `json:"inferenceGeography"`
}

// UnreadableAiProviderCredential is a credential row that could not be decrypted.
type UnreadableAiProviderCredential struct {
	UUID  string `json:"uuid"`
	Label string `json:"label"`
}

// LegacyBedrockCredential is pre-named-credential Bedrock config surfaced in the list API.
type LegacyBedrockCredential struct {
	APIKeyHint    string   `json:"apiKeyHint"`
	AllowedModels []string `json:"allowedModels"`
	Region        string   `json:"region"`
}

// AiProviderCredentialsList is the results object from GET /api/v1/ai/provider-credentials.
type AiProviderCredentialsList struct {
	Credentials           []AiProviderCredential           `json:"credentials"`
	UnreadableCredentials []UnreadableAiProviderCredential `json:"unreadableCredentials"`
	LegacyBedrock         *LegacyBedrockCredential         `json:"legacyBedrock"`
}

// CreateAiProviderCredentialRequest is the POST body for creating a credential.
type CreateAiProviderCredentialRequest struct {
	Provider           string   `json:"provider"`
	Label              string   `json:"label"`
	Region             string   `json:"region"`
	AllowedModels      []string `json:"allowedModels"`
	APIKey             string   `json:"apiKey"` // #nosec G117 -- Lightdash BYOK credential payload
	InferenceGeography *string  `json:"inferenceGeography,omitempty"`
}

// UpdateAiProviderCredentialRequest is the PATCH body for updating a credential.
type UpdateAiProviderCredentialRequest struct {
	Label              *string  `json:"label,omitempty"`
	Region             *string  `json:"region,omitempty"`
	AllowedModels      []string `json:"allowedModels,omitempty"`
	APIKey             *string  `json:"apiKey,omitempty"` // #nosec G117 -- Lightdash BYOK credential payload
	InferenceGeography *string  `json:"inferenceGeography,omitempty"`
}

// ProjectAiCredentialSelection is the project-level credential binding.
type ProjectAiCredentialSelection struct {
	CredentialUUID *string `json:"credentialUuid"`
}
