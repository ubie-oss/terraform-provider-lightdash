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
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/controllers"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
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
	client          *api.Client
	spaceController *controllers.SpaceController
}

// spaceResourceModel describes the resource data model.
type spaceResourceModel struct {
	ID types.String `tfsdk:"id"`
	// The response from the API does not contain the organization UUID right now.
	// OrganizationUUID types.String `tfsdk:"organization_uuid"`
	ProjectUUID      types.String                  `tfsdk:"project_uuid"`
	ParentSpaceUUID  types.String                  `tfsdk:"parent_space_uuid"`
	SpaceUUID        types.String                  `tfsdk:"space_uuid"`
	IsPrivate        types.Bool                    `tfsdk:"is_private"`
	SpaceName        types.String                  `tfsdk:"name"`
	DeleteProtection types.Bool                    `tfsdk:"deletion_protection"`
	CreatedAt        types.String                  `tfsdk:"created_at"`
	LastUpdated      types.String                  `tfsdk:"last_updated"`
	MemberAccessList []spaceMemberAccessBlockModel `tfsdk:"access"`
	GroupAccessList  []spaceGroupAccessBlockModel  `tfsdk:"group_access"`
}

type spaceMemberAccessBlockModel struct {
	UserUUID            types.String `tfsdk:"user_uuid"`
	SpaceRole           types.String `tfsdk:"space_role"`
	IsOrganizationAdmin types.Bool   `tfsdk:"is_organization_admin"`
}

type spaceGroupAccessBlockModel struct {
	GroupUUID types.String `tfsdk:"group_uuid"`
	SpaceRole types.String `tfsdk:"space_role"`
}

