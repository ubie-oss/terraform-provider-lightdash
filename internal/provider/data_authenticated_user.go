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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &authenticatedUserDataSource{}
	_ datasource.DataSourceWithConfigure = &authenticatedUserDataSource{}
)

func NewAuthenticatedUserDataSource() datasource.DataSource {
	return &authenticatedUserDataSource{}
}

// lightdashProjectDataSource defines the data source implementation.
type authenticatedUserDataSource struct {
	client *api.Client
}

// LightdashProjectDataSourceModel describes the data source data model.
type authenticatedUserDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationUuid types.String `tfsdk:"organization_uuid"`
	UserUuid         types.String `tfsdk:"user_uuid"`
}

func (d *authenticatedUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authenticated_user"
}

func (d *authenticatedUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_authenticated_user.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Authenticated data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash organization UUID",
				Computed:            true,
			},
			"user_uuid": schema.StringAttribute{
				MarkdownDescription: "Lightdash authenticated user UUID",
				Computed:            true,
			},
		},
	}
}

func (d *authenticatedUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *authenticatedUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state authenticatedUserDataSourceModel

	authenticatedUser, err := d.client.GetAuthenticatedUserV1()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get Lightdash authenticated user",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Authenticated user UUID: %s", authenticatedUser.UserUUID))

	// Map response body to model
	state.OrganizationUuid = types.StringValue(authenticatedUser.OrganizationUUID)
	state.UserUuid = types.StringValue(authenticatedUser.UserUUID)

	// Set resource ID
	state.ID = types.StringValue(fmt.Sprintf("organizations/%s/authenticated-users/%s", authenticatedUser.OrganizationUUID, authenticatedUser.UserUUID))

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
