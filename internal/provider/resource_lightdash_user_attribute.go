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
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

var (
	_ resource.Resource                   = &userAttributeResource{}
	_ resource.ResourceWithConfigure      = &userAttributeResource{}
	_ resource.ResourceWithImportState    = &userAttributeResource{}
	_ resource.ResourceWithValidateConfig = &userAttributeResource{}
)

var userAttributeGroupValueObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"group_uuid": types.StringType,
		"values":     types.SetType{ElemType: types.StringType},
	},
}

func NewUserAttributeResource() resource.Resource {
	return &userAttributeResource{}
}

type userAttributeResource struct {
	client  *api.Client
	service *services.UserAttributesService
}

type userAttributeResourceModel struct {
	ID                types.String `tfsdk:"id"`
	OrganizationUUID  types.String `tfsdk:"organization_uuid"`
	UserAttributeUUID types.String `tfsdk:"user_attribute_uuid"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	AttributeDefaults types.Set    `tfsdk:"attribute_defaults"`
	Groups            types.Set    `tfsdk:"groups"`
}

type userAttributeGroupModel struct {
	GroupUUID types.String `tfsdk:"group_uuid"`
	Values    types.Set    `tfsdk:"values"`
}

func (r *userAttributeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_attribute"
}

func (r *userAttributeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_user_attribute.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Manages a Lightdash user attribute",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The resource identifier. It is computed as `organizations/<organization_uuid>/user-attributes/<user_attribute_uuid>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization that the user attribute belongs to.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_attribute_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash user attribute.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user attribute. It is the name referenced from dbt, such as `required_attributes` and `${lightdash.attributes.<name>}`.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the user attribute. It defaults to an empty string.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"attribute_defaults": schema.SetAttribute{
				MarkdownDescription: "The default values for users who don't have their own value in this user attribute. Omit it to have no default values. It must not be an empty set.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"groups": schema.SetNestedAttribute{
				MarkdownDescription: "The values of the user attribute assigned to groups. This is authoritative, so group values that are not in the configuration are removed.",
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(types.SetValueMust(userAttributeGroupValueObjectType, []attr.Value{})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group_uuid": schema.StringAttribute{
							MarkdownDescription: "The UUID of the Lightdash group, which must belong to the organization. It must be written in lowercase.",
							Required:            true,
						},
						"values": schema.SetAttribute{
							MarkdownDescription: "The values of the user attribute for the group. At least one value is required.",
							ElementType:         types.StringType,
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (r *userAttributeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
	r.service = services.NewUserAttributesService(client)
}

func (r *userAttributeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config userAttributeResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateUserAttributeConfig(ctx, config)...)
}

func validateUserAttributeConfig(ctx context.Context, config userAttributeResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if !config.Name.IsNull() && !config.Name.IsUnknown() && strings.TrimSpace(config.Name.ValueString()) == "" {
		diags.AddAttributeError(path.Root("name"), "Empty name", "name must not be empty.")
	}
	if isKnownEmptySet(config.AttributeDefaults) {
		diags.AddAttributeError(path.Root("attribute_defaults"), "Empty attribute_defaults", "Omit attribute_defaults to have no default values. It must not be an empty set.")
	}
	if hasNullElement(config.AttributeDefaults) {
		diags.AddAttributeError(path.Root("attribute_defaults"), "Null value", "attribute_defaults must not contain null.")
	}

	groups, groupDiags := userAttributeGroupsFromSet(ctx, config.Groups)
	diags.Append(groupDiags...)
	if groupDiags.HasError() {
		return diags
	}

	uuids := make([]string, 0, len(groups))
	for _, group := range groups {
		if !group.GroupUUID.IsNull() && !group.GroupUUID.IsUnknown() {
			uuid := group.GroupUUID.ValueString()
			if uuid != strings.ToLower(uuid) {
				diags.AddAttributeError(path.Root("groups"), "Invalid group_uuid", fmt.Sprintf("The group %q must be written in lowercase, as Lightdash returns it.", uuid))
			}
			uuids = append(uuids, uuid)
		}
		if isKnownEmptySet(group.Values) {
			diags.AddAttributeError(path.Root("groups"), "Empty values", "Each group must have at least one value. Remove the group to assign no value.")
		}
		if hasNullElement(group.Values) {
			diags.AddAttributeError(path.Root("groups"), "Null value", "A group must not have null in its values.")
		}
	}
	for _, duplicate := range findDuplicateStrings(uuids) {
		diags.AddAttributeError(path.Root("groups"), "Duplicate group_uuid", fmt.Sprintf("The group %q appears more than once. Merge its values into a single element.", duplicate))
	}
	return diags
}

func (r *userAttributeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userAttributeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	desired, diags := userAttributeFromResourceModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.checkGroupsBelongToOrganization(ctx, desired.Groups)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.service.Create(ctx, desired)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Lightdash user attribute", err.Error())
		return
	}
	state, diags := userAttributeToResourceModel(ctx, *created)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userAttributeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userAttributeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attribute, err := r.service.GetByUUID(ctx, state.OrganizationUUID.ValueString(), state.UserAttributeUUID.ValueString())
	if err != nil {
		if errors.Is(err, services.ErrUserAttributeNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read Lightdash user attribute", err.Error())
		return
	}

	updated, diags := userAttributeToResourceModel(ctx, *attribute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
}

func (r *userAttributeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userAttributeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state userAttributeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	desired, diags := userAttributeFromResourceModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	desired.OrganizationUUID = state.OrganizationUUID.ValueString()
	desired.UserAttributeUUID = state.UserAttributeUUID.ValueString()
	resp.Diagnostics.Append(r.checkGroupsBelongToOrganization(ctx, desired.Groups)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.service.Update(ctx, desired.OrganizationUUID, desired)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Lightdash user attribute", err.Error())
		return
	}
	next, diags := userAttributeToResourceModel(ctx, *updated)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

func (r *userAttributeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userAttributeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.service.Delete(ctx, state.OrganizationUUID.ValueString(), state.UserAttributeUUID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete Lightdash user attribute", err.Error())
	}
}

func (r *userAttributeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	organizationUUID, userAttributeUUID, err := extractUserAttributeResourceID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	tokenOrganizationUUID, err := services.GetOrganizationUUID(ctx, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the organization", err.Error())
		return
	}
	if !strings.EqualFold(tokenOrganizationUUID, organizationUUID) {
		resp.Diagnostics.AddError(
			"User attribute is in another organization",
			fmt.Sprintf("Import ID organization %s does not match the organization %s of the API token.", organizationUUID, tokenOrganizationUUID),
		)
		return
	}

	attribute, err := r.service.GetByUUID(ctx, tokenOrganizationUUID, userAttributeUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to import Lightdash user attribute", err.Error())
		return
	}
	state, diags := userAttributeToResourceModel(ctx, *attribute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userAttributeResource) checkGroupsBelongToOrganization(ctx context.Context, groups []models.UserAttributeValue) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(groups) == 0 {
		return diags
	}
	knownGroups, err := services.NewOrganizationGroupsService(r.client).GetOrganizationGroups(ctx)
	if err != nil {
		diags.AddError("Unable to read organization groups", err.Error())
		return diags
	}
	known := make(map[string]struct{}, len(knownGroups))
	for _, group := range knownGroups {
		known[strings.ToLower(group.GroupUUID)] = struct{}{}
	}
	missing := make([]string, 0)
	for _, group := range groups {
		if _, ok := known[strings.ToLower(group.UUID)]; !ok {
			missing = append(missing, group.UUID)
		}
	}
	if len(missing) > 0 {
		diags.AddAttributeError(path.Root("groups"), "Unknown group_uuid", fmt.Sprintf("These groups are not in the organization: %s.", strings.Join(missing, ", ")))
	}
	return diags
}

func userAttributeGroupsFromSet(ctx context.Context, set types.Set) ([]userAttributeGroupModel, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}
	var groups []userAttributeGroupModel
	diags := set.ElementsAs(ctx, &groups, false)
	return groups, diags
}

func userAttributeFromResourceModel(ctx context.Context, model userAttributeResourceModel) (models.UserAttribute, diag.Diagnostics) {
	var diags diag.Diagnostics
	attribute := models.UserAttribute{
		OrganizationUUID:  model.OrganizationUUID.ValueString(),
		UserAttributeUUID: model.UserAttributeUUID.ValueString(),
		Name:              model.Name.ValueString(),
		Description:       model.Description.ValueString(),
		Groups:            []models.UserAttributeValue{},
	}
	if !model.AttributeDefaults.IsNull() && !model.AttributeDefaults.IsUnknown() {
		defaults, defaultDiags := stringSetToStringSlice(ctx, model.AttributeDefaults)
		diags.Append(defaultDiags...)
		attribute.AttributeDefaults = defaults
	}
	groups, groupDiags := userAttributeGroupsFromSet(ctx, model.Groups)
	diags.Append(groupDiags...)
	for _, group := range groups {
		values, valueDiags := stringSetToStringSlice(ctx, group.Values)
		diags.Append(valueDiags...)
		attribute.Groups = append(attribute.Groups, models.UserAttributeValue{
			UUID:   group.GroupUUID.ValueString(),
			Values: values,
		})
	}
	return attribute, diags
}

func userAttributeToResourceModel(ctx context.Context, attribute models.UserAttribute) (userAttributeResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	defaults := types.SetNull(types.StringType)
	if len(attribute.AttributeDefaults) > 0 {
		var defaultDiags diag.Diagnostics
		defaults, defaultDiags = stringSliceToStringSet(ctx, uniqueStrings(attribute.AttributeDefaults))
		diags.Append(defaultDiags...)
	}
	groups, groupDiags := userAttributeGroupsToSet(ctx, attribute.Groups)
	diags.Append(groupDiags...)
	model := userAttributeResourceModel{
		ID:                types.StringValue(userAttributeResourceID(attribute.OrganizationUUID, attribute.UserAttributeUUID)),
		OrganizationUUID:  types.StringValue(attribute.OrganizationUUID),
		UserAttributeUUID: types.StringValue(attribute.UserAttributeUUID),
		Name:              types.StringValue(attribute.Name),
		Description:       types.StringValue(attribute.Description),
		AttributeDefaults: defaults,
		Groups:            groups,
	}
	return model, diags
}

func userAttributeGroupsToSet(ctx context.Context, groups []models.UserAttributeValue) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	merged := make(map[string][]string)
	order := make([]string, 0, len(groups))
	for _, group := range groups {
		if _, ok := merged[group.UUID]; !ok {
			order = append(order, group.UUID)
		}
		merged[group.UUID] = append(merged[group.UUID], group.Values...)
	}

	elements := make([]attr.Value, 0, len(order))
	for _, groupUUID := range order {
		values, valueDiags := stringSliceToStringSet(ctx, uniqueStrings(merged[groupUUID]))
		diags.Append(valueDiags...)
		element, elementDiags := types.ObjectValue(userAttributeGroupValueObjectType.AttrTypes, map[string]attr.Value{
			"group_uuid": types.StringValue(groupUUID),
			"values":     values,
		})
		diags.Append(elementDiags...)
		elements = append(elements, element)
	}
	if diags.HasError() {
		return types.SetNull(userAttributeGroupValueObjectType), diags
	}
	set, setDiags := types.SetValue(userAttributeGroupValueObjectType, elements)
	diags.Append(setDiags...)
	return set, diags
}

func userAttributeResourceID(organizationUUID string, userAttributeUUID string) string {
	return fmt.Sprintf("organizations/%s/user-attributes/%s", organizationUUID, userAttributeUUID)
}

func extractUserAttributeResourceID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 4 || parts[0] != "organizations" || parts[2] != "user-attributes" || parts[1] == "" || parts[3] == "" {
		return "", "", fmt.Errorf("expected organizations/<organization_uuid>/user-attributes/<user_attribute_uuid>, got %q", id)
	}
	return strings.ToLower(parts[1]), strings.ToLower(parts[3]), nil
}
