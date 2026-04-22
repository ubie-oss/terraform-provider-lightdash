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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &userAttributeResource{}
	_ resource.ResourceWithConfigure   = &userAttributeResource{}
	_ resource.ResourceWithImportState = &userAttributeResource{}
)

func NewUserAttributeResource() resource.Resource {
	return &userAttributeResource{}
}

// userAttributeResource defines the resource implementation.
type userAttributeResource struct {
	client *api.Client
}

// userAttributeAssignmentModel represents one user assignment in the resource model.
type userAttributeAssignmentModel struct {
	UserUUID types.String `tfsdk:"user_uuid"`
	Value    types.String `tfsdk:"value"`
}

// groupAttributeAssignmentModel represents one group assignment in the resource model.
type groupAttributeAssignmentModel struct {
	GroupUUID types.String `tfsdk:"group_uuid"`
	Value     types.String `tfsdk:"value"`
}

// userAttributeResourceModel describes the resource data model.
type userAttributeResourceModel struct {
	ID                types.String `tfsdk:"id"`
	OrganizationUUID  types.String `tfsdk:"organization_uuid"`
	UserAttributeUUID types.String `tfsdk:"user_attribute_uuid"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	AttributeDefault  types.String `tfsdk:"attribute_default"`
	Users             types.Set    `tfsdk:"users"`
	Groups            types.Set    `tfsdk:"groups"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

func userAttributeUserObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"user_uuid": types.StringType,
			"value":     types.StringType,
		},
	}
}

func userAttributeGroupObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"group_uuid": types.StringType,
			"value":      types.StringType,
		},
	}
}

func (r *userAttributeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_attribute"
}

func (r *userAttributeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_lightdash_user_attribute.md")
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
				Computed:            true,
				MarkdownDescription: "The resource identifier. It is computed as `organizations/<organization_uuid>/attributes/<user_attribute_uuid>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
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
				MarkdownDescription: "The name of the user attribute.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the user attribute.",
				Optional:            true,
			},
			"attribute_default": schema.StringAttribute{
				MarkdownDescription: "The default value of the user attribute. Set to `null` to have no default.",
				Optional:            true,
			},
			"users": schema.SetNestedAttribute{
				MarkdownDescription: "A set of user-specific values for this attribute.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_uuid": schema.StringAttribute{
							MarkdownDescription: "The UUID of the Lightdash user.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The attribute value assigned to the user.",
							Required:            true,
						},
					},
				},
			},
			"groups": schema.SetNestedAttribute{
				MarkdownDescription: "A set of group-specific values for this attribute.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group_uuid": schema.StringAttribute{
							MarkdownDescription: "The UUID of the Lightdash group.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The attribute value assigned to the group.",
							Required:            true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the user attribute was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userAttributeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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
}

func (r *userAttributeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan userAttributeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the API request from the plan
	createRequest, err := buildCreateUserAttributeRequest(ctx, &plan, &resp.Diagnostics)
	if err != nil || resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Creating user attribute with name: %s", plan.Name.ValueString()))
	created, apiErr := r.client.CreateUserAttributeV1(createRequest)
	if apiErr != nil {
		resp.Diagnostics.AddError(
			"Error creating user attribute",
			"Could not create user attribute, unexpected error: "+apiErr.Error(),
		)
		return
	}

	// Map the API response back into state
	if mapErr := applyUserAttributeToState(ctx, created, &plan, &resp.Diagnostics); mapErr != nil {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userAttributeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userAttributeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userAttributeUuid := state.UserAttributeUUID.ValueString()

	found, err := r.client.GetUserAttributeV1(userAttributeUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading user attribute",
			"Could not read user attribute ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// If the user attribute is not found, remove it from state
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if mapErr := applyUserAttributeToState(ctx, found, &state, &resp.Diagnostics); mapErr != nil {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userAttributeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state userAttributeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userAttributeUuid := state.UserAttributeUUID.ValueString()

	updateRequest, err := buildCreateUserAttributeRequest(ctx, &plan, &resp.Diagnostics)
	if err != nil || resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Updating user attribute %s", userAttributeUuid))
	updated, apiErr := r.client.UpdateUserAttributeV1(userAttributeUuid, updateRequest)
	if apiErr != nil {
		resp.Diagnostics.AddError(
			"Error Updating user attribute",
			fmt.Sprintf("Could not update user attribute with UUID '%s', unexpected error: %s", userAttributeUuid, apiErr.Error()),
		)
		return
	}

	if mapErr := applyUserAttributeToState(ctx, updated, &plan, &resp.Diagnostics); mapErr != nil {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userAttributeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userAttributeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userAttributeUuid := state.UserAttributeUUID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting user attribute %s", userAttributeUuid))
	if err := r.client.DeleteUserAttributeV1(userAttributeUuid); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting user attribute",
			"Could not delete user attribute, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *userAttributeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	extracted, err := extractUserAttributeResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	organizationUuid := extracted[0]
	userAttributeUuid := extracted[1]

	imported, err := r.client.GetUserAttributeV1(userAttributeUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting user attribute",
			fmt.Sprintf("Could not get user attribute with UUID %s, unexpected error: %s", userAttributeUuid, err.Error()),
		)
		return
	}
	if imported == nil {
		resp.Diagnostics.AddError(
			"User attribute not found",
			fmt.Sprintf("No user attribute found with UUID %s in organization %s", userAttributeUuid, organizationUuid),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), getUserAttributeResourceId(imported.OrganizationUUID, imported.UUID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_uuid"), imported.OrganizationUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_attribute_uuid"), imported.UUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), imported.Name)...)

	if imported.Description != nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), *imported.Description)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), types.StringNull())...)
	}

	if imported.AttributeDefault != nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("attribute_default"), *imported.AttributeDefault)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("attribute_default"), types.StringNull())...)
	}

	usersSet, diagUsers := buildUsersSet(ctx, imported.Users)
	resp.Diagnostics.Append(diagUsers...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("users"), usersSet)...)

	groupsSet, diagGroups := buildGroupsSet(ctx, imported.Groups)
	resp.Diagnostics.Append(diagGroups...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("groups"), groupsSet)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), imported.CreatedAt)...)
}

