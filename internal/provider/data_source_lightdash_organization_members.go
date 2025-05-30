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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &organizationMembersDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationMembersDataSource{}
)

func NewOrganizationMembersDataSource() datasource.DataSource {
	return &organizationMembersDataSource{}
}

// lightdashOrganizationMemberDataSource defines the data source implementation.
type organizationMembersDataSource struct {
	client *api.Client
}

// organizationMemberDataSourceModel describes the data source data model.
type organizationMemberModel struct {
	UserUuid         types.String                  `tfsdk:"user_uuid"`
	Email            types.String                  `tfsdk:"email"`
	OrganizationRole models.OrganizationMemberRole `tfsdk:"role"`
}

// lightdashOrganizationMemberDataSourceModel describes the data source data model.
type organizationMembersDataSourceModel struct {
	ID               types.String              `tfsdk:"id"`
	OrganizationUuid types.String              `tfsdk:"organization_uuid"`
	Members          []organizationMemberModel `tfsdk:"members"`
}

func (d *organizationMembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_members"
}

func (d *organizationMembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_lightdash_organization_members.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Lightdash organization members data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The data source identifier. It is computed as `organizations/<organization_uuid>/users`.",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
				Computed:            true,
			},
			"members": schema.ListNestedAttribute{
				MarkdownDescription: "A list of organization members.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_uuid": schema.StringAttribute{
							MarkdownDescription: "The UUID of the Lightdash user.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "The email address of the Lightdash user.",
							Computed:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "The organization role of the user.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *organizationMembersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state organizationMembersDataSourceModel

	// Get information of the organization
	organization, err := d.client.GetMyOrganizationV1()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get organization",
			err.Error(),
		)
		return
	}

	// Get all members in the organization
	service := services.GetOrganizationMembersService(d.client)
	members, err := service.GetOrganizationMembers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get organization member",
			err.Error(),
		)
		return
	}

	// log the number of members
	tflog.Info(ctx, fmt.Sprintf("Fetched organization members: %d", len(members)))

	// Map response body to model
	newMembers := []organizationMemberModel{}
	for _, member := range members {
		fetchedMember := organizationMemberModel{
			UserUuid:         types.StringValue(member.UserUUID),
			Email:            types.StringValue(member.Email),
			OrganizationRole: member.OrganizationRole,
		}
		newMembers = append(newMembers, fetchedMember)
	}

	// Sort the members by user UUID
	sort.Slice(newMembers, func(i, j int) bool {
		return newMembers[i].UserUuid.ValueString() < newMembers[j].UserUuid.ValueString()
	})

	// log the number of new members
	tflog.Info(ctx, fmt.Sprintf("Updated organization members: %d", len(newMembers)))

	// Set state
	state.OrganizationUuid = types.StringValue(organization.OrganizationUUID)
	state.Members = newMembers

	// Set resource ID
	state.ID = types.StringValue(fmt.Sprintf("organizations/%s/users", organization.OrganizationUUID))

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
