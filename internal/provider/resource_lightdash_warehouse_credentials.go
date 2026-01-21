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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &warehouseCredentialsResource{}
	_ resource.ResourceWithConfigure   = &warehouseCredentialsResource{}
	_ resource.ResourceWithImportState = &warehouseCredentialsResource{}
)

func NewWarehouseCredentialsResource() resource.Resource {
	return &warehouseCredentialsResource{}
}

// warehouseCredentialsResource defines the resource implementation.
type warehouseCredentialsResource struct {
	client *api.Client
}

// warehouseCredentialsResourceModel describes the resource data model.
type warehouseCredentialsResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	OrganizationUUID          types.String `tfsdk:"organization_uuid"`
	OrganizationWarehouseUUID types.String `tfsdk:"organization_warehouse_uuid"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	WarehouseType             types.String `tfsdk:"warehouse_type"`
	Project                   types.String `tfsdk:"project"`
	Dataset                   types.String `tfsdk:"dataset"`
	KeyfileContents           types.String `tfsdk:"keyfile_contents"`
	Location                  types.String `tfsdk:"location"`
	TimeoutSeconds            types.Int64  `tfsdk:"timeout_seconds"`
	MaximumBytesBilled        types.Int64  `tfsdk:"maximum_bytes_billed"`
	Priority                  types.String `tfsdk:"priority"`
	Retries                   types.Int64  `tfsdk:"retries"`
	StartOfWeek               types.Int64  `tfsdk:"start_of_week"`
}

func (r *warehouseCredentialsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_warehouse_credentials"
}

func (r *warehouseCredentialsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Lightdash warehouse credentials for BigQuery with service account key authentication.",
		Description:         "Manages a Lightdash warehouse credentials",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource identifier. It is computed as `organizations/<organization_uuid>/warehouse-credentials/<organization_warehouse_uuid>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_warehouse_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the warehouse credentials.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the warehouse credentials.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the warehouse credentials.",
				Optional:            true,
			},
			"warehouse_type": schema.StringAttribute{
				MarkdownDescription: "The type of the warehouse. Currently only 'bigquery' is supported.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project": schema.StringAttribute{
				MarkdownDescription: "The GCP project ID for BigQuery.",
				Required:            true,
			},
			"dataset": schema.StringAttribute{
				MarkdownDescription: "The BigQuery dataset name.",
				Optional:            true,
			},
			"keyfile_contents": schema.StringAttribute{
				MarkdownDescription: "The contents of the service account key file in JSON format.",
				Required:            true,
				Sensitive:           true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "The location of the BigQuery dataset.",
				Optional:            true,
			},
			"timeout_seconds": schema.Int64Attribute{
				MarkdownDescription: "The timeout for BigQuery queries in seconds.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"maximum_bytes_billed": schema.Int64Attribute{
				MarkdownDescription: "The maximum bytes that can be billed for a query.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"priority": schema.StringAttribute{
				MarkdownDescription: "The priority for BigQuery jobs (INTERACTIVE or BATCH).",
				Optional:            true,
			},
			"retries": schema.Int64Attribute{
				MarkdownDescription: "The number of retries for failed queries.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"start_of_week": schema.Int64Attribute{
				MarkdownDescription: "The start of week (0 = Sunday, 1 = Monday, etc.).",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *warehouseCredentialsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *warehouseCredentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan warehouseCredentialsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build BigQuery credentials
	credentials := models.BigQueryCredentials{
		Type:            plan.WarehouseType.ValueString(),
		Project:         plan.Project.ValueString(),
		KeyfileContents: plan.KeyfileContents.ValueString(),
	}

	if !plan.Dataset.IsNull() {
		dataset := plan.Dataset.ValueString()
		credentials.Dataset = &dataset
	}

	if !plan.Location.IsNull() {
		location := plan.Location.ValueString()
		credentials.Location = &location
	}

	if !plan.TimeoutSeconds.IsNull() {
		timeout := int(plan.TimeoutSeconds.ValueInt64())
		credentials.TimeoutSeconds = &timeout
	}

	if !plan.MaximumBytesBilled.IsNull() {
		maxBytes := plan.MaximumBytesBilled.ValueInt64()
		credentials.MaximumBytesBilled = &maxBytes
	}

	if !plan.Priority.IsNull() {
		priority := plan.Priority.ValueString()
		credentials.Priority = &priority
	}

	if !plan.Retries.IsNull() {
		retries := int(plan.Retries.ValueInt64())
		credentials.Retries = &retries
	}

	if !plan.StartOfWeek.IsNull() {
		startOfWeek := int(plan.StartOfWeek.ValueInt64())
		credentials.StartOfWeek = &startOfWeek
	}

	var description *string
	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		description = &desc
	}

	// Create warehouse credentials
	createdCreds, err := r.client.CreateWarehouseCredentialsV1(
		plan.Name.ValueString(),
		credentials,
		description,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating warehouse credentials",
			"Could not create warehouse credentials, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state
	organizationUUID := plan.OrganizationUUID.ValueString()
	stateId := getWarehouseCredentialsResourceId(organizationUUID, createdCreds.OrganizationWarehouseUUID)
	plan.ID = types.StringValue(stateId)
	plan.OrganizationWarehouseUUID = types.StringValue(createdCreds.OrganizationWarehouseUUID)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *warehouseCredentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state warehouseCredentialsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get warehouse credentials
	creds, err := r.client.GetWarehouseCredentialsV1(state.OrganizationWarehouseUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading warehouse credentials",
			"Could not read warehouse credentials ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state
	state.Name = types.StringValue(creds.Name)
	state.WarehouseType = types.StringValue(creds.WarehouseType)
	state.OrganizationUUID = types.StringValue(creds.OrganizationUUID)

	if creds.Description != nil {
		state.Description = types.StringValue(*creds.Description)
	} else {
		state.Description = types.StringNull()
	}

	// Note: credentials are not returned in the API response for security reasons
	// We keep the existing values from the state

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *warehouseCredentialsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state warehouseCredentialsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build BigQuery credentials
	credentials := models.BigQueryCredentials{
		Type:            plan.WarehouseType.ValueString(),
		Project:         plan.Project.ValueString(),
		KeyfileContents: plan.KeyfileContents.ValueString(),
	}

	if !plan.Dataset.IsNull() {
		dataset := plan.Dataset.ValueString()
		credentials.Dataset = &dataset
	}

	if !plan.Location.IsNull() {
		location := plan.Location.ValueString()
		credentials.Location = &location
	}

	if !plan.TimeoutSeconds.IsNull() {
		timeout := int(plan.TimeoutSeconds.ValueInt64())
		credentials.TimeoutSeconds = &timeout
	}

	if !plan.MaximumBytesBilled.IsNull() {
		maxBytes := plan.MaximumBytesBilled.ValueInt64()
		credentials.MaximumBytesBilled = &maxBytes
	}

	if !plan.Priority.IsNull() {
		priority := plan.Priority.ValueString()
		credentials.Priority = &priority
	}

	if !plan.Retries.IsNull() {
		retries := int(plan.Retries.ValueInt64())
		credentials.Retries = &retries
	}

	if !plan.StartOfWeek.IsNull() {
		startOfWeek := int(plan.StartOfWeek.ValueInt64())
		credentials.StartOfWeek = &startOfWeek
	}

	var description *string
	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		description = &desc
	}

	// Update warehouse credentials
	tflog.Info(ctx, fmt.Sprintf("Updating warehouse credentials %s", plan.OrganizationWarehouseUUID.ValueString()))
	updatedCreds, err := r.client.UpdateWarehouseCredentialsV1(
		plan.OrganizationWarehouseUUID.ValueString(),
		plan.Name.ValueString(),
		credentials,
		description,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating warehouse credentials",
			fmt.Sprintf("Could not update warehouse credentials with UUID '%s', unexpected error: %s", plan.OrganizationWarehouseUUID.ValueString(), err.Error()),
		)
		return
	}

	// Update state
	plan.Name = types.StringValue(updatedCreds.Name)
	if updatedCreds.Description != nil {
		plan.Description = types.StringValue(*updatedCreds.Description)
	} else {
		plan.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *warehouseCredentialsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state warehouseCredentialsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete warehouse credentials
	uuid := state.OrganizationWarehouseUUID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting warehouse credentials %s", uuid))
	err := r.client.DeleteWarehouseCredentialsV1(uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting warehouse credentials",
			"Could not delete warehouse credentials, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *warehouseCredentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Extract the resource ID
	extractedStrings, err := extractWarehouseCredentialsResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	organizationUUID := extractedStrings[0]
	warehouseUUID := extractedStrings[1]

	// Get the imported warehouse credentials
	importedCreds, err := r.client.GetWarehouseCredentialsV1(warehouseUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting warehouse credentials",
			fmt.Sprintf("Could not get warehouse credentials with organization UUID %s and warehouse UUID %s, unexpected error: %s", organizationUUID, warehouseUUID, err.Error()),
		)
		return
	}

	// Set the resource attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_uuid"), importedCreds.OrganizationUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_warehouse_uuid"), importedCreds.OrganizationWarehouseUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), importedCreds.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("warehouse_type"), importedCreds.WarehouseType)...)

	if importedCreds.Description != nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), *importedCreds.Description)...)
	}

	// Note: credentials are not returned in the API response for security reasons
	// User will need to set them manually in the configuration
	resp.Diagnostics.AddWarning(
		"Credentials not imported",
		"The warehouse credentials (keyfile_contents, project, etc.) are not returned by the API for security reasons. You must manually configure these in your Terraform configuration.",
	)
}

func getWarehouseCredentialsResourceId(organizationUUID string, warehouseUUID string) string {
	return fmt.Sprintf("organizations/%s/warehouse-credentials/%s", organizationUUID, warehouseUUID)
}

func extractWarehouseCredentialsResourceId(input string) ([]string, error) {
	pattern := `^organizations/([^/]+)/warehouse-credentials/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	organizationUUID := groups[0]
	warehouseUUID := groups[1]
	return []string{organizationUUID, warehouseUUID}, nil
}
