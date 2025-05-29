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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type ProjectService struct {
	client *api.Client
}

func NewProjectService(client *api.Client) *ProjectService {
	return &ProjectService{client: client}
}

// TODO refactoring the returned data type
func (s *ProjectService) GetProjectMembers(ctx context.Context, projectUuid string) ([]models.ProjectMember, error) {
	apiMembers, err := s.client.GetProjectAccessListV1(projectUuid)
	if err != nil {
		return nil, err
	}

	members := make([]models.ProjectMember, len(apiMembers))
	for i, apiMember := range apiMembers {
		members[i] = models.ProjectMember{
			ProjectUUID: apiMember.ProjectUUID,
			UserUUID:    apiMember.UserUUID,
			Email:       &apiMember.Email,
			ProjectRole: apiMember.ProjectRole,
		}
	}

	return members, nil
}
