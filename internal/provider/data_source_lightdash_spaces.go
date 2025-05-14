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
	_ datasource.DataSource              = &spacesDataSource{}
	_ datasource.DataSourceWithConfigure = &spacesDataSource{}
)

func NewSpacesDataSource() datasource.DataSource {
	return &spacesDataSource{}
}

// projectDataSource defines the data source implementation.
type spacesDataSource struct {
	client *api.Client
}

type spaceModel struct {
	ParentSpaceUUID types.String `tfsdk:"parent_space_uuid"`
	SpaceUUID       types.String `tfsdk:"space_uuid"`
	SpaceName       types.String `tfsdk:"name"`
	IsPrivate       types.Bool   `tfsdk:"is_private"`
}

// projectDataSourceModel describes the data source data model.
type spacesDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationUUID types.String `tfsdk:"organization_uuid"`
	ProjectUUID      types.String `tfsdk:"project_uuid"`
	Spaces           []spaceModel `tfsdk:"spaces"`
}

func (d *spacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spaces"
}

func (d *spacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lightdash spaces data source",
		Description:         "Lightdash spaces data source",
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
				Description: "Organization UUID of the Lightdash project.",
				Required:    true,
			},
			"spaces": schema.ListNestedAttribute{
				Description: "List of spaces.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"parent_space_uuid": schema.StringAttribute{
							Description: "Parent space UUID of the Lightdash space. This attribute is nullable and will be empty if the space has no parent.",
							Computed:    true,
							Optional:    true,
						},
						"space_uuid": schema.StringAttribute{
							Description: "Space UUID of the Lightdash space.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the Lightdash space.",
							Computed:    true,
						},
						"is_private": schema.BoolAttribute{
							Description: "Is the Lightdash space private.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *spacesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *spacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state spacesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project_uuid := state.ProjectUUID.ValueString()
	spaces, err := d.client.ListSpacesInProjectV1(project_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lightdash spaces",
			err.Error(),
		)
		return
	}

	// Map response body to model
	updatedSpaces := []spaceModel{}
	for _, space := range spaces {
		// Even if the parent space UUID is empty, we want to set it to null
		parentSpace := types.StringNull()
		if space.ParentSpaceUUID != nil {
			parentSpace = types.StringValue(*space.ParentSpaceUUID)
		}

		spaceState := spaceModel{
			ParentSpaceUUID: parentSpace,
			SpaceUUID:       types.StringValue(space.SpaceUUID),
			SpaceName:       types.StringValue(space.SpaceName),
			IsPrivate:       types.BoolValue(space.IsPrivate),
		}
		updatedSpaces = append(updatedSpaces, spaceState)
	}
	// Sort the spaces by space UUID
	sort.Slice(updatedSpaces, func(i, j int) bool {
		return updatedSpaces[i].SpaceUUID.ValueString() < updatedSpaces[j].SpaceUUID.ValueString()
	})
	state.Spaces = updatedSpaces

	// Set resource ID
	state_id := fmt.Sprintf("organizations/%s/projects/%s/spaces",
		state.OrganizationUUID.ValueString(), state.ProjectUUID.ValueString())
	state.ID = types.StringValue(state_id)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
