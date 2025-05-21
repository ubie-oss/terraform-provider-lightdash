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
	"slices"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &organizationMembersByEmailsDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationMembersByEmailsDataSource{}
)

func NewOrganizationMembersByEmailsDataSource() datasource.DataSource {
	return &organizationMembersByEmailsDataSource{}
}

// lightdashOrganizationMemberDataSource defines the data source implementation.
type organizationMembersByEmailsDataSource struct {
	client *api.Client
}

// lightdashOrganizationMemberDataSourceModel describes the data source data model.
type organizationMembersByEmailsDataSourceModel struct {
	ID               types.String              `tfsdk:"id"`
	OrganizationUuid types.String              `tfsdk:"organization_uuid"`
	Emails           types.List                `tfsdk:"emails"`
	Members          []organizationMemberModel `tfsdk:"members"`
}

func (d *organizationMembersByEmailsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_members_by_emails"
}

func (d *organizationMembersByEmailsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches Lightdash organization members filtered by a list of emails.",
		Description:         "Fetches Lightdash organization members filtered by a list of emails.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the data source, computed as `organizations/<organization_uuid>/users`.",
				Computed:            true,
			},
			"emails": schema.ListAttribute{
				MarkdownDescription: "A list of email addresses to filter the organization members by. Only members with an email in this list will be returned.",
				Required:            true,
				ElementType:         types.StringType,
				Sensitive:           true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the organization the members belong to.",
				Computed:            true,
			},
			"members": schema.ListNestedAttribute{
				Description: "A list of organization members matching the provided emails, sorted by user UUID.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_uuid": schema.StringAttribute{
							MarkdownDescription: "The unique identifier of the Lightdash user.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "The email address of the Lightdash user.",
							Computed:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "The organization role of the Lightdash user (e.g., `viewer`, `editor`, `admin`).",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *organizationMembersByEmailsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationMembersByEmailsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state organizationMembersByEmailsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

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
	members, err := service.GetOrganizationMembersByCache()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get organization member",
			err.Error(),
		)
		return
	}

	// log the number of members
	tflog.Info(ctx, fmt.Sprintf("(organization_members_by_emails) Fetched organization members: %d", len(members)))

	// Convert types.List of emails to Go slice of strings
	var emailList []string
	diags := state.Emails.ElementsAs(ctx, &emailList, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to model
	newMembers := []organizationMemberModel{}
	for _, member := range members {
		// Check if the email is in the list of emails
		if !slices.Contains(emailList, member.Email) {
			tflog.Debug(ctx, fmt.Sprintf("(organization_members_by_emails) Skipping member %s because it is not in the list of emails", member.Email))
			continue
		}

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
	tflog.Info(ctx, fmt.Sprintf("(organization_members_by_emails) Updated organization members: %d", len(newMembers)))

	// Set state
	state.OrganizationUuid = types.StringValue(organization.OrganizationUUID)
	state.Members = newMembers

	// Set resource ID
	state.ID = types.StringValue(fmt.Sprintf("organizations/%s/users", organization.OrganizationUUID))

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
