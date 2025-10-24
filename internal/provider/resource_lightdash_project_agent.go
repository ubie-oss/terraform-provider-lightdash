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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &projectAgentResource{}
	_ resource.ResourceWithConfigure   = &projectAgentResource{}
	_ resource.ResourceWithImportState = &projectAgentResource{}
)

func NewProjectAgentResource() resource.Resource {
	return &projectAgentResource{}
}

// projectAgentResource defines the resource implementation.
type projectAgentResource struct {
	client *api.Client
}

// projectAgentResourceModel describes the resource data model.
type projectAgentResourceModel struct {
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
	DeleteProtection      types.Bool   `tfsdk:"deletion_protection"`
	Integrations          types.List   `tfsdk:"integrations"`
	Version               types.Int64  `tfsdk:"version"`
}

type integrationObjectModel struct {
	Type      types.String `tfsdk:"type"`
	ChannelID types.String `tfsdk:"channel_id"`
}

func (r *projectAgentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_agent"
}

func (r *projectAgentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_lightdash_project_agent.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Manages a Lightdash project agent",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource identifier. It is computed as `organizations/<organization_uuid>/projects/<project_uuid>/agents/<agent_uuid>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
				Required:            true,
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash project.",
				Required:            true,
			},
			"agent_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash agent.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Lightdash agent.",
				Required:            true,
				Validators: []validator.String{
					ValidateNonEmptyString{},
				},
			},
			"instruction": schema.StringAttribute{
				MarkdownDescription: "Custom instruction (system prompt) for the agent (max 8192 chars).",
				Required:            true,
				Validators: []validator.String{
					ValidateNonEmptyString{},
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Tags associated with the agent.",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
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
				Default:             booldefault.StaticBool(true),
			},
			"enable_self_improvement": schema.BoolAttribute{
				MarkdownDescription: "Whether the agent can improve itself based on user interactions.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"group_access": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "UUIDs of user groups with access.",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"user_access": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "UUIDs of individual users with access.",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: "When set to `true`, prevents the destruction of the project agent resource by Terraform. Defaults to `false`.",
				Required:            true,
			},
			"integrations": schema.ListNestedAttribute{
				MarkdownDescription: "List of integrations for the agent.",
				Optional:            true,
				Computed:            true,
				Default: listdefault.StaticValue(types.ListValueMust(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"type":       types.StringType,
						"channel_id": types.StringType,
					},
				}, []attr.Value{})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of integration (e.g., `slack`).",
							Required:            true,
						},
						"channel_id": schema.StringAttribute{
							MarkdownDescription: "The channel ID for the integration.",
							Optional:            true,
						},
					},
				},
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "The version of the agent.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(2),
			},
		},
	}
}

