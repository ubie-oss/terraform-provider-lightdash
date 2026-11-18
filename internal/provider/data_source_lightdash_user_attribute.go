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

var (
	_ datasource.DataSource              = &userAttributeDataSource{}
	_ datasource.DataSourceWithConfigure = &userAttributeDataSource{}
)

func NewUserAttributeDataSource() datasource.DataSource {
	return &userAttributeDataSource{}
}

type userAttributeDataSource struct {
	service *services.UserAttributesService
}

func (d *userAttributeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_attribute"
}

func (d *userAttributeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_lightdash_user_attribute.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Reads a Lightdash user attribute by name",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The data source identifier. It is computed as `organizations/<organization_uuid>/user-attributes/<user_attribute_uuid>`.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user attribute to look up.",
				Required:            true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization that the user attribute belongs to.",
				Computed:            true,
			},
			"user_attribute_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash user attribute.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the user attribute.",
				Computed:            true,
			},
			"attribute_defaults": schema.SetAttribute{
				MarkdownDescription: "The default values of the user attribute. It is null when there are none.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"groups": schema.SetNestedAttribute{
				MarkdownDescription: "The values of the user attribute assigned to groups.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group_uuid": schema.StringAttribute{
							MarkdownDescription: "The UUID of the Lightdash group.",
							Computed:            true,
						},
						"values": schema.SetAttribute{
							MarkdownDescription: "The values of the user attribute for the group.",
							ElementType:         types.StringType,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *userAttributeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.service = services.NewUserAttributesService(client)
}

func (d *userAttributeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config userAttributeResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attribute, err := d.service.GetByName(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Lightdash user attribute",
			fmt.Sprintf("Could not read the user attribute named %q: %s", config.Name.ValueString(), err.Error()),
		)
		return
	}

	state, diags := userAttributeToResourceModel(ctx, *attribute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
