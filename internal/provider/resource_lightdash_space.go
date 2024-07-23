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
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
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
	UserUUID            types.String `tfsdk:"user_uuid"`
	SpaceRole           types.String `tfsdk:"space_role"`
	IsOrganizationAdmin types.Bool   `tfsdk:"is_organization_admin"`
	// TODO support last_updated
	// LastUpdated types.String `tfsdk:"last_updated"`
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

	// Add space access
	accessList := []spaceResourceAccessBlockModel{}
	var errors []error
	organizationMembersService := services.NewOrganizationMembersService(r.client)
	for _, access := range plan.AccessList {
		// Organization admins shouldn't be managed in Terraform states,
		// because they have access to all spaces in the organization by default.
		isOrganizationAdmin, err := organizationMembersService.IsOrganizationAdmin(access.UserUUID.ValueString())
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Error checking if user %s is an organization admin: %s", access.UserUUID, err.Error()))
			errors = append(errors, err)
			continue
		}
		if isOrganizationAdmin {
			tflog.Info(ctx, fmt.Sprintf("Skipping adding access for organization admin user: %s because organization admins inherently have access to all spaces by default, making explicit access management unnecessary.", access.UserUUID))
		}

		// Add space access
		spaceRole := models.SpaceMemberRole(access.SpaceRole.ValueString())
		err = services.GrantSpaceAccess(
			r.client, project_uuid, created_space.SpaceUUID,
			access.UserUUID.ValueString(), spaceRole)
		if err != nil {
			tflog.Debug(ctx, fmt.Sprintf("Error adding space access %s: %s", access.UserUUID, err.Error()))
			errors = append(errors, err)
		} else {
			accessList = append(accessList, spaceResourceAccessBlockModel{
				UserUUID:            access.UserUUID,
				SpaceRole:           access.SpaceRole,
				IsOrganizationAdmin: types.BoolValue(isOrganizationAdmin),
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

	// Assign the plan values to the state
	state_id := getSpaceResourceId(created_space.ProjectUUID, created_space.SpaceUUID)
	plan.ID = types.StringValue(state_id)
	plan.ProjectUUID = types.StringValue(created_space.ProjectUUID)
	plan.SpaceUUID = types.StringValue(created_space.SpaceUUID)
	plan.IsPrivate = types.BoolValue(created_space.IsPrivate)
	plan.DeleteProtection = types.BoolValue(plan.DeleteProtection.ValueBool())
	plan.AccessList = accessList
	plan.CreatedAt = types.StringValue(time.Now().Format(time.RFC850))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *spaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Declare variables to import from state
	var projectUuid string
	var spaceUuid string
	var access []spaceResourceAccessBlockModel
	var created_at string
	var last_updated string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project_uuid"), &projectUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("space_uuid"), &spaceUuid)...)
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
	projectUuid = state.ProjectUUID.ValueString()
	spaceUuid = state.SpaceUUID.ValueString()
	space, err := r.client.GetSpaceV1(projectUuid, spaceUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading space",
			"Could not read space ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Get space members
	organizationMembersService := services.NewOrganizationMembersService(r.client)
	var members []spaceResourceAccessBlockModel
	for _, access := range state.AccessList {
		// Continue if the user no longer exists in the organization
		_, err := organizationMembersService.GetOrganizationMemberByUserUuid(access.UserUUID.ValueString())
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("User %s no longer exists in the organization. Skipping reading access to the space.", access.UserUUID))
			continue
		}
		// Check if the user who isn't is an organization admin
		isOrganizationAdmin, err := organizationMembersService.IsOrganizationAdmin(access.UserUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error checking if user is an organization admin",
				"Could not check if user is an organization admin: "+err.Error(),
			)
		}
		// Skip if the user who isn't in the state is an organization admin
		if isOrganizationAdmin {
			tflog.Info(ctx, fmt.Sprintf("Organization admin %s is registered in Terraform states. However, granting and revoking operations for organization admins are not executed because organization admins inherently have access to all spaces by default, making explicit access management unnecessary.", access.UserUUID))
		}
		// Append the user to the members list
		members = append(members, spaceResourceAccessBlockModel{
			UserUUID:            access.UserUUID,
			SpaceRole:           access.SpaceRole,
			IsOrganizationAdmin: types.BoolValue(isOrganizationAdmin),
		})
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
	updatedSpace, err := r.client.UpdateSpaceV1(project_uuid, space_uuid, space_name, is_private)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating space",
			"Could not update space, unexpected error: "+err.Error(),
		)
		return
	}

	// Get the space to get the access members
	space, err := r.client.GetSpaceV1(project_uuid, space_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting space",
			"Could not get space, unexpected error: "+err.Error(),
		)
		return
	}

	// Revoke access from users not managed by Terraform
	// NOTE: Manually added users will be removed
	organizationMembersService := services.NewOrganizationMembersService(r.client)
	for _, existingAccess := range space.SpaceAccess {
		// Check if the user is an organization admin
		// It is not possible to revoke space access from organization admins
		isOrganizationAdmin, err := organizationMembersService.IsOrganizationAdmin(existingAccess.UserUUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error checking if user is an organization admin",
				"Could not check if user is an organization admin: "+err.Error(),
			)
		}
		if isOrganizationAdmin {
			tflog.Debug(ctx, fmt.Sprintf("Skipping user %s who is an organization admin. Organization admins have space access to all spaces by default and cannot be managed individually.", existingAccess.UserUUID))
			continue
		}
		// Check if the user is managed by Terraform
		found := false
		for _, access := range plan.AccessList {
			if access.UserUUID.ValueString() == existingAccess.UserUUID {
				found = true
				break
			}
		}
		// Revoke access if the user is not in the access list
		if !found {
			err := r.client.RevokeSpaceAccessV1(
				project_uuid, plan.SpaceUUID.ValueString(), existingAccess.UserUUID)
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
	var updatedAccessList []spaceResourceAccessBlockModel
	for _, access := range plan.AccessList {
		// Skip if the user doesn't exist in the organization
		_, err := organizationMembersService.GetOrganizationMemberByUserUuid(access.UserUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddWarning(
				"User no longer exists in the organization",
				fmt.Sprintf("User %s no longer exists in the organization. Skipping granting access to the space.", access.UserUUID),
			)
			continue
		}
		// Check if the user is an organization admin
		// It is not possible to revoke space access from organization admins
		isOrganizationAdmin, err := organizationMembersService.IsOrganizationAdmin(access.UserUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error checking if user is an organization admin",
				"Could not check if user is an organization admin: "+err.Error(),
			)
		}
		// Skip if the user is an organization admin
		if isOrganizationAdmin {
			tflog.Debug(ctx, fmt.Sprintf("Skipping user %s who is an organization admin because organization admins inherently have access to all spaces by default, making individual space access management unnecessary.", access.UserUUID))
		}
		// Skip if the user is already in the access list
		found := false
		for _, existingAccess := range space.SpaceAccess {
			if access.UserUUID.ValueString() == existingAccess.UserUUID {
				found = true
				break
			}
		}
		// Grant access if the user is not in the access list
		if !found {
			spaceRole := models.SpaceMemberRole(access.SpaceRole.ValueString())
			err := services.GrantSpaceAccess(
				r.client, project_uuid, plan.SpaceUUID.ValueString(), access.UserUUID.ValueString(), spaceRole)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Granting space access",
					"Could not grant space access, unexpected error: "+err.Error(),
				)
				return
			}
		}
		// Append the user to the managed members list
		updatedAccessList = append(
			updatedAccessList,
			spaceResourceAccessBlockModel{
				UserUUID:            access.UserUUID,
				SpaceRole:           types.StringValue(access.SpaceRole.ValueString()),
				IsOrganizationAdmin: types.BoolValue(isOrganizationAdmin),
			})
	}

	// Update the state
	plan.SpaceName = types.StringValue(updatedSpace.SpaceName)
	plan.IsPrivate = types.BoolValue(updatedSpace.IsPrivate)
	plan.AccessList = updatedAccessList
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
			fmt.Sprintf("Could not delete space with UUID %s, deletion protection is enabled", state.SpaceUUID),
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

	// Get the importedSpace
	importedSpace, err := r.client.GetSpaceV1(project_uuid, space_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting space",
			fmt.Sprintf("Could not get space with project UUID %s and space UUID %s, unexpected error: %s", project_uuid, space_uuid, err.Error()),
		)
		return
	}

	// Get the space members
	accessList := []spaceResourceAccessBlockModel{}
	organizationMembersService := services.NewOrganizationMembersService(r.client)
	for _, access := range importedSpace.SpaceAccess {
		isOrganizationAdmin, err := organizationMembersService.IsOrganizationAdmin(access.UserUUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error checking if user is an organization admin",
				"Could not check if user is an organization admin: "+err.Error(),
			)
		}
		// Append each element to the slice
		accessList = append(accessList, spaceResourceAccessBlockModel{
			UserUUID:            types.StringValue(access.UserUUID),
			SpaceRole:           types.StringValue(access.SpaceRole.String()),
			IsOrganizationAdmin: types.BoolValue(isOrganizationAdmin),
		})
	}

	// Set the resource attributes
	stateId := getSpaceResourceId(importedSpace.ProjectUUID, importedSpace.SpaceUUID)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), stateId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), importedSpace.ProjectUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_uuid"), importedSpace.SpaceUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), importedSpace.SpaceName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("is_private"), importedSpace.IsPrivate)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access"), accessList)...)

	// resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access"), accessList)...)
	// Note We put the current time as the last updated time because we don't know when the space was last updated.
	currentTime := types.StringValue(time.Now().Format(time.RFC850))
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), currentTime)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("last_updated"), currentTime)...)
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
