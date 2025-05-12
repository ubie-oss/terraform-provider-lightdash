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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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

// Helper function to convert models.SpaceAccessMember to spaceMemberAccessBlockModel for access_all
func convertToAllMemberAccessBlockModels(apiMembers []models.SpaceAccessMember) []spaceMemberAccessBlockModel {
	var blockModels []spaceMemberAccessBlockModel
	if apiMembers == nil {
		return blockModels
	}
	for _, member := range apiMembers {
		var hasDirectAccessVal = types.BoolValue(member.HasDirectAccess)

		var inheritedFromVal types.String
		if member.InheritedFrom != "" {
			inheritedFromVal = types.StringValue(member.InheritedFrom)
		} else {
			inheritedFromVal = types.StringNull()
		}

		blockModels = append(blockModels, spaceMemberAccessBlockModel{
			UserUUID:        types.StringValue(member.UserUUID),
			SpaceRole:       types.StringValue(string(member.SpaceRole)),
			HasDirectAccess: hasDirectAccessVal,
			InheritedFrom:   inheritedFromVal,
		})
	}
	return blockModels
}

// Helper function to convert models.SpaceAccessGroup to spaceGroupAccessBlockModel
func convertToGroupAccessBlockModels(apiGroups []models.SpaceAccessGroup) []spaceGroupAccessBlockModel {
	var blockModels []spaceGroupAccessBlockModel
	if apiGroups == nil {
		return blockModels
	}
	for _, group := range apiGroups {
		blockModels = append(blockModels, spaceGroupAccessBlockModel{
			GroupUUID: types.StringValue(group.GroupUUID),
			SpaceRole: types.StringValue(string(group.SpaceRole)),
		})
	}
	return blockModels
}

// Helper function to convert Controller SpaceGroupAccess to spaceGroupAccessBlockModel
func convertControllerGroupAccessToBlockModels(controllerGroups []controllers.SpaceGroupAccess) []spaceGroupAccessBlockModel {
	var blockModels []spaceGroupAccessBlockModel
	if controllerGroups == nil {
		return blockModels
	}
	for _, group := range controllerGroups {
		blockModels = append(blockModels, spaceGroupAccessBlockModel{
			GroupUUID: types.StringValue(group.GroupUUID),
			SpaceRole: types.StringValue(string(group.SpaceRole)),
		})
	}
	return blockModels
}

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
	ProjectUUID         types.String                  `tfsdk:"project_uuid"`
	ParentSpaceUUID     types.String                  `tfsdk:"parent_space_uuid"`
	SpaceUUID           types.String                  `tfsdk:"space_uuid"`
	IsPrivate           types.Bool                    `tfsdk:"is_private"`
	SpaceName           types.String                  `tfsdk:"name"`
	DeleteProtection    types.Bool                    `tfsdk:"deletion_protection"`
	CreatedAt           types.String                  `tfsdk:"created_at"`
	LastUpdated         types.String                  `tfsdk:"last_updated"`
	MemberAccessList    []spaceMemberAccessBlockModel `tfsdk:"access"`
	MemberAccessListAll []spaceMemberAccessBlockModel `tfsdk:"access_all"`
	GroupAccessList     []spaceGroupAccessBlockModel  `tfsdk:"group_access"`
	GroupAccessListAll  []spaceGroupAccessBlockModel  `tfsdk:"group_access_all"`
}

