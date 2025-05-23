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
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/controllers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &spaceDataSource{}
	_ datasource.DataSourceWithConfigure = &spaceDataSource{}
)

func NewSpaceDataSource() datasource.DataSource {
	return &spaceDataSource{}
}

// spaceDataSource defines the data source implementation.
type spaceDataSource struct {
	client *api.Client
}

type spaceAccessMember struct {
	UserUUID        types.String `tfsdk:"user_uuid"`
	SpaceRole       types.String `tfsdk:"space_role"`
	HasDirectAccess types.Bool   `tfsdk:"has_direct_access"`
	InheritedRole   types.String `tfsdk:"inherited_role"`
	InheritedFrom   types.String `tfsdk:"inherited_from"`
	ProjectRole     types.String `tfsdk:"project_role"`
}

type spaceAccessGroup struct {
	GroupUUID types.String `tfsdk:"group_uuid"`
	SpaceRole types.String `tfsdk:"space_role"`
}

type spaceDetailedModel struct {
	ID               types.String        `tfsdk:"id"`
	OrganizationUUID types.String        `tfsdk:"organization_uuid"`
	ProjectUUID      types.String        `tfsdk:"project_uuid"`
	ParentSpaceUUID  types.String        `tfsdk:"parent_space_uuid"`
	SpaceUUID        types.String        `tfsdk:"space_uuid"`
	SpaceName        types.String        `tfsdk:"name"`
	IsPrivate        types.Bool          `tfsdk:"is_private"`
	Access           []spaceAccessMember `tfsdk:"access"`
	AccessGroups     []spaceAccessGroup  `tfsdk:"access_groups"`
}

func (d *spaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (d *spaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lightdash space data source",
		Description:         "Lightdash space data source",
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
			"space_uuid": schema.StringAttribute{
				Description: "Space UUID of the Lightdash space.",
				Required:    true,
			},
			"parent_space_uuid": schema.StringAttribute{
				Description: "Parent space UUID of the Lightdash space. This attribute is nullable and will be empty if the space has no parent.",
				Computed:    true,
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the Lightdash space.",
				Computed:    true,
			},
			"is_private": schema.BoolAttribute{
				Description: "Is the Lightdash space private.",
				Computed:    true,
			},
			"access": schema.ListNestedAttribute{
				Description: "List of members of the Lightdash space.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_uuid": schema.StringAttribute{
							Description: "User UUID of the Lightdash space member.",
							Computed:    true,
						},
						"space_role": schema.StringAttribute{
							Description: "Role of the Lightdash space member.",
							Computed:    true,
						},
						"has_direct_access": schema.BoolAttribute{
							Description: "Whether the user has direct access to the space.",
							Computed:    true,
						},
						"inherited_role": schema.StringAttribute{
							Description: "Inherited role of the Lightdash space member.",
							Computed:    true,
						},
						"inherited_from": schema.StringAttribute{
							Description: "Inherited from of the Lightdash space member.",
							Computed:    true,
						},
						"project_role": schema.StringAttribute{
							Description: "Project role of the Lightdash space member.",
							Computed:    true,
						},
					},
				},
			},
			"access_groups": schema.ListNestedAttribute{
				Description: "List of groups of the Lightdash space.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group_uuid": schema.StringAttribute{
							Description: "Group UUID of the Lightdash space group.",
							Computed:    true,
						},
						"space_role": schema.StringAttribute{
							Description: "Role of the Lightdash space group.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *spaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *spaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state spaceDetailedModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the space controller
	spaceController := controllers.NewSpaceController(d.client)

	projectUuid := state.ProjectUUID.ValueString()
	spaceUuid := state.SpaceUUID.ValueString()

	space, err := spaceController.GetSpace(ctx, projectUuid, spaceUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lightdash space",
			fmt.Sprintf("Unable to read space with Project UUID %q and Space UUID %q: %s", projectUuid, spaceUuid, err.Error()),
		)
		return
	}

	// Map response to state model
	state.ID = types.StringValue(fmt.Sprintf("organizations/%s/projects/%s/space/%s",
		state.OrganizationUUID.ValueString(), space.ProjectUUID, space.SpaceUUID))
	if space.ParentSpaceUUID != nil {
		state.ParentSpaceUUID = types.StringValue(*space.ParentSpaceUUID)
	} else {
		state.ParentSpaceUUID = types.StringNull()
	}

	state.ProjectUUID = types.StringValue(space.ProjectUUID)
	state.SpaceUUID = types.StringValue(space.SpaceUUID)
	state.SpaceName = types.StringValue(space.SpaceName)
	state.IsPrivate = types.BoolValue(space.IsPrivate)

	// Map SpaceAccessMembers
	accessList := make([]spaceAccessMember, len(space.SpaceAccessMembers))
	for i, access := range space.SpaceAccessMembers {
		accessList[i] = spaceAccessMember{
			UserUUID:        types.StringValue(access.UserUUID),
			SpaceRole:       types.StringValue(string(access.SpaceRole)),
			HasDirectAccess: types.BoolValue(*access.HasDirectAccess),
			InheritedRole:   types.StringValue(*access.InheritedRole),
			InheritedFrom:   types.StringValue(*access.InheritedFrom),
			ProjectRole:     types.StringValue(*access.ProjectRole),
		}
	}
	state.Access = accessList

	// Map SpaceAccessGroups
	accessGroupsList := make([]spaceAccessGroup, len(space.SpaceAccessGroups))
	for i, access := range space.SpaceAccessGroups {
		accessGroupsList[i] = spaceAccessGroup{
			GroupUUID: types.StringValue(access.GroupUUID),
			SpaceRole: types.StringValue(string(access.SpaceRole)),
		}
	}
	state.AccessGroups = accessGroupsList

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
