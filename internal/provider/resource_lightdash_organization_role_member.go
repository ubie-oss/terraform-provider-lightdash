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

	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
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
	client      *api.Client
	roleService *services.RoleService
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
			},
		},
	}
}

func (r *organizationRoleMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.roleService = services.GetRoleService(client)
}

func (r *organizationRoleMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationRoleMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userUUID := plan.UserUUID.ValueString()
	orgUUID := plan.OrganizationUUID.ValueString()
	role := models.OrganizationMemberRole(plan.OrganizationRole.ValueString())

	assignment, err := r.roleService.AssignOrgUserRole(ctx, orgUUID, userUUID, role.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating organization member",
			"Could not update organization member, unexpected error: "+err.Error(),
		)
		return
	}

	orgRole, err := services.TerraformOrganizationRoleFromAssignment(assignment)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping organization role", err.Error())
		return
	}

	user, err := apiv1.GetOrganizationMemberByUuidV1(r.client, userUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading organization member",
			"Could not read organization member ID "+userUUID+": "+err.Error(),
		)
		return
	}

	plan.OrganizationUUID = types.StringValue(orgUUID)
	plan.UserUUID = types.StringValue(userUUID)
	plan.Email = types.StringValue(user.Email)
	plan.OrganizationRole = types.StringValue(orgRole.String())
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.ID = types.StringValue(getOrganizationRoleMemberResourceId(orgUUID, userUUID))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *organizationRoleMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var organizationUUID string
	var userUUID string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("organization_uuid"), &organizationUUID)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_uuid"), &userUUID)...)

	var state organizationRoleMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	assignment, err := r.roleService.GetOrgUserAssignment(ctx, organizationUUID, userUUID)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Warning Reading organization member",
			"Could not read organization role assignment for user "+userUUID+": "+err.Error(),
		)
		return
	}

	orgRole, err := services.TerraformOrganizationRoleFromAssignment(assignment)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping organization role", err.Error())
		return
	}

	user, err := apiv1.GetOrganizationMemberByUuidV1(r.client, userUUID)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Warning Reading organization member",
			"Could not read organization member ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.OrganizationUUID = types.StringValue(user.OrganizationUUID)
	state.UserUUID = types.StringValue(user.UserUUID)
	state.Email = types.StringValue(user.Email)
	state.OrganizationRole = types.StringValue(orgRole.String())

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *organizationRoleMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationRoleMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userUUID := plan.UserUUID.ValueString()
	orgUUID := plan.OrganizationUUID.ValueString()
	role := models.OrganizationMemberRole(plan.OrganizationRole.ValueString())

	assignment, err := r.roleService.AssignOrgUserRole(ctx, orgUUID, userUUID, role.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating organization member",
			"Could not update organization member, unexpected error: "+err.Error(),
		)
		return
	}

	orgRole, err := services.TerraformOrganizationRoleFromAssignment(assignment)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping organization role", err.Error())
		return
	}

	user, err := apiv1.GetOrganizationMemberByUuidV1(r.client, userUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading organization member",
			"Could not read organization member ID "+userUUID+": "+err.Error(),
		)
		return
	}

	plan.OrganizationUUID = types.StringValue(user.OrganizationUUID)
	plan.UserUUID = types.StringValue(user.UserUUID)
	plan.Email = types.StringValue(user.Email)
	plan.OrganizationRole = types.StringValue(orgRole.String())
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *organizationRoleMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationRoleMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userUUID := state.UserUUID.ValueString()
	orgUUID := state.OrganizationUUID.ValueString()
	assignment, err := r.roleService.AssignOrgUserRole(ctx, orgUUID, userUUID, models.ORGANIZATION_MEMBER_ROLE.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating organization member",
			"Could not update organization member, unexpected error: "+err.Error(),
		)
		return
	}

	orgRole, err := services.TerraformOrganizationRoleFromAssignment(assignment)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping organization role", err.Error())
		return
	}
	if orgRole != models.ORGANIZATION_MEMBER_ROLE {
		resp.Diagnostics.AddError(
			"Error Updating organization member",
			fmt.Sprintf("expected role %q after delete, got %q", models.ORGANIZATION_MEMBER_ROLE, orgRole),
		)
		return
	}
}

func (r *organizationRoleMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	extractedStrings, err := extractOrganizationRoleMemberResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	organizationUUID := extractedStrings[0]
	userUUID := extractedStrings[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_uuid"), organizationUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_uuid"), userUUID)...)
}

func getOrganizationRoleMemberResourceId(organization_uuid string, user_uuid string) string {
	return fmt.Sprintf("organizations/%s/users/%s", organization_uuid, user_uuid)
}

func extractOrganizationRoleMemberResourceId(input string) ([]string, error) {
	pattern := `^organizations/([^/]+)/users/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	return []string{groups[0], groups[1]}, nil
}
