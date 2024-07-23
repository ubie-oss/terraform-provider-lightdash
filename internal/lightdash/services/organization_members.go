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
	"fmt"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type OrganizationMembersService struct {
	client  *api.Client
	members []api.GetOrganizationMembersV1Results
}

func NewOrganizationMembersService(client *api.Client) *OrganizationMembersService {
	return &OrganizationMembersService{
		client:  client,
		members: []api.GetOrganizationMembersV1Results{},
	}
}

// GetOrganizationMembers retrieves the members of an organization.
// It checks if the members list is already populated to avoid unnecessary API calls.
// If the members list is empty, it fetches the members from the organization using the API client.
func (s *OrganizationMembersService) GetOrganizationMembers() ([]api.GetOrganizationMembersV1Results, error) {
	// Check if the members list is already populated
	if len(s.members) == 0 {
		page := 1
		pageSize := 100
		for {
			// Fetch the members from the organization using the API client
			members, err := s.client.GetOrganizationMembersV1(0, pageSize, page, "")
			if err != nil {
				return nil, err
			}
			if len(members) == 0 {
				break
			}
			s.members = append(s.members, members...)
			page++
		}
	}
	// Return the list of members
	return s.members, nil
}

// GetOrganizationAdmins retrieves the admins of an organization.
// It leverages the GetOrganizationMembers method to fetch all members and filters out non-admins.
func (s *OrganizationMembersService) GetOrganizationMembersByRole(role models.OrganizationMemberRole) ([]api.GetOrganizationMembersV1Results, error) {
	// Retrieve all members of the organization
	allMembers, err := s.GetOrganizationMembers()
	if err != nil {
		return nil, err
	}
	// Initialize a slice to hold members with the specified role
	var admins []api.GetOrganizationMembersV1Results
	for _, member := range allMembers {
		// Check if the member's role matches the specified role
		if member.OrganizationRole == role {
			admins = append(admins, member)
		}
	}
	// Return the filtered list of members with the specified role
	return admins, nil
}

// GetOrganizationMemberByUserUuid retrieves a member of an organization by their UUID.
func (s *OrganizationMembersService) GetOrganizationMemberByUserUuid(userUuid string) (*api.GetOrganizationMembersV1Results, error) {
	// Retrieve all organization members
	organizationMembers, err := s.GetOrganizationMembers()
	if err != nil {
		return nil, err
	}
	// Iterate through all members to find the one with the matching UUID
	for _, member := range organizationMembers {
		if member.UserUUID == userUuid {
			return &member, nil
		}
	}
	// Return an error if no member with the specified UUID is found
	return nil, fmt.Errorf("member with UUID %s not found", userUuid)
}

// Check if a member with the passed user UUID is an admin of the organization.
func (s *OrganizationMembersService) IsOrganizationAdmin(userUuid string) (bool, error) {
	// Retrieve the organization member by their UUID
	member, err := s.GetOrganizationMemberByUserUuid(userUuid)
	if err != nil {
		return false, err
	}
	// Check if the retrieved member's role is ORGANIZATION_ADMIN_ROLE
	if member.OrganizationRole == models.ORGANIZATION_ADMIN_ROLE {
		// If the member is an admin, return true
		return true, nil
	}
	// If the member is not an admin, return false
	return false, nil
}
