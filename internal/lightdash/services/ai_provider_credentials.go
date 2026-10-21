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
	"errors"
	"fmt"
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

var ErrAiProviderCredentialNotFound = errors.New("ai provider credential not found")

type AiProviderCredentialsService struct {
	client *api.Client
}

func NewAiProviderCredentialsService(client *api.Client) *AiProviderCredentialsService {
	return &AiProviderCredentialsService{client: client}
}

func (s *AiProviderCredentialsService) List(ctx context.Context) (*models.AiProviderCredentialsList, error) {
	_ = ctx
	list, err := apiv1.ListAiProviderCredentialsV1(s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to list AI provider credentials: %w", err)
	}
	return list, nil
}

func (s *AiProviderCredentialsService) GetByUUID(ctx context.Context, credentialUUID string) (*models.AiProviderCredential, error) {
	_ = ctx
	if strings.TrimSpace(credentialUUID) == "" {
		return nil, fmt.Errorf("credential UUID is empty")
	}

	list, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range list.Credentials {
		if list.Credentials[i].UUID == credentialUUID {
			return &list.Credentials[i], nil
		}
	}

	return nil, fmt.Errorf("%w: credential UUID %q", ErrAiProviderCredentialNotFound, credentialUUID)
}
