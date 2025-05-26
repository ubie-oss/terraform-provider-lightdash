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
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &organizationRoleMemberResource{}
	_ resource.ResourceWithConfigure   = &organizationRoleMemberResource{}
	_ resource.ResourceWithImportState = &organizationRoleMemberResource{}
)

func NewOrganizationRoleMemberResource() resource.Resource {
	return &organizationRoleMemberResource{}
}

// organizationRoleMemberResource defines the resource implementation.
type organizationRoleMemberResource struct {
	client *api.Client
}

// organizationRoleMemberResourceModel describes the resource data model.
type organizationRoleMemberResourceModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationUUID types.String `tfsdk:"organization_uuid"`
	UserUUID         types.String `tfsdk:"user_uuid"`
	Email            types.String `tfsdk:"email"`
	OrganizationRole types.String `tfsdk:"role"`
	LastUpdated      types.String `tfsdk:"last_updated"`
}

func (r *organizationRoleMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_role_member"
}

func (r *organizationRoleMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_organization_role_member.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: markdownDescription,
		Description:         "Lightash the role of a member at organization level",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource identifier. It is computed as `organizations/<organization_uuid>/users/<user_uuid>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
				Required:            true,
			},
			"user_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash user.",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address of the Lightdash user.",
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The organization role assigned to the user.",
				Required:            true,
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "The timestamp of the last Terraform update applied to the organization role member.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *organizationRoleMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	r.client = client
}

func (r *organizationRoleMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// We assume the an user already has any role in the organization.
	// So, we change the role of the user.
	var plan organizationRoleMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing organization member
	user_uuid := plan.UserUUID.ValueString()
	role := models.OrganizationMemberRole(plan.OrganizationRole.ValueString())
	user, err := r.client.UpdateOrganizationMemberV1(user_uuid, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating organization member",
			"Could not update organization member, unexpected error: "+err.Error(),
		)
		return
	}
	// Update the state
	plan.OrganizationUUID = types.StringValue(user.OrganizationUUID)
	plan.UserUUID = types.StringValue(user.UserUUID)
	plan.Email = types.StringValue(user.Email)
	plan.OrganizationRole = types.StringValue(user.OrganizationRole.String())
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set resource ID
	state_id := getOrganizationRoleMemberResourceId(user.OrganizationUUID, user.UserUUID)
	plan.ID = types.StringValue(state_id)

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *organizationRoleMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Declare variables to import from state
	var organization_uuid string
	var user_uuid string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("organization_uuid"), &organization_uuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_uuid"), &user_uuid)...)

	// Get current state
	var state organizationRoleMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get space
	user_uuid = state.UserUUID.ValueString()
	user, err := r.client.GetOrganizationMemberByUuidV1(user_uuid)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Warning Reading organization member",
			"Could not read organization member ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Set the state values
	state.OrganizationUUID = types.StringValue(user.OrganizationUUID)
	state.UserUUID = types.StringValue(user.UserUUID)
	state.Email = types.StringValue(user.Email)
	state.OrganizationRole = types.StringValue(user.OrganizationRole.String())

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *organizationRoleMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan organizationRoleMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing organization member
	user_uuid := plan.UserUUID.ValueString()
	role := models.OrganizationMemberRole(plan.OrganizationRole.ValueString())
	user, err := r.client.UpdateOrganizationMemberV1(user_uuid, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating organization member",
			"Could not update organization member, unexpected error: "+err.Error(),
		)
		return
	}
	// Update the state
	plan.OrganizationUUID = types.StringValue(user.OrganizationUUID)
	plan.UserUUID = types.StringValue(user.UserUUID)
	plan.Email = types.StringValue(user.Email)
	plan.OrganizationRole = types.StringValue(user.OrganizationRole.String())
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *organizationRoleMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state organizationRoleMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the role of the user to "member"
	user_uuid := state.UserUUID.ValueString()
	role := models.ORGANIZATION_MEMBER_ROLE
	user, err := r.client.UpdateOrganizationMemberV1(user_uuid, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating organization member",
			"Could not update organization member, unexpected error: "+err.Error(),
		)
		return
	}
	// Check if the updated role is "member"
	if user.OrganizationRole != models.ORGANIZATION_MEMBER_ROLE {
		resp.Diagnostics.AddError(
			"Error Updating organization member",
			"Could not update organization member, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *organizationRoleMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Extract the resource ID
	extracted_strings, err := extractOrganizationRoleMemberResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	project_uuid := extracted_strings[0]
	user_uuid := extracted_strings[1]

	// Set the resource attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_uuid"), project_uuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_uuid"), user_uuid)...)
}

func getOrganizationRoleMemberResourceId(organization_uuid string, user_uuid string) string {
	// Return the resource ID
	return fmt.Sprintf("organizations/%s/users/%s", organization_uuid, user_uuid)
}

func extractOrganizationRoleMemberResourceId(input string) ([]string, error) {
	// Extract the captured groups
	pattern := `^organizations/([^/]+)/users/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	// Return the captured strings
	organization_uuid := groups[0]
	user_uuid := groups[1]
	return []string{organization_uuid, user_uuid}, nil
}
