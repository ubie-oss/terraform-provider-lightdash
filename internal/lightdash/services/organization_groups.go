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
	"fmt"

	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type OrganizationGroupsService struct {
	client *api.Client
}

func NewOrganizationGroupsService(client *api.Client) *OrganizationGroupsService {
	return &OrganizationGroupsService{
		client: client,
	}
}

// GetGroup retrieves a single group by UUID
func (s *OrganizationGroupsService) GetGroup(ctx context.Context, groupUUID string) (*models.OrganizationGroup, error) {
	// Get the group from the API
	group, err := apiv1.GetGroupV1(s.client, groupUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group with UUID %s: %w", groupUUID, err)
	}

	// Convert to the model
	result := &models.OrganizationGroup{
		OrganizationUUID: group.OrganizationUUID,
		GroupUUID:        group.GroupUUID,
		Name:             group.Name,
		CreatedAt:        group.CreatedAt,
	}

	return result, nil
}

func (s *OrganizationGroupsService) GetOrganizationGroups(ctx context.Context) ([]models.OrganizationGroup, error) {
	groupMap := make(map[string]models.OrganizationGroup)
	page := 0
	pageSize := 100

	for {
		// Fetch the groups from the organization using the API client
		groups, err := apiv1.GetOrganizationGroupsV1(s.client, float64(page), float64(pageSize), 0, "")
		if err != nil {
			return nil, err
		}
		if len(groups) == 0 {
			break
		}

		// Convert API response to models.OrganizationGroup and store in map to deduplicate
		for _, group := range groups {
			newGroup := models.OrganizationGroup{
				OrganizationUUID: group.OrganizationUUID,
				Name:             group.Name,
				GroupUUID:        group.GroupUUID,
				CreatedAt:        group.CreatedAt,
			}

			// Use GroupUUID as the key to ensure uniqueness
			key := fmt.Sprintf("%s/%s", newGroup.OrganizationUUID, newGroup.GroupUUID)
			groupMap[key] = newGroup
		}

		page++
	}

	// Convert map values to slice
	allGroups := make([]models.OrganizationGroup, 0, len(groupMap))
	for _, group := range groupMap {
		allGroups = append(allGroups, group)
	}

	// Check duplicates
	seen := make(map[string]bool)
	duplicateUUIDs := []string{}
	for _, group := range allGroups {
		key := fmt.Sprintf("%s/%s", group.OrganizationUUID, group.GroupUUID)
		if seen[key] {
			duplicateUUIDs = append(duplicateUUIDs, group.GroupUUID)
		}
		seen[key] = true
	}
	if len(duplicateUUIDs) > 0 {
		return nil, fmt.Errorf("duplicated group_uuid(s): %v", duplicateUUIDs)
	}

	return allGroups, nil
}
