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
	_ datasource.DataSource              = &groupMembersDataSource{}
	_ datasource.DataSourceWithConfigure = &groupMembersDataSource{}
)

func NewGroupMembersDataSource() datasource.DataSource {
	return &groupMembersDataSource{}
}

// projectGroupsDataSource defines the data source implementation.
type groupMembersDataSource struct {
	client *api.Client
}

type groupMemberModelForGroupMembers struct {
	// NOTE Those aren't exposed, as they are sensitive data.
	// LastName  types.String `tfsdk:"last_name"`
	// FirstName types.String `tfsdk:"first_name"`
	// Email     types.String `tfsdk:"email"`
	UserUUID types.String `tfsdk:"user_uuid"`
}

// groupMembersDataSourceModel describes the data source data model.
type groupMembersDataSourceModel struct {
	ID               types.String                      `tfsdk:"id"`
	OrganizationUUID types.String                      `tfsdk:"organization_uuid"`
	ProjectUUID      types.String                      `tfsdk:"project_uuid"`
	GroupUUID        types.String                      `tfsdk:"group_uuid"`
	Members          []groupMemberModelForGroupMembers `tfsdk:"members"`
}

func (d *groupMembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_members"
}

func (d *groupMembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lightdash group members data source",
		Description:         "Data source for Lightdash group members.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				Description: "Organization UUID of the Lightdash group.",
				Required:    true,
			},
			"project_uuid": schema.StringAttribute{
				Description: "UUID of the Lightdash project.",
				Required:    true,
			},
			"group_uuid": schema.StringAttribute{
				Description: "Group UUID of the Lightdash group.",
				Required:    true,
			},
			"members": schema.ListNestedAttribute{
				Description: "List of group members.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_uuid": schema.StringAttribute{
							Description: "User UUID of the Lightdash group member.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *groupMembersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *groupMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupMembersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group_uuid := state.GroupUUID.ValueString()
	members, err := d.client.GetGroupMembersV1(group_uuid)
	updatedMembers := []groupMemberModelForGroupMembers{}
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Lightdash group members for group UUID: "+group_uuid,
			"Error: "+err.Error(),
		)
		return
	}
	for _, member := range members {
		member := groupMemberModelForGroupMembers{
			UserUUID: types.StringValue(member.UserUUID),
		}
		updatedMembers = append(updatedMembers, member)
	}
	// Sort the members by user UUID
	sort.Slice(updatedMembers, func(i, j int) bool {
		return updatedMembers[i].UserUUID.ValueString() < updatedMembers[j].UserUUID.ValueString()
	})
	state.Members = updatedMembers

	// Set resource ID
	state_id := fmt.Sprintf("organizations/%s/groups/%s/members",
		state.OrganizationUUID.ValueString(), state.GroupUUID.ValueString())
	state.ID = types.StringValue(state_id)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
