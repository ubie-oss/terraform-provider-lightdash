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

// UserAttributeGroupValue represents a group attribute value
type UserAttributeGroupValue struct {
	GroupUUID string `json:"groupUuid"`
	Value     string `json:"value"`
}

// UserAttributeUserValue represents a user attribute value returned by the API
type UserAttributeUserValue struct {
	UserUUID string `json:"userUuid"`
	Email    string `json:"email"`
	Value    string `json:"value"`
}

// CreateUserAttributeUserValue represents a user attribute value in a create/update request
type CreateUserAttributeUserValue struct {
	UserUUID string `json:"userUuid"`
	Value    string `json:"value"`
}

// UserAttribute represents a user attribute in Lightdash
type UserAttribute struct {
	UUID             string                    `json:"uuid"`
	Name             string                    `json:"name"`
	Description      *string                   `json:"description"`
	OrganizationUUID string                    `json:"organizationUuid"`
	AttributeDefault *string                   `json:"attributeDefault"`
	Users            []UserAttributeUserValue  `json:"users"`
	Groups           []UserAttributeGroupValue `json:"groups"`
	CreatedAt        string                    `json:"createdAt"`
}

// CreateUserAttribute represents the request body for creating or updating a user attribute
type CreateUserAttribute struct {
	Name             string                         `json:"name"`
	Description      *string                        `json:"description,omitempty"`
	AttributeDefault *string                        `json:"attributeDefault"`
	Groups           []UserAttributeGroupValue      `json:"groups"`
	Users            []CreateUserAttributeUserValue `json:"users"`
}
