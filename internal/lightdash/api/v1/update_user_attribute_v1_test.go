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

import "testing"

func TestUpdateUserAttributeV1RejectsEmptyUUID(t *testing.T) {
	_, err := UpdateUserAttributeV1(nil, "  ", CreateUserAttributeV1Request{Name: "region"})
	if err == nil {
		t.Fatal("expected an error for an empty user attribute UUID")
	}
}

func TestUpdateUserAttributeV1RejectsEmptyName(t *testing.T) {
	_, err := UpdateUserAttributeV1(nil, "attr-1", CreateUserAttributeV1Request{})
	if err == nil {
		t.Fatal("expected an error for an empty name")
	}
}