func (r *spaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (r *spaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Lightdash space is a powerful feature of the Lightdash platform that allows you to create and manage spaces for your analytics projects. " +
			"With Lightdash space, you can organize your data, dashboards, and reports into separate spaces, providing a logical separation and access control for different teams or projects. " +
			"Each space has a unique identifier (UUID) and can be associated with a project. " +
			"You can specify whether a space is private or not, allowing you to control who can access the space. " +
			"Additionally, you can enable deletion protection for a space, preventing accidental deletion during Terraform operations. " +
			"The created_at and last_updated attributes provide timestamps for the creation and last update of the space, respectively. " +
			"The access block allows you to define the users who have access to the space by specifying their user UUIDs. " +
			"Each access block should contain a single attribute 'user_uuid' which is a required string attribute representing the user UUID. " +
			"Lightdash space is a flexible and scalable solution for managing your analytics projects and ensuring data security and access control.",
		Description: "Lightdash space",
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
			"parent_space_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash parent space UUID",
				Optional:            true,
				Computed:            true,
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
				MarkdownDescription: "This block represents the access settings for the Lightdash space. " +
					"It allows you to define the users who have access to the space by specifying their user UUIDs. " +
					"Each access block should contain a single attribute 'user_uuid' which is a required string attribute representing the user UUID.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"user_uuid": schema.StringAttribute{
							MarkdownDescription: "User UUID",
							Required:            true,
						},
						"space_role": schema.StringAttribute{
							MarkdownDescription: "Lightdash space role",
							Required:            true,
						},
						"is_organization_admin": schema.BoolAttribute{
							MarkdownDescription: "Indicates if the user's access is managed by Terraform." +
								"Note: It is not possible to manage space access for organization admins through Terraform because they inherently have access to all spaces within the organization by default.",
							Computed: true,
						},
					},
				},
			},
			"group_access": schema.SetNestedBlock{
				MarkdownDescription: "This block represents the access settings for the Lightdash space. " +
					"It allows you to define the groups who have access to the space by specifying their group UUIDs. " +
					"Each group access block should contain a single attribute 'group_uuid' which is a required string attribute representing the group UUID.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group_uuid": schema.StringAttribute{
							MarkdownDescription: "Group UUID",
							Required:            true,
						},
						"space_role": schema.StringAttribute{
							MarkdownDescription: "Lightdash space role",
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
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	r.client = client
	r.spaceController = controllers.NewSpaceController(client)
}

func (r *spaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan spaceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare data for controller
	projectUUID := plan.ProjectUUID.ValueString()
	spaceName := plan.SpaceName.ValueString()
	isPrivate := plan.IsPrivate.ValueBool()
	var parentSpaceUUID *string
	if !plan.ParentSpaceUUID.IsNull() && !plan.ParentSpaceUUID.IsUnknown() {
		parentSpaceUUID = plan.ParentSpaceUUID.ValueStringPointer()
	}

	// Convert member access list from plan to controller format.
	// The controller will handle whether to apply access based on is_organization_admin status.
	memberAccess := []controllers.SpaceMemberAccess{}
	for _, access := range plan.MemberAccessList {
		memberAccess = append(memberAccess, controllers.SpaceMemberAccess{
			UserUUID:  access.UserUUID.ValueString(),
			SpaceRole: models.SpaceMemberRole(access.SpaceRole.ValueString()),
			// Pass the is_organization_admin status from the plan to the controller.
			// The controller should decide whether to use this info or rely on API checks.
			IsOrganizationAdmin: access.IsOrganizationAdmin.ValueBool(),
		})
	}

	// Convert group access list to controller format
	groupAccess := []controllers.SpaceGroupAccess{}
	for _, access := range plan.GroupAccessList {
		groupAccess = append(groupAccess, controllers.SpaceGroupAccess{
			GroupUUID: access.GroupUUID.ValueString(),
			SpaceRole: models.SpaceMemberRole(access.SpaceRole.ValueString()),
		})
	}

	// Create space using controller
	spaceDetails, errors := r.spaceController.CreateSpace(
		projectUUID,
		spaceName,
		isPrivate,
		parentSpaceUUID,
		memberAccess,
		groupAccess,
	)

	// Handle errors from controller
	if len(errors) > 0 {
		for _, err := range errors {
			resp.Diagnostics.AddWarning("Warning during space creation", err.Error())
		}
	}

	if spaceDetails == nil {
		resp.Diagnostics.AddError(
			"Error creating space",
			"Could not create space, controller returned nil result",
		)
		return
	}

	// Populate the state with values returned by the controller (which reflects the API response)
	var state spaceResourceModel
	state.ID = types.StringValue(fmt.Sprintf("projects/%s/spaces/%s", spaceDetails.ProjectUUID, spaceDetails.SpaceUUID))
	state.ProjectUUID = types.StringValue(spaceDetails.ProjectUUID)
	state.SpaceUUID = types.StringValue(spaceDetails.SpaceUUID)
	state.SpaceName = types.StringValue(spaceDetails.SpaceName)
	state.IsPrivate = types.BoolValue(spaceDetails.IsPrivate)

	// Handle parent space UUID from controller result
	if spaceDetails.ParentSpaceUUID != nil {
		state.ParentSpaceUUID = types.StringValue(*spaceDetails.ParentSpaceUUID)
	} else {
		state.ParentSpaceUUID = types.StringNull()
	}

	// Preserve deletion protection from plan - this is a Terraform setting
	state.DeleteProtection = plan.DeleteProtection

	// Set timestamps - these are typically set by the provider on creation and update
	state.CreatedAt = types.StringValue(time.Now().Format(time.RFC850))
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Populate member access list from controller result (which reflects the API)
	memberAccessList := []spaceMemberAccessBlockModel{}
	for _, member := range spaceDetails.MemberAccess {
		memberAccessList = append(memberAccessList, spaceMemberAccessBlockModel{
			UserUUID:            types.StringValue(member.UserUUID),
			SpaceRole:           types.StringValue(string(member.SpaceRole)),
			IsOrganizationAdmin: types.BoolValue(member.IsOrganizationAdmin),
		})
	}
	state.MemberAccessList = memberAccessList

	// Populate group access list from controller result
	groupAccessList := []spaceGroupAccessBlockModel{}
	for _, group := range spaceDetails.GroupAccess {
		groupAccessList = append(groupAccessList, spaceGroupAccessBlockModel{
			GroupUUID: types.StringValue(group.GroupUUID),
			SpaceRole: types.StringValue(string(group.SpaceRole)),
		})
	}
	state.GroupAccessList = groupAccessList

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *spaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state. This is needed to preserve Terraform-managed attributes.
	var state spaceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get space details from controller (API)
	projectUUID := state.ProjectUUID.ValueString()
	spaceUUID := state.SpaceUUID.ValueString()

	spaceDetails, err := r.spaceController.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading space",
			"Could not read space ID "+state.ID.ValueString()+": "+err.Error(),
		)
		resp.Diagnostics.AddWarning("Read failed, potentially due to space deletion",
			fmt.Sprintf("Lightdash space with ID %s not found. It may have been deleted outside of Terraform.", state.ID.ValueString()))
		// Mark the resource as removed from the state
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with values from controller for space attributes
	state.ProjectUUID = types.StringValue(spaceDetails.ProjectUUID)
	state.SpaceUUID = types.StringValue(spaceDetails.SpaceUUID)
	state.SpaceName = types.StringValue(spaceDetails.SpaceName)
	state.IsPrivate = types.BoolValue(spaceDetails.IsPrivate)

	// Handle parent space UUID
	if spaceDetails.ParentSpaceUUID != nil {
		state.ParentSpaceUUID = types.StringValue(*spaceDetails.ParentSpaceUUID)
	} else {
		state.ParentSpaceUUID = types.StringNull()
	}

	// Populate member access list directly from API response.
	// The API response should provide the current state including is_organization_admin status.
	memberAccessList := []spaceMemberAccessBlockModel{}
	for _, member := range spaceDetails.MemberAccess {
		memberAccessList = append(memberAccessList, spaceMemberAccessBlockModel{
			UserUUID:            types.StringValue(member.UserUUID),
			SpaceRole:           types.StringValue(string(member.SpaceRole)),
			IsOrganizationAdmin: types.BoolValue(member.IsOrganizationAdmin),
		})
	}
	state.MemberAccessList = memberAccessList

	// Populate group access list directly from API response
	groupAccessList := []spaceGroupAccessBlockModel{}
	for _, group := range spaceDetails.GroupAccess {
		groupAccessList = append(groupAccessList, spaceGroupAccessBlockModel{
			GroupUUID: types.StringValue(group.GroupUUID),
			SpaceRole: types.StringValue(string(group.SpaceRole)),
		})
	}
	state.GroupAccessList = groupAccessList

	// Preserve deletion protection, created_at, and last_updated from the *original* state.
	// These fields are managed by Terraform, not directly by the API in this context, and should only change
	// when the user modifies the Terraform configuration.
	// We already read the original state at the beginning of the function.

	// Restore original Terraform-managed attributes
	// Note: 'state' now contains the API-fetched data for most fields, but we overwrite
	// the Terraform-managed ones with values from the state *before* the API call.
	state.DeleteProtection = state.DeleteProtection // This line doesn't do anything, retaining original state value implicitly
	state.CreatedAt = state.CreatedAt
	state.LastUpdated = state.LastUpdated

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	// Check for inconsistencies after setting state
	if resp.Diagnostics.HasError() {
		return
	}

	// Optional: If needed for debugging, uncomment the following to see the state being set
	// tflog.Debug(ctx, "Read method finished, state set")
}

