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
	"slices"
	"sort"
	"sync"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// Package-level variables for the singleton instance and sync.Once
var organizationMembersServiceInstance *OrganizationMembersService
var once sync.Once

type OrganizationMembersService struct {
	client *api.Client
	// members are cached results from GetOrganizationMembers
	members []api.GetOrganizationMembersV1Results
}

// GetOrganizationMembersService returns the singleton instance of OrganizationMembersService.
// It initializes the instance the first first time it is called in a thread-safe manner.
// We don't add and remove members in the terraform provider at the moment.
// So, we can cache the results of GetOrganizationMembers.
func GetOrganizationMembersService(client *api.Client) *OrganizationMembersService {
	once.Do(func() {
		organizationMembersServiceInstance = &OrganizationMembersService{
			client:  client,
			members: []api.GetOrganizationMembersV1Results{}, // Initialize empty cache
		}
	})
	return organizationMembersServiceInstance
}

// Fetch the members from the organization using the API client
func (s *OrganizationMembersService) GetOrganizationMembers() ([]api.GetOrganizationMembersV1Results, error) {
	pageSize := 100
	members := []api.GetOrganizationMembersV1Results{}

	// Fetch the members from the organization using the API client
	page := 0
	for {
		// Fetch the members from the organization using the API client
		pageMembers, err := s.client.GetOrganizationMembersV1(0, pageSize, page, "")
		if err != nil {
			return nil, err
		}
		// If no members are returned, break the loop
		if len(pageMembers) == 0 {
			break
		}
		// Append a member if it's not already in the list
		for _, member := range pageMembers {
			if !slices.Contains(members, member) {
				members = append(members, member)
			}
		}
		// Increment the page number
		page++
	}
	// Sort the members by email
	sort.Slice(members, func(i, j int) bool {
		return members[i].Email < members[j].Email
	})
	return members, nil
}

// Fetch the members from the organization using the API client and cache the results
// If the members list is already populated, it returns the cached results
func (s *OrganizationMembersService) GetOrganizationMembersByCache() ([]api.GetOrganizationMembersV1Results, error) {
	// Check if the members list is already populated
	if len(s.members) == 0 {
		// Fetch the members from the organization using the API client
		members, err := s.GetOrganizationMembers()
		if err != nil {
			return nil, err
		}
		s.members = members
	}
	// Return the list of members
	return s.members, nil
}

// GetOrganizationAdmins retrieves the admins of an organization.
// It leverages the GetOrganizationMembers method to fetch all members and filters out non-admins.
func (s *OrganizationMembersService) GetOrganizationMembersByRole(role models.OrganizationMemberRole) ([]api.GetOrganizationMembersV1Results, error) {
	// Retrieve all members of the organization
	allMembers, err := s.GetOrganizationMembersByCache()
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
	organizationMembers, err := s.GetOrganizationMembersByCache()
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

// GetOrganizationMemberByEmail retrieves a member of an organization by their email.
func (s *OrganizationMembersService) GetOrganizationMemberByEmail(email string) (*api.GetOrganizationMembersV1Results, error) {
	// Retrieve all organization members
	organizationMembers, err := s.GetOrganizationMembersByCache()
	if err != nil {
		return nil, err
	}
	// Iterate through all members to find the one with the matching email
	for _, member := range organizationMembers {
		if member.Email == email {
			return &member, nil
		}
	}
	// Return an error if no member with the specified email is found
	return nil, fmt.Errorf("member with email %s not found", email)
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
