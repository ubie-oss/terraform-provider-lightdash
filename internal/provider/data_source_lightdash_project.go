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
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

// projectDataSource defines the data source implementation.
type projectDataSource struct {
	client *api.Client
}

// projectDataSourceModel describes the data source data model.
type projectDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationUuid types.String `tfsdk:"organization_uuid"`
	ProjectUuid      types.String `tfsdk:"project_uuid"`
	ProjectType      types.String `tfsdk:"project_type"`
	Name             types.String `tfsdk:"name"`
}

func (d *projectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lightdash project data source",
		Description:         "Lightdash project data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash organization UUID",
				Optional:            true,
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash project UUID",
				Required:            true,
			},
			"project_type": schema.StringAttribute{
				MarkdownDescription: "Lightdash project type",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Lightdash project name",
				Computed:            true,
			},
		},
	}
}

func (d *projectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	d.client = client
}

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project_uuid := state.ProjectUuid.ValueString()
	project, err := d.client.GetProjectV1(project_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lightdash project",
			err.Error(),
		)
		return
	}

	// Map response body to model
	state.OrganizationUuid = types.StringValue(project.OrganizationUUID)
	state.ProjectUuid = types.StringValue(project.ProjectUUID)
	state.ProjectType = types.StringValue(project.ProjectType)
	state.Name = types.StringValue(project.ProjectName)

	// Set resource ID
	state_id := fmt.Sprintf("organizations/%s/projects/%s", project.OrganizationUUID, project.ProjectUUID)
	state.ID = types.StringValue(state_id)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
