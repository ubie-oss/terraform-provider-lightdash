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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

// groupResource defines the resource implementation.
type groupResource struct {
	client *api.Client
}

// groupResourceModel describes the resource data model.
type groupResourceModel struct {
	ID               types.String               `tfsdk:"id"`
	OrganizationUUID types.String               `tfsdk:"organization_uuid"`
	GroupUUID        types.String               `tfsdk:"group_uuid"`
	Name             types.String               `tfsdk:"name"`
	Members          []groupMemberModelForGroup `tfsdk:"members"`
}

type groupMemberModelForGroup struct {
	UserUUID types.String `tfsdk:"user_uuid"`
}

func (r *groupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Lightdash group represents a set of users within an organization, " +
			"providing a mechanism to manage permissions for projects and resources. " +
			"Each group is uniquely identified by a UUID and is associated with a specific organization. " +
			"Group membership is defined by the user UUIDs of the members.",
		Description: "Manages a Lightdash group",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier for the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization to which the group belongs.",
				Required:            true,
			},
			"group_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash group.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Lightdash group.",
				Required:            true,
			},
			// TODO check if values of userUUID are unique
			"members": schema.ListNestedAttribute{
				Description: "List of users.",
				Required:    false,
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_uuid": schema.StringAttribute{
							MarkdownDescription: "Lightdash user UUID",
							Required:            true,
						},
					},
				},
				Default: listdefault.StaticValue(
					types.ListValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"user_uuid": types.StringType,
							},
						},
						[]attr.Value{},
					),
				),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *groupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan groupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new group
	organization_uuid := plan.OrganizationUUID.ValueString()
	group_name := plan.Name.ValueString()
	members := make([]api.CreateGroupInOrganizationV1Member, 0, len(plan.Members))
	if plan.Members != nil {
		for _, member := range plan.Members {
			members = append(members, api.CreateGroupInOrganizationV1Member{
				UserUUID: member.UserUUID.ValueString(),
			})
		}
	}
	createdGroup, err := r.client.CreateGroupInOrganizationV1(organization_uuid, group_name, members)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating group",
			"Could not create group, unexpected error: "+err.Error(),
		)
		return
	}

	// Assign the plan values to the state
	stateId := getGroupResourceId(organization_uuid, createdGroup.GroupUUID)
	plan.ID = types.StringValue(stateId)
	plan.GroupUUID = types.StringValue(createdGroup.GroupUUID)
	plan.Name = types.StringValue(createdGroup.Name)
	// plan.Members = plan.Members

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Declare variables to import from state
	var organizationUuid string
	var groupUuid string
	var groupName string
	var members []groupMemberModelForGroup
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("organization_uuid"), &organizationUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("group_uuid"), &groupUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("name"), &groupName)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("members"), &members)...)

	// Get current state
	var state groupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get group
	group, err := r.client.GetGroupV1(groupUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading group",
			"Could not read group ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Get group members
	fetchedGroupMembers, getGroupMembersError := r.client.GetGroupMembersV1(group.GroupUUID)
	if getGroupMembersError != nil {
		resp.Diagnostics.AddError(
			"Error Getting group members",
			"Could not get members for group "+group.GroupUUID+": "+getGroupMembersError.Error(),
		)
		return
	}
	// Convert members to the correct type before assignment
	stateMembers := make([]groupMemberModelForGroup, len(fetchedGroupMembers))
	for i, member := range fetchedGroupMembers {
		stateMembers[i] = groupMemberModelForGroup{
			UserUUID: types.StringValue(member.UserUUID),
		}
	}

	// Set the state values
	state.OrganizationUUID = types.StringValue(group.OrganizationUUID)
	state.GroupUUID = types.StringValue(group.GroupUUID)
	state.Name = types.StringValue(group.Name)
	if len(stateMembers) > 0 {
		state.Members = stateMembers
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan, state groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the information from the state
	groupUuid := plan.GroupUUID.ValueString()
	groupName := plan.Name.ValueString()

	updatedMembers := []groupMemberModelForGroup{}
	removedMembers := []groupMemberModelForGroup{}

	// The service to get the organization members
	organizationMembersService := services.NewOrganizationMembersService(r.client)

	// Select removed members
	for _, memberInState := range state.Members {
		// Check if the user still exists in the organization
		_, err := organizationMembersService.GetOrganizationMemberByUserUuid(memberInState.UserUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddWarning(
				"User no longer exists in the organization",
				fmt.Sprintf("User %s no longer exists in the organization. Skipping adding access to the group.", memberInState.UserUUID.ValueString()),
			)
			continue
		}
		// Check if the user is in the plan
		isRemoved := true
		for _, memberInPlan := range plan.Members {
			if memberInState.UserUUID == memberInPlan.UserUUID {
				isRemoved = false
				break
			}
		}
		if isRemoved {
			removedMembers = append(removedMembers, memberInState)
		}
	}
	// Check if the user is in the plan
	for _, memberInPlan := range plan.Members {
		// Check if the user still exists in the organization
		_, err := organizationMembersService.GetOrganizationMemberByUserUuid(memberInPlan.UserUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddWarning(
				"User no longer exists in the organization",
				fmt.Sprintf("User %s no longer exists in the organization. Skipping adding access to the group.", memberInPlan.UserUUID.ValueString()),
			)
			continue
		}
		updatedMembers = append(updatedMembers, memberInPlan)
	}

	// Revoke access to removed members
	for _, member := range removedMembers {
		tflog.Info(ctx, fmt.Sprintf("Revoking access to group %s for user %s", groupUuid, member.UserUUID.ValueString()))
		err := r.client.RemoveUserFromGroupV1(groupUuid, member.UserUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Revoking access to group",
				fmt.Sprintf("Could not revoke access to group %s for user %s, unexpected error: %s", groupUuid, member.UserUUID, err.Error()),
			)
		}
	}

	// Update the group
	tflog.Info(ctx, fmt.Sprintf("Updating group %s", groupUuid))
	updateMembersUUIDs := make([]api.UpdateGroupV1Member, len(updatedMembers))
	for i, member := range updatedMembers {
		updateMembersUUIDs[i] = api.UpdateGroupV1Member{
			UserUUID: member.UserUUID.ValueString(),
		}
	}
	updatedGroup, err := r.client.UpdateGroupV1(groupUuid, groupName, updateMembersUUIDs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating group",
			fmt.Sprintf("Could not update group with UUID '%s' and name '%s', unexpected error: %s", plan.GroupUUID, plan.Name, err.Error()),
		)
		return
	}

	// Update the state
	plan.GroupUUID = types.StringValue(updatedGroup.GroupUUID)
	plan.Name = types.StringValue(updatedGroup.Name)
	plan.Members = updatedMembers

	// Set state
	diags := resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state groupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing group
	groupUuid := state.GroupUUID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting group %s", groupUuid))
	err := r.client.DeleteGroupV1(groupUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting group",
			"Could not delete group, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Extract the resource ID
	extracted_strings, err := extractGroupResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	organization_uuid := extracted_strings[0]
	groupUuid := extracted_strings[1]

	// Get the imported group
	importedGroup, err := r.client.GetGroupV1(groupUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting group",
			fmt.Sprintf("Could not get group with organization UUID %s and group UUID %s, unexpected error: %s", organization_uuid, groupUuid, err.Error()),
		)
		return
	}

	// Get the members of the group
	importedMembers, err := r.client.GetGroupMembersV1(importedGroup.GroupUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting group members",
			fmt.Sprintf("Could not get group members with group UUID %s, unexpected error: %s", importedGroup.GroupUUID, err.Error()),
		)
	}

	// Get the group memberBlock
	memberBlock := make([]groupMemberModelForGroup, len(importedMembers))
	for i, member := range importedMembers {
		// Update each element in the slice
		memberBlock[i] = groupMemberModelForGroup{
			UserUUID: types.StringValue(member.UserUUID),
		}
	}

	// Set the resource attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), groupUuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_uuid"), importedGroup.OrganizationUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_uuid"), importedGroup.GroupUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), importedGroup.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("members"), memberBlock)...)
}

func getGroupResourceId(organization_uuid string, group_uuid string) string {
	// Return the resource ID
	return fmt.Sprintf("organizations/%s/groups/%s", organization_uuid, group_uuid)
}

func extractGroupResourceId(input string) ([]string, error) {
	// Extract the captured groups
	pattern := `^organizations/([^/]+)/groups/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	// Return the captured strings
	organization_uuid := groups[0]
	group_uuid := groups[1]
	return []string{organization_uuid, group_uuid}, nil
}
