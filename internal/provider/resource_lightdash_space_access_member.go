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
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &spaceAccessMemberResource{}
	_ resource.ResourceWithConfigure   = &spaceAccessMemberResource{}
	_ resource.ResourceWithImportState = &spaceAccessMemberResource{}
)

func NewSpaceAccessMemberResource() resource.Resource {
	return &spaceAccessMemberResource{}
}

// spaceResource defines the resource implementation.
type spaceAccessMemberResource struct {
	client *api.Client
}

// spaceResourceModel describes the resource data model.
type spaceAccessMemberResourceModel struct {
	ID types.String `tfsdk:"id"`
	// The response from the API does not contain the organization UUID right now.
	// OrganizationUUID types.String `tfsdk:"organization_uuid"`
	ProjectUUID types.String `tfsdk:"project_uuid"`
	SpaceUUID   types.String `tfsdk:"space_uuid"`
	UserUUID    types.String `tfsdk:"user_uuid"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (r *spaceAccessMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_access_member"
}

func (r *spaceAccessMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Lightash space access member",
		Description:         "Lightash space access member",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// The response from the API does not contain the organization UUID right now.
			// "organization_uuid": schema.StringAttribute{
			// 	MarkdownDescription: "Lightdash organization UUID",
			// 	Computed:            true,
			// },
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash project UUID",
				Required:            true,
			},
			"space_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash space UUID",
				Required:            true,
			},
			"user_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash user UUID",
				Required:            true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the order.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *spaceAccessMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *spaceAccessMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan spaceAccessMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	project_uuid := plan.ProjectUUID.ValueString()
	space_uuid := plan.SpaceUUID.ValueString()
	user_uuid := plan.UserUUID.ValueString()

	// Check if the member is a memmber of the project.
	_, err := r.client.GetProjectMemberByUuidV1(project_uuid, user_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting project member",
			"Could not get project member, unexpected error: "+err.Error(),
		)
		return
	}

	// Create new space
	err = r.client.AddSpaceShareToUserV1(project_uuid, space_uuid, user_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding space access",
			"Could not add space access, unexpected error: "+err.Error(),
		)
		return
	}

	// Set resources
	// plan.OrganizationUUID = types.StringValue(created_space.OrganizationUUID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set resource ID
	state_id := getSpaceAccessMemberResourceId(
		plan.ProjectUUID.ValueString(),
		plan.SpaceUUID.ValueString(),
		plan.UserUUID.ValueString())
	plan.ID = types.StringValue(state_id)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *spaceAccessMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Declare variables to import from state
	var project_uuid string
	var space_uuid string
	var user_uuid string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project_uuid"), &project_uuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("space_uuid"), &space_uuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_uuid"), &user_uuid)...)

	// Get current state
	var state spaceAccessMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the space member
	project_uuid = state.ProjectUUID.ValueString()
	space_uuid = state.SpaceUUID.ValueString()
	user_uuid = state.UserUUID.ValueString()
	_, err := r.client.GetSpaceMemberV1(project_uuid, space_uuid, user_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading space",
			"Could not read space ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Set the state values
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *spaceAccessMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// We don't have to implement this because we don't support updates
	// Get current state
	var state spaceAccessMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *spaceAccessMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state spaceAccessMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing space
	project_uuid := state.ProjectUUID.ValueString()
	space_uuid := state.SpaceUUID.ValueString()
	user_uuid := state.UserUUID.ValueString()
	err := r.client.RevokeSpaceAccessV1(project_uuid, space_uuid, user_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting space",
			"Could not delete order, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *spaceAccessMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Extract the resource ID
	extracted_strings, err := extractSpaceAccessMemberResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	project_uuid := extracted_strings[0]
	space_uuid := extracted_strings[1]
	user_uuid := extracted_strings[2]

	// Set the resource attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), project_uuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_uuid"), space_uuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_uuid"), user_uuid)...)
}

func getSpaceAccessMemberResourceId(project_uuid string, space_uuid string, user_uuid string) string {
	// Return the resource ID
	return fmt.Sprintf("projects/%s/spaces/%s/access/%s", project_uuid, space_uuid, user_uuid)
}

func extractSpaceAccessMemberResourceId(input string) ([]string, error) {
	// Extract the captured groups
	pattern := `^projects/([^/]+)/spaces/([^/]+)/access/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	// Return the captured strings
	project_uuid := groups[0]
	space_uuid := groups[1]
	user_uuid := groups[1]
	return []string{project_uuid, space_uuid, user_uuid}, nil
}
