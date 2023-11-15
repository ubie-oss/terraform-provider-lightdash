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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &spaceResource{}
	_ resource.ResourceWithConfigure   = &spaceResource{}
	_ resource.ResourceWithImportState = &spaceResource{}
)

func NewSpaceResource() resource.Resource {
	return &spaceResource{}
}

// spaceResource defines the resource implementation.
type spaceResource struct {
	client *api.Client
}

// spaceResourceModel describes the resource data model.
type spaceResourceModel struct {
	ID types.String `tfsdk:"id"`
	// The response from the API does not contain the organization UUID right now.
	// OrganizationUUID types.String `tfsdk:"organization_uuid"`
	ProjectUUID      types.String                    `tfsdk:"project_uuid"`
	SpaceUUID        types.String                    `tfsdk:"space_uuid"`
	IsPrivate        types.Bool                      `tfsdk:"is_private"`
	SpaceName        types.String                    `tfsdk:"name"`
	DeleteProtection types.Bool                      `tfsdk:"deletion_protection"`
	CreatedAt        types.String                    `tfsdk:"created_at"`
	LastUpdated      types.String                    `tfsdk:"last_updated"`
	AccessList       []spaceResourceAccessBlockModel `tfsdk:"access"`
}

type spaceResourceAccessBlockModel struct {
	UserUUID types.String `tfsdk:"user_uuid"`
	// TODO support last_updated
	// LastUpdated types.String `tfsdk:"last_updated"`
}

