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

type projectGroupServiceV1 struct {
	client *api.Client
}

func NewProjectGroupServiceV1(client *api.Client) ProjectGroupService {
	return &projectGroupServiceV1{
		client: client,
	}
}

func (s *projectGroupServiceV1) AddProjectAccessToGroup(ctx context.Context, projectUUID, groupUUID string, role models.ProjectMemberRole) error {
	_, err := apiv1.AddProjectAccessToGroupV1(s.client, projectUUID, groupUUID, role)
	return err
}

func (s *projectGroupServiceV1) UpdateProjectAccessForGroup(ctx context.Context, projectUUID, groupUUID string, role models.ProjectMemberRole) error {
	_, err := apiv1.UpdateProjectAccessForGroupV1(s.client, projectUUID, groupUUID, role)
	return err
}

func (s *projectGroupServiceV1) RemoveProjectAccessFromGroup(ctx context.Context, projectUUID, groupUUID string) error {
	return apiv1.RemoveProjectAccessFromGroupV1(s.client, projectUUID, groupUUID)
}

func (s *projectGroupServiceV1) GetProjectGroupAccesses(ctx context.Context, projectUUID string) ([]ProjectGroupAccess, error) {
	results, err := apiv1.GetProjectGroupAccessesV1(s.client, projectUUID)
	if err != nil {
		return nil, err
	}

	accesses := make([]ProjectGroupAccess, len(results))
	for i, r := range results {
		accesses[i] = ProjectGroupAccess{
			GroupUUID: r.GroupUUID,
			Role:      r.ProjectRole,
		}
	}
	return accesses, nil
}
