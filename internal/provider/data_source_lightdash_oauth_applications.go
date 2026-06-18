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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

var (
	_ datasource.DataSource              = &oauthApplicationsDataSource{}
	_ datasource.DataSourceWithConfigure = &oauthApplicationsDataSource{}
)

func NewOAuthApplicationsDataSource() datasource.DataSource {
	return &oauthApplicationsDataSource{}
}

type oauthApplicationsDataSource struct {
	client *api.Client
}

type oauthApplicationListItemModel struct {
	ClientID          types.String `tfsdk:"client_id"`
	ClientName        types.String `tfsdk:"client_name"`
	RedirectURIs      types.List   `tfsdk:"redirect_uris"`
	Scopes            types.List   `tfsdk:"scopes"`
	CreatedAt         types.String `tfsdk:"created_at"`
	CreatedByUserUUID types.String `tfsdk:"created_by_user_uuid"`
}

type oauthApplicationsDataSourceModel struct {
	ID               types.String                    `tfsdk:"id"`
	OrganizationUUID types.String                    `tfsdk:"organization_uuid"`
	Applications     []oauthApplicationListItemModel `tfsdk:"applications"`
}

func (d *oauthApplicationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_applications"
}

func (d *oauthApplicationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_lightdash_oauth_applications.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Lightdash OAuth applications data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The data source identifier. It is computed as `organizations/<organization_uuid>/oauth_applications`.",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
				Required:            true,
			},
			"applications": schema.ListNestedAttribute{
				MarkdownDescription: "OAuth applications registered in the organization.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"client_id": schema.StringAttribute{
							MarkdownDescription: "The OAuth client ID.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *oauthApplicationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *oauthApplicationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state oauthApplicationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := services.NewOAuthApplicationsService(d.client)
	clients, err := service.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list OAuth applications", err.Error())
		return
	}

	applications := make([]oauthApplicationListItemModel, 0, len(clients))
	for _, client := range clients {
		item, diags := oauthClientV1ToListItemModel(ctx, client)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		applications = append(applications, item)
	}

	sort.Slice(applications, func(i, j int) bool {
		return applications[i].ClientID.ValueString() < applications[j].ClientID.ValueString()
	})

	state.ID = types.StringValue(fmt.Sprintf("organizations/%s/oauth_applications", state.OrganizationUUID.ValueString()))
	state.Applications = applications

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func oauthClientV1ToListItemModel(ctx context.Context, client apiv1.OAuthClientV1) (oauthApplicationListItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	redirectURIs, redirectDiags := stringSliceToStringList(ctx, client.RedirectURIs)
	diags.Append(redirectDiags...)

	scopes, scopesDiags := stringSliceToStringList(ctx, client.Scopes)
	diags.Append(scopesDiags...)

	createdBy := types.StringNull()
	if client.CreatedByUserUUID != nil {
		createdBy = types.StringValue(*client.CreatedByUserUUID)
	}

	return oauthApplicationListItemModel{
		ClientID:          types.StringValue(client.ClientID),
		ClientName:        types.StringValue(client.ClientName),
		RedirectURIs:      redirectURIs,
		Scopes:            scopes,
		CreatedAt:         types.StringValue(client.CreatedAt),
		CreatedByUserUUID: createdBy,
	}, diags
}
