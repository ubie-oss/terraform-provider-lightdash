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
	apiv2 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v2"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type organizationMemberServiceV2 struct {
	client     *api.Client
	roleMapper *RoleMappingService
	orgUUID    string
}

func NewOrganizationMemberServiceV2(client *api.Client) OrganizationMemberService {
	return &organizationMemberServiceV2{
		client:     client,
		roleMapper: GetRoleMappingService(client),
	}
}

func (s *organizationMemberServiceV2) getOrgUUID(ctx context.Context) (string, error) {
	if s.orgUUID == "" {
		user, err := apiv1.GetAuthenticatedUserV1(s.client)
		if err != nil {
			return "", err
		}
		s.orgUUID = user.OrganizationUUID
	}
	return s.orgUUID, nil
}

func (s *organizationMemberServiceV2) UpdateOrganizationMember(ctx context.Context, userUUID string, role models.OrganizationMemberRole) (*OrganizationMember, error) {
	orgUUID, err := s.getOrgUUID(ctx)
	if err != nil {
		return nil, err
	}

	roleUUID, err := s.roleMapper.GetRoleUUID(ctx, role.String())
	if err != nil {
		return nil, err
	}

	res, err := apiv2.UpdateOrganizationMemberV2(s.client, orgUUID, userUUID, roleUUID)
	if err != nil {
		return nil, err
	}

	return &OrganizationMember{
		UserUUID:         res.UserUUID,
		OrganizationUUID: res.OrganizationUUID,
		Email:            res.Email,
		OrganizationRole: role, // V2 response might use UUID/internal name, we map it back to the input role for Terraform compatibility
	}, nil
}

func (s *organizationMemberServiceV2) GetOrganizationMemberByUUID(ctx context.Context, userUUID string) (*OrganizationMember, error) {
	// For Read, we can still use V1 or implement a V2 version if available.
	// Since we need it for Terraform Read, let's use GetOrganizationMemberByUuidV1 which returns the role name.
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
