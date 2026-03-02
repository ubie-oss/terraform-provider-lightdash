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
	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type projectMemberServiceV1 struct {
	client *api.Client
}

func NewProjectMemberServiceV1(client *api.Client) ProjectMemberService {
	return &projectMemberServiceV1{
		client: client,
	}
}

func (s *projectMemberServiceV1) GrantProjectAccess(ctx context.Context, projectUUID, email string, role models.ProjectMemberRole, sendEmail bool) error {
	return apiv1.GrantProjectAccessToUserV1(s.client, projectUUID, email, role, sendEmail)
}

func (s *projectMemberServiceV1) UpdateProjectAccess(ctx context.Context, projectUUID, userUUID string, role models.ProjectMemberRole) error {
	return apiv1.UpdateProjectAccessToUserV1(s.client, projectUUID, userUUID, role)
}

func (s *projectMemberServiceV1) RevokeProjectAccess(ctx context.Context, projectUUID, userUUID string) error {
	return apiv1.RevokeProjectAccessV1(s.client, projectUUID, userUUID)
}

func (s *projectMemberServiceV1) GetProjectMembers(ctx context.Context, projectUUID string) ([]ProjectMember, error) {
	// API v1 doesn't seem to have a direct list project members that returns DTOs easily without more calls.
	// For now, this is used in Read if needed.
	// Implementation can be added if required.
	return nil, nil
}
