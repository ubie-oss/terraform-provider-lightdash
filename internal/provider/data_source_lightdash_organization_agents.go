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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource = &organizationAgentsDataSource{}
)

func NewOrganizationAgentsDataSource() datasource.DataSource {
	return &organizationAgentsDataSource{}
}

// organizationAgentsDataSource defines the data source implementation.
type organizationAgentsDataSource struct {
	client *api.Client
}

// organizationAgentsDataSourceModel describes the data source data model.
type organizationAgentsDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationUUID types.String `tfsdk:"organization_uuid"`
	Agents           []agentModel `tfsdk:"agents"`
}

// agentModel describes the data source data model for a Lightdash agent.
type agentModel struct {
	AgentUUID        types.String `tfsdk:"agent_uuid"`
	OrganizationUUID types.String `tfsdk:"organization_uuid"`
	ProjectUUID      types.String `tfsdk:"project_uuid"`
	Name             types.String `tfsdk:"name"`
	Tags             types.List   `tfsdk:"tags"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	CreatedAt        types.String `tfsdk:"created_at"`
	ImageURL         types.String `tfsdk:"image_url"`
	EnableDataAccess types.Bool   `tfsdk:"enable_data_access"`
	GroupAccess      types.List   `tfsdk:"group_access"`
	UserAccess       types.List   `tfsdk:"user_access"`
}

func (d *organizationAgentsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_agents"
}

func (d *organizationAgentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Lightdash organization agents data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source ID.",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "Organization UUID.",
				Required:            true,
			},
			"agents": schema.ListNestedAttribute{
				MarkdownDescription: "List of agents.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"agent_uuid": schema.StringAttribute{
							MarkdownDescription: "Agent UUID.",
							Computed:            true,
						},
						"organization_uuid": schema.StringAttribute{
							MarkdownDescription: "Organization UUID.",
							Computed:            true,
						},
						"project_uuid": schema.StringAttribute{
							MarkdownDescription: "Project UUID.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Agent name.",
							Computed:            true,
						},
						"tags": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "Tags associated with the agent.",
							Optional:            true,
							Computed:            true,
						},
						"updated_at": schema.StringAttribute{
							MarkdownDescription: "Timestamp of the last update.",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Timestamp of creation.",
							Computed:            true,
						},
						"image_url": schema.StringAttribute{
							MarkdownDescription: "URL for the agent's icon/image.",
							Optional:            true,
							Computed:            true,
						},
						"enable_data_access": schema.BoolAttribute{
							MarkdownDescription: "Whether the agent can access underlying project data.",
							Optional:            true,
							Computed:            true,
						},
						"group_access": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "UUIDs of user groups with access.",
							Optional:            true,
							Computed:            true,
						},
						"user_access": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "UUIDs of individual users with access.",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *organizationAgentsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationAgentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data organizationAgentsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agentService := services.NewAgentService(d.client)
	agents, err := agentService.GetAllAgents(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read agents, got error: %s", err))
		return
	}

	// Map agents data to model
	for _, agent := range agents {
		// Convert tags slice to Terraform List
		tagsList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, agent.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Convert group access slice to Terraform List
		groupAccessList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, agent.GroupAccess)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Convert user access slice to Terraform List
		userAccessList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, agent.UserAccess)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Handle nullable ImageURL
		imageURL := types.StringNull()
		if agent.ImageURL != nil {
			imageURL = types.StringValue(*agent.ImageURL)
		}

		agentState := agentModel{
			AgentUUID:        types.StringValue(agent.AgentUUID),
			OrganizationUUID: types.StringValue(agent.OrganizationUUID),
			ProjectUUID:      types.StringValue(agent.ProjectUUID),
			Name:             types.StringValue(agent.Name),
			Tags:             tagsList,
			UpdatedAt:        types.StringValue(agent.UpdatedAt),
			CreatedAt:        types.StringValue(agent.CreatedAt),
			ImageURL:         imageURL,
			EnableDataAccess: types.BoolValue(agent.EnableDataAccess),
			GroupAccess:      groupAccessList,
			UserAccess:       userAccessList,
		}
		data.Agents = append(data.Agents, agentState)
	}

	data.ID = types.StringValue(data.OrganizationUUID.ValueString())

	tflog.Trace(ctx, "read a data source resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
