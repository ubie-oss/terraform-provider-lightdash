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

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type SpaceAccessMember struct {
	UserUUID  string                 `json:"userUuid"`
	SpaceRole models.SpaceMemberRole `json:"role"`
	// Additional fields from GetSpaceV1 API response
	HasDirectAccess bool   `json:"hasDirectAccess"`
	InheritedRole   string `json:"inheritedRole"`
	InheritedFrom   string `json:"inheritedFrom"`
	ProjectRole     string `json:"projectRole"`
}

type SpaceAccessGroup struct {
	GroupUUID string                 `json:"groupUuid"`
	GroupName string                 `json:"groupName"`
	SpaceRole models.SpaceMemberRole `json:"spaceRole"`
}

type ChildSpace struct {
	OrganizationUuid string `json:"organizationUuid"`
	ProjectUuid      string `json:"projectUuid"`
	SpaceUUID        string `json:"uuid"`
	Name             string `json:"name"`
	IsPrivate        bool   `json:"isPrivate"`
}

type GetSpaceV1Results struct {
	// The response doesn't contain the OrganizationUUID right now
	// OrganizationUUID string              `json:"organizationUuid"`
	ProjectUUID        string              `json:"projectUuid"`
	ParentSpaceUUID    *string             `json:"parentSpaceUuid,omitempty"`
	SpaceUUID          string              `json:"uuid"`
	SpaceName          string              `json:"name"`
	IsPrivate          bool                `json:"isPrivate"`
	ChildSpaces        []ChildSpace        `json:"childSpaces"`
	SpaceAccessMembers []SpaceAccessMember `json:"access"`
	SpaceAccessGroups  []SpaceAccessGroup  `json:"groupsAccess"`
}

type GetSpaceV1Response struct {
	Results GetSpaceV1Results `json:"results,omitempty"`
	Status  string            `json:"status"`
}

func (c *Client) GetSpaceV1(projectUuid string, spaceUuid string) (*GetSpaceV1Results, error) {
	// Validate the arguments
	if len(strings.TrimSpace(projectUuid)) == 0 {
		return nil, fmt.Errorf("project UUID is empty")
	}
	if len(strings.TrimSpace(spaceUuid)) == 0 {
		return nil, fmt.Errorf("space UUID is empty")
	}

	// Make a request
	path := fmt.Sprintf("%s/api/v1/projects/%s/spaces/%s", c.HostUrl, projectUuid, spaceUuid)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request for space: %w", err)
	}
	// Do the request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request for space: %w", err)
	}
	// Parse the response
	response := GetSpaceV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling space response: %w", err)
	}
	// Make sure if the organization is not nil
	if len(strings.TrimSpace(response.Results.SpaceUUID)) == 0 {
		return nil, fmt.Errorf("space UUID is nil")
	}

	return &response.Results, nil
}

func (c *Client) GetSpaceMemberV1(projectUuid string, spaceUuid string, userUuid string) (*SpaceAccessMember, error) {
	// Validate the arguments
	if len(strings.TrimSpace(userUuid)) == 0 {
		return nil, fmt.Errorf("user UUID is empty")
	}

	// Get the space
	space, err := c.GetSpaceV1(projectUuid, spaceUuid)
	if err != nil {
		return nil, err
	}

	// Find the user in the space
	for _, member := range space.SpaceAccessMembers {
		if member.UserUUID == userUuid {
			return &member, nil
		}
	}

	return nil, fmt.Errorf("user %s is not found in the space %s", userUuid, spaceUuid)
}