func (r *projectAgentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *projectAgentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read plan data into model
	var plan projectAgentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract plan values
	projectUuid := plan.ProjectUUID.ValueString()

	// Validate required fields for creation
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing required argument",
			"The name argument is required when creating a Lightdash project agent.",
		)
		return
	}

	// Get optional values
	var imageUrl *string
	if !plan.ImageURL.IsUnknown() && !plan.ImageURL.IsNull() {
		imageUrlVal := plan.ImageURL.ValueString()
		imageUrl = &imageUrlVal
	}

	var instruction *string
	if !plan.Instruction.IsUnknown() && !plan.Instruction.IsNull() {
		instructionVal := plan.Instruction.ValueString()
		instruction = &instructionVal
	}

	// Convert tags list to slice
	var tags []string
	if !plan.Tags.IsUnknown() && !plan.Tags.IsNull() {
		diags = plan.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Convert group access list to slice
	var groupAccess []string
	if !plan.GroupAccess.IsUnknown() && !plan.GroupAccess.IsNull() {
		diags = plan.GroupAccess.ElementsAs(ctx, &groupAccess, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Convert user access list to slice
	var userAccess []string
	if !plan.UserAccess.IsUnknown() && !plan.UserAccess.IsNull() {
		diags = plan.UserAccess.ElementsAs(ctx, &userAccess, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Convert integrations list to slice
	var integrations []models.AgentIntegration
	if !plan.Integrations.IsUnknown() && !plan.Integrations.IsNull() {
		var integrationObjects []integrationObjectModel
		diags = plan.Integrations.ElementsAs(ctx, &integrationObjects, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, obj := range integrationObjects {
			integrations = append(integrations, models.AgentIntegration{
				Type:      obj.Type.ValueString(),
				ChannelID: obj.ChannelID.ValueString(),
			})
		}
	}

	// Get enable data access (defaults to false if not set)
	enableDataAccess := plan.EnableDataAccess.ValueBool()

	// Get enable self improvement (defaults to false if not set)
	enableSelfImprovement := plan.EnableSelfImprovement.ValueBool()

	// Get version (defaults to 2 if not set)
	version := plan.Version.ValueInt64()

	// Create agent via service
	agentService := services.NewAgentService(r.client)

	// Ensure all required fields have proper defaults
	if tags == nil {
		tags = []string{}
	}
	if groupAccess == nil {
		groupAccess = []string{}
	}
	if userAccess == nil {
		userAccess = []string{}
	}

	agent, err := agentService.CreateAgent(ctx, projectUuid, plan.Name.ValueString(), instruction, imageUrl, tags, integrations, groupAccess, userAccess, enableDataAccess, enableSelfImprovement, version)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Lightdash project agent",
			fmt.Sprintf("Could not create agent in project %q: %s", projectUuid, err.Error()),
		)
		return
	}

	// Map response to state model
	plan.ID = types.StringValue(fmt.Sprintf("organizations/%s/projects/%s/agents/%s",
		agent.OrganizationUUID, agent.ProjectUUID, agent.AgentUUID))
	plan.OrganizationUUID = types.StringValue(agent.OrganizationUUID)
	plan.ProjectUUID = types.StringValue(agent.ProjectUUID)
	plan.AgentUUID = types.StringValue(agent.AgentUUID)
	plan.Name = types.StringValue(agent.Name)

	// Handle integrations
	if agent.Integrations != nil {
		integrationObjects := []integrationObjectModel{}
		for _, integration := range agent.Integrations {
			integrationObjects = append(integrationObjects, integrationObjectModel{
				Type:      types.StringValue(integration.Type),
				ChannelID: types.StringValue(integration.ChannelID),
			})
		}
		integrationsList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":       types.StringType,
				"channel_id": types.StringType,
			},
		}, integrationObjects)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Integrations = integrationsList
	} else {
		plan.Integrations = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":       types.StringType,
				"channel_id": types.StringType,
			},
		})
	}

	// Handle Instruction - required field, cannot be null or empty
	var instructionTf types.String
	if agent.Instruction == nil || *agent.Instruction == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Agent instruction cannot be null or empty in API response, but the API returned null or empty value. This indicates an issue with the Lightdash API or the agent configuration.",
		)
		return
	}
	instructionTf = types.StringValue(*agent.Instruction)
	plan.Instruction = instructionTf

	// Convert tags slice to Terraform List (ensure never null)
	tagsVal := agent.Tags
	if tagsVal == nil {
		tagsVal = []string{}
	}
	tagsList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, tagsVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Tags = tagsList

	plan.UpdatedAt = types.StringValue(agent.UpdatedAt)
	plan.CreatedAt = types.StringValue(agent.CreatedAt)

	// Handle nullable ImageURL
	imageURL := types.StringNull()
	if agent.ImageURL != nil {
		imageURL = types.StringValue(*agent.ImageURL)
	}
	plan.ImageURL = imageURL

	plan.EnableDataAccess = types.BoolValue(agent.EnableDataAccess)
	plan.EnableSelfImprovement = types.BoolValue(agent.EnableSelfImprovement)

	// Convert group access slice to Terraform List (ensure never null)
	groupAccessVal := agent.GroupAccess
	if groupAccessVal == nil {
		groupAccessVal = []string{}
	}
	groupAccessList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, groupAccessVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.GroupAccess = groupAccessList

	// Convert user access slice to Terraform List (ensure never null)
	userAccessVal := agent.UserAccess
	if userAccessVal == nil {
		userAccessVal = []string{}
	}
	userAccessList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, userAccessVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.UserAccess = userAccessList

	// Set version (should be the same as what was sent)
	plan.Version = types.Int64Value(version)

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectAgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Declare variables to import from state
	var organizationUuid string
	var projectUuid string
	var agentUuid string

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("organization_uuid"), &organizationUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project_uuid"), &projectUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("agent_uuid"), &agentUuid)...)

	// Get current state
	var state projectAgentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get agent from service
	agentService := services.NewAgentService(r.client)
	agent, err := agentService.GetAgent(ctx, projectUuid, agentUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lightdash project agent",
			fmt.Sprintf("Unable to read agent with Project UUID %q and Agent UUID %q: %s", projectUuid, agentUuid, err.Error()),
		)
		return
	}

	// Map response to state model
	state.ID = types.StringValue(fmt.Sprintf("organizations/%s/projects/%s/agents/%s",
		agent.OrganizationUUID, agent.ProjectUUID, agent.AgentUUID))
	state.OrganizationUUID = types.StringValue(agent.OrganizationUUID)
	state.ProjectUUID = types.StringValue(agent.ProjectUUID)
	state.AgentUUID = types.StringValue(agent.AgentUUID)
	state.Name = types.StringValue(agent.Name)

	// Handle integrations
	if agent.Integrations != nil {
		integrationObjects := []integrationObjectModel{}
		for _, integration := range agent.Integrations {
			integrationObjects = append(integrationObjects, integrationObjectModel{
				Type:      types.StringValue(integration.Type),
				ChannelID: types.StringValue(integration.ChannelID),
			})
		}
		integrationsList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":       types.StringType,
				"channel_id": types.StringType,
			},
		}, integrationObjects)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Integrations = integrationsList
	} else {
		state.Integrations = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":       types.StringType,
				"channel_id": types.StringType,
			},
		})
	}

	// Handle Instruction - required field, cannot be null or empty
	var instructionTf types.String
	if agent.Instruction == nil || *agent.Instruction == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Agent instruction cannot be null or empty in API response, but the API returned null or empty value. This indicates an issue with the Lightdash API or the agent configuration.",
		)
		return
	}
	instructionTf = types.StringValue(*agent.Instruction)
	state.Instruction = instructionTf

	// Convert tags slice to Terraform List (ensure never null)
	tagsVal := agent.Tags
	if tagsVal == nil {
		tagsVal = []string{}
	}
	tagsList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, tagsVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tags = tagsList

	state.UpdatedAt = types.StringValue(agent.UpdatedAt)
	state.CreatedAt = types.StringValue(agent.CreatedAt)

	// Handle nullable ImageURL
	imageURL := types.StringNull()
	if agent.ImageURL != nil {
		imageURL = types.StringValue(*agent.ImageURL)
	}
	state.ImageURL = imageURL

	state.EnableDataAccess = types.BoolValue(agent.EnableDataAccess)
	state.EnableSelfImprovement = types.BoolValue(agent.EnableSelfImprovement)

	// Convert group access slice to Terraform List (ensure never null)
	groupAccessVal := agent.GroupAccess
	if groupAccessVal == nil {
		groupAccessVal = []string{}
	}
	groupAccessList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, groupAccessVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.GroupAccess = groupAccessList

	// Convert user access slice to Terraform List (ensure never null)
	userAccessVal := agent.UserAccess
	if userAccessVal == nil {
		userAccessVal = []string{}
	}
	userAccessList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, userAccessVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.UserAccess = userAccessList

	// Preserve deletion protection from current state - this is a Terraform setting
	// (already populated by req.State.Get(ctx, &state))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectAgentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read plan and state data
	var plan, state projectAgentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract values from state
	projectUuid := state.ProjectUUID.ValueString()
	agentUuid := state.AgentUUID.ValueString()

	// Prepare update request - send all fields from plan
	// Always send name (required field)
	nameVal := plan.Name.ValueString()
	name := &nameVal

	// Always send instruction (required field)
	instructionVal := plan.Instruction.ValueString()
	instruction := &instructionVal

	// Handle optional imageUrl
	var imageUrl *string
	if !plan.ImageURL.IsNull() {
		imageUrlVal := plan.ImageURL.ValueString()
		imageUrl = &imageUrlVal
	}

	// Always send tags
	var tags []string
	if !plan.Tags.IsNull() {
		diags := plan.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		tags = []string{}
	}

	// For integrations, since they're not exposed in the schema, we pass empty slice
	var integrations []models.AgentIntegration
	if !plan.Integrations.IsUnknown() && !plan.Integrations.IsNull() {
		var integrationObjects []integrationObjectModel
		diags := plan.Integrations.ElementsAs(ctx, &integrationObjects, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, obj := range integrationObjects {
			integrations = append(integrations, models.AgentIntegration{
				Type:      obj.Type.ValueString(),
				ChannelID: obj.ChannelID.ValueString(),
			})
		}
	}

	// Always send groupAccess
	var groupAccess []string
	if !plan.GroupAccess.IsNull() {
		diags := plan.GroupAccess.ElementsAs(ctx, &groupAccess, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		groupAccess = []string{}
	}

	// Always send userAccess
	var userAccess []string
	if !plan.UserAccess.IsNull() {
		diags := plan.UserAccess.ElementsAs(ctx, &userAccess, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		userAccess = []string{}
	}

	// Always include enableDataAccess in updates since it's a required field
	enableDataAccessVal := plan.EnableDataAccess.ValueBool()
	enableDataAccess := &enableDataAccessVal

	// Always include enableSelfImprovement in updates since it's a required field
	enableSelfImprovementVal := plan.EnableSelfImprovement.ValueBool()
	enableSelfImprovement := &enableSelfImprovementVal

	// Always include version in updates
	versionVal := plan.Version.ValueInt64()
	version := versionVal

	// Update agent via service
	agentService := services.NewAgentService(r.client)
	agent, err := agentService.UpdateAgent(ctx, projectUuid, agentUuid, name, instruction, imageUrl, tags, integrations, groupAccess, userAccess, enableDataAccess, enableSelfImprovement, version)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Lightdash project agent",
			fmt.Sprintf("Could not update agent %q in project %q: %s", agentUuid, projectUuid, err.Error()),
		)
		return
	}

	// Update the plan with the response data
	plan.ID = types.StringValue(fmt.Sprintf("organizations/%s/projects/%s/agents/%s",
		agent.OrganizationUUID, agent.ProjectUUID, agent.AgentUUID))
	plan.OrganizationUUID = types.StringValue(agent.OrganizationUUID)
	plan.ProjectUUID = types.StringValue(agent.ProjectUUID)
	plan.AgentUUID = types.StringValue(agent.AgentUUID)
	plan.Name = types.StringValue(agent.Name)

	// Handle integrations
	if agent.Integrations != nil {
		integrationObjects := []integrationObjectModel{}
		for _, integration := range agent.Integrations {
			integrationObjects = append(integrationObjects, integrationObjectModel{
				Type:      types.StringValue(integration.Type),
				ChannelID: types.StringValue(integration.ChannelID),
			})
		}
		integrationsList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":       types.StringType,
				"channel_id": types.StringType,
			},
		}, integrationObjects)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Integrations = integrationsList
	} else {
		plan.Integrations = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":       types.StringType,
				"channel_id": types.StringType,
			},
		})
	}

	// Handle Instruction - required field, cannot be null or empty
	var instructionTf types.String
	if agent.Instruction == nil || *agent.Instruction == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Agent instruction cannot be null or empty in API response, but the API returned null or empty value. This indicates an issue with the Lightdash API or the agent configuration.",
		)
		return
	}
	instructionTf = types.StringValue(*agent.Instruction)
	plan.Instruction = instructionTf

	// Convert tags slice to Terraform List (ensure never null)
	tagsVal := agent.Tags
	if tagsVal == nil {
		tagsVal = []string{}
	}
	tagsList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, tagsVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Tags = tagsList

	plan.UpdatedAt = types.StringValue(agent.UpdatedAt)
	plan.CreatedAt = types.StringValue(agent.CreatedAt)

	// Handle nullable ImageURL
	imageURL := types.StringNull()
	if agent.ImageURL != nil {
		imageURL = types.StringValue(*agent.ImageURL)
	}
	plan.ImageURL = imageURL

	plan.EnableDataAccess = types.BoolValue(agent.EnableDataAccess)
	plan.EnableSelfImprovement = types.BoolValue(agent.EnableSelfImprovement)

	// Convert group access slice to Terraform List (ensure never null)
	groupAccessVal := agent.GroupAccess
	if groupAccessVal == nil {
		groupAccessVal = []string{}
	}
	groupAccessList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, groupAccessVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.GroupAccess = groupAccessList

	// Convert user access slice to Terraform List (ensure never null)
	userAccessVal := agent.UserAccess
	if userAccessVal == nil {
		userAccessVal = []string{}
	}
	userAccessList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, userAccessVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.UserAccess = userAccessList

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectAgentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state projectAgentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract values from state
	projectUuid := state.ProjectUUID.ValueString()
	agentUuid := state.AgentUUID.ValueString()
	deletionProtection := state.DeleteProtection.ValueBool()

	// Check deletion protection
	if deletionProtection {
		resp.Diagnostics.AddError(
			"Deletion Protection Enabled",
			"Cannot delete project agent because deletion_protection is set to true. Set deletion_protection to false to allow deletion.",
		)
		return
	}

	// Delete agent via service
	agentService := services.NewAgentService(r.client)
	err := agentService.DeleteAgent(ctx, projectUuid, agentUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Lightdash project agent",
			fmt.Sprintf("Could not delete agent %q in project %q: %s", agentUuid, projectUuid, err.Error()),
		)
		return
	}
}

