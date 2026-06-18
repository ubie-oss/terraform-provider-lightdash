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

package models

import "time"

const (
	AssigneeTypeUser  = "user"
	AssigneeTypeGroup = "group"
)

// Role is an organization role from GET /api/v2/orgs/{orgUuid}/roles.
// System roles may include a scopes list (RoleWithScopes shape).
type Role struct {
	RoleUUID         string     `json:"roleUuid"`
	Name             string     `json:"name"`
	Description      *string    `json:"description"`
	OwnerType        string     `json:"ownerType"`
	OrganizationUUID *string    `json:"organizationUuid"`
	CreatedAt        *time.Time `json:"createdAt"`
	UpdatedAt        *time.Time `json:"updatedAt"`
	CreatedBy        *string    `json:"createdBy"`
	Scopes           []string   `json:"scopes,omitempty"`
}

// RoleAssignment is a role assignment from v2 assignment list or upsert responses.
type RoleAssignment struct {
	RoleID         string    `json:"roleId"`
	RoleName       string    `json:"roleName"`
	AssigneeType   string    `json:"assigneeType"`
	AssigneeID     string    `json:"assigneeId"`
	AssigneeName   string    `json:"assigneeName,omitempty"`
	OwnerType      string    `json:"ownerType"`
	ProjectID      string    `json:"projectId,omitempty"`
	OrganizationID string    `json:"organizationId,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
