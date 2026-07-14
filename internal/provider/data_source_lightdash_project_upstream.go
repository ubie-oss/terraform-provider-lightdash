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
	_ datasource.DataSource              = &projectUpstreamDataSource{}
	_ datasource.DataSourceWithConfigure = &projectUpstreamDataSource{}
)

func NewProjectUpstreamDataSource() datasource.DataSource {
	return &projectUpstreamDataSource{}
}

// projectUpstreamDataSource defines the data source implementation.
type projectUpstreamDataSource struct {
	client *api.Client
}

// projectUpstreamDataSourceModel describes the data source data model.
type projectUpstreamDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	OrganizationUUID    types.String `tfsdk:"organization_uuid"`
	ProjectUUID         types.String `tfsdk:"project_uuid"`
	UpstreamProjectUUID types.String `tfsdk:"upstream_project_uuid"`
}

func (d *projectUpstreamDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_upstream"
}

func (d *projectUpstreamDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_project_upstream.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Lightdash project upstream data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The data source identifier. It is computed as `organizations/<organization_uuid>/projects/<project_uuid>/upstream`.",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
				Required:            true,
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash project whose upstream link should be read.",
				Required:            true,
			},
			"upstream_project_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the upstream Lightdash project used for content promotion. Null when no upstream is configured.",
				Computed:            true,
			},
		},
	}
}

func (d *projectUpstreamDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectUpstreamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectUpstreamDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectUUID := state.ProjectUUID.ValueString()
	upstreamService := services.NewProjectUpstreamService(d.client, projectUUID)
	upstreamUUID, err := upstreamService.GetProjectUpstream(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Lightdash project upstream (project UUID: %s)", projectUUID),
			err.Error(),
		)
		return
	}

	if upstreamUUID == nil {
		state.UpstreamProjectUUID = types.StringNull()
	} else {
		state.UpstreamProjectUUID = types.StringValue(*upstreamUUID)
	}

	state.ID = types.StringValue(getProjectUpstreamResourceId(
		state.OrganizationUUID.ValueString(), projectUUID))

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