func (r *projectAgentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Extract the resource ID
	extracted_strings, err := extractProjectAgentResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	organization_uuid := extracted_strings[0]
	project_uuid := extracted_strings[1]
	agent_uuid := extracted_strings[2]

	tflog.Info(ctx, fmt.Sprintf("Importing agent with Organization UUID %s, Project UUID %s, Agent UUID %s", organization_uuid, project_uuid, agent_uuid))

	// Fetch the agent data from the API
	agentService := services.NewAgentService(r.client)
	agent, err := agentService.GetAgent(ctx, project_uuid, agent_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing Lightdash project agent",
			fmt.Sprintf("Could not retrieve agent %q in project %q: %s", agent_uuid, project_uuid, err.Error()),
		)
		return
	}

	// Set all the resource attributes from the fetched agent data
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_uuid"), agent.OrganizationUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), agent.ProjectUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("agent_uuid"), agent.AgentUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), agent.Name)...)

	// Handle Instruction - required field, cannot be null or empty
	if agent.Instruction == nil || *agent.Instruction == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Agent instruction cannot be null or empty in API response, but the API returned null or empty value. This indicates an issue with the Lightdash API or the agent configuration.",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instruction"), *agent.Instruction)...)

	// Convert tags slice to Terraform List (ensure never null)
	tagsVal := agent.Tags
	if tagsVal == nil {
		tagsVal = []string{}
	}
	tagsList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, tagsVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tags"), tagsList)...)

	// Handle integrations
	if agent.Integrations != nil {
		integrationObjects := []integrationObjectModel{}
		for _, integration := range agent.Integrations {
			integrationObjects = append(integrationObjects, integrationObjectModel{
				Type:      types.StringValue(integration.Type),
				ChannelID: types.StringValue(integration.ChannelID),
			})
		}
		integrationsList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":       types.StringType,
				"channel_id": types.StringType,
			},
		}, integrationObjects)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("integrations"), integrationsList)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("integrations"), types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":       types.StringType,
				"channel_id": types.StringType,
			},
		}))...)
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("updated_at"), agent.UpdatedAt)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), agent.CreatedAt)...)

	// Handle nullable ImageURL
	if agent.ImageURL != nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("image_url"), *agent.ImageURL)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("image_url"), types.StringNull())...)
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enable_data_access"), agent.EnableDataAccess)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enable_self_improvement"), agent.EnableSelfImprovement)...)

	// Convert group access slice to Terraform List (ensure never null)
	groupAccessVal := agent.GroupAccess
	if groupAccessVal == nil {
		groupAccessVal = []string{}
	}
	groupAccessList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, groupAccessVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_access"), groupAccessList)...)

	// Convert user access slice to Terraform List (ensure never null)
	userAccessVal := agent.UserAccess
	if userAccessVal == nil {
		userAccessVal = []string{}
	}
	userAccessList, diags := basetypes.NewListValueFrom(context.Background(), types.StringType, userAccessVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_access"), userAccessList)...)

	// Set deletion protection to false by default for imported resources (matches schema default)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("deletion_protection"), false)...)
	// Set version to default value for imported resources
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), int64(2))...)
}

func getProjectAgentResourceId(organization_uuid string, project_uuid string, agent_uuid string) string {
	// Return the resource ID
	return fmt.Sprintf("organizations/%s/projects/%s/agents/%s", organization_uuid, project_uuid, agent_uuid)
}

func extractProjectAgentResourceId(input string) ([]string, error) {
	// Extract the captured groups
	pattern := `^organizations/([^/]+)/projects/([^/]+)/agents/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	// Return the captured strings
	organization_uuid := groups[0]
	project_uuid := groups[1]
	agent_uuid := groups[2]
	return []string{organization_uuid, project_uuid, agent_uuid}, nil
}
