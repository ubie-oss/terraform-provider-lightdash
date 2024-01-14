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
	_ datasource.DataSource              = &organizationGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationGroupsDataSource{}
)

func NewOrganizationGroupsDataSource() datasource.DataSource {
	return &organizationGroupsDataSource{}
}

// lightdashOrganizationMemberDataSource defines the data source implementation.
type organizationGroupsDataSource struct {
	client *api.Client
}

// organizationGroupModel describes the data source data model for a Lightdash group.
type organizationGroupModel struct {
	GroupUuid types.String `tfsdk:"group_uuid"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
}

// organizationGroupsDataSourceModel describes the data source data model.
type organizationGroupsDataSourceModel struct {
	ID               types.String             `tfsdk:"id"`
	OrganizationUuid types.String             `tfsdk:"organization_uuid"`
	Groups           []organizationGroupModel `tfsdk:"groups"`
}

func (d *organizationGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_groups"
}

func (d *organizationGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lightdash organization groups data source",
		Description:         "Lightdash organization groups data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash organization UUID",
				Required:            true,
			},
			"groups": schema.ListNestedAttribute{
				Description: "List of organization groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group_uuid": schema.StringAttribute{
							MarkdownDescription: "Lightdash group UUID",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Lightdash group name",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Timestamp when the group was created",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *organizationGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state organizationGroupsDataSourceModel

	// Get information of the organization
	organization, err := d.client.GetMyOrganizationV1()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get organization",
			err.Error(),
		)
		return
	}

	// Get all groups in the organization
	groups, err := d.client.GetOrganizationGroupsV1()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get organization groups",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, group := range groups {
		fetchedGroup := organizationGroupModel{
			GroupUuid: types.StringValue(group.GroupUUID),
			Name:      types.StringValue(group.Name),
			CreatedAt: types.StringValue(group.CreatedAt),
		}
		state.Groups = append(state.Groups, fetchedGroup)
	}
	state.OrganizationUuid = types.StringValue(organization.OrganizationUUID)

	// Set resource ID
	state.ID = types.StringValue(fmt.Sprintf("organizations/%s/groups", organization.OrganizationUUID))

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
