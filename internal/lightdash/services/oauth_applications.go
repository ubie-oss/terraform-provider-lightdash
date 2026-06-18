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

	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

var ErrOAuthApplicationNotFound = errors.New("oauth application not found")

type OAuthApplicationsService struct {
	client *api.Client
}

func NewOAuthApplicationsService(client *api.Client) *OAuthApplicationsService {
	return &OAuthApplicationsService{client: client}
}

func (s *OAuthApplicationsService) List(ctx context.Context) ([]apiv1.OAuthClientV1, error) {
	_ = ctx
	clients, err := apiv1.ListOAuthClientsV1(s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to list OAuth applications: %w", err)
	}
	return clients, nil
}

func (s *OAuthApplicationsService) GetByClientID(ctx context.Context, clientID string) (*apiv1.OAuthClientV1, error) {
	_ = ctx
	if strings.TrimSpace(clientID) == "" {
		return nil, fmt.Errorf("client ID is empty")
	}

	client, err := apiv1.GetOAuthClientV1(s.client, clientID)
	if err == nil {
		return client, nil
	}

	if strings.Contains(err.Error(), "status code: 404") {
		clients, listErr := s.List(ctx)
		if listErr != nil {
			return nil, err
		}
		for i := range clients {
			if clients[i].ClientID == clientID {
				return &clients[i], nil
			}
		}
		return nil, fmt.Errorf("%w: client ID %q", ErrOAuthApplicationNotFound, clientID)
	}

	return nil, err
}