func (r *spaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan and state
	var plan, state spaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare data for controller
	projectUUID := plan.ProjectUUID.ValueString()
	spaceUUID := plan.SpaceUUID.ValueString()
	spaceName := plan.SpaceName.ValueString()
	isPrivate := plan.IsPrivate.ValueBool()
	var parentSpaceUUID *string
	if !plan.ParentSpaceUUID.IsNull() && !plan.ParentSpaceUUID.IsUnknown() {
		parentSpaceUUID = plan.ParentSpaceUUID.ValueStringPointer()
	}

	// Convert member access list from plan to controller format.
	// Pass the is_organization_admin status from the plan.
	memberAccess := []controllers.SpaceMemberAccess{}
	for _, access := range plan.MemberAccessList {
		memberAccess = append(memberAccess, controllers.SpaceMemberAccess{
			UserUUID:            access.UserUUID.ValueString(),
			SpaceRole:           models.SpaceMemberRole(access.SpaceRole.ValueString()),
			IsOrganizationAdmin: access.IsOrganizationAdmin.ValueBool(),
		})
	}

	// Convert group access list to controller format
	groupAccess := []controllers.SpaceGroupAccess{}
	for _, access := range plan.GroupAccessList {
		groupAccess = append(groupAccess, controllers.SpaceGroupAccess{
			GroupUUID: access.GroupUUID.ValueString(),
			SpaceRole: models.SpaceMemberRole(access.SpaceRole.ValueString()),
		})
	}

	// Update space using controller
	spaceDetails, errors := r.spaceController.UpdateSpace(
		projectUUID,
		spaceUUID,
		spaceName,
		&isPrivate,
		parentSpaceUUID,
		memberAccess,
		groupAccess,
	)

	// Handle errors from controller
	if len(errors) > 0 {
		for _, err := range errors {
			resp.Diagnostics.AddWarning("Warning during space update", err.Error())
		}
	}

	if spaceDetails == nil {
		resp.Diagnostics.AddError(
			"Error updating space",
			"Could not update space, controller returned nil result",
		)
		return
	}

	// Populate the state with values returned by the controller (which reflects the API response)
	state.ProjectUUID = types.StringValue(spaceDetails.ProjectUUID)
	state.SpaceUUID = types.StringValue(spaceDetails.SpaceUUID)
	state.SpaceName = types.StringValue(spaceDetails.SpaceName)
	state.IsPrivate = types.BoolValue(spaceDetails.IsPrivate)

	// Handle parent space UUID from controller result
	if spaceDetails.ParentSpaceUUID != nil {
		state.ParentSpaceUUID = types.StringValue(*spaceDetails.ParentSpaceUUID)
	} else {
		state.ParentSpaceUUID = types.StringNull()
	}

	// Preserve deletion protection from plan
	state.DeleteProtection = plan.DeleteProtection

	// Set last updated timestamp - this is typically set by the provider on creation and update
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Populate member access list from controller result (which reflects the API)
	memberAccessList := []spaceMemberAccessBlockModel{}
	for _, member := range spaceDetails.MemberAccess {
		memberAccessList = append(memberAccessList, spaceMemberAccessBlockModel{
			UserUUID:            types.StringValue(member.UserUUID),
			SpaceRole:           types.StringValue(string(member.SpaceRole)),
			IsOrganizationAdmin: types.BoolValue(member.IsOrganizationAdmin),
		})
	}
	state.MemberAccessList = memberAccessList

	// Populate group access list from controller result
	groupAccessList := []spaceGroupAccessBlockModel{}
	for _, group := range spaceDetails.GroupAccess {
		groupAccessList = append(groupAccessList, spaceGroupAccessBlockModel{
			GroupUUID: types.StringValue(group.GroupUUID),
			SpaceRole: types.StringValue(string(group.SpaceRole)),
		})
	}
	state.GroupAccessList = groupAccessList

	// Set state
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *spaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state spaceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete space using controller
	projectUUID := state.ProjectUUID.ValueString()
	spaceUUID := state.SpaceUUID.ValueString()
	deletionProtection := state.DeleteProtection.ValueBool()

	err := r.spaceController.DeleteSpace(projectUUID, spaceUUID, deletionProtection)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting space",
			"Could not delete space: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Deleted space %s", spaceUUID))
}

