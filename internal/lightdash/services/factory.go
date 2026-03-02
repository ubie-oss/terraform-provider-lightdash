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
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

// ServiceRegistry holds all the services used by the provider.
type ServiceRegistry struct {
	ProjectMember      ProjectMemberService
	OrganizationMember OrganizationMemberService
	ProjectGroup       ProjectGroupService
}

// NewServiceRegistry creates a new ServiceRegistry with default (V2) implementations.
func NewServiceRegistry(client *api.Client) *ServiceRegistry {
	return &ServiceRegistry{
		ProjectMember:      NewProjectMemberServiceV2(client),
		OrganizationMember: NewOrganizationMemberServiceV2(client),
		ProjectGroup:       NewProjectGroupServiceV2(client),
	}
}

// NewServiceRegistryV1 creates a new ServiceRegistry with V1 implementations.
func NewServiceRegistryV1(client *api.Client) *ServiceRegistry {
	return &ServiceRegistry{
		ProjectMember:      NewProjectMemberServiceV1(client),
		OrganizationMember: NewOrganizationMemberServiceV1(client),
		ProjectGroup:       NewProjectGroupServiceV1(client),
	}
}
