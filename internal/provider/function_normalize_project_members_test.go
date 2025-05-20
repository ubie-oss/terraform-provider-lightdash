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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestNormalizeProjectMembersRun_simple(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project_role_group")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Get the function config
	functionConfig010, err := ReadAccTestResource([]string{"function_normalize_project_members", "simple", "010_simple.tf"})
	if err != nil {
		t.Fatalf("Failed to get functionConfig: %v", err)
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + functionConfig010,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("normalized_project_members_admins_json", "[\"admin1\"]"),
					resource.TestCheckOutput("normalized_project_members_developers_json", "[\"dev1\"]"),
					resource.TestCheckOutput("normalized_project_members_editors_json", "[\"editor1\"]"),
					resource.TestCheckOutput("normalized_project_members_interactive_viewers_json", "[\"interactive_viewer1\"]"),
					resource.TestCheckOutput("normalized_project_members_viewers_json", "[\"viewer1\"]"),
				),
			},
		},
	})
}

func TestNormalizeProjectMembersRun_complicated(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project_role_group")
	}

	// Get the provider config
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	// Get the function config
	functionConfig010, err := ReadAccTestResource([]string{"function_normalize_project_members", "complicated", "010_complicated.tf"})
	if err != nil {
		t.Fatalf("Failed to get functionConfig: %v", err)
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + functionConfig010,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("normalized_project_members_admins_json", "[\"admin1\",\"admin2\"]"),
					resource.TestCheckOutput("normalized_project_members_developers_json", "[\"dev1\",\"dev2\"]"),
					resource.TestCheckOutput("normalized_project_members_editors_json", "[\"editor1\",\"editor2\"]"),
					resource.TestCheckOutput("normalized_project_members_interactive_viewers_json", "[\"interactive_viewer1\",\"interactive_viewer2\"]"),
					resource.TestCheckOutput("normalized_project_members_viewers_json", "[\"viewer1\",\"viewer2\"]"),
				),
			},
		},
	})
}

func TestNormalizeMembers_normalizeMembers(t *testing.T) {
	tests := []struct {
		name                       string
		admins                     []string
		developers                 []string
		editors                    []string
		interactiveViewers         []string
		viewers                    []string
		expectedAdmins             []string
		expectedDevelopers         []string
		expectedEditors            []string
		expectedInteractiveViewers []string
		expectedViewers            []string
	}{
		{
			name:                       "simple case - no overlapping roles",
			admins:                     []string{"admin1"},
			developers:                 []string{"dev1"},
			editors:                    []string{"editor1"},
			interactiveViewers:         []string{"iv1"},
			viewers:                    []string{"viewer1"},
			expectedAdmins:             []string{"admin1"},
			expectedDevelopers:         []string{"dev1"},
			expectedEditors:            []string{"editor1"},
			expectedInteractiveViewers: []string{"iv1"},
			expectedViewers:            []string{"viewer1"},
		},
		{
			name:                       "overlapping roles - member in multiple roles",
			admins:                     []string{"admin1", "admin2"},
			developers:                 []string{"dev1", "admin1"},
			editors:                    []string{"editor1", "dev1"},
			interactiveViewers:         []string{"iv1", "editor1"},
			viewers:                    []string{"viewer1", "iv1"},
			expectedAdmins:             []string{"admin1", "admin2"},
			expectedDevelopers:         []string{"dev1"},
			expectedEditors:            []string{"editor1"},
			expectedInteractiveViewers: []string{"iv1"},
			expectedViewers:            []string{"viewer1"},
		},
		{
			name:                       "empty input",
			admins:                     []string{},
			developers:                 []string{},
			editors:                    []string{},
			interactiveViewers:         []string{},
			viewers:                    []string{},
			expectedAdmins:             []string{},
			expectedDevelopers:         []string{},
			expectedEditors:            []string{},
			expectedInteractiveViewers: []string{},
			expectedViewers:            []string{},
		},
		{
			name:                       "mixed input - some roles empty",
			admins:                     []string{"admin1"},
			developers:                 []string{},
			editors:                    []string{"editor1"},
			interactiveViewers:         []string{},
			viewers:                    []string{"viewer1"},
			expectedAdmins:             []string{"admin1"},
			expectedDevelopers:         []string{},
			expectedEditors:            []string{"editor1"},
			expectedInteractiveViewers: []string{},
			expectedViewers:            []string{"viewer1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &NormalizeProjectMembersFunction{}
			admins, developers, editors, interactiveViewers, viewers := f.normalizeMembers(
				tt.admins,
				tt.developers,
				tt.editors,
				tt.interactiveViewers,
				tt.viewers,
			)

			// Compare the results with expected values
			if !equalStringSlices(admins, tt.expectedAdmins) {
				t.Errorf("admins = %v, expected %v", admins, tt.expectedAdmins)
			}
			if !equalStringSlices(developers, tt.expectedDevelopers) {
				t.Errorf("developers = %v, expected %v", developers, tt.expectedDevelopers)
			}
			if !equalStringSlices(editors, tt.expectedEditors) {
				t.Errorf("editors = %v, expected %v", editors, tt.expectedEditors)
			}
			if !equalStringSlices(interactiveViewers, tt.expectedInteractiveViewers) {
				t.Errorf("interactiveViewers = %v, expected %v", interactiveViewers, tt.expectedInteractiveViewers)
			}
			if !equalStringSlices(viewers, tt.expectedViewers) {
				t.Errorf("viewers = %v, expected %v", viewers, tt.expectedViewers)
			}
		})
	}
}

// Helper function to compare two string slices for equality
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
