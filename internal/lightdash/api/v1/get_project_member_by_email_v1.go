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

package v1

import (
	"fmt"
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

type FindProjectMemberByEmailResponse struct {
	Results GetProjectAccessListV1Results `json:"results,omitempty"`
	Status  string                        `json:"status"`
}

// Theere is no API to get a specific member of a project right now.
// So, we have to find the member from the list of all members.
func FindProjectMemberByEmail(c *api.Client, projectUuid string, email string) (*GetProjectAccessListV1Results, error) {
	// Validate the arguments
	if len(strings.TrimSpace(email)) == 0 {
		return nil, fmt.Errorf("user's email is empty")
	}

	// Get all members
	members, err := GetProjectAccessListV1(c, projectUuid)
	if err != nil {
		return nil, fmt.Errorf("error retrieving project access list for UUID %s: %w", projectUuid, err)
	}

	// Find the matchedMember
	var matchedMember *GetProjectAccessListV1Results
	for _, member := range members {
		if member.Email == email {
			// To avoid the exportloopref violation
			matched := member
			matchedMember = &matched
		}
	}
	if matchedMember == nil {
		return nil, fmt.Errorf("no member found")
	}

	// Parse the response
	response := FindProjectMemberByEmailResponse{
		Results: *matchedMember,
		Status:  "ok",
	}

	return &response.Results, nil
}
