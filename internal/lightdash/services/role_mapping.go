// Copyright 2024 Ubie, inc.
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
	"fmt"
	"strings"
	"sync"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"
	apiv2 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v2"
)

var (
	roleMappingServiceInstance *RoleMappingService
	roleMappingOnce            sync.Once
)

type RoleMappingService struct {
	client        *api.Client
	roleMap       map[string]string // roleName/slug -> roleUUID
	orgUUID       string
	isInitialized bool
	mu            sync.RWMutex
}

func GetRoleMappingService(client *api.Client) *RoleMappingService {
	roleMappingOnce.Do(func() {
		roleMappingServiceInstance = &RoleMappingService{
			client:  client,
			roleMap: make(map[string]string),
		}
	})
	return roleMappingServiceInstance
}

func (s *RoleMappingService) initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isInitialized {
		return nil
	}

	// Fetch orgUUID if not already set
	if s.orgUUID == "" {
		user, err := apiv1.GetAuthenticatedUserV1(s.client)
		if err != nil {
			return fmt.Errorf("error getting authenticated user for role mapping: %w", err)
		}
		s.orgUUID = user.OrganizationUUID
	}

	// Fetch roles from API v2
	roles, err := apiv2.GetOrganizationRolesV2(s.client, s.orgUUID)
	if err != nil {
		return fmt.Errorf("error listing organization roles for mapping: %w", err)
	}

	for _, r := range roles {
		// Map both Name and Slug for robustness
		s.roleMap[strings.ToLower(r.Name)] = r.UUID
		if r.Slug != "" {
			s.roleMap[strings.ToLower(r.Slug)] = r.UUID
		}
	}

	s.isInitialized = true
	return nil
}

func (s *RoleMappingService) GetRoleUUID(ctx context.Context, roleName string) (string, error) {
	s.mu.RLock()
	if !s.isInitialized {
		s.mu.RUnlock()
		if err := s.initialize(ctx); err != nil {
			return "", err
		}
		s.mu.RLock()
	}

	uuid, ok := s.roleMap[strings.ToLower(roleName)]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("role name/slug '%s' not found in organization roles", roleName)
	}

	return uuid, nil
}
