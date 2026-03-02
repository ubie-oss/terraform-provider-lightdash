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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// ProjectMemberService defines the interface for project member management.
type ProjectMemberService interface {
	GrantProjectAccess(ctx context.Context, projectUUID, email string, role models.ProjectMemberRole, sendEmail bool) error
	UpdateProjectAccess(ctx context.Context, projectUUID, userUUID string, role models.ProjectMemberRole) error
	RevokeProjectAccess(ctx context.Context, projectUUID, userUUID string) error
	GetProjectMembers(ctx context.Context, projectUUID string) ([]ProjectMember, error)
}

// OrganizationMemberService defines the interface for organization member management.
type OrganizationMemberService interface {
	UpdateOrganizationMember(ctx context.Context, userUUID string, role models.OrganizationMemberRole) (*OrganizationMember, error)
	GetOrganizationMemberByUUID(ctx context.Context, userUUID string) (*OrganizationMember, error)
}

// ProjectGroupService defines the interface for project group management.
type ProjectGroupService interface {
	AddProjectAccessToGroup(ctx context.Context, projectUUID, groupUUID string, role models.ProjectMemberRole) error
	UpdateProjectAccessForGroup(ctx context.Context, projectUUID, groupUUID string, role models.ProjectMemberRole) error
	RemoveProjectAccessFromGroup(ctx context.Context, projectUUID, groupUUID string) error
	GetProjectGroupAccesses(ctx context.Context, projectUUID string) ([]ProjectGroupAccess, error)
}

// DTOs for service layer communication

type ProjectMember struct {
	UserUUID string
	Email    string
	Role     models.ProjectMemberRole
}

type OrganizationMember struct {
	UserUUID         string
	OrganizationUUID string
	Email            string
	OrganizationRole models.OrganizationMemberRole
}

type ProjectGroupAccess struct {
	GroupUUID string
	Role      models.ProjectMemberRole
}
