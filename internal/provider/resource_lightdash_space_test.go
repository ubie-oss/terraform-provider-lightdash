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

package provider

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestGetSpaceResourceId(t *testing.T) {
	project_uuid := "abc-123"
	space_uuid := "xyz-234"
	results := getSpaceResourceId(project_uuid, space_uuid)
	expected := "projects/abc-123/spaces/xyz-234"

	if results != expected {
		t.Errorf("Expected: %s, Got: %s", expected, results)
	}
}

func TestExtractSpaceResourceId(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
		wantErr  bool
	}{
		{
			input:    "projects/xyz-567/spaces/abc-123",
			expected: []string{"xyz-567", "abc-123"},
			wantErr:  false,
		},
		{
			input:    "projects/123/spaces/456",
			expected: []string{"123", "456"},
			wantErr:  false,
		},
		{
			input:    "projects/xyz/spaces/",
			expected: nil,
			wantErr:  true,
		},
		{
			input:    "projects/xyz/spaces/abc/def",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		info, err := extractSpaceResourceId(test.input)

		if (err != nil) != test.wantErr {
			t.Errorf("Input: %s, Expected error: %v, Got error: %v", test.input, test.wantErr, err)
		}

		if !reflect.DeepEqual(info, test.expected) {
			t.Errorf("Input: %s, Expected: %v, Got: %v", test.input, test.expected, info)
		}
	}
}

func TestAccLightdashSpaceResource(t *testing.T) {
	api_key, err := getLightdashApiKey()
	if isIntegrationTestMode() && err != nil {
		t.Errorf("Error retrieving LIGHTDASH_API_KEY environment variable: %v", err)
	}
	providerConfig := getTestProviderConfig(api_key)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccLightdashSpaceResourceConfig("xxx-xxx-xxx", "test-space", true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_project.test", "project_uuid", "example-id"),
					resource.TestCheckResourceAttr("lightdash_project.test", "name", "one"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLightdashSpaceResourceConfig(
	projectUuid string, name string, is_private bool, delete_protection bool) string {
	resource_template := `
resource "lightdash_space" "test" {
  project_uuid = "%[1]s"
  name = "%[2]s"
  is_private = %[3]t
  deletion_protection = %[4]t
}`
	return fmt.Sprintf(resource_template, projectUuid, name, is_private, delete_protection)
}