// buildCreateUserAttributeRequest converts a resource model into an API create/update request.
func buildCreateUserAttributeRequest(ctx context.Context, plan *userAttributeResourceModel, diags *diag.Diagnostics) (*models.CreateUserAttribute, error) {
	req := &models.CreateUserAttribute{
		Name:   plan.Name.ValueString(),
		Users:  []models.CreateUserAttributeUserValue{},
		Groups: []models.UserAttributeGroupValue{},
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		desc := plan.Description.ValueString()
		req.Description = &desc
	}

	if !plan.AttributeDefault.IsNull() && !plan.AttributeDefault.IsUnknown() {
		def := plan.AttributeDefault.ValueString()
		req.AttributeDefault = &def
	}

	if !plan.Users.IsNull() && !plan.Users.IsUnknown() {
		var users []userAttributeAssignmentModel
		diags.Append(plan.Users.ElementsAs(ctx, &users, false)...)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract users")
		}
		for _, u := range users {
			req.Users = append(req.Users, models.CreateUserAttributeUserValue{
				UserUUID: u.UserUUID.ValueString(),
				Value:    u.Value.ValueString(),
			})
		}
	}

	if !plan.Groups.IsNull() && !plan.Groups.IsUnknown() {
		var groups []groupAttributeAssignmentModel
		diags.Append(plan.Groups.ElementsAs(ctx, &groups, false)...)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to extract groups")
		}
		for _, g := range groups {
			req.Groups = append(req.Groups, models.UserAttributeGroupValue{
				GroupUUID: g.GroupUUID.ValueString(),
				Value:     g.Value.ValueString(),
			})
		}
	}

	return req, nil
}

// applyUserAttributeToState maps the API response into the resource model.
func applyUserAttributeToState(ctx context.Context, ua *models.UserAttribute, state *userAttributeResourceModel, diags *diag.Diagnostics) error {
	state.ID = types.StringValue(getUserAttributeResourceId(ua.OrganizationUUID, ua.UUID))
	state.OrganizationUUID = types.StringValue(ua.OrganizationUUID)
	state.UserAttributeUUID = types.StringValue(ua.UUID)
	state.Name = types.StringValue(ua.Name)

	if ua.Description != nil {
		state.Description = types.StringValue(*ua.Description)
	} else {
		state.Description = types.StringNull()
	}

	if ua.AttributeDefault != nil {
		state.AttributeDefault = types.StringValue(*ua.AttributeDefault)
	} else {
		state.AttributeDefault = types.StringNull()
	}

	usersSet, diagUsers := buildUsersSet(ctx, ua.Users)
	diags.Append(diagUsers...)
	if diags.HasError() {
		return fmt.Errorf("failed to build users set")
	}
	state.Users = usersSet

	groupsSet, diagGroups := buildGroupsSet(ctx, ua.Groups)
	diags.Append(diagGroups...)
	if diags.HasError() {
		return fmt.Errorf("failed to build groups set")
	}
	state.Groups = groupsSet

	state.CreatedAt = types.StringValue(ua.CreatedAt)
	return nil
}

func buildUsersSet(ctx context.Context, users []models.UserAttributeUserValue) (types.Set, diag.Diagnostics) {
	elements := make([]userAttributeAssignmentModel, 0, len(users))
	for _, u := range users {
		elements = append(elements, userAttributeAssignmentModel{
			UserUUID: types.StringValue(u.UserUUID),
			Value:    types.StringValue(u.Value),
		})
	}
	return types.SetValueFrom(ctx, userAttributeUserObjectType(), elements)
}

func buildGroupsSet(ctx context.Context, groups []models.UserAttributeGroupValue) (types.Set, diag.Diagnostics) {
	elements := make([]groupAttributeAssignmentModel, 0, len(groups))
	for _, g := range groups {
		elements = append(elements, groupAttributeAssignmentModel{
			GroupUUID: types.StringValue(g.GroupUUID),
			Value:     types.StringValue(g.Value),
		})
	}
	return types.SetValueFrom(ctx, userAttributeGroupObjectType(), elements)
}

func getUserAttributeResourceId(organizationUuid string, userAttributeUuid string) string {
	return fmt.Sprintf("organizations/%s/attributes/%s", organizationUuid, userAttributeUuid)
}

func extractUserAttributeResourceId(input string) ([]string, error) {
	pattern := `^organizations/([^/]+)/attributes/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}
	return []string{groups[0], groups[1]}, nil
}
