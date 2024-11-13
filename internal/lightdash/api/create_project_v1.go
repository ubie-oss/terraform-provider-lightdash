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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type DbtConnection struct {
	Type           string `json:"type"`
	Repository     string `json:"repository"`
	Branch         string `json:"branch"`
	ProjectSubPath string `json:"project_sub_path"`
	HostDomain     string `json:"host_domain"`
}

type WarehouseConnection struct {
	Type                   models.WarehouseType `json:"type"`
	Account                string               `json:"account,omitempty"`
	Role                   string               `json:"role,omitempty"`
	Database               string               `json:"database,omitempty"`
	Warehouse              string               `json:"warehouse,omitempty"`
	Schema                 string               `json:"schema,omitempty"`
	ClientSessionKeepAlive bool                 `json:"clientSessionKeepAlive,omitempty"`
	Threads                int32                `json:"threads,omitempty"`
	ServerHostName         string               `json:"serverHostName,omitempty"`
	HTTPPath               string               `json:"httpPath,omitempty"`
	PersonalAccessToken    string               `json:"personalAccessToken,omitempty"`
	Catalog                string               `json:"catalog,omitempty"`
}

type CreateProjectV1Request struct {
	OrganisationUUID    string              `json:"organizationUuid"`
	Name                string              `json:"name"`
	Type                models.ProjectType  `json:"type"`
	DbtConnection       DbtConnection       `json:"dbtConnection"`
	WarehouseConnection WarehouseConnection `json:"warehouseConnection"`
}

type CreateProjectV1Results struct {
	ProjectUUID         string              `json:"projectUuid"`
	Name                string              `json:"name,omitempty"`
	OrganisationUUID    string              `json:"organizationUuid"`
	Type                models.ProjectType  `json:"type"`
	DbtConnection       DbtConnection       `json:"dbtConnection"`
	WarehouseConnection WarehouseConnection `json:"warehouseConnection"`
}

type CreateProjectV1ResponseResults struct {
	Project        CreateProjectV1Results `json:"project"`
	HasContentCopy bool                   `json:"hasContentCopy"`
}

type CreateProjectV1Response struct {
	Results CreateProjectV1ResponseResults `json:"results,omitempty"`
	Status  string                         `json:"status"`
}

func (c *Client) CreateProjectV1(organisationUUID string, name string, projectType models.ProjectType, dbtConnection DbtConnection, warehouseConnection WarehouseConnection) (*CreateProjectV1Results, error) {
	// Create the request body
	data := CreateProjectV1Request{
		OrganisationUUID:    organisationUUID,
		Name:                name,
		Type:                projectType,
		DbtConnection:       dbtConnection,
		WarehouseConnection: warehouseConnection,
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("error marshalling request: %s", err)
	}
	// Create the request
	path := fmt.Sprintf("%s/api/v1/org/projects", c.HostUrl)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("error creating new request: %v", err)
	}
	// Do request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %v", err)
	}
	// Marshal the response
	response := CreateProjectV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}
	return &response.Results.Project, nil
}
