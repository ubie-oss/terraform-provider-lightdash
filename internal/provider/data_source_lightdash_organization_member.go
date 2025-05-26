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
	_ datasource.DataSource              = &organizationMemberDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationMemberDataSource{}
)

func NewOrganizationMemberDataSource() datasource.DataSource {
	return &organizationMemberDataSource{}
}

// organizationMemberDataSourceModel describes the data source data model.
type organizationMemberDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationUUID types.String `tfsdk:"organization_uuid"`
	UserUUID         types.String `tfsdk:"user_uuid"`
	Email            types.String `tfsdk:"email"`
	OrganizationRole types.String `tfsdk:"role"`
}

// organizationMemberDataSource defines the data source implementation.
type organizationMemberDataSource struct {
	client *api.Client
}

func (d *organizationMemberDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_member"
}

func (d *organizationMemberDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_lightdash_organization_member.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Lightdash organization member data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The data source identifier. It is computed as `organizations/<organization_uuid>/users/<user_uuid>`.",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
				Computed:            true,
			},
			"user_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash user.",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address of the Lightdash user.",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The organization role of the user.",
				Computed:            true,
			},
		},
	}
}

func (d *organizationMemberDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationMemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state organizationMemberDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all members in the organization
	email := state.Email.ValueString()
	service := services.GetOrganizationMembersService(d.client)
	member, err := service.GetOrganizationMemberByEmail(email)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get organization member",
			err.Error(),
		)
		return
	}

	// Map response body to model
	state.OrganizationUUID = types.StringValue(member.OrganizationUUID)
	state.UserUUID = types.StringValue(member.UserUUID)
	state.Email = types.StringValue(member.Email)
	state.OrganizationRole = types.StringValue(member.OrganizationRole.String())

	// Set resource ID
	state.ID = types.StringValue(fmt.Sprintf("organizations/%s/users/%s",
		member.OrganizationUUID, member.UserUUID))

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
