// Copyright 2023 Ubie, inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

func TestCreateSpaceV1Request_JSON_privateSpace(t *testing.T) {
	isPrivate := true
	req := CreateSpaceV1Request{
		Name:                     "Private",
		InheritParentPermissions: models.InheritParentPermissionsFromTerraformIsPrivate(&isPrivate),
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"name":"Private","inheritParentPermissions":false}`
	if string(b) != want {
		t.Fatalf("json mismatch\n got:  %s\n want: %s", string(b), want)
	}
}

func TestCreateSpaceV1Request_JSON_publicSpace(t *testing.T) {
	isPrivate := false
	req := CreateSpaceV1Request{
		Name:                     "Public",
		InheritParentPermissions: models.InheritParentPermissionsFromTerraformIsPrivate(&isPrivate),
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"name":"Public","inheritParentPermissions":true}`
	if string(b) != want {
		t.Fatalf("json mismatch\n got:  %s\n want: %s", string(b), want)
	}
}

func TestCreateSpaceV1Request_JSON_omitsInheritWhenIsPrivateUnknown(t *testing.T) {
	req := CreateSpaceV1Request{
		Name:                     "Default",
		InheritParentPermissions: models.InheritParentPermissionsFromTerraformIsPrivate(nil),
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"name":"Default"}`
	if string(b) != want {
		t.Fatalf("json mismatch\n got:  %s\n want: %s", string(b), want)
	}
}

func TestUpdateSpaceV1Request_JSON_inheritOnly(t *testing.T) {
	inherit := false
	req := UpdateSpaceV1Request{
		Name:                     "Renamed",
		InheritParentPermissions: &inherit,
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"name":"Renamed","inheritParentPermissions":false}`
	if string(b) != want {
		t.Fatalf("json mismatch\n got:  %s\n want: %s", string(b), want)
	}
}
