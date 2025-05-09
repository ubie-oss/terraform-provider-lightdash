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
				MarkdownDescription: "This block represents the access settings for the Lightdash space. " +
					"It allows you to define the users who have access to the space by specifying their user UUIDs. " +
					"Each access block should contain a 'user_uuid' and 'space_role' attributes. " +
					"Note: Organization administrators in Lightdash inherently have access to all spaces. " +
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
						"has_direct_access": schema.BoolAttribute{
							MarkdownDescription: "Indicates if the user has direct access to the space.",
							Computed:            true,
							Optional:            true,
						},
						"inherited_from": schema.StringAttribute{
							MarkdownDescription: "Indicates where the user's access is inherited from (e.g., organization, project).",
							Computed:            true,
							Optional:            true,
						},
					},
				},
			},
			"group_access": schema.SetNestedBlock{
				MarkdownDescription: "This block represents the group access settings for the Lightdash space. " +
					"It allows you to define the groups who have access to the space by specifying their group UUIDs. " +
					"Each group access block should contain 'group_uuid' and 'space_role' attributes. " +
					"IMPORTANT: This block is ignored for nested spaces, as they inherit access from their parent space.",
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

	// Determine if the created space is nested
	isNestedSpace = spaceDetails.ParentSpaceUUID != nil

	// Populate member access list from controller result (API response)
	memberAccessList := []spaceMemberAccessBlockModel{}
	for _, member := range spaceDetails.MemberAccess {
		memberAccessList = append(memberAccessList, spaceMemberAccessBlockModel{
			UserUUID:  types.StringValue(member.UserUUID),
			SpaceRole: types.StringValue(string(member.SpaceRole)),
		})
	}
	state.MemberAccessList = memberAccessList

	// Populate group access list from controller result (API response)
	groupAccessList := []spaceGroupAccessBlockModel{}
	for _, group := range spaceDetails.GroupAccess {
		groupAccessList = append(groupAccessList, spaceGroupAccessBlockModel{
			GroupUUID: types.StringValue(group.GroupUUID),
			SpaceRole: types.StringValue(string(group.SpaceRole)),
		})
	}
	state.GroupAccessList = groupAccessList

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
	var state spaceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Show the state
	tflog.Debug(ctx, "************* State (Read start) *************", map[string]any{"state": state})
	// Show the space name
	tflog.Debug(ctx, "************* Space Name (Read start): ", map[string]any{"spaceName": state.SpaceName.ValueString()})
	// Show length of member access list
	tflog.Debug(ctx, "************* Member Access List (Read start): ", map[string]any{"length": len(state.MemberAccessList)})
	// Show length of group access list
	tflog.Debug(ctx, "************* Group Access List (Read start): ", map[string]any{"length": len(state.GroupAccessList)})

	// Get space details from controller (API)
	projectUUID := state.ProjectUUID.ValueString()
	spaceUUID := state.SpaceUUID.ValueString()

	spaceDetails, err := r.spaceController.GetSpace(projectUUID, spaceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading space",
			"Could not read space ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// Convert member access list to the terraform model
	memberAccessList := []spaceMemberAccessBlockModel{}
	for _, member := range spaceDetails.MemberAccess {
		memberAccessList = append(memberAccessList, spaceMemberAccessBlockModel{
			UserUUID:  types.StringValue(member.UserUUID),
			SpaceRole: types.StringValue(string(member.SpaceRole)),
		})
	}
	// Convert group access list to the terraform model
	groupAccessList := []spaceGroupAccessBlockModel{}
	for _, group := range spaceDetails.GroupAccess {
		groupAccessList = append(groupAccessList, spaceGroupAccessBlockModel{
			GroupUUID: types.StringValue(group.GroupUUID),
			SpaceRole: types.StringValue(string(group.SpaceRole)),
		})
	}

	// Update state from controller response
	state.SpaceName = types.StringValue(spaceDetails.SpaceName)
	state.IsPrivate = types.BoolValue(spaceDetails.IsPrivate)
	// Handle parent space UUID from controller result
	if spaceDetails.ParentSpaceUUID != nil {
		state.ParentSpaceUUID = types.StringValue(*spaceDetails.ParentSpaceUUID)
	} else {
		state.ParentSpaceUUID = types.StringNull()
	}
	state.MemberAccessList = memberAccessList
	state.GroupAccessList = groupAccessList

	// Show the state
	tflog.Debug(ctx, "************* State (Read end) *************", map[string]any{"state": state})
	// Show the space name
	tflog.Debug(ctx, "************* Space Name (Read end): ", map[string]any{"spaceName": state.SpaceName.ValueString()})
	// Show length of member access list
	tflog.Debug(ctx, "************* Member Access List (Read end): ", map[string]any{"length": len(state.MemberAccessList)})
	// Show length of group access list
	tflog.Debug(ctx, "************* Group Access List (Read end): ", map[string]any{"length": len(state.GroupAccessList)})

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
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
	var isPrivate *bool
	if !plan.IsPrivate.IsNull() && !plan.IsPrivate.IsUnknown() {
		isPrivateVal := plan.IsPrivate.ValueBool()
		isPrivate = &isPrivateVal
	}
	var parentSpaceUUID *string
	if !plan.ParentSpaceUUID.IsNull() && !plan.ParentSpaceUUID.IsUnknown() {
		parentSpaceUUID = plan.ParentSpaceUUID.ValueStringPointer()
	}

	// Determine if the space is currently or becoming nested from the ParentSpaceUUIDs
	isCurrentlyNestedSpace := !state.ParentSpaceUUID.IsNull() && !state.ParentSpaceUUID.IsUnknown()
	isBecomingNestedSpace := parentSpaceUUID != nil

	// Convert member access list from plan to controller format
	newMemberAccess := []controllers.SpaceAccessMemberRequest{}
	// Only process access for root spaces or spaces becoming root spaces
	if !isBecomingNestedSpace {
		for _, access := range plan.MemberAccessList {
			newMemberAccess = append(newMemberAccess, controllers.SpaceAccessMemberRequest{
				BaseSpaceAccessMember: controllers.BaseSpaceAccessMember{
					UserUUID:  access.UserUUID.ValueString(),
					SpaceRole: models.SpaceMemberRole(access.SpaceRole.ValueString()),
				},
				IsOrganizationAdmin: false,
			})
		}
	}

	// Convert group access list to controller format
	newGroupAccess := []controllers.SpaceGroupAccess{}
	// Only process group access for root spaces or spaces becoming root spaces
	if !isBecomingNestedSpace {
		for _, access := range plan.GroupAccessList {
			newGroupAccess = append(newGroupAccess, controllers.SpaceGroupAccess{
				GroupUUID: access.GroupUUID.ValueString(),
				SpaceRole: models.SpaceMemberRole(access.SpaceRole.ValueString()),
			})
		}
	}

	// Log transition information for debugging
	tflog.Debug(ctx, "Space transition details", map[string]any{
		"currentIsNested":  isCurrentlyNestedSpace,
		"isBecomingNested": isBecomingNestedSpace,
		"parentInPlan":     plan.ParentSpaceUUID.ValueString(),
		"parentInState":    state.ParentSpaceUUID.ValueString(),
		"accessListCount":  len(plan.MemberAccessList),
		"groupAccessCount": len(plan.GroupAccessList),
	})

	// Update space using controller
	spaceDetails, errors := r.spaceController.UpdateSpace(
		projectUUID,
		spaceUUID,
		spaceName,
		isPrivate,
		parentSpaceUUID,
		newMemberAccess,
		newGroupAccess,
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

	// Log new space status
	tflog.Debug(ctx, "Updated space details", map[string]any{
		"parentSpaceUUID":   spaceDetails.ParentSpaceUUID,
		"memberAccessCount": len(spaceDetails.MemberAccess),
		"groupAccessCount":  len(spaceDetails.GroupAccess),
	})

	// Update state with values from controller for space attributes
	var updatedState spaceResourceModel
	updatedState.ID = types.StringValue(fmt.Sprintf("projects/%s/spaces/%s", spaceDetails.ProjectUUID, spaceDetails.SpaceUUID))
	updatedState.ProjectUUID = types.StringValue(spaceDetails.ProjectUUID)
	updatedState.SpaceUUID = types.StringValue(spaceDetails.SpaceUUID)
	updatedState.SpaceName = types.StringValue(spaceDetails.SpaceName)
	updatedState.IsPrivate = types.BoolValue(spaceDetails.IsPrivate)

	// Handle parent space UUID
	if spaceDetails.ParentSpaceUUID != nil {
		updatedState.ParentSpaceUUID = types.StringValue(*spaceDetails.ParentSpaceUUID)
	} else {
		updatedState.ParentSpaceUUID = types.StringNull()
	}

	// Preserve deletion protection from state
	updatedState.DeleteProtection = state.DeleteProtection

	// Update timestamps - We only update lastUpdated, keep createdAt from state
	updatedState.CreatedAt = state.CreatedAt
	updatedState.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Determine if the updated space is nested
	isNestedSpace := spaceDetails.ParentSpaceUUID != nil

	// Populate member access list from controller result
	memberAccessList := []spaceMemberAccessBlockModel{}
	tflog.Debug(ctx, "Processing member access list from controller result in Update")

	// For nested spaces, we don't populate access lists as they can't be managed
	if !isNestedSpace {
		for _, member := range spaceDetails.MemberAccess {
			tflog.Debug(ctx, "Processing member", map[string]any{
				"userUuid":        member.UserUUID,
				"hasDirectAccess": member.HasDirectAccess,
				"spaceRole":       member.SpaceRole,
				"inheritedFrom":   member.InheritedFrom,
			})

			// Convert pointer fields to types.Bool/String, handling nil
			var hasDirectAccess types.Bool
			if member.HasDirectAccess != nil {
				hasDirectAccess = types.BoolValue(*member.HasDirectAccess)
			} else {
				hasDirectAccess = types.BoolNull()
			}

			var inheritedFrom types.String
			if member.InheritedFrom != nil {
				inheritedFrom = types.StringValue(*member.InheritedFrom)
			} else {
				inheritedFrom = types.StringNull()
			}

			// Only include members based on the controller's filtered list (which now includes only managed types)
			memberAccessList = append(memberAccessList, spaceMemberAccessBlockModel{
				UserUUID:        types.StringValue(member.UserUUID),
				SpaceRole:       types.StringValue(string(member.SpaceRole)),
				HasDirectAccess: hasDirectAccess,
				InheritedFrom:   inheritedFrom,
			})
		}
		tflog.Debug(ctx, "Filtered member access list after setting state in Update", map[string]any{"count": len(memberAccessList)})

		// Populate group access list from controller result
		groupAccessList := []spaceGroupAccessBlockModel{}
		tflog.Debug(ctx, "Processing group access list from controller result in Update")
		for _, group := range spaceDetails.GroupAccess {
			tflog.Debug(ctx, "Processing group", map[string]any{"groupUuid": group.GroupUUID, "spaceRole": group.SpaceRole})
			groupAccessList = append(groupAccessList, spaceGroupAccessBlockModel{
				GroupUUID: types.StringValue(group.GroupUUID),
				SpaceRole: types.StringValue(string(group.SpaceRole)),
			})
		}
		updatedState.GroupAccessList = groupAccessList
		tflog.Debug(ctx, "Group access list after setting state in Update", map[string]any{"count": len(updatedState.GroupAccessList)})
	} else {
		tflog.Debug(ctx, "Skipping access lists for nested space in Update")
		// Empty the access lists for nested spaces since they can't be managed
		updatedState.GroupAccessList = []spaceGroupAccessBlockModel{}
	}

	updatedState.MemberAccessList = memberAccessList

	// Set state
	diags := resp.State.Set(ctx, updatedState)
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

	// Determine if the imported space is nested
	isNestedSpace := spaceDetails.ParentSpaceUUID != nil

	// Only set access information for root (non-nested) spaces
	if !isNestedSpace {
		// Set member access list
		memberAccessList := []spaceMemberAccessBlockModel{}
		for _, member := range spaceDetails.MemberAccess {
			// Convert pointer fields to types.Bool/String, handling nil
			var hasDirectAccess types.Bool
			if member.HasDirectAccess != nil {
				hasDirectAccess = types.BoolValue(*member.HasDirectAccess)
			} else {
				hasDirectAccess = types.BoolNull()
			}

			var inheritedFrom types.String
			if member.InheritedFrom != nil {
				inheritedFrom = types.StringValue(*member.InheritedFrom)
			} else {
				inheritedFrom = types.StringNull()
			}

			// Only include members based on the controller's filtered list (which now includes only managed types)
			memberAccessList = append(memberAccessList, spaceMemberAccessBlockModel{
				UserUUID:        types.StringValue(member.UserUUID),
				SpaceRole:       types.StringValue(string(member.SpaceRole)),
				HasDirectAccess: hasDirectAccess,
				InheritedFrom:   inheritedFrom,
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
	} else {
		// For nested spaces, set empty access lists
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access"), []spaceMemberAccessBlockModel{})...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_access"), []spaceGroupAccessBlockModel{})...)
	}

	// Set timestamps
	currentTime := types.StringValue(time.Now().Format(time.RFC850))
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), currentTime)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("last_updated"), currentTime)...)
}
