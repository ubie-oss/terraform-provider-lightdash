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
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// orgRoleMemberTestEmailEnvVar names a DEDICATED test user in the organization. Never point it at a
// real employee: this test changes that user's organization role (member -> viewer -> member).
const orgRoleMemberTestEmailEnvVar = "LIGHTDASH_TEST_ORG_MEMBER_EMAIL"

// Using the shared testAccPreCheck and testAccProtoV6ProviderFactories from provider_test.go

func TestAccOrganizationRoleMemberResource_member(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for lightdash_organization_role_member")
	}
	email := strings.TrimSpace(os.Getenv(orgRoleMemberTestEmailEnvVar))
	if email == "" {
		t.Skipf("Skipping acceptance test: %s is not set", orgRoleMemberTestEmailEnvVar)
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}
	memberConfig, err := ReadAccTestResource([]string{"resources", "lightdash_organization_role_member", "member", "010_member.tf"})
	if err != nil {
		t.Fatalf("Failed to read member config: %v", err)
	}
	viewerConfig, err := ReadAccTestResource([]string{"resources", "lightdash_organization_role_member", "member", "020_viewer.tf"})
	if err != nil {
		t.Fatalf("Failed to read viewer config: %v", err)
	}

	// Terraform reads the variable from the environment, so the email is never written to a file.
	t.Setenv("TF_VAR_org_member_email", email)
	target := resolveOrgRoleMemberTestTarget(t, email)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create with role "member": the path that failed in issue #461.
				Config: providerConfig + memberConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_organization_role_member.test", "role", "member"),
				),
			},
			{
				// Update to a built-in role.
				Config: providerConfig + viewerConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_organization_role_member.test", "role", "viewer"),
				),
			},
		},
		// Destroy runs the Delete path, which resets the user to "member".
		CheckDestroy: checkOrgRoleMemberDefaultAfterDestroy(target),
	})
}

type orgRoleMemberTestTarget struct {
	client   *api.Client
	orgUUID  string
	userUUID string
}

// resolveOrgRoleMemberTestTarget looks up the organization and user UUIDs for the test email.
// Errors never include the email.
func resolveOrgRoleMemberTestTarget(t *testing.T, email string) orgRoleMemberTestTarget {
	t.Helper()

	host, err := getLightdashUrl()
	if err != nil {
		t.Fatalf("%v", err)
	}
	token, err := getLightdashApiKey()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client, err := api.NewClient(host, token, nil)
	if err != nil {
		t.Fatalf("failed to create API client: %v", err)
	}

	ctx := context.Background()
	orgUUID, err := services.GetOrganizationUUID(ctx, client)
	if err != nil {
		t.Fatalf("failed to get organization UUID: %v", err)
	}
	member, err := services.GetOrganizationMembersService(client).GetOrganizationMemberByEmail(ctx, email)
	if err != nil {
		t.Fatalf("no organization member matches %s (email not printed)", orgRoleMemberTestEmailEnvVar)
	}
	return orgRoleMemberTestTarget{client: client, orgUUID: orgUUID, userUUID: member.UserUUID}
}

// checkOrgRoleMemberDefaultAfterDestroy checks the Delete path: the user must be back on "member".
func checkOrgRoleMemberDefaultAfterDestroy(target orgRoleMemberTestTarget) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		assignment, err := services.GetRoleService(target.client).GetOrgUserAssignment(context.Background(), target.orgUUID, target.userUUID)
		if err != nil {
			return fmt.Errorf("failed to read organization role after destroy: %w", err)
		}
		role, err := services.TerraformOrganizationRoleFromAssignment(assignment)
		if err != nil {
			return err
		}
		if role != models.ORGANIZATION_MEMBER_ROLE {
			return fmt.Errorf("organization role after destroy is %q, want %q", role, models.ORGANIZATION_MEMBER_ROLE)
		}
		return nil
	}
}
