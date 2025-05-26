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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	ProjectUUID      types.String `tfsdk:"project_uuid"`
	ParentSpaceUUID  types.String `tfsdk:"parent_space_uuid"`
	SpaceUUID        types.String `tfsdk:"space_uuid"`
	IsPrivate        types.Bool   `tfsdk:"is_private"`
	SpaceName        types.String `tfsdk:"name"`
	DeleteProtection types.Bool   `tfsdk:"deletion_protection"`
	CreatedAt        types.String `tfsdk:"created_at"`
	LastUpdated      types.String `tfsdk:"last_updated"`
	MemberAccessList types.Set    `tfsdk:"access"`
	GroupAccessList  types.Set    `tfsdk:"group_access"`
}

// spaceMemberAccessBlockModel maps the member access data for the user input schema (access block)
type spaceMemberAccessBlockModel struct {
	UserUUID  types.String `tfsdk:"user_uuid"`
	SpaceRole types.String `tfsdk:"space_role"`
	// These fields were removed as they're only for the access_all block
}

// spaceMemberAccessAllBlockModel maps the member access data from the API to the output-only access_all block
// This type is primarily kept for documentation purposes. We're now using types.List for the attribute instead,
// but this model shows the schema of each item in the access_all list.
//
//nolint:unused
type spaceMemberAccessAllBlockModel struct {
	UserUUID        types.String `tfsdk:"user_uuid"`
	SpaceRole       types.String `tfsdk:"space_role"`
	HasDirectAccess types.Bool   `tfsdk:"has_direct_access"`
	InheritedFrom   types.String `tfsdk:"inherited_from"`
	ProjectRole     types.String `tfsdk:"project_role"`
}

type spaceGroupAccessBlockModel struct {
	GroupUUID types.String `tfsdk:"group_uuid"`
	SpaceRole types.String `tfsdk:"space_role"`
}