// spaceMemberAccessBlockModel maps the member access data from the API to the Terraform schema.
type spaceMemberAccessBlockModel struct {
	UserUUID  types.String `tfsdk:"user_uuid"`
	SpaceRole types.String `tfsdk:"space_role"`
	// These fields are output-only from the API and reflect the *actual* access
	// including inheritance and organization admin status, but are not managed
	// directly by this resource configuration.
	HasDirectAccess types.Bool   `tfsdk:"has_direct_access"`
	InheritedFrom   types.String `tfsdk:"inherited_from"`
	// InheritedRole is not currently exposed in the API response for space access members.
	// ProjectRole is not currently exposed in the API response for space access members.
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
			"\n\n**IMPORTANT: Nested spaces (with parent_space_uuid) have significant limitations:** " +
			"\n- Access controls for nested spaces are inherited from their parent space and cannot be managed individually " +
			"\n- Visibility (public/private) for nested spaces is inherited from their parent space and cannot be changed " +
			"\n- When a space is moved from root level to nested (or vice versa), its access controls will change accordingly " +
			"\n- For nested spaces, the `access` and `group_access` blocks will be empty in Terraform state as they cannot be managed",
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
				MarkdownDescription: "Lightdash parent space UUID. If set, creates a nested space that inherits access controls and visibility from its parent space.",
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
				MarkdownDescription: "Lightdash space is private or not. Note: This setting is ignored for nested spaces as they inherit visibility from their parent.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Lightdash space name",
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
				MarkdownDescription: "Use this block to define the users who have access to the Lightdash space. " +
					"Specify each user's UUID and their role within the space. " +
					"Note: Organization administrators in Lightdash automatically have access to all spaces. " +
					"IMPORTANT: This block is ignored for nested spaces, as they inherit access from their parent space.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"user_uuid": schema.StringAttribute{
							MarkdownDescription: "User UUID",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"space_role": schema.StringAttribute{
							MarkdownDescription: "Lightdash space role: 'admin' (Full Access), 'editor' (Can Edit), or 'viewer' (Can View)",
							Required:            true,
						},
					},
				},
			},
			"access_all": schema.SetNestedBlock{
				MarkdownDescription: "This block displays the complete list of users with access to the space, including those with direct access, inherited access, and organization administrators." +
					"It mirrors the API response for space access members and is read-only.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"user_uuid": schema.StringAttribute{
							MarkdownDescription: "User UUID",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"space_role": schema.StringAttribute{
							MarkdownDescription: "Lightdash space role: 'admin' (Full Access), 'editor' (Can Edit), or 'viewer' (Can View)",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"has_direct_access": schema.BoolAttribute{
							MarkdownDescription: "Indicates if the user has direct access to the space.",
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"inherited_from": schema.StringAttribute{
							MarkdownDescription: "Indicates where the user's access is inherited from (e.g., organization, project).",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"group_access": schema.SetNestedBlock{
				MarkdownDescription: "Use this block to define the groups who have access to the Lightdash space by specifying their group UUIDs and their role within the space.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group_uuid": schema.StringAttribute{
							MarkdownDescription: "Group UUID",
							Required:            true,
						},
						"space_role": schema.StringAttribute{
							MarkdownDescription: "Lightdash space role: 'admin' (Full Access), 'editor' (Can Edit), or 'viewer' (Can View)",
							Required:            true,
						},
					},
				},
			},
			"group_access_all": schema.SetNestedBlock{
				MarkdownDescription: "This block displays the complete list of groups with access to the space, including those with direct access and inherited access." +
					"It mirrors the API response for space access groups and is read-only.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group_uuid": schema.StringAttribute{
							MarkdownDescription: "Group UUID",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"space_role": schema.StringAttribute{
							MarkdownDescription: "Lightdash space role: 'admin' (Full Access), 'editor' (Can Edit), or 'viewer' (Can View)",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
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

// TODO implement the config validation
// func (r *spaceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
// 	// Retrieve values from plan
// 	var config spaceResourceModel
// 	diags := req.Config.Get(ctx, &config)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}
// }

func (r *spaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan spaceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Show the plan
	tflog.Debug(ctx, "************* Plan (Create start) *************", map[string]any{"plan": plan})
	// Show the space name
	tflog.Debug(ctx, "************* Space Name (Create start): ", map[string]any{"spaceName": plan.SpaceName.ValueString()})
	// Show length of member access list
	tflog.Debug(ctx, "************* Member Access List (Create start): ", map[string]any{"length": len(plan.MemberAccessList)})
	// Show length of group access list
	tflog.Debug(ctx, "************* Group Access List (Create start): ", map[string]any{"length": len(plan.GroupAccessList)})

	// Prepare data for controller
	projectUUID := plan.ProjectUUID.ValueString()
	spaceName := plan.SpaceName.ValueString()
	isPrivate := plan.IsPrivate.ValueBool()
	var parentSpaceUUID *string
	if !plan.ParentSpaceUUID.IsNull() && !plan.ParentSpaceUUID.IsUnknown() {
		parentSpaceUUID = plan.ParentSpaceUUID.ValueStringPointer()
	}

	// Determine if this will be a nested space from the plan's parent_space_uuid
	isNestedSpace := parentSpaceUUID != nil

	// Convert member access list from plan to controller format
	memberAccess := []controllers.SpaceAccessMemberRequest{}
	if !isNestedSpace { // Only process access for root spaces
		for _, access := range plan.MemberAccessList {
			memberAccess = append(memberAccess, controllers.SpaceAccessMemberRequest{
				BaseSpaceAccessMember: controllers.BaseSpaceAccessMember{
					UserUUID:  access.UserUUID.ValueString(),
					SpaceRole: models.SpaceMemberRole(access.SpaceRole.ValueString()),
				},
				IsOrganizationAdmin: false,
			})
		}
	}

	// Convert group access list to controller format
	groupAccess := []controllers.SpaceGroupAccess{}
	if !isNestedSpace { // Only process group access for root spaces
		for _, access := range plan.GroupAccessList {
			groupAccess = append(groupAccess, controllers.SpaceGroupAccess{
				GroupUUID: access.GroupUUID.ValueString(),
				SpaceRole: models.SpaceMemberRole(access.SpaceRole.ValueString()),
			})
		}
	}

	// log the parameters of CreateSpace
	tflog.Debug(ctx, "************* CreateSpace parameters *************",
		map[string]any{
			"projectUUID":     projectUUID,
			"spaceName":       spaceName,
			"isPrivate":       isPrivate,
			"parentSpaceUUID": parentSpaceUUID,
			"memberAccess":    memberAccess,
			"groupAccess":     groupAccess,
		})

	// Create space using controller
	createdSpaceDetails, controllerErrors := r.spaceController.CreateSpace(
		projectUUID,
		spaceName,
		isPrivate,
		parentSpaceUUID,
		memberAccess,
		groupAccess,
	)

	// Handle errors from controller
	if len(controllerErrors) > 0 {
		for _, err := range controllerErrors {
			resp.Diagnostics.AddError("Error during space creation", err.Error())
		}
		return
	}

	if createdSpaceDetails == nil {
		resp.Diagnostics.AddError("Error during space creation", "Controller returned nil space details")
		return
	}

	// Get the space to fetch all the space access
	fetchedSpaceDetails, err := r.spaceController.GetSpace(projectUUID, createdSpaceDetails.SpaceUUID)
	if err != nil {
		resp.Diagnostics.AddError("Error during space creation", "Controller returned nil space details")
		return
	}

	// Populate the state with values returned by the controller (which reflects the API response)
	var state spaceResourceModel
	state.ID = types.StringValue(fmt.Sprintf("projects/%s/spaces/%s", createdSpaceDetails.ProjectUUID, createdSpaceDetails.SpaceUUID))
	state.ProjectUUID = types.StringValue(createdSpaceDetails.ProjectUUID)
	state.SpaceUUID = types.StringValue(createdSpaceDetails.SpaceUUID)
	state.SpaceName = types.StringValue(createdSpaceDetails.SpaceName)
	state.IsPrivate = types.BoolValue(createdSpaceDetails.IsPrivate)
	// Handle parent space UUID from controller result
	if createdSpaceDetails.ParentSpaceUUID != nil {
		state.ParentSpaceUUID = types.StringValue(*createdSpaceDetails.ParentSpaceUUID)
	} else {
		state.ParentSpaceUUID = types.StringNull()
	}
	// Preserve deletion protection from plan - this is a Terraform setting
	state.DeleteProtection = plan.DeleteProtection
	// Set timestamps
	// Use the CreatedAt from the controller's SpaceDetails, and set LastUpdated to now
	currentTime := types.StringValue(time.Now().Format(time.RFC850))
	state.CreatedAt = currentTime
	state.LastUpdated = currentTime
	// Populate MemberAccessList (access attribute) directly from plan
	state.MemberAccessList = plan.MemberAccessList
	// Populate GroupAccessList from controller result (what was actually applied)
	state.GroupAccessList = plan.GroupAccessList
	// Populate MemberAccessListAll and GroupAccessListAll from raw data returned by controller
	state.MemberAccessListAll = convertToAllMemberAccessBlockModels(fetchedSpaceDetails.SpaceAccessMembers)
	state.GroupAccessListAll = convertToGroupAccessBlockModels(fetchedSpaceDetails.SpaceAccessGroups)

	// Show the state
	tflog.Debug(ctx, "************* State (Create end) *************", map[string]any{"state": state})
	// Show the space name
	tflog.Debug(ctx, "************* Space Name (Create end): ", map[string]any{"spaceName": state.SpaceName.ValueString()})
	// Show length of member access list
	tflog.Debug(ctx, "************* Member Access List (Create end): ", map[string]any{"length": len(state.MemberAccessList)})
	// Show length of group access list
	tflog.Debug(ctx, "************* Group Access List (Create end): ", map[string]any{"length": len(state.GroupAccessList)})

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *spaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state. This is needed to preserve Terraform-managed attributes.
	var currentState spaceResourceModel
	diags := req.State.Get(ctx, &currentState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "************* Current State (Read start) *************", map[string]any{"currentState": currentState})

	// Get space details from API using controller
	projectUUID := currentState.ProjectUUID.ValueString()
	spaceUUID := currentState.SpaceUUID.ValueString()

	// Use controller's GetSpace which returns SpaceDetails with raw lists
	fetchedSpaceDetails, err := r.spaceController.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		// If the space is not found, remove it from state
		// TODO: Check for specific "not found" error type if available
		tflog.Warn(ctx, fmt.Sprintf("Space %s not found during Read, removing from state", spaceUUID))
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state from controller response
	var newState spaceResourceModel

	// Preserve fields that are computed once or managed by Terraform
	newState.ID = currentState.ID
	newState.ProjectUUID = currentState.ProjectUUID
	newState.SpaceUUID = currentState.SpaceUUID
	newState.CreatedAt = currentState.CreatedAt
	newState.DeleteProtection = currentState.DeleteProtection
	newState.LastUpdated = currentState.LastUpdated // Read does not update this TF-managed field

	// Update fields from controller response
	newState.SpaceName = types.StringValue(fetchedSpaceDetails.SpaceName)
	newState.IsPrivate = types.BoolValue(fetchedSpaceDetails.IsPrivate)
	if fetchedSpaceDetails.ParentSpaceUUID != nil {
		newState.ParentSpaceUUID = types.StringValue(*fetchedSpaceDetails.ParentSpaceUUID)
	} else {
		newState.ParentSpaceUUID = types.StringNull()
	}

	// Preserve MemberAccessList (access attribute) from current state (user's config)
	newState.MemberAccessList = currentState.MemberAccessList

	// Preserve GroupAccessList (group_access attribute) from current state (user's config)
	newState.GroupAccessList = currentState.GroupAccessList

	// Populate MemberAccessListAll and GroupAccessListAll from raw data in controller response
	newState.MemberAccessListAll = convertToAllMemberAccessBlockModels(fetchedSpaceDetails.SpaceAccessMembers)
	newState.GroupAccessListAll = convertToGroupAccessBlockModels(fetchedSpaceDetails.SpaceAccessGroups)

	tflog.Debug(ctx, "************* New State (Read end) *************", map[string]any{"newState": newState})

	// Set state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *spaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan and state
	var plan, oldState spaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare data for controller
	projectUUID := plan.ProjectUUID.ValueString() // Should come from plan or state, must not change for update
	spaceUUID := oldState.SpaceUUID.ValueString() // Must come from state
	spaceName := plan.SpaceName.ValueString()
	var isPrivate *bool
	if !plan.IsPrivate.IsNull() && !plan.IsPrivate.IsUnknown() {
		isPrivateVal := plan.IsPrivate.ValueBool()
		isPrivate = &isPrivateVal
	}
	var parentSpaceUUID *string
	if !plan.ParentSpaceUUID.IsNull() && !plan.ParentSpaceUUID.IsUnknown() {
		parentSpaceUUID = plan.ParentSpaceUUID.ValueStringPointer()
	} else if plan.ParentSpaceUUID.IsNull() { // Explicitly setting to null means make it root
		parentSpaceUUID = nil
	}

	// Determine if the planned space will be nested
	// This check is needed to decide whether to send access lists to the controller
	isPlannedSpaceNested := parentSpaceUUID != nil

	newMemberAccess := []controllers.SpaceAccessMemberRequest{}
	// Only process access for root spaces in the plan
	if !isPlannedSpaceNested {
		for _, access := range plan.MemberAccessList {
			newMemberAccess = append(newMemberAccess, controllers.SpaceAccessMemberRequest{
				BaseSpaceAccessMember: controllers.BaseSpaceAccessMember{
					UserUUID:  access.UserUUID.ValueString(),
					SpaceRole: models.SpaceMemberRole(access.SpaceRole.ValueString()),
				},
				IsOrganizationAdmin: false, // Assuming this is always false for direct management
			})
		}
	}

	newGroupAccess := []controllers.SpaceGroupAccess{}
	// Only process group access for root spaces in the plan
	if !isPlannedSpaceNested {
		for _, access := range plan.GroupAccessList {
			newGroupAccess = append(newGroupAccess, controllers.SpaceGroupAccess{
				GroupUUID: access.GroupUUID.ValueString(),
				SpaceRole: models.SpaceMemberRole(access.SpaceRole.ValueString()),
			})
		}
	}

	tflog.Debug(ctx, "Space update parameters", map[string]any{
		"projectUUID":     projectUUID,
		"spaceUUID":       spaceUUID,
		"spaceName":       spaceName,
		"isPrivate":       isPrivate,
		"parentSpaceUUID": parentSpaceUUID,
		"newMemberAccess": newMemberAccess,
		"newGroupAccess":  newGroupAccess,
	})

	// Update space using controller
	// The controller will handle the logic of moving the space and managing access based on space type
	updatedSpaceDetails, controllerErrors := r.spaceController.UpdateSpace(
		projectUUID,
		spaceUUID,
		spaceName,
		isPrivate,
		parentSpaceUUID,
		newMemberAccess,
		newGroupAccess,
	)

	if len(controllerErrors) > 0 {
		for _, err := range controllerErrors {
			resp.Diagnostics.AddError("Error during space update", err.Error())
		}
		return // Stop if controller reported errors
	}

	if updatedSpaceDetails == nil {
		resp.Diagnostics.AddError("Error updating space", "Controller returned nil result after update")
		return
	}

	// Populate the state with values returned by the controller (which reflect the final API state)
	var updatedState spaceResourceModel
	updatedState.ID = oldState.ID
	updatedState.ProjectUUID = oldState.ProjectUUID // ProjectUUID cannot change
	updatedState.SpaceUUID = oldState.SpaceUUID     // SpaceUUID cannot change

	updatedState.SpaceName = types.StringValue(updatedSpaceDetails.SpaceName)
	updatedState.IsPrivate = types.BoolValue(updatedSpaceDetails.IsPrivate)

	if updatedSpaceDetails.ParentSpaceUUID != nil {
		updatedState.ParentSpaceUUID = types.StringValue(*updatedSpaceDetails.ParentSpaceUUID)
	} else {
		updatedState.ParentSpaceUUID = types.StringNull()
	}

	updatedState.DeleteProtection = plan.DeleteProtection // From plan
	updatedState.CreatedAt = oldState.CreatedAt           // Preserve creation timestamp
	updatedState.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Populate MemberAccessList (access attribute) directly from plan
	updatedState.MemberAccessList = plan.MemberAccessList

	// Populate GroupAccessList (group_access attribute) directly from plan
	updatedState.GroupAccessList = plan.GroupAccessList

	// Populate MemberAccessListAll and GroupAccessListAll from raw data in controller response
	updatedState.MemberAccessListAll = convertToAllMemberAccessBlockModels(updatedSpaceDetails.SpaceAccessMembers)
	updatedState.GroupAccessListAll = convertToGroupAccessBlockModels(updatedSpaceDetails.SpaceAccessGroups)

	tflog.Debug(ctx, "Updated state after API calls", map[string]any{"updatedState": updatedState})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, updatedState)...)
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

// ImportSpace imports an existing space by its resource ID.
func (r *spaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Fetch space details from the controller
	spaceDetailsFromController, err := r.spaceController.ImportSpace(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing space",
			fmt.Sprintf("Could not retrieve space for ID %s: %s", req.ID, err.Error()),
		)
		return
	}

	// Set the resource attributes from the controller's SpaceDetails
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), spaceDetailsFromController.ProjectUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_uuid"), spaceDetailsFromController.SpaceUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), spaceDetailsFromController.SpaceName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("is_private"), spaceDetailsFromController.IsPrivate)...)

	// Set deletion protection to true by default for imported resources
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("deletion_protection"), true)...)

	// Handle parent space UUID
	// The ParentSpaceUUID in SpaceDetails is *string. Convert to types.String
	var parentSpaceUUID types.String
	if spaceDetailsFromController.ParentSpaceUUID != nil {
		parentSpaceUUID = types.StringValue(*spaceDetailsFromController.ParentSpaceUUID)
	} else {
		parentSpaceUUID = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("parent_space_uuid"), parentSpaceUUID)...)

	// Determine if the imported space is nested based on the ParentSpaceUUID
	// Check if parentSpaceUUID (types.String) is not null and not unknown
	isImportedSpaceNested := !parentSpaceUUID.IsNull() && !parentSpaceUUID.IsUnknown()

	// Populate 'access' with direct members (only relevant for root spaces)
	directMemberAccessListForImport := []spaceMemberAccessBlockModel{}
	// This condition was incorrect, using the raw *string type instead of types.String
	// if !spaceDetailsFromController.ParentSpaceUUID.IsNull() && !spaceDetailsFromController.ParentSpaceUUID.IsUnknown() && spaceDetailsFromController.SpaceAccessMembers != nil {
	// Corrected condition using the types.String variable and the correct raw list field
	if !isImportedSpaceNested && spaceDetailsFromController.SpaceAccessMembers != nil {
		for _, member := range spaceDetailsFromController.SpaceAccessMembers {
			if member.HasDirectAccess {
				directMemberAccessListForImport = append(directMemberAccessListForImport, spaceMemberAccessBlockModel{
					UserUUID:  types.StringValue(member.UserUUID),
					SpaceRole: types.StringValue(string(member.SpaceRole)),
					// HasDirectAccess and InheritedFrom are not part of 'access' input schema, so not set here.
				})
			}
		}
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access"), directMemberAccessListForImport)...)

	// Populate 'access_all' with all members
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access_all"), convertToAllMemberAccessBlockModels(spaceDetailsFromController.SpaceAccessMembers))...)

	// Populate 'group_access' (only relevant for root spaces or if API provides explicit flag)
	// Assuming API doesn't distinguish explicit group grants, populate with all groups from API result for root spaces.
	// For nested spaces, this block should be empty as access is inherited.
	groupAccessListForImport := []spaceGroupAccessBlockModel{}
	if !isImportedSpaceNested && spaceDetailsFromController.SpaceAccessGroups != nil {
		// For root spaces, populate group_access with all groups returned by the API
		groupAccessListForImport = convertToGroupAccessBlockModels(spaceDetailsFromController.SpaceAccessGroups)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_access"), groupAccessListForImport)...)

	// Populate 'group_access_all' with all groups from API result
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_access_all"), convertToGroupAccessBlockModels(spaceDetailsFromController.SpaceAccessGroups))...)

	// Set timestamps
	// Use the CreatedAt from the controller's SpaceDetails, and set LastUpdated to now
	// As mentioned, CreatedAt is not expected from Lightdash, so we'll use current time
	currentTime := types.StringValue(time.Now().Format(time.RFC850))
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), currentTime)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("last_updated"), currentTime)...)
}
