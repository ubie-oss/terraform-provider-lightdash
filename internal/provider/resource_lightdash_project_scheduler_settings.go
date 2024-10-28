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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &projectSchedulerSettingsResource{}
	_ resource.ResourceWithConfigure   = &projectSchedulerSettingsResource{}
	_ resource.ResourceWithImportState = &projectSchedulerSettingsResource{}
)

func NewProjectSchedulerSettingsResource() resource.Resource {
	return &projectSchedulerSettingsResource{}
}

// projectSchedulerSettingsResource defines the resource implementation.
type projectSchedulerSettingsResource struct {
	client *api.Client
}

// projectSchedulerSettingsResourceModel describes the resource data model.
type projectSchedulerSettingsResourceModel struct {
	ID                types.String `tfsdk:"id"`
	OrganizationUUID  types.String `tfsdk:"organization_uuid"`
	ProjectUUID       types.String `tfsdk:"project_uuid"`
	SchedulerTimezone types.String `tfsdk:"scheduler_timezone"`
}

func (r *projectSchedulerSettingsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_scheduler_settings"
}

func (r *projectSchedulerSettingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Lightdash scheduler settings resource manages the scheduling configurations for projects and resources within an organization.",
		Description:         "Manages Lightdash scheduler settings",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier for the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization to which the settings belong.",
				Required:            true,
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash project to which the settings belong.",
				Required:            true,
			},
			"scheduler_timezone": schema.StringAttribute{
				MarkdownDescription: "The timezone of the Lightdash scheduler settings.",
				Required:            true,
			},
		},
	}
}

func (r *projectSchedulerSettingsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectSchedulerSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan projectSchedulerSettingsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new scheduler settings
	organization_uuid := plan.OrganizationUUID.ValueString()
	settings_name := plan.Name.ValueString()
	enabled := plan.Enabled.ValueBool()

	createdSettings, err := r.client.CreateSchedulerSettingsV1(organization_uuid, settings_name, enabled)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating scheduler settings",
			"Could not create scheduler settings, unexpected error: "+err.Error(),
		)
		return
	}

	// Assign the plan values to the state
	stateId := getSchedulerSettingsResourceId(organization_uuid, createdSettings.SettingsUUID)
	plan.ID = types.StringValue(stateId)
	plan.OrganizationUUID = types.StringValue(organization_uuid)
	plan.ProjectUUID = types.StringValue(createdSettings.ProjectUUID)
	plan.SchedulerTimezone = types.StringValue(createdSettings.SchedulerTimezone)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectSchedulerSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Declare variables to import from state
	var organizationUuid string
	var settingsUuid string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("organization_uuid"), &organizationUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project_uuid"), &projectUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("scheduler_timezone"), &schedulerTimezone)...)

	// Get current state
	var state projectSchedulerSettingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get scheduler settings
	settings, err := services.GetProjectSchedulerSettings(r.client, projectUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading scheduler settings",
			"Could not read settings ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Set the state values
	state.OrganizationUUID = types.StringValue(settings.OrganizationUUID)
	state.ProjectUUID = types.StringValue(settings.ProjectUUID)
	state.SchedulerTimezone = types.StringValue(settings.SchedulerTimezone)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectSchedulerSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan, state projectSchedulerSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the information from the plan
	projectUuid := plan.ProjectUUID.ValueString()
	schedulerTimezone := plan.SchedulerTimezone.ValueString()

	// Update the scheduler settings
	updatedSettings, err := r.client.UpdateSchedulerSettingsV1(projectUuid, schedulerTimezone)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating scheduler settings",
			fmt.Sprintf("Could not update scheduler settings with UUID '%s' and name '%s', unexpected error: %s", plan.SettingsUUID, plan.Name, err.Error()),
		)
		return
	}

	// Update the state
	plan.ProjectUUID = types.StringValue(updatedSettings.ProjectUUID)
	plan.SchedulerTimezone = types.StringValue(updatedSettings.SchedulerTimezone)

	// Set state
	diags := resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectSchedulerSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state projectSchedulerSettingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing scheduler settings
	settingsUuid := state.SettingsUUID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting scheduler settings %s", projectUuid))
	err := r.client.DeleteSchedulerSettingsV1(projectUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting scheduler settings",
			"Could not delete scheduler settings, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *projectSchedulerSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Extract the resource ID
	extracted_strings, err := extractSchedulerSettingsResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	organization_uuid := extracted_strings[0]
	projectUuid := extracted_strings[1]

	// Get the imported scheduler settings
	importedSettings, err := services.GetProjectSchedulerSettings(r.client, projectUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting scheduler settings",
			fmt.Sprintf("Could not get scheduler settings with organization UUID %s and project UUID %s, unexpected error: %s", organization_uuid, projectUuid, err.Error()),
		)
		return
	}

	// Set the resource attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), projectUuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_uuid"), importedSettings.OrganizationUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), importedSettings.ProjectUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scheduler_timezone"), importedSettings.SchedulerTimezone)...)
}

func getSchedulerSettingsResourceId(organization_uuid string, settings_uuid string) string {
	// Return the resource ID
	return fmt.Sprintf("organizations/%s/scheduler_settings/%s", organization_uuid, settings_uuid)
}

func extractSchedulerSettingsResourceId(input string) ([]string, error) {
	// Extract the captured groups
	pattern := `^organizations/([^/]+)/scheduler_settings/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	// Return the captured strings
	organization_uuid := groups[0]
	settings_uuid := groups[1]
	return []string{organization_uuid, settings_uuid}, nil
}
