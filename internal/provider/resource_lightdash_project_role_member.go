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

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &projectRoleMemberResource{}
var _ resource.ResourceWithImportState = &projectRoleMemberResource{}

func NewProjectRoleMemberResource() resource.Resource {
	return &projectRoleMemberResource{}
}

// LightdashProjectResource defines the resource implementation.
type projectRoleMemberResource struct {
	client *api.Client
}

// LightdashProjectResourceModel describes the resource data model.
type projectMemberResourceModel struct {
	ID          types.String             `tfsdk:"id"`
	ProjectUUID types.String             `tfsdk:"project_uuid"`
	UserUUID    types.String             `tfsdk:"user_uuid"`
	Email       types.String             `tfsdk:"email"`
	ProjectRole models.ProjectMemberRole `tfsdk:"role"`
	SendEmail   types.Bool               `tfsdk:"send_email"`
	LastUpdated types.String             `tfsdk:"last_updated"`
}

func (r *projectRoleMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_role_member"
}

func (r *projectRoleMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Lightash project role member",
		Description:         "Lightash project role member",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash project UUID.",
				Required:            true,
			},
			"user_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash user UUID.",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Lightdash user email.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Lightdash user's role.",
				Required:            true,
			},
			"send_email": schema.BoolAttribute{
				MarkdownDescription: "Send email to the user.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "Last updated time.",
				Computed:            true,
			},
		},
	}
}

func (r *projectRoleMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectRoleMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan projectMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the member
	user_uuid := plan.UserUUID.ValueString()
	projectMember, err := r.client.GetOrganizationMemberByUuidV1(user_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading organization member",
			"Could not find organization member ID "+user_uuid+": "+err.Error(),
		)
		return
	}

	// Grant the project role to the user
	project_uuid := plan.ProjectUUID.ValueString()
	project_role := plan.ProjectRole
	email := projectMember.Email
	send_email := plan.SendEmail.ValueBool()
	err = r.client.GrantProjectAccessToUserV1(project_uuid, email, project_role, send_email)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error granting project role",
			"Could not grant project role, unexpected error: "+err.Error(),
		)
		return
	}

	// Set resources
	// plan.OrganizationUUID = types.StringValue(created_space.OrganizationUUID)
	plan.UserUUID = types.StringValue(projectMember.UserUUID)
	plan.Email = types.StringValue(plan.Email.String())
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set resource ID
	state_id := fmt.Sprintf("projects/%s/access/%s",
		plan.ProjectUUID.ValueString(), plan.UserUUID.ValueString())
	plan.ID = types.StringValue(state_id)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectRoleMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state projectMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get a member from the API
	user_uuid := state.UserUUID.ValueString()
	projectMember, err := r.client.GetOrganizationMemberByUuidV1(user_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading organization member",
			"Could not read organization member ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// Set the state values
	state.UserUUID = types.StringValue(projectMember.UserUUID)
	state.Email = types.StringValue(projectMember.Email)
	state.ProjectRole = models.ProjectMemberRole(state.ProjectRole)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectRoleMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan projectMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing order
	project_uuid := plan.ProjectUUID.ValueString()
	user_uuid := plan.UserUUID.ValueString()
	role := models.ProjectMemberRole(plan.ProjectRole)
	err := r.client.UpdateProjectAccessToUserV1(project_uuid, user_uuid, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating project member's role",
			"Could not update project member's role, unexpected error: "+err.Error(),
		)
		return
	}

	// Update the state
	plan.ProjectRole = models.ProjectMemberRole(plan.ProjectRole)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectRoleMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state projectMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	project_uuid := state.ProjectUUID.ValueString()
	user_uuid := state.UserUUID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Revoking project role of %s", user_uuid))
	err := r.client.RevokeProjectAccessV1(project_uuid, user_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Revoking project role",
			"Could not revoke project role, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *projectRoleMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
