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
	"fmt"
	"strings"
)

type GetProjectAccessMemberV1Response struct {
	Results GetProjectAccessListV1Results `json:"results,omitempty"`
	Status  string                        `json:"status"`
}

// Theere is no API to get a specific member of a project right now.
// So, we have to find the member from the list of all members.
func (c *Client) GetProjectMemberByUuidV1(projectUuid string, userUuid string) (*GetProjectAccessListV1Results, error) {
	// Validate the arguments
	if len(strings.TrimSpace(userUuid)) == 0 {
		return nil, fmt.Errorf("user UUID is empty")
	}

	// Get all members
	members, err := c.GetProjectAccessListV1(projectUuid)
	if err != nil {
		return nil, err
	}

	// Find the searchedMember
	var searchedMember *GetProjectAccessListV1Results
	for _, member := range members {
		if member.UserUUID == userUuid {
			searchedMember = &member
		}
	}
	if searchedMember == nil {
		return nil, fmt.Errorf("member not found")
	}

	// Parse the response
	response := GetProjectAccessMemberV1Response{
		Results: *searchedMember,
		Status:  "ok",
	}

	return &response.Results, nil
}
