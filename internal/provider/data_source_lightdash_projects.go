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
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &projectsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectsDataSource{}
)

func NewProjectsDataSource() datasource.DataSource {
	return &projectsDataSource{}
}

// projectDataSource defines the data source implementation.
type projectsDataSource struct {
	client *api.Client
}

type nestedProjectModel struct {
	ProjectUUID types.String `tfsdk:"project_uuid"`
	ProjectName types.String `tfsdk:"name"`
	ProjectType types.String `tfsdk:"type"`
}

// projectDataSourceModel describes the data source data model.
type projectsDataSourceModel struct {
	ID               types.String         `tfsdk:"id"`
	OrganizationUUID types.String         `tfsdk:"organization_uuid"`
	Projects         []nestedProjectModel `tfsdk:"projects"`
}

func (d *projectsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *projectsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_lightdash_projects.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Lightdash projects data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				Description: "Organization UUID of the Lightdash project.",
				Optional:    true,
				Computed:    true,
			},
			"projects": schema.ListNestedAttribute{
				Description: "List of projects.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project_uuid": schema.StringAttribute{
							Description: "Project UUID of the Lightdash project.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Lightdash project name.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Lightdash project type.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *projectsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projects, err := d.client.ListOrganizationProjectsV1()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lightdash project",
			err.Error(),
		)
		return
	}

	// Map response body to model
	updatedProjects := []nestedProjectModel{}
	for _, project := range projects {
		projectState := nestedProjectModel{
			ProjectUUID: types.StringValue(project.ProjectUUID),
			ProjectName: types.StringValue(project.ProjectName),
			ProjectType: types.StringValue(project.ProjectType),
		}
		updatedProjects = append(updatedProjects, projectState)
	}
	// Sort the projects by project UUID
	sort.Slice(updatedProjects, func(i, j int) bool {
		return updatedProjects[i].ProjectUUID.ValueString() < updatedProjects[j].ProjectUUID.ValueString()
	})
	state.Projects = updatedProjects

	// Set resource ID
	state_id := fmt.Sprintf("organizations/%s/projects", state.OrganizationUUID)
	state.ID = types.StringValue(state_id)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