func (r *spaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (r *spaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "docs/resources/resource_space.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read markdown description",
			"Could not read the markdown description file for the space resource: "+err.Error(),
		)
		return // Stop processing if documentation can't be loaded
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
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
				// Computed:            true,
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
						},
						"space_role": schema.StringAttribute{
							MarkdownDescription: "Lightdash space role: 'admin' (Full Access), 'editor' (Can Edit), or 'viewer' (Can View)",
							Required:            true,
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
func (r *spaceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var errors []error

	// Retrieve values from plan
	var config spaceResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate configuration for space visibility
	errors = append(errors, r.validateSpaceVisibilityConfig(ctx, config)...)

	// Validate configuration for nested spaces
	for _, error := range r.validateNestedSpaceConfig(ctx, config) {
		errors = append(errors, error)
	}

	// TODO validate MemberAccessList and GroupAccessList
	// NOTE WE don't understand how to instantiate the elements of the set in ValidateConfig

	// Add errors to the response
	for _, error := range errors {
		resp.Diagnostics.AddError("Error during space creation", error.Error())
	}
}

func (r *spaceResource) validateNestedSpaceConfig(_ context.Context, config spaceResourceModel) []error {
	var errors []error
	// The available options are different for root and nested spaces.
	if !config.ParentSpaceUUID.IsNull() {
		// Nested spaces inherit visibility from the parent space.
		// So, it is impossible to set is_private for nested spaces.
		if !config.IsPrivate.IsNull() {
			errors = append(errors, fmt.Errorf("parent space UUID is set, is_private must be empty"))
		}

		// Nested spaces inherit access from the parent space.
		// So, it is impossible to set access or group_access when parent_space_uuid is set.
		if !config.MemberAccessList.IsNull() && len(config.MemberAccessList.Elements()) > 0 {
			errors = append(errors, fmt.Errorf("parent space UUID is set, member access list must be empty"))
		}

		// Nested spaces inherit access from the parent space.
		// So, it is impossible to set access or group_access when parent_space_uuid is set.
		if !config.GroupAccessList.IsNull() && len(config.GroupAccessList.Elements()) > 0 {
			errors = append(errors, fmt.Errorf("parent space UUID is set, group access list must be empty"))
		}
	}
	return errors
}

func (r *spaceResource) validateSpaceVisibilityConfig(_ context.Context, config spaceResourceModel) []error {
	var errors []error
	// A public space shouldn't have access lists
	if !config.IsPrivate.IsNull() && !config.IsPrivate.ValueBool() {
		if !config.MemberAccessList.IsNull() || len(config.MemberAccessList.Elements()) > 0 {
			errors = append(errors, fmt.Errorf("is_private is not set, member access list must be empty"))
		}
		if !config.GroupAccessList.IsNull() || len(config.GroupAccessList.Elements()) > 0 {
			errors = append(errors, fmt.Errorf("is_private is not set, group access list must be empty"))
		}
	}
	return errors
}

func (r *spaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan spaceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract member access details from the 'access' block
	memberAccess := []spaceMemberAccessBlockModel{}
	diags = plan.MemberAccessList.ElementsAs(ctx, &memberAccess, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract group access details from the 'group_access' block
	groupAccess := []spaceGroupAccessBlockModel{}
	diags = plan.GroupAccessList.ElementsAs(ctx, &groupAccess, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Construct the options for the create operation
	createOptions := controllers.CreateSpaceOptions{
		ProjectUUID:     plan.ProjectUUID.ValueString(),
		SpaceName:       plan.SpaceName.ValueString(),
		IsPrivate:       plan.IsPrivate.ValueBoolPointer(),
		ParentSpaceUUID: plan.ParentSpaceUUID.ValueStringPointer(),
		MemberAccess:    convertToControllerMemberAccess(memberAccess), // Convert to controller format
		GroupAccess:     convertToControllerGroupAccess(groupAccess),   // Convert to controller format
	}

	// Create the space via the controller
	createdSpaceDetails, controllerErrors := r.spaceController.CreateSpace(ctx, createOptions)

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
	tflog.Debug(ctx, "Space created", map[string]any{"spaceDetails": createdSpaceDetails})

	// Get the space to fetch all the space access
	tflog.Debug(ctx, "Fetching space", map[string]any{"projectUUID": createdSpaceDetails.ProjectUUID, "spaceUUID": createdSpaceDetails.SpaceUUID})
	fetchedSpaceDetails, err := r.spaceController.GetSpace(ctx, createdSpaceDetails.ProjectUUID, createdSpaceDetails.SpaceUUID)
	if err != nil {
		resp.Diagnostics.AddError("Error during space creation", "Controller returned nil space details")
		return
	}

	// Populate the state with values returned by the controller (which reflects the API response)
	var state spaceResourceModel
	state.ID = types.StringValue(fmt.Sprintf("projects/%s/spaces/%s", createdSpaceDetails.ProjectUUID, createdSpaceDetails.SpaceUUID))
	state.ProjectUUID = types.StringValue(fetchedSpaceDetails.ProjectUUID)
	state.SpaceUUID = types.StringValue(fetchedSpaceDetails.SpaceUUID)
	state.SpaceName = types.StringValue(fetchedSpaceDetails.SpaceName)
	state.IsPrivate = types.BoolValue(fetchedSpaceDetails.IsPrivate)
	// Handle parent space UUID from controller result
	if fetchedSpaceDetails.ParentSpaceUUID != nil {
		state.ParentSpaceUUID = types.StringValue(*fetchedSpaceDetails.ParentSpaceUUID)
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

	// Populate MemberAccessList (direct access from plan)
	state.MemberAccessList = r.populateMemberAccessListSet(ctx, memberAccess, &resp.Diagnostics)

	// Populate GroupAccessList (direct access from plan)
	state.GroupAccessList = r.populateGroupAccessListSet(ctx, groupAccess, &resp.Diagnostics)

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

	// Get space details from API using controller
	projectUUID := currentState.ProjectUUID.ValueString()
	spaceUUID := currentState.SpaceUUID.ValueString()

	tflog.Debug(ctx, "Fetching space", map[string]any{"projectUUID": projectUUID, "spaceUUID": spaceUUID})

	// Use controller's GetSpace which returns SpaceDetails with raw lists
	fetchedSpaceDetails, err := r.spaceController.GetSpace(ctx, projectUUID, spaceUUID)
	if err != nil {
		// If the space is not found, remove it from state
		// TODO: Check for specific "not found" error type if available
		tflog.Warn(ctx, fmt.Sprintf("Space %s not found during Read, removing from state", spaceUUID))
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, "Fetched space details", map[string]any{"spaceDetails": fetchedSpaceDetails})

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

	// Populate MemberAccessList (direct access - filtered in controller/service)
	// Assuming a method exists in models.SpaceMemberAccess to check for direct access
	// If not, manual filtering is needed here.
	// Restore the MemberAccessList from the current state (representing the plan)
	newState.MemberAccessList = currentState.MemberAccessList

	// Populate GroupAccessList (direct access - assuming all groups from API are direct for root spaces)
	// If Lightdash API provides a way to distinguish direct group access, filter here.
	// For now, assuming all groups returned for a root space are 'managed' by this resource.
	// Restore the GroupAccessList from the current state (representing the plan)
	newState.GroupAccessList = currentState.GroupAccessList

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

	tflog.Debug(ctx, "(spaceResource.Update) Updating space", map[string]any{"plan": plan, "oldState": oldState})

	// Extract member access details from the 'access' block
	memberAccess := []spaceMemberAccessBlockModel{}
	diags := plan.MemberAccessList.ElementsAs(ctx, &memberAccess, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract group access details from the 'group_access' block
	groupAccess := []spaceGroupAccessBlockModel{}
	diags = plan.GroupAccessList.ElementsAs(ctx, &groupAccess, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Construct the options for the update operation
	updateOptions := controllers.UpdateSpaceOptions{
		ProjectUUID:     plan.ProjectUUID.ValueString(),
		SpaceUUID:       plan.SpaceUUID.ValueString(),
		SpaceName:       plan.SpaceName.ValueString(),
		IsPrivate:       plan.IsPrivate.ValueBoolPointer(),
		ParentSpaceUUID: plan.ParentSpaceUUID.ValueStringPointer(),
		MemberAccess:    convertToControllerMemberAccess(memberAccess), // Convert to controller format
		GroupAccess:     convertToControllerGroupAccess(groupAccess),   // Convert to controller format
	}

	// Update space using controller
	// The controller will handle the logic of moving the space and managing access based on space type
	updatedSpaceDetails, controllerErrors := r.spaceController.UpdateSpace(ctx, updateOptions)
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

	tflog.Debug(ctx, "(spaceResource.Update) Space updated", map[string]any{
		"spaceDetails": updatedSpaceDetails,
		"isPrivate":    updatedSpaceDetails.IsPrivate,
	})

	// Populate the state with values returned by the controller (which reflect the final API state)
	var updatedState spaceResourceModel
	updatedState.ID = oldState.ID
	updatedState.ProjectUUID = oldState.ProjectUUID // ProjectUUID cannot change
	updatedState.SpaceUUID = oldState.SpaceUUID     // SpaceUUID cannot change

	updatedState.SpaceName = types.StringValue(updatedSpaceDetails.SpaceName)
	updatedState.IsPrivate = types.BoolValue(updatedSpaceDetails.IsPrivate)

	if plan.ParentSpaceUUID.IsNull() {
		updatedState.ParentSpaceUUID = types.StringNull()
	} else {
		updatedState.ParentSpaceUUID = types.StringValue(plan.ParentSpaceUUID.ValueString())
	}

	updatedState.DeleteProtection = plan.DeleteProtection // From plan
	updatedState.CreatedAt = oldState.CreatedAt           // Preserve creation timestamp
	updatedState.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Convert the slices to Set types for state
	updatedState.MemberAccessList = r.populateMemberAccessListSet(ctx, memberAccess, &resp.Diagnostics)
	updatedState.GroupAccessList = r.populateGroupAccessListSet(ctx, groupAccess, &resp.Diagnostics)

	// Set state
	diags = resp.State.Set(ctx, updatedState)
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

	tflog.Debug(ctx, "Deleting space", map[string]any{"projectUUID": projectUUID, "spaceUUID": spaceUUID, "deletionProtection": deletionProtection})

	err := r.spaceController.DeleteSpace(
		ctx,
		controllers.DeleteSpaceOptions{
			ProjectUUID:        projectUUID,
			SpaceUUID:          spaceUUID,
			DeletionProtection: deletionProtection,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting space",
			"Could not delete space: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Space deleted", map[string]any{"projectUUID": projectUUID, "spaceUUID": spaceUUID})
}

// ImportSpace imports an existing space by its resource ID.
func (r *spaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importSpaceOptions := controllers.ImportSpaceOptions{
		ResourceID: req.ID,
	}
	tflog.Debug(ctx, "Importing space", map[string]any{"importSpaceOptions": importSpaceOptions})

	// Fetch space details from the controller
	spaceDetailsFromController, err := r.spaceController.ImportSpace(ctx, importSpaceOptions)
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
	if !isImportedSpaceNested && spaceDetailsFromController.SpaceAccessMembers != nil {
		for _, member := range spaceDetailsFromController.SpaceAccessMembers {
			if member.HasDirectSpaceMemberAccess() {
				directMemberAccessListForImport = append(directMemberAccessListForImport, spaceMemberAccessBlockModel{
					UserUUID:  types.StringValue(member.UserUUID),
					SpaceRole: types.StringValue(string(member.SpaceRole)),
					// HasDirectAccess and InheritedFrom are not part of 'access' input schema, so not set here.
				})
			}
		}
	}

	// Convert direct member access list to Set for import
	memberAccessList, memberDiags := convertToMemberAccessSet(directMemberAccessListForImport)
	resp.Diagnostics.Append(memberDiags...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access"), memberAccessList)...)
	}

	// Populate 'group_access' (only relevant for root spaces or if API provides explicit flag)
	// Assuming API doesn't distinguish explicit group grants, populate with all groups from API result for root spaces.
	// For nested spaces, this block should be empty as access is inherited.
	groupAccessListForImport := []spaceGroupAccessBlockModel{}
	if !isImportedSpaceNested && spaceDetailsFromController.SpaceAccessGroups != nil {
		// For root spaces, populate group_access with all groups returned by the API
		for _, group := range spaceDetailsFromController.SpaceAccessGroups {
			groupAccessListForImport = append(groupAccessListForImport, spaceGroupAccessBlockModel{
				GroupUUID: types.StringValue(group.GroupUUID),
				SpaceRole: types.StringValue(string(group.SpaceRole)),
			})
		}
	}

	// Convert group access list to Set for import
	groupAccessSet, groupAccessSetDiags := convertToGroupAccessSet(groupAccessListForImport)
	resp.Diagnostics.Append(groupAccessSetDiags...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_access"), groupAccessSet)...)
	}

	// Set timestamps
	// Use the CreatedAt from the controller's SpaceDetails, and set LastUpdated to now
	// As mentioned, CreatedAt is not expected from Lightdash, so we'll use current time
	currentTime := types.StringValue(time.Now().Format(time.RFC850))
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), currentTime)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("last_updated"), currentTime)...)
}

func convertToControllerMemberAccess(memberAccess []spaceMemberAccessBlockModel) []models.SpaceAccessMember {
	controllerAccess := make([]models.SpaceAccessMember, 0, len(memberAccess))
	for _, member := range memberAccess {
		controllerAccess = append(controllerAccess, models.SpaceAccessMember{
			UserUUID:  member.UserUUID.ValueString(),
			SpaceRole: models.SpaceMemberRole(member.SpaceRole.ValueString()),
		})
	}
	return controllerAccess
}

func convertToControllerGroupAccess(groupAccess []spaceGroupAccessBlockModel) []models.SpaceAccessGroup {
	controllerAccess := make([]models.SpaceAccessGroup, 0, len(groupAccess))
	for _, group := range groupAccess {
		controllerAccess = append(controllerAccess, models.SpaceAccessGroup{
			GroupUUID: group.GroupUUID.ValueString(),
			SpaceRole: models.SpaceMemberRole(group.SpaceRole.ValueString()),
		})
	}
	return controllerAccess
}

// Remove the _ parameter since we no longer pass context
func convertToMemberAccessSet(memberAccess []spaceMemberAccessBlockModel) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Define the element type for the set
	elementType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"user_uuid":  types.StringType,
			"space_role": types.StringType,
		},
	}

	elements := make([]attr.Value, 0, len(memberAccess))
	for _, access := range memberAccess {
		element, elemDiags := types.ObjectValue(
			elementType.AttrTypes,
			map[string]attr.Value{
				"user_uuid":  access.UserUUID,
				"space_role": access.SpaceRole,
			},
		)
		diags.Append(elemDiags...)
		if elemDiags.HasError() {
			continue
		}

		elements = append(elements, element)
	}

	return types.SetValueMust(elementType, elements), diags
}

