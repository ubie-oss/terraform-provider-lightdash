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
	"encoding/json"
	"testing"
)

func TestCreateProjectV1Request_JSON_noneConnection(t *testing.T) {
	data := CreateProjectV1Request{
		Name:       "api-create-verify",
		Type:       "DEFAULT",
		DbtVersion: "latest",
		DbtConnection: ProjectDbtConnectionV1{
			Type: dbtConnectionTypeNone,
		},
	}
	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded["name"] != "api-create-verify" || decoded["type"] != "DEFAULT" || decoded["dbtVersion"] != "latest" {
		t.Fatalf("request fields = %#v", decoded)
	}
	connection, ok := decoded["dbtConnection"].(map[string]any)
	if !ok || connection["type"] != "none" {
		t.Fatalf("dbtConnection = %#v, want type none", decoded["dbtConnection"])
	}
}

func TestCreateProjectV1Response_JSON(t *testing.T) {
	const payload = `{
		"status": "ok",
		"results": {
			"hasContentCopy": false,
			"project": {
				"organizationUuid": "org-1",
				"projectUuid": "proj-1",
				"slug": "api-create-verify",
				"name": "api-create-verify",
				"type": "DEFAULT",
				"dbtConnection": {"type": "none"},
				"dbtVersion": "latest"
			}
		}
	}`
	var response CreateProjectV1Response
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	project := response.Results.Project
	if project.ProjectUUID != "proj-1" || project.Slug != "api-create-verify" || project.Name != "api-create-verify" {
		t.Fatalf("project = %#v", project)
	}
	if project.Type != "DEFAULT" || project.DbtVersion != "latest" || project.DbtConnection.Type != "none" {
		t.Fatalf("project connection = %#v", project)
	}
	if response.Results.HasContentCopy {
		t.Fatal("hasContentCopy = true, want false")
	}
}

func TestCreateProjectV1_rejectsEmptyName(t *testing.T) {
	_, err := CreateProjectV1(nil, "  ", "DEFAULT", "latest")
	if err == nil {
		t.Fatal("expected error for empty project name")
	}
}

func TestGetProjectV1Results_JSON_dbtConnection(t *testing.T) {
	const payload = `{
		"organizationUuid": "org-1",
		"projectUuid": "proj-1",
		"slug": "api-create-verify",
		"name": "api-create-verify",
		"type": "DEFAULT",
		"dbtVersion": "latest",
		"dbtConnection": {"type": "none", "target": null},
		"schedulerTimezone": "UTC",
		"upstreamProjectUuid": null
	}`
	var results GetProjectV1Results
	if err := json.Unmarshal([]byte(payload), &results); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if results.Slug != "api-create-verify" || results.DbtVersion != "latest" {
		t.Fatalf("slug/dbtVersion = %q %q", results.Slug, results.DbtVersion)
	}
	if results.DbtConnection == nil || results.DbtConnection.Type != "none" {
		t.Fatalf("dbtConnection = %#v, want type none", results.DbtConnection)
	}
	if results.UpstreamProjectUUID != nil {
		t.Fatalf("upstreamProjectUuid = %v, want nil", results.UpstreamProjectUUID)
	}
}
