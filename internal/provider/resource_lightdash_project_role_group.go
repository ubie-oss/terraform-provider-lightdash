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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &projectRoleGroupResource{}
var _ resource.ResourceWithImportState = &projectRoleGroupResource{}

func NewProjectRoleGroupResource() resource.Resource {
	return &projectRoleGroupResource{}
}

// projectRoleGroupResource defines the resource implementation.
type projectRoleGroupResource struct {
	client *api.Client
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
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Assigns the role of a group at project level",
		Description:         "Assigns the role of a group at project level",
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
			"group_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash group UUID.",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Lightdash group's role.",
				Required:            true,
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "The timestamp of the last update.",
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
	group_uuid := plan.GroupUUID.ValueString()
	_, err := r.client.GetGroupV1(group_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading group",
			"Could not find group ID "+group_uuid+": "+err.Error(),
		)
		return
	}

	// Grant the project role to the group
	project_uuid := plan.ProjectUUID.ValueString()
	project_role := plan.ProjectRole
	grantedGroup, err := r.client.AddProjectAccessToGroupV1(project_uuid, group_uuid, project_role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error granting project role to group",
			"Could not grant project role to group, unexpected error: "+err.Error(),
		)
		return
	}

	// Set resources
	state_id := getProjectRoleGroupResourceId(plan.ProjectUUID.ValueString(), plan.GroupUUID.ValueString())
	plan.ID = types.StringValue(state_id)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	plan.ProjectUUID = types.StringValue(grantedGroup.ProjectUUID)
	plan.GroupUUID = types.StringValue(grantedGroup.GroupUUID)
	plan.ProjectRole = models.ProjectMemberRole(grantedGroup.Role)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectRoleGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Declare variables to import from state
	var projectUuid string
	var groupUuid string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project_uuid"), &projectUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("group_uuid"), &groupUuid)...)

	// Get current state
	var state projectGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve group details using the API
	groupUuid = state.GroupUUID.ValueString()
	groupsInProject, err := r.client.GetProjectGroupAccessesV1(projectUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading group",
			"Could not read group with UUID "+groupUuid+": "+err.Error(),
		)
		return
	}

	// Find the group in the groups of the project
	var group *api.GetProjectGroupAccessesV1Results
	found := false
	for i := range groupsInProject {
		if groupsInProject[i].GroupUUID == groupUuid {
			group = &groupsInProject[i]
			found = true
			break
		}
	}
	if !found {
		resp.Diagnostics.AddError(
			"Error Reading group",
			"Group with UUID "+groupUuid+" not found in project",
		)
		return
	}

	// Set the state values
	state.GroupUUID = types.StringValue(group.GroupUUID)
	state.ProjectRole = group.ProjectRole

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectRoleGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan projectGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing project group role
	project_uuid := plan.ProjectUUID.ValueString()
	group_uuid := plan.GroupUUID.ValueString()
	role := plan.ProjectRole
	updatedGroupAccess, err := r.client.UpdateProjectAccessForGroupV1(project_uuid, group_uuid, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating project group's role",
			"Could not update project group's role, unexpected error: "+err.Error(),
		)
		return
	}

	// Update the state
	plan.GroupUUID = types.StringValue(updatedGroupAccess.GroupUUID)
	plan.ProjectRole = models.ProjectMemberRole(updatedGroupAccess.Role)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectRoleGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state projectGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing project group role
	project_uuid := state.ProjectUUID.ValueString()
	group_uuid := state.GroupUUID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Revoking project role for group %s", group_uuid))
	err := r.client.RemoveProjectAccessFromGroupV1(project_uuid, group_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Revoking project group role: "+group_uuid+", project: "+project_uuid,
			"Could not revoke project group role, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *projectRoleGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Extract the resource ID
	extracted_strings, err := extractProjectRoleGroupResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	project_uuid := extracted_strings[0]
	group_uuid := extracted_strings[1]

	// Retrieve group details using the API
	groupUuid := group_uuid
	groupsInProject, err := r.client.GetProjectGroupAccessesV1(project_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading group",
			"Could not read group with UUID "+groupUuid+": "+err.Error(),
		)
		return
	}

	// Find the group in the groups of the project
	var group *api.GetProjectGroupAccessesV1Results
	found := false
	for i := range groupsInProject {
		if groupsInProject[i].GroupUUID == groupUuid {
			group = &groupsInProject[i]
			found = true
			break
		}
	}
	if !found {
		resp.Diagnostics.AddError(
			"Error Reading group",
			"Group with UUID "+groupUuid+" not found in project",
		)
		return
	}

	// Set the resource attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), project_uuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_uuid"), group_uuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role"), group.ProjectRole)...)
}

func getProjectRoleGroupResourceId(project_uuid string, group_uuid string) string {
	// Return the resource ID
	return fmt.Sprintf("projects/%s/access-groups/%s", project_uuid, group_uuid)
}

func extractProjectRoleGroupResourceId(input string) ([]string, error) {
	// Extract the captured groups
	pattern := `^projects/([^/]+)/access-groups/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	// Return the captured strings
	project_uuid := groups[0]
	group_uuid := groups[1]
	return []string{project_uuid, group_uuid}, nil
}
