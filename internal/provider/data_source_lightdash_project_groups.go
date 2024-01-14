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
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &projectGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectGroupsDataSource{}
)

func NewProjectGroupsDataSource() datasource.DataSource {
	return &projectGroupsDataSource{}
}

// projectGroupsDataSource defines the data source implementation.
type projectGroupsDataSource struct {
	client *api.Client
}

type nestedProjectGroupsModel struct {
	ProjectUUID types.String             `tfsdk:"project_uuid"`
	GroupUUID   types.String             `tfsdk:"group_uuid"`
	Role        models.ProjectMemberRole `tfsdk:"role"`
}

// projectGroupsDataSourceModel describes the data source data model.
type projectGroupsDataSourceModel struct {
	ID               types.String               `tfsdk:"id"`
	OrganizationUUID types.String               `tfsdk:"organization_uuid"`
	ProjectUUID      types.String               `tfsdk:"project_uuid"`
	Groups           []nestedProjectGroupsModel `tfsdk:"groups"`
}

func (d *projectGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_groups"
}

func (d *projectGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lightdash project group accesses data source",
		Description:         "Lightdash project group accesses data source",
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
			"groups": schema.ListNestedAttribute{
				Description: "List of group accesses.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project_uuid": schema.StringAttribute{
							Description: "Project UUID of the Lightdash project.",
							Computed:    true,
						},
						"group_uuid": schema.StringAttribute{
							Description: "Group UUID of the Lightdash group.",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "Role of the group in the Lightdash project.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *projectGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project_uuid := state.ProjectUUID.ValueString()
	groupAccesses, err := d.client.GetProjectGroupAccessesV1(project_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lightdash project group accesses",
			"Error: "+err.Error(),
		)
		return
	}

	// Map response body to model
	var groupAccessesList = []nestedProjectGroupsModel{}
	for _, groupAccess := range groupAccesses {
		accessState := nestedProjectGroupsModel{
			ProjectUUID: types.StringValue(groupAccess.ProjectUUID),
			GroupUUID:   types.StringValue(groupAccess.GroupUUID),
			Role:        groupAccess.Role,
		}
		groupAccessesList = append(groupAccessesList, accessState)
	}
	state.Groups = groupAccessesList

	// Set resource ID
	state_id := fmt.Sprintf("organizations/%s/projects/%s/groups",
		state.OrganizationUUID.ValueString(), state.ProjectUUID.ValueString())
	state.ID = types.StringValue(state_id)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