// Remove the _ parameter since we no longer pass context
func convertToGroupAccessSet(groupAccess []spaceGroupAccessBlockModel) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Define the element type for the set
	elementType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"group_uuid": types.StringType,
			"space_role": types.StringType,
		},
	}

	elements := make([]attr.Value, 0, len(groupAccess))
	for _, access := range groupAccess {
		element, elemDiags := types.ObjectValue(
			elementType.AttrTypes,
			map[string]attr.Value{
				"group_uuid": access.GroupUUID,
				"space_role": access.SpaceRole,
			},
		)
		diags.Append(elemDiags...)
		if elemDiags.HasError() {
			continue
		}

		elements = append(elements, element)
	}

	return types.SetValueMust(elementType, elements), diags
}

// populateMemberAccessListSet converts a slice of spaceMemberAccessBlockModel to a types.Set.
// This is used for the 'access' block, representing directly assigned member access.
func (r *spaceResource) populateMemberAccessListSet(_ context.Context, members []spaceMemberAccessBlockModel, diags *diag.Diagnostics) types.Set {
	if len(members) == 0 {
		// Return empty set with correct element type if input is empty
		elementType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"user_uuid":  types.StringType,
				"space_role": types.StringType,
			},
		}
		return types.SetValueMust(elementType, nil)
	}

	memberAccessSet, conversionDiags := convertToMemberAccessSet(members)
	diags.Append(conversionDiags...)
	if conversionDiags.HasError() {
		// Ensure empty set with correct type if conversion failed
		elementType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"user_uuid":  types.StringType,
				"space_role": types.StringType,
			},
		}
		return types.SetValueMust(elementType, nil)
	}
	return memberAccessSet
}

// populateGroupAccessListSet converts a slice of spaceGroupAccessBlockModel to a types.Set.
// This is used for the 'group_access' block, representing directly assigned group access.
func (r *spaceResource) populateGroupAccessListSet(ctx context.Context, groups []spaceGroupAccessBlockModel, diags *diag.Diagnostics) types.Set {
	if len(groups) == 0 {
		tflog.Debug(ctx, "populateGroupAccessListSet: No groups to populate")

		// Return empty set with correct element type if input is empty
		elementType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"group_uuid": types.StringType,
				"space_role": types.StringType,
			},
		}
		return types.SetValueMust(elementType, nil)
	}

	groupAccessSet, conversionDiags := convertToGroupAccessSet(groups)
	diags.Append(conversionDiags...)
	if conversionDiags.HasError() {
		// Ensure empty set with correct type if conversion failed
		elementType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"group_uuid": types.StringType,
				"space_role": types.StringType,
			},
		}
		return types.SetValueMust(elementType, nil)
	}
	tflog.Debug(ctx, "populateGroupAccessListSet: Populated group access list set", map[string]any{"groupAccessSet": groupAccessSet})

	return groupAccessSet
}