func (r *spaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (r *spaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Lightash space",
		Description:         "Lightash space",
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
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_private": schema.BoolAttribute{
				MarkdownDescription: "Lightdash project is private or not",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Lightdash project name",
				Required:            true,
			},
			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: "Whether or not to allow Terraform to destroy the instance. Unless this field is set to false in Terraform state, a terraform destroy or terraform apply that would delete the instance will fail.",
				Required:            true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp of the creation of the space",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the space.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"access": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"user_uuid": schema.StringAttribute{
							MarkdownDescription: "User UUID",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (r *spaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *spaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan spaceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new space
	project_uuid := plan.ProjectUUID.ValueString()
	space_name := plan.SpaceName.ValueString()
	is_private := plan.IsPrivate.ValueBool()
	created_space, err := r.client.CreateSpaceV1(project_uuid, space_name, is_private)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating space",
			"Could not create space, unexpected error: "+err.Error(),
		)
		return
	}

	// plan.OrganizationUUID = types.StringValue(created_space.OrganizationUUID)
	plan.ProjectUUID = types.StringValue(created_space.ProjectUUID)
	plan.SpaceUUID = types.StringValue(created_space.SpaceUUID)
	plan.IsPrivate = types.BoolValue(created_space.IsPrivate)
	plan.DeleteProtection = types.BoolValue(plan.DeleteProtection.ValueBool())
	plan.CreatedAt = types.StringValue(time.Now().Format(time.RFC850))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Add space access
	accessList := []spaceResourceAccessBlockModel{}
	var errors []error
	for _, access := range plan.AccessList {
		err := services.GrantSpaceAccess(
			r.client, project_uuid, plan.SpaceUUID.ValueString(), access.UserUUID.ValueString())
		if err != nil {
			errors = append(errors, err)
		} else {
			accessList = append(accessList, spaceResourceAccessBlockModel{
				UserUUID: access.UserUUID,
			})
		}
	}
	if len(errors) > 0 {
		for _, err := range errors {
			resp.Diagnostics.AddError(
				"Error adding space access",
				"Could not add space access, unexpected error: "+err.Error(),
			)
		}
	}

	// Set resource ID
	state_id := getSpaceResourceId(created_space.ProjectUUID, created_space.SpaceUUID)
	plan.ID = types.StringValue(state_id)
	plan.AccessList = accessList

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *spaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Declare variables to import from state
	var project_uuid string
	var space_uuid string
	var access []spaceResourceAccessBlockModel
	var created_at string
	var last_updated string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project_uuid"), &project_uuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("space_uuid"), &space_uuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("access"), &access)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("created_at"), &created_at)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("last_updated"), &last_updated)...)

	// Get current state
	var state spaceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get space
	project_uuid = state.ProjectUUID.ValueString()
	space_uuid = state.SpaceUUID.ValueString()
	space, err := r.client.GetSpaceV1(project_uuid, space_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading space",
			"Could not read space ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Get space members
	members := make([]spaceResourceAccessBlockModel, len(state.AccessList))
	for i, access := range state.AccessList {
		members[i] = spaceResourceAccessBlockModel{
			UserUUID: access.UserUUID,
		}
	}

	// Set the state values
	state.ProjectUUID = types.StringValue(space.ProjectUUID)
	state.SpaceUUID = types.StringValue(space.SpaceUUID)
	state.AccessList = members

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *spaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get current state
	var currentState spaceResourceModel
	currentStateDiags := req.State.Get(ctx, &currentState)
	resp.Diagnostics.Append(currentStateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Retrieve values from plan
	var plan spaceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing space
	project_uuid := plan.ProjectUUID.ValueString()
	space_uuid := plan.SpaceUUID.ValueString()
	space_name := plan.SpaceName.ValueString()
	is_private := plan.IsPrivate.ValueBool()
	tflog.Info(ctx, fmt.Sprintf("Updating space %s", space_uuid))
	space, err := r.client.UpdateSpaceV1(project_uuid, space_uuid, space_name, is_private)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating space",
			"Could not update space, unexpected error: "+err.Error(),
		)
		return
	}

	// Revoke access from removed users
	for _, access := range currentState.AccessList {
		found := false
		for _, accessNew := range plan.AccessList {
			if access.UserUUID.ValueString() == accessNew.UserUUID.ValueString() {
				found = true
				break
			}
		}
		if !found {
			err := r.client.RevokeSpaceAccessV1(
				project_uuid, currentState.SpaceUUID.ValueString(), access.UserUUID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Revoking space access",
					"Could not revoke space access, unexpected error: "+err.Error(),
				)
				return
			}
		}
	}

	// Grant access to new users
	for _, access := range plan.AccessList {
		found := false
		for _, accessCurrent := range currentState.AccessList {
			if access.UserUUID.ValueString() == accessCurrent.UserUUID.ValueString() {
				found = true
				break
			}
		}
		if !found {
			err := services.GrantSpaceAccess(
				r.client, project_uuid, plan.SpaceUUID.ValueString(), access.UserUUID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Granting space access",
					"Could not grant space access, unexpected error: "+err.Error(),
				)
				return
			}
		}
	}

	// Update the state
	plan.SpaceName = types.StringValue(space.SpaceName)
	plan.IsPrivate = types.BoolValue(space.IsPrivate)
	// TODO Update the last_updated field
	//
	// We can't update the last_updated field because of the following error:
	// When applying changes to lightdash_space.test_private, provider "provider[\"github.com/ubie-oss/lightdash\"]" produced an unexpected new
	// value: .last_updated: was cty.StringVal("Wednesday, 15-Nov-23 13:08:36 JST"), but now cty.StringVal("Wednesday, 15-Nov-23 13:09:43 JST").
	// plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *spaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state spaceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if deletion protection is enabled
	if state.DeleteProtection.ValueBool() {
		resp.Diagnostics.AddError(
			"Error Deleting space",
			"Could not delete space, deletion protection is enabled",
		)
		return
	}

	// Delete existing space
	project_uuid := state.ProjectUUID.ValueString()
	space_uuid := state.SpaceUUID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting space %s", space_uuid))
	err := r.client.DeleteSpaceV1(project_uuid, space_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting space",
			"Could not delete space, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *spaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Extract the resource ID
	extracted_strings, err := extractSpaceResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	project_uuid := extracted_strings[0]
	space_uuid := extracted_strings[1]

	// Set the resource attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), project_uuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_uuid"), space_uuid)...)
}

func getSpaceResourceId(project_uuid string, space_uuid string) string {
	// Return the resource ID
	return fmt.Sprintf("projects/%s/spaces/%s", project_uuid, space_uuid)
}

func extractSpaceResourceId(input string) ([]string, error) {
	// Extract the captured groups
	pattern := `^projects/([^/]+)/spaces/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	// Return the captured strings
	project_uuid := groups[0]
	space_uuid := groups[1]
	return []string{project_uuid, space_uuid}, nil
}
