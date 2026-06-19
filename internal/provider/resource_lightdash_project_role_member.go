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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &projectRoleMemberResource{}
	_ resource.ResourceWithConfigure   = &projectRoleMemberResource{}
	_ resource.ResourceWithImportState = &projectRoleMemberResource{}
)

func NewProjectRoleMemberResource() resource.Resource {
	return &projectRoleMemberResource{}
}

// projectRoleMemberResource defines the resource implementation.
type projectRoleMemberResource struct {
	client      *api.Client
	roleService *services.RoleService
}

// projectMemberResourceModel describes the resource data model.
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
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_project_role_member.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Manages the role of a member at the project level.",
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

func (r *projectRoleMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userUUID := plan.UserUUID.ValueString()
	orgMember, err := apiv1.GetOrganizationMemberByUuidV1(r.client, userUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading organization member",
			"Could not find organization member ID "+userUUID+": "+err.Error(),
		)
		return
	}
	if !orgMember.IsActive {
		resp.Diagnostics.AddError(
			"Error Reading organization member",
			"Organization member ID "+userUUID+" is not active",
		)
		return
	}

	projectUUID := plan.ProjectUUID.ValueString()
	projectRole, err := r.assignProjectUserRole(
		ctx, orgMember.OrganizationUUID, projectUUID, userUUID,
		plan.ProjectRole.String(),
		plan.SendEmail.ValueBool(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error granting project role",
			"Could not grant project role, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ProjectRole = projectRole
	plan.UserUUID = types.StringValue(orgMember.UserUUID)
	plan.Email = types.StringValue(orgMember.Email)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.ID = types.StringValue(getProjectRoleMemberResourceId(projectUUID, userUUID))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *projectRoleMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectUUID := state.ProjectUUID.ValueString()
	userUUID := state.UserUUID.ValueString()

	assignment, err := r.roleService.GetProjectUserAssignment(ctx, projectUUID, userUUID)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Warning Reading project role assignment",
			"Could not read project role assignment for user "+userUUID+": "+err.Error(),
		)
		return
	}

	projectRole, err := services.TerraformProjectRoleFromAssignment(assignment)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping project role", err.Error())
		return
	}

	orgMember, err := apiv1.GetOrganizationMemberByUuidV1(r.client, userUUID)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Warning Reading organization member",
			"Could not read organization member ID "+userUUID+": "+err.Error(),
		)
		return
	}

	state.UserUUID = types.StringValue(orgMember.UserUUID)
	state.Email = types.StringValue(orgMember.Email)
	state.ProjectRole = projectRole

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *projectRoleMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgUUID, err := r.roleService.OrganizationUUID(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving organization", err.Error())
		return
	}

	projectUUID := plan.ProjectUUID.ValueString()
	userUUID := plan.UserUUID.ValueString()
	// Update must not send notification emails; only Create honors send_email (AGENTS.md).
	projectRole, err := r.assignProjectUserRole(ctx, orgUUID, projectUUID, userUUID, plan.ProjectRole.String(), false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating project member's role",
			"Could not update project member's role, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ProjectRole = projectRole
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *projectRoleMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectUUID := state.ProjectUUID.ValueString()
	userUUID := state.UserUUID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Revoking project role of %s", userUUID))
	err := r.roleService.RemoveProjectUserRole(ctx, projectUUID, userUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Revoking project role",
			"Could not revoke project role, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *projectRoleMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	extractedStrings, err := extractProjectRoleMemberResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	projectUUID := extractedStrings[0]
	userUUID := extractedStrings[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), projectUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_uuid"), userUUID)...)
}

func getProjectRoleMemberResourceId(project_uuid string, user_uuid string) string {
	return fmt.Sprintf("projects/%s/access/%s", project_uuid, user_uuid)
}

func (r *projectRoleMemberResource) assignProjectUserRole(
	ctx context.Context,
	orgUUID, projectUUID, userUUID, role string,
	sendEmail bool,
) (models.ProjectMemberRole, error) {
	assignment, err := r.roleService.AssignProjectUserRole(ctx, orgUUID, projectUUID, userUUID, role, sendEmail)
	if err != nil {
		return "", err
	}
	return services.TerraformProjectRoleFromAssignment(assignment)
}

func extractProjectRoleMemberResourceId(input string) ([]string, error) {
	pattern := `^projects/([^/]+)/access/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	return []string{groups[0], groups[1]}, nil
}
