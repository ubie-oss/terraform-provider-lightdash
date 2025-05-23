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
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &projectMembersDataSource{}
	_ datasource.DataSourceWithConfigure = &projectMembersDataSource{}
)

func NewProjectMembersDataSource() datasource.DataSource {
	return &projectMembersDataSource{}
}

// projectDataSource defines the data source implementation.
type projectMembersDataSource struct {
	client *api.Client
}

type projectMemberModel struct {
	UserUUID    types.String             `tfsdk:"user_uuid"`
	Email       types.String             `tfsdk:"email"`
	ProjectRole models.ProjectMemberRole `tfsdk:"role"`
}

// projectDataSourceModel describes the data source data model.
type projectMembersDataSourceModel struct {
	ID          types.String         `tfsdk:"id"`
	ProjectUUID types.String         `tfsdk:"project_uuid"`
	Members     []projectMemberModel `tfsdk:"members"`
}

func (d *projectMembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_members"
}

func (d *projectMembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_lightdash_project_members.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Lightdash project member data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"project_uuid": schema.StringAttribute{
				Description: "Project UUID.",
				Required:    true,
			},
			"members": schema.ListNestedAttribute{
				Description: "List of members.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_uuid": schema.StringAttribute{
							Description: "Lightdash user UUID.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "Lightdash user email.",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "Lightdash project role.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *projectMembersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectMembersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get project members
	project_uuid := state.ProjectUUID.ValueString()
	members, err := d.client.GetProjectAccessListV1(project_uuid)
	updatedMembers := []projectMemberModel{}
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lightdash project",
			err.Error(),
		)
		return
	}
	// Map response body to model
	for _, member := range members {
		projectState := projectMemberModel{
			UserUUID:    types.StringValue(member.UserUUID),
			Email:       types.StringValue(member.Email),
			ProjectRole: member.ProjectRole,
		}
		updatedMembers = append(updatedMembers, projectState)
	}
	state.ProjectUUID = types.StringValue(project_uuid)
	// Sort the members by user UUID
	sort.Slice(updatedMembers, func(i, j int) bool {
		return updatedMembers[i].UserUUID.ValueString() < updatedMembers[j].UserUUID.ValueString()
	})
	state.Members = updatedMembers

	// Set resource ID
	state_id := fmt.Sprintf("projects/%s/access", state.ProjectUUID.ValueString())
	state.ID = types.StringValue(state_id)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
