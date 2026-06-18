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

	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

var (
	_ datasource.DataSource              = &oauthApplicationDataSource{}
	_ datasource.DataSourceWithConfigure = &oauthApplicationDataSource{}
)

func NewOAuthApplicationDataSource() datasource.DataSource {
	return &oauthApplicationDataSource{}
}

type oauthApplicationDataSource struct {
	client *api.Client
}

type oauthApplicationDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	OrganizationUUID  types.String `tfsdk:"organization_uuid"`
	ClientID          types.String `tfsdk:"client_id"`
	ClientName        types.String `tfsdk:"client_name"`
	RedirectURIs      types.List   `tfsdk:"redirect_uris"`
	Scopes            types.List   `tfsdk:"scopes"`
	CreatedAt         types.String `tfsdk:"created_at"`
	CreatedByUserUUID types.String `tfsdk:"created_by_user_uuid"`
}

func (d *oauthApplicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_application"
}

func (d *oauthApplicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_lightdash_oauth_application.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Lightdash OAuth application data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The data source identifier. It is computed as `organizations/<organization_uuid>/oauth_applications/<client_id>`.",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization for the authenticated token.",
				Computed:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The OAuth client ID to look up.",
				Required:            true,
			},
			"client_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the OAuth application.",
				Computed:            true,
			},
			"redirect_uris": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Allowed OAuth redirect URIs.",
				Computed:            true,
			},
			"scopes": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "OAuth scopes stored on the client.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "ISO 8601 timestamp when the client was created.",
				Computed:            true,
			},
			"created_by_user_uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the user who created the client, if known.",
				Computed:            true,
			},
		},
	}
}

func (d *oauthApplicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *oauthApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state oauthApplicationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := state.ClientID.ValueString()

	organization, err := apiv1.GetMyOrganizationV1(d.client)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read organization", err.Error())
		return
	}

	service := services.NewOAuthApplicationsService(d.client)
	client, err := service.GetByClientID(ctx, clientID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read OAuth application", err.Error())
		return
	}

	item, diags := oauthClientV1ToListItemModel(ctx, *client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.OrganizationUUID = types.StringValue(organization.OrganizationUUID)
	state.ID = types.StringValue(getOAuthApplicationResourceID(organization.OrganizationUUID, clientID))
	state.ClientName = item.ClientName
	state.RedirectURIs = item.RedirectURIs
	state.Scopes = item.Scopes
	state.CreatedAt = item.CreatedAt
	state.CreatedByUserUUID = item.CreatedByUserUUID

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
