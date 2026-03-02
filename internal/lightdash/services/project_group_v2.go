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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv2 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v2"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type projectGroupServiceV2 struct {
	client     *api.Client
	roleMapper *RoleMappingService
}

func NewProjectGroupServiceV2(client *api.Client) ProjectGroupService {
	return &projectGroupServiceV2{
		client:     client,
		roleMapper: GetRoleMappingService(client),
	}
}

func (s *projectGroupServiceV2) AddProjectAccessToGroup(ctx context.Context, projectUUID, groupUUID string, role models.ProjectMemberRole) error {
	roleUUID, err := s.roleMapper.GetRoleUUID(ctx, role.String())
	if err != nil {
		return err
	}

	_, err = apiv2.AddProjectGroupV2(s.client, projectUUID, groupUUID, roleUUID)
	return err
}

func (s *projectGroupServiceV2) UpdateProjectAccessForGroup(ctx context.Context, projectUUID, groupUUID string, role models.ProjectMemberRole) error {
	roleUUID, err := s.roleMapper.GetRoleUUID(ctx, role.String())
	if err != nil {
		return err
	}

	_, err = apiv2.UpdateProjectGroupV2(s.client, projectUUID, groupUUID, roleUUID)
	return err
}

func (s *projectGroupServiceV2) RemoveProjectAccessFromGroup(ctx context.Context, projectUUID, groupUUID string) error {
	return apiv2.RemoveProjectGroupV2(s.client, projectUUID, groupUUID)
}

func (s *projectGroupServiceV2) GetProjectGroupAccesses(ctx context.Context, projectUUID string) ([]ProjectGroupAccess, error) {
	results, err := apiv2.GetProjectGroupAccessesV2(s.client, projectUUID)
	if err != nil {
		return nil, err
	}

	accesses := make([]ProjectGroupAccess, len(results))
	for i, r := range results {
		accesses[i] = ProjectGroupAccess{
			GroupUUID: r.GroupUUID,
			Role:      models.ProjectMemberRole(r.RoleName),
		}
	}
	return accesses, nil
}
