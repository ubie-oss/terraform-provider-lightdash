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

type organizationMemberServiceV1 struct {
	client *api.Client
}

func NewOrganizationMemberServiceV1(client *api.Client) OrganizationMemberService {
	return &organizationMemberServiceV1{
		client: client,
	}
}

func (s *organizationMemberServiceV1) UpdateOrganizationMember(ctx context.Context, userUUID string, role models.OrganizationMemberRole) (*OrganizationMember, error) {
	res, err := apiv1.UpdateOrganizationMemberV1(s.client, userUUID, role)
	if err != nil {
		return nil, err
	}
	return &OrganizationMember{
		UserUUID:         res.UserUUID,
		OrganizationUUID: res.OrganizationUUID,
		Email:            res.Email,
		OrganizationRole: res.OrganizationRole,
	}, nil
}

func (s *organizationMemberServiceV1) GetOrganizationMemberByUUID(ctx context.Context, userUUID string) (*OrganizationMember, error) {
	res, err := apiv1.GetOrganizationMemberByUuidV1(s.client, userUUID)
	if err != nil {
		return nil, err
	}
	return &OrganizationMember{
		UserUUID:         res.UserUUID,
		OrganizationUUID: res.OrganizationUUID,
		Email:            res.Email,
		OrganizationRole: res.OrganizationRole,
	}, nil
}
