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

type projectMemberServiceV2 struct {
	client     *api.Client
	roleMapper *RoleMappingService
}

func NewProjectMemberServiceV2(client *api.Client) ProjectMemberService {
	return &projectMemberServiceV2{
		client:     client,
		roleMapper: GetRoleMappingService(client),
	}
}

func (s *projectMemberServiceV2) GrantProjectAccess(ctx context.Context, projectUUID, email string, role models.ProjectMemberRole, sendEmail bool) error {
	roleUUID, err := s.roleMapper.GetRoleUUID(ctx, role.String())
	if err != nil {
		return err
	}

	_, err = apiv2.GrantProjectMemberV2(s.client, projectUUID, email, roleUUID, sendEmail)
	return err
}

func (s *projectMemberServiceV2) UpdateProjectAccess(ctx context.Context, projectUUID, userUUID string, role models.ProjectMemberRole) error {
	roleUUID, err := s.roleMapper.GetRoleUUID(ctx, role.String())
	if err != nil {
		return err
	}

	_, err = apiv2.UpdateProjectMemberV2(s.client, projectUUID, userUUID, roleUUID)
	return err
}

func (s *projectMemberServiceV2) RevokeProjectAccess(ctx context.Context, projectUUID, userUUID string) error {
	return apiv2.RevokeProjectMemberV2(s.client, projectUUID, userUUID)
}

func (s *projectMemberServiceV2) GetProjectMembers(ctx context.Context, projectUUID string) ([]ProjectMember, error) {
	results, err := apiv2.GetProjectMembersV2(s.client, projectUUID)
	if err != nil {
		return nil, err
	}

	members := make([]ProjectMember, len(results))
	for i, r := range results {
		members[i] = ProjectMember{
			UserUUID: r.UserUUID,
			Email:    r.Email,
			Role:     models.ProjectMemberRole(r.RoleName),
		}
	}
	return members, nil
}
