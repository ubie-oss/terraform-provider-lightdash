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
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

func TestUserAttributeResourceID(t *testing.T) {
	got := userAttributeResourceID("org-uuid", "attribute-uuid")
	expected := "organizations/org-uuid/user-attributes/attribute-uuid"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}

	organizationUUID, userAttributeUUID, err := extractUserAttributeResourceID(expected)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if organizationUUID != "org-uuid" || userAttributeUUID != "attribute-uuid" {
		t.Fatalf("unexpected id parts: %s %s", organizationUUID, userAttributeUUID)
	}

	organizationUUID, userAttributeUUID, err = extractUserAttributeResourceID("organizations/ORG-UUID/user-attributes/ATTRIBUTE-UUID")
	if err != nil {
		t.Fatalf("extract mixed case: %v", err)
	}
	if organizationUUID != "org-uuid" || userAttributeUUID != "attribute-uuid" {
		t.Fatalf("expected lowercase ids, got %s %s", organizationUUID, userAttributeUUID)
	}
}

func TestExtractUserAttributeResourceIDInvalid(t *testing.T) {
	for _, id := range []string{"", "attribute-uuid", "organizations/org-uuid/groups/group-uuid", "organizations/org-uuid/user-attributes/", "organizations/org-uuid/user-attributes/attribute-uuid/extra"} {
		if _, _, err := extractUserAttributeResourceID(id); err == nil {
			t.Errorf("expected an error for %q", id)
		}
	}
}

func TestValidateUserAttributeConfig(t *testing.T) {
	ctx := context.Background()
	valid := userAttributeResourceModel{
		Name:              types.StringValue("region"),
		Description:       types.StringValue(""),
		AttributeDefaults: types.SetNull(types.StringType),
		Groups:            types.SetNull(userAttributeGroupValueObjectType),
	}
	if diags := validateUserAttributeConfig(ctx, valid); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	invalid := userAttributeResourceModel{
		Name:              types.StringValue("  "),
		AttributeDefaults: types.SetValueMust(types.StringType, nil),
		Groups:            types.SetValueMust(userAttributeGroupValueObjectType, nil),
	}
	// The empty groups set is valid. The empty name and empty defaults are not.
	// A group with an uppercase UUID and a duplicate is covered below.
	summaries := diagnosticSummaries(validateUserAttributeConfig(ctx, invalid))
	if !reflect.DeepEqual(summaries, []string{"Empty name", "Empty attribute_defaults"}) {
		t.Fatalf("unexpected diagnostics: %+v", summaries)
	}
}

func TestValidateUserAttributeConfigGroups(t *testing.T) {
	ctx := context.Background()
	us, diags := stringSliceToStringSet(ctx, []string{"us"})
	if diags.HasError() {
		t.Fatalf("values: %v", diags)
	}
	eu, diags := stringSliceToStringSet(ctx, []string{"eu"})
	if diags.HasError() {
		t.Fatalf("values: %v", diags)
	}
	emptyValues := types.SetValueMust(types.StringType, nil)
	withNull := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("us"), types.StringNull()})

	set := types.SetValueMust(userAttributeGroupValueObjectType, []attr.Value{
		mustGroupObject("group-1", us),
		mustGroupObject("group-1", eu),
		mustGroupObject("GROUP-2", us),
		mustGroupObject("group-3", emptyValues),
		mustGroupObject("group-4", withNull),
	})
	config := userAttributeResourceModel{
		Name:              types.StringValue("region"),
		AttributeDefaults: types.SetNull(types.StringType),
		Groups:            set,
	}
	summaries := diagnosticSummaries(validateUserAttributeConfig(ctx, config))
	sort.Strings(summaries)
	expected := []string{"Duplicate group_uuid", "Empty values", "Invalid group_uuid", "Null value"}
	if !reflect.DeepEqual(summaries, expected) {
		t.Fatalf("unexpected diagnostics: %+v", summaries)
	}
}

