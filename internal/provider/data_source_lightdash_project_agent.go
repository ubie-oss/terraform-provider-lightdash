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
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &projectAgentDataSource{}
	_ datasource.DataSourceWithConfigure = &projectAgentDataSource{}
)

func NewProjectAgentDataSource() datasource.DataSource {
	return &projectAgentDataSource{}
}

// projectAgentDataSource defines the data source implementation.
type projectAgentDataSource struct {
	client *api.Client
}

// projectAgentDataSourceModel describes the data source data model.
type projectAgentDataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	OrganizationUUID      types.String `tfsdk:"organization_uuid"`
	ProjectUUID           types.String `tfsdk:"project_uuid"`
	AgentUUID             types.String `tfsdk:"agent_uuid"`
	Name                  types.String `tfsdk:"name"`
	Instruction           types.String `tfsdk:"instruction"`
	Tags                  types.List   `tfsdk:"tags"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
	CreatedAt             types.String `tfsdk:"created_at"`
	ImageURL              types.String `tfsdk:"image_url"`
	EnableDataAccess      types.Bool   `tfsdk:"enable_data_access"`
	EnableSelfImprovement types.Bool   `tfsdk:"enable_self_improvement"`
	GroupAccess           types.List   `tfsdk:"group_access"`
	UserAccess            types.List   `tfsdk:"user_access"`
	Version               types.Int64  `tfsdk:"version"`
}

func (d *projectAgentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_agent"
}

func (d *projectAgentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_lightdash_project_agent.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Lightdash project agent data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The data source identifier. It is computed as `organizations/<organization_uuid>/projects/<project_uuid>/agents/<agent_uuid>`.",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
				Computed:            true,
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash project.",
				Required:            true,
			},
			"agent_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash agent.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Lightdash agent.",
				Computed:            true,
			},
			"instruction": schema.StringAttribute{
				MarkdownDescription: "Custom instruction (system prompt) for the agent.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Tags associated with the agent.",
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
				Computed:            true,
			},
			"enable_data_access": schema.BoolAttribute{
				MarkdownDescription: "Whether the agent can access underlying project data.",
				Computed:            true,
			},
			"enable_self_improvement": schema.BoolAttribute{
				MarkdownDescription: "Whether the agent can improve itself based on user interactions.",
				Computed:            true,
			},
			"group_access": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "UUIDs of user groups with access.",
				Computed:            true,
			},
			"user_access": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "UUIDs of individual users with access.",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Version of the agent.",
				Computed:            true,
			},
		},
	}
}

func (d *projectAgentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectAgentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectAgentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectUUID := config.ProjectUUID.ValueString()
	agentUUID := config.AgentUUID.ValueString()

	agent, err := d.client.GetAgentV1(projectUUID, agentUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lightdash project agent",
			err.Error(),
		)
		return
	}

	// Map response body to model
	config.OrganizationUUID = types.StringValue(agent.OrganizationUUID)
	config.ProjectUUID = types.StringValue(agent.ProjectUUID)
	config.AgentUUID = types.StringValue(agent.UUID)
	config.Name = types.StringValue(agent.Name)
	config.Version = types.Int64Value(agent.Version)

	// Handle optional instruction
	if agent.Instruction != nil {
		config.Instruction = types.StringValue(*agent.Instruction)
	} else {
		config.Instruction = types.StringNull()
	}

	// Handle tags
	tags, diags := types.ListValueFrom(ctx, types.StringType, agent.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Tags = tags

	config.UpdatedAt = types.StringValue(agent.UpdatedAt)
	config.CreatedAt = types.StringValue(agent.CreatedAt)

	// Handle optional image URL
	if agent.ImageURL != nil {
		config.ImageURL = types.StringValue(*agent.ImageURL)
	} else {
		config.ImageURL = types.StringNull()
	}

	config.EnableDataAccess = types.BoolValue(agent.EnableDataAccess)
	config.EnableSelfImprovement = types.BoolValue(agent.EnableSelfImprovement)

	// Handle group access
	groupAccess, diags := types.ListValueFrom(ctx, types.StringType, agent.GroupAccess)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.GroupAccess = groupAccess

	// Handle user access
	userAccess, diags := types.ListValueFrom(ctx, types.StringType, agent.UserAccess)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.UserAccess = userAccess

	// Set resource ID
	stateID := fmt.Sprintf("organizations/%s/projects/%s/agents/%s", agent.OrganizationUUID, agent.ProjectUUID, agent.UUID)
	config.ID = types.StringValue(stateID)

	// Set state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
