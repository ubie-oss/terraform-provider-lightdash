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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &projectRoleGroupResource{}
var _ resource.ResourceWithImportState = &projectRoleGroupResource{}

func NewProjectRoleGroupResource() resource.Resource {
	return &projectRoleGroupResource{}
}

// projectRoleGroupResource defines the resource implementation.
type projectRoleGroupResource struct {
	client      *api.Client
	roleService *services.RoleService
}

// projectGroupResourceModel describes the resource data model.
type projectGroupResourceModel struct {
	ID          types.String             `tfsdk:"id"`
	ProjectUUID types.String             `tfsdk:"project_uuid"`
	GroupUUID   types.String             `tfsdk:"group_uuid"`
	ProjectRole models.ProjectMemberRole `tfsdk:"role"`
	LastUpdated types.String             `tfsdk:"last_updated"`
}

func (r *projectRoleGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_role_group"
}

func (r *projectRoleGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_project_role_group.md")
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
		Description:         "Assigns the role of a group at project level",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource identifier. It is computed as `projects/<project_uuid>/groups/<group_uuid>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash project.",
				Required:            true,
			},
			"group_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash group.",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role assigned to the group within the project.",
				Required:            true,
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "The timestamp of the last Terraform update applied to the project role group resource.",
				Computed:            true,
			},
		},
	}
}

func (r *projectRoleGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.roleService = services.GetRoleService(client)
}

func (r *projectRoleGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan projectGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the group
	groupUUID := plan.GroupUUID.ValueString()
	_, err := apiv1.GetGroupV1(r.client, groupUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading group",
			"Could not find group ID "+groupUUID+": "+err.Error(),
		)
		return
	}

	orgUUID, err := r.roleService.OrganizationUUID(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving organization", err.Error())
		return
	}

	projectUUID := plan.ProjectUUID.ValueString()
	assignment, err := r.roleService.AssignProjectGroupRole(ctx, orgUUID, projectUUID, groupUUID, plan.ProjectRole.String(), false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error granting project role to group",
			"Could not grant project role to group, unexpected error: "+err.Error(),
		)
		return
	}

	projectRole, err := services.TerraformProjectRoleFromAssignment(assignment)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping project role", err.Error())
		return
	}

	stateID := getProjectRoleGroupResourceId(projectUUID, groupUUID)
	plan.ID = types.StringValue(stateID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	plan.ProjectUUID = types.StringValue(projectUUID)
	plan.GroupUUID = types.StringValue(groupUUID)
	plan.ProjectRole = projectRole

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *projectRoleGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var projectUUID string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project_uuid"), &projectUUID)...)

	var state projectGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupUUID := state.GroupUUID.ValueString()
	assignment, err := r.roleService.GetProjectGroupAssignment(ctx, projectUUID, groupUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading group",
			"Could not read group with UUID "+groupUUID+": "+err.Error(),
		)
		return
	}

	projectRole, err := services.TerraformProjectRoleFromAssignment(assignment)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping project role", err.Error())
		return
	}

	state.GroupUUID = types.StringValue(groupUUID)
	state.ProjectRole = projectRole

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *projectRoleGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectGroupResourceModel
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
	groupUUID := plan.GroupUUID.ValueString()
	assignment, err := r.roleService.UpdateProjectGroupRole(ctx, orgUUID, projectUUID, groupUUID, plan.ProjectRole.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating project group's role",
			"Could not update project group's role, unexpected error: "+err.Error(),
		)
		return
	}

	projectRole, err := services.TerraformProjectRoleFromAssignment(assignment)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping project role", err.Error())
		return
	}

	plan.GroupUUID = types.StringValue(groupUUID)
	plan.ProjectRole = projectRole
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *projectRoleGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectUUID := state.ProjectUUID.ValueString()
	groupUUID := state.GroupUUID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Revoking project role for group %s", groupUUID))
	err := r.roleService.RemoveProjectGroupRole(ctx, projectUUID, groupUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Revoking project group role: "+groupUUID+", project: "+projectUUID,
			"Could not revoke project group role, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *projectRoleGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	extractedStrings, err := extractProjectRoleGroupResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	projectUUID := extractedStrings[0]
	groupUUID := extractedStrings[1]

	assignment, err := r.roleService.GetProjectGroupAssignment(ctx, projectUUID, groupUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading group",
			"Could not read group with UUID "+groupUUID+": "+err.Error(),
		)
		return
	}

	projectRole, err := services.TerraformProjectRoleFromAssignment(assignment)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping project role", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), projectUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_uuid"), groupUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role"), projectRole)...)
}

func getProjectRoleGroupResourceId(project_uuid string, group_uuid string) string {
	return fmt.Sprintf("projects/%s/access-groups/%s", project_uuid, group_uuid)
}

func extractProjectRoleGroupResourceId(input string) ([]string, error) {
	pattern := `^projects/([^/]+)/access-groups/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	return []string{groups[0], groups[1]}, nil
}
