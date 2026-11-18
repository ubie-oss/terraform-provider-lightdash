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

// UserAttribute is a Lightdash user attribute in the organization of the API token.
// AttributeDefaults is nil when the attribute has no default.
// Users holds values assigned to individual users. The Terraform resource does not manage them,
// but updates must send them back because the API replaces the whole attribute.
type UserAttribute struct {
	OrganizationUUID  string
	UserAttributeUUID string
	Name              string
	Description       string
	AttributeDefaults []string
	Groups            []UserAttributeValue
	Users             []UserAttributeValue
	CreatedAt         string
}

// UserAttributeValue is the values of one group or one user.
// UUID is a group UUID or a user UUID, depending on the slice it belongs to.
type UserAttributeValue struct {
	UUID   string
	Values []string
}