func TestUserAttributeResourceModelRoundTrip(t *testing.T) {
	ctx := context.Background()
	attribute := models.UserAttribute{
		OrganizationUUID:  "org-uuid",
		UserAttributeUUID: "attribute-uuid",
		Name:              "region",
		Description:       "The sales region",
		AttributeDefaults: []string{"jp"},
		Groups: []models.UserAttributeValue{
			{UUID: "group-1", Values: []string{"us"}},
			{UUID: "group-2", Values: []string{"eu"}},
		},
		Users: []models.UserAttributeValue{{UUID: "user-1", Values: []string{"uk"}}},
	}

	model, diags := userAttributeToResourceModel(ctx, attribute)
	if diags.HasError() {
		t.Fatalf("to model: %v", diags)
	}
	if model.ID.ValueString() != "organizations/org-uuid/user-attributes/attribute-uuid" {
		t.Fatalf("unexpected id: %s", model.ID.ValueString())
	}

	converted, diags := userAttributeFromResourceModel(ctx, model)
	if diags.HasError() {
		t.Fatalf("from model: %v", diags)
	}
	// User values are not part of the Terraform model, so they are not sent from it.
	if len(converted.Users) != 0 {
		t.Fatalf("expected no users in the resource model, got %+v", converted.Users)
	}
	sortUserAttributeGroups(&converted)
	if converted.Name != attribute.Name || converted.Description != attribute.Description {
		t.Fatalf("unexpected attribute: %+v", converted)
	}
	if !reflect.DeepEqual(converted.AttributeDefaults, []string{"jp"}) {
		t.Fatalf("unexpected defaults: %+v", converted.AttributeDefaults)
	}
	if len(converted.Groups) != 2 {
		t.Fatalf("unexpected groups: %+v", converted.Groups)
	}
}

func TestUserAttributeToResourceModelEmpty(t *testing.T) {
	ctx := context.Background()
	model, diags := userAttributeToResourceModel(ctx, models.UserAttribute{
		OrganizationUUID:  "org-uuid",
		UserAttributeUUID: "attribute-uuid",
		Name:              "region",
	})
	if diags.HasError() {
		t.Fatalf("to model: %v", diags)
	}
	if !model.AttributeDefaults.IsNull() {
		t.Fatalf("expected null defaults, got %s", model.AttributeDefaults)
	}
	if model.Groups.IsNull() || len(model.Groups.Elements()) != 0 {
		t.Fatalf("expected an empty groups set, got %s", model.Groups)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("expected an empty description, got %s", model.Description)
	}
}

func TestUserAttributeToResourceModelDuplicateValues(t *testing.T) {
	ctx := context.Background()
	model, diags := userAttributeToResourceModel(ctx, models.UserAttribute{
		OrganizationUUID:  "org-uuid",
		UserAttributeUUID: "attribute-uuid",
		Name:              "region",
		Groups:            []models.UserAttributeValue{{UUID: "group-1", Values: []string{"b", "a", "a"}}},
	})
	if diags.HasError() {
		t.Fatalf("to model: %v", diags)
	}
	converted, diags := userAttributeFromResourceModel(ctx, model)
	if diags.HasError() {
		t.Fatalf("from model: %v", diags)
	}
	sortUserAttributeGroups(&converted)
	expected := []models.UserAttributeValue{{UUID: "group-1", Values: []string{"a", "b"}}}
	if !reflect.DeepEqual(converted.Groups, expected) {
		t.Fatalf("unexpected groups: %+v", converted.Groups)
	}
}

func diagnosticSummaries(diags diag.Diagnostics) []string {
	summaries := make([]string, 0)
	for _, diagnostic := range diags.Errors() {
		summaries = append(summaries, diagnostic.Summary())
	}
	return summaries
}

func sortUserAttributeGroups(attribute *models.UserAttribute) {
	sort.Strings(attribute.AttributeDefaults)
	for i := range attribute.Groups {
		sort.Strings(attribute.Groups[i].Values)
	}
	sort.Slice(attribute.Groups, func(i, j int) bool {
		return attribute.Groups[i].UUID < attribute.Groups[j].UUID
	})
}

func mustGroupObject(uuid string, values types.Set) types.Object {
	return types.ObjectValueMust(userAttributeGroupValueObjectType.AttrTypes, map[string]attr.Value{
		"group_uuid": types.StringValue(uuid),
		"values":     values,
	})
}

// The acceptance tests assign values only to groups created by the tests.
// They never assign values to users, because a user's values change their data access.

