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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &projectSchedulerSettingsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectSchedulerSettingsDataSource{}
)

func NewProjectSchedulerSettingsDataSource() datasource.DataSource {
	return &projectSchedulerSettingsDataSource{}
}

// projectSchedulerSettingsDataSource defines the data source implementation.
type projectSchedulerSettingsDataSource struct {
	client *api.Client
}

// projectSchedulerSettingsDataSourceModel describes the data source data model.
type projectSchedulerSettingsDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	OrganizationUUID  types.String `tfsdk:"organization_uuid"`
	ProjectUUID       types.String `tfsdk:"project_uuid"`
	SchedulerTimezone types.String `tfsdk:"scheduler_timezone"`
}

// Metadata defines the metadata for the data source.
func (d *projectSchedulerSettingsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_scheduler_settings"
}

// Schema defines the schema for the data source.
func (d *projectSchedulerSettingsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lightdash project scheduler settings data source",
		Description:         "Lightdash project scheduler settings data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				Description: "Organization UUID of the Lightdash project.",
				Required:    true,
			},
			"project_uuid": schema.StringAttribute{
				Description: "Project UUID of the Lightdash project.",
				Required:    true,
			},
			"scheduler_timezone": schema.StringAttribute{
				Description: "Timezone for the Lightdash project scheduler.",
				Required:    false,
				Computed:    true,
			},
		},
	}
}

// Configure configures the data source.
func (d *projectSchedulerSettingsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	d.client = client
}

// Read reads the data source.
func (d *projectSchedulerSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectSchedulerSettingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch scheduler settings from the API
	schedulerSettingsService := services.NewProjectSchedulerSettingsService(d.client, state.ProjectUUID.ValueString())
	settings, err := schedulerSettingsService.GetProjectSchedulerSettings(state.ProjectUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Lightdash project scheduler settings (project UUID: %s)", state.ProjectUUID.ValueString()),
			err.Error(),
		)
		return
	}

	// Set the state with fetched settings
	state.SchedulerTimezone = types.StringValue(settings.GetSchedulerTimezone()) // Change to use the correct field

	// Set resource ID
	state.ID = types.StringValue(fmt.Sprintf("organizations/%s/projects/%s/scheduler_settings",
		state.OrganizationUUID.ValueString(), state.ProjectUUID.ValueString()))

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
