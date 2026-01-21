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

type CredentialsDetail struct {
	Type string `json:"type"`
	User string `json:"user,omitempty"`
}

// BigQueryCredentials represents BigQuery warehouse credentials with service account key
type BigQueryCredentials struct {
	Type                       string  `json:"type"`
	Project                    string  `json:"project"`
	Dataset                    *string `json:"dataset,omitempty"`
	KeyfileContents            string  `json:"keyfileContents"`
	Location                   *string `json:"location,omitempty"`
	TimeoutSeconds             *int    `json:"timeoutSeconds,omitempty"`
	MaximumBytesBilled         *int64  `json:"maximumBytesBilled,omitempty"`
	Priority                   *string `json:"priority,omitempty"`
	Retries                    *int    `json:"retries,omitempty"`
	StartOfWeek                *int    `json:"startOfWeek,omitempty"`
	RequireUserCredentials     *bool   `json:"requireUserCredentials,omitempty"`
	OrganizationWarehouseUUID  *string `json:"organizationWarehouseCredentialsUuid,omitempty"`
}

type WarehouseCredentials struct {
	Credentials               interface{} `json:"credentials"`
	UpdatedAt                 time.Time   `json:"updatedAt"`
	CreatedAt                 time.Time   `json:"createdAt"`
	CreatedByUserUUID         string      `json:"createdByUserUuid,omitempty"`
	Name                      string      `json:"name"`
	Description               *string     `json:"description,omitempty"`
	WarehouseType             string      `json:"warehouseType"`
	OrganizationUUID          string      `json:"organizationUuid"`
	OrganizationWarehouseUUID string      `json:"organizationWarehouseCredentialsUuid"`
	UserUUID                  string      `json:"userUuid,omitempty"`
	UUID                      string      `json:"uuid,omitempty"` // Deprecated: use OrganizationWarehouseUUID
}