func TestAccUserAttributeResource(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_user_attribute")
	}
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}
	createConfig, err := ReadAccTestResource([]string{"resources", "lightdash_user_attribute", "create", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get createConfig: %v", err)
	}
	updateConfig, err := ReadAccTestResource([]string{"resources", "lightdash_user_attribute", "create", "020_update.tf"})
	if err != nil {
		t.Fatalf("Failed to get updateConfig: %v", err)
	}
	clearConfig, err := ReadAccTestResource([]string{"resources", "lightdash_user_attribute", "create", "030_clear.tf"})
	if err != nil {
		t.Fatalf("Failed to get clearConfig: %v", err)
	}
	removeConfig, err := ReadAccTestResource([]string{"resources", "lightdash_user_attribute", "create", "040_remove.tf"})
	if err != nil {
		t.Fatalf("Failed to get removeConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserAttributesGone("zzz_acc_test_user_attribute_create", "zzz_acc_test_user_attribute_update"),
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "name", "zzz_acc_test_user_attribute_create"),
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "description", "Acceptance Test - create"),
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "attribute_defaults.#", "1"),
					resource.TestCheckTypeSetElemAttr("lightdash_user_attribute.test", "attribute_defaults.*", "default"),
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "groups.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("lightdash_user_attribute.test", "groups.*.group_uuid", "lightdash_group.test_group1", "group_uuid"),
					resource.TestCheckResourceAttrSet("lightdash_user_attribute.test", "user_attribute_uuid"),
					resource.TestCheckResourceAttrPair("data.lightdash_user_attribute.check", "user_attribute_uuid", "lightdash_user_attribute.test", "user_attribute_uuid"),
					resource.TestCheckResourceAttr("data.lightdash_user_attribute.check", "description", "Acceptance Test - create"),
					resource.TestCheckResourceAttr("data.lightdash_user_attribute.check", "groups.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.lightdash_user_attribute.check", "groups.*.values.*", "us"),
				),
			},
			{
				Config: providerConfig + updateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "name", "zzz_acc_test_user_attribute_update"),
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "description", "Acceptance Test - update"),
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "attribute_defaults.#", "2"),
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "groups.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("lightdash_user_attribute.test", "groups.*.group_uuid", "lightdash_group.test_group1", "group_uuid"),
					resource.TestCheckTypeSetElemAttrPair("lightdash_user_attribute.test", "groups.*.group_uuid", "lightdash_group.test_group2", "group_uuid"),
					resource.TestCheckResourceAttrPair("data.lightdash_user_attribute.check", "user_attribute_uuid", "lightdash_user_attribute.test", "user_attribute_uuid"),
					resource.TestCheckResourceAttr("data.lightdash_user_attribute.check", "name", "zzz_acc_test_user_attribute_update"),
					resource.TestCheckResourceAttr("data.lightdash_user_attribute.check", "groups.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.lightdash_user_attribute.check", "groups.*.values.*", "eu"),
					resource.TestCheckTypeSetElemAttr("data.lightdash_user_attribute.check", "groups.*.values.*", "jp"),
				),
			},
			{
				Config: providerConfig + clearConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "name", "zzz_acc_test_user_attribute_update"),
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "attribute_defaults.#", "0"),
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "groups.#", "0"),
					resource.TestCheckResourceAttr("data.lightdash_user_attribute.check", "groups.#", "0"),
				),
			},
			{
				Config: providerConfig + removeConfig,
			},
		},
	})
}

func TestAccUserAttributeResourceImport(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_user_attribute")
	}
	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}
	importConfig, err := ReadAccTestResource([]string{"resources", "lightdash_user_attribute", "import", "010_import.tf"})
	if err != nil {
		t.Fatalf("Failed to get importConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserAttributesGone("zzz_acc_test_user_attribute_import"),
		Steps: []resource.TestStep{
			{
				Config: providerConfig + importConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "name", "zzz_acc_test_user_attribute_import"),
					resource.TestCheckResourceAttr("lightdash_user_attribute.test", "groups.#", "1"),
				),
			},
			{
				Config:            providerConfig + importConfig,
				ResourceName:      "lightdash_user_attribute.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: userAttributeImportID,
			},
		},
	})
}

func userAttributeImportID(state *terraform.State) (string, error) {
	res, ok := state.RootModule().Resources["lightdash_user_attribute.test"]
	if !ok {
		return "", fmt.Errorf("resource not found in state for import")
	}
	organizationUUID := res.Primary.Attributes["organization_uuid"]
	userAttributeUUID := res.Primary.Attributes["user_attribute_uuid"]
	if organizationUUID == "" || userAttributeUUID == "" {
		return "", fmt.Errorf("organization_uuid or user_attribute_uuid is missing from state")
	}
	return userAttributeResourceID(organizationUUID, userAttributeUUID), nil
}

func testAccCheckUserAttributesGone(names ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		host, err := getLightdashUrl()
		if err != nil {
			return err
		}
		token, err := getLightdashApiKey()
		if err != nil {
			return err
		}
		client, err := api.NewClient(host, token, nil)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}
		attributes, err := services.NewUserAttributesService(client).List(context.Background())
		if err != nil {
			return err
		}
		for _, attribute := range attributes {
			for _, name := range names {
				if attribute.Name == name {
					return fmt.Errorf("user attribute %q still exists with UUID %s", name, attribute.UserAttributeUUID)
				}
			}
		}
		return nil
	}
}