func (r *spaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import space using controller
	spaceDetails, err := r.spaceController.ImportSpace(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing space",
			"Could not import space: "+err.Error(),
		)
		return
	}

	// Set the resource attributes
	resourceID := fmt.Sprintf("projects/%s/spaces/%s", spaceDetails.ProjectUUID, spaceDetails.SpaceUUID)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), resourceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), spaceDetails.ProjectUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_uuid"), spaceDetails.SpaceUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), spaceDetails.SpaceName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("is_private"), spaceDetails.IsPrivate)...)

	// Set deletion protection to true by default for imported resources
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("deletion_protection"), true)...)

	// Handle parent space UUID
	if spaceDetails.ParentSpaceUUID != nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("parent_space_uuid"), *spaceDetails.ParentSpaceUUID)...)
	}

	// Set member access list
	memberAccessList := []spaceMemberAccessBlockModel{}
	for _, member := range spaceDetails.MemberAccess {
		memberAccessList = append(memberAccessList, spaceMemberAccessBlockModel{
			UserUUID:            types.StringValue(member.UserUUID),
			SpaceRole:           types.StringValue(string(member.SpaceRole)),
			IsOrganizationAdmin: types.BoolValue(member.IsOrganizationAdmin),
		})
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access"), memberAccessList)...)

	// Set group access list
	groupAccessList := []spaceGroupAccessBlockModel{}
	for _, group := range spaceDetails.GroupAccess {
		groupAccessList = append(groupAccessList, spaceGroupAccessBlockModel{
			GroupUUID: types.StringValue(group.GroupUUID),
			SpaceRole: types.StringValue(string(group.SpaceRole)),
		})
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_access"), groupAccessList)...)

	// Set timestamps
	currentTime := types.StringValue(time.Now().Format(time.RFC850))
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), currentTime)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("last_updated"), currentTime)...)
}
