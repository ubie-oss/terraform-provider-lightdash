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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &projectAgentEvaluationsResource{}
	_ resource.ResourceWithConfigure   = &projectAgentEvaluationsResource{}
	_ resource.ResourceWithImportState = &projectAgentEvaluationsResource{}
)

func NewProjectAgentEvaluationsResource() resource.Resource {
	return &projectAgentEvaluationsResource{}
}

// projectAgentEvaluationsResource defines the resource implementation.
type projectAgentEvaluationsResource struct {
	client *api.Client
}

// projectAgentEvaluationsResourceModel describes the resource data model.
type projectAgentEvaluationsResourceModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationUUID types.String `tfsdk:"organization_uuid"`
	ProjectUUID      types.String `tfsdk:"project_uuid"`
	AgentUUID        types.String `tfsdk:"agent_uuid"`
	EvaluationUUID   types.String `tfsdk:"evaluation_uuid"`
	Title            types.String `tfsdk:"title"`
	Description      types.String `tfsdk:"description"`
	Prompts          types.List   `tfsdk:"prompts"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeleteProtection types.Bool   `tfsdk:"deletion_protection"`
}

func (r *projectAgentEvaluationsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_agent_evaluations"
}

func (r *projectAgentEvaluationsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_lightdash_project_agent_evaluations.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Manages a Lightdash project agent evaluation",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource identifier. It is computed as `organizations/<organization_uuid>/projects/<project_uuid>/agents/<agent_uuid>/evaluations/<evaluation_uuid>`.",
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
				Required:            true,
			},
			"evaluation_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the evaluation.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "The title of the evaluation.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description of the evaluation.",
				Optional:            true,
				Computed:            true,
			},
			"prompts": schema.ListNestedAttribute{
				MarkdownDescription: "List of evaluation prompts.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"prompt": schema.StringAttribute{
							MarkdownDescription: "The prompt text.",
							Required:            true,
						},
						"eval_prompt_uuid": schema.StringAttribute{
							MarkdownDescription: "The UUID of the evaluation prompt.",
							Optional:            true,
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the prompt.",
							Optional:            true,
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Timestamp of creation.",
							Computed:            true,
						},
					},
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the last update.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp of creation.",
				Computed:            true,
			},
			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: "When set to `true`, prevents the destruction of the project agent evaluation resource by Terraform. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *projectAgentEvaluationsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectAgentEvaluationsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read plan data into model
	var plan projectAgentEvaluationsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract plan values
	projectUuid := plan.ProjectUUID.ValueString()
	agentUuid := plan.AgentUUID.ValueString()
	title := plan.Title.ValueString()

	// Get optional description
	var description *string
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		desc := plan.Description.ValueString()
		description = &desc
	}

	// Convert prompts list to slice of strings
	var prompts []string
	if !plan.Prompts.IsUnknown() && !plan.Prompts.IsNull() {
		var promptObjects []types.Object
		diags = plan.Prompts.ElementsAs(ctx, &promptObjects, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, obj := range promptObjects {
			// Extract the prompt field from the object
			promptAttr := obj.Attributes()["prompt"]
			if promptAttr != nil {
				promptValue, ok := promptAttr.(types.String)
				if ok {
					prompts = append(prompts, promptValue.ValueString())
				}
			}
		}
	}

	// Create evaluation via service
	evaluationService := services.NewAgentEvaluationsService(r.client)
	evaluation, err := evaluationService.CreateEvaluations(ctx, projectUuid, agentUuid, title, description, prompts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Lightdash project agent evaluation",
			fmt.Sprintf("Could not create evaluation for agent %q in project %q: %s", agentUuid, projectUuid, err.Error()),
		)
		return
	}

	// Map response to state model
	plan.ID = types.StringValue(fmt.Sprintf("organizations/%s/projects/%s/agents/%s/evaluations/%s",
		plan.OrganizationUUID.ValueString(), projectUuid, evaluation.AgentUUID, evaluation.EvalUUID))
	plan.ProjectUUID = types.StringValue(projectUuid)
	plan.AgentUUID = types.StringValue(agentUuid)
	plan.EvaluationUUID = types.StringValue(evaluation.EvalUUID)
	plan.Title = types.StringValue(evaluation.Title)

	// Handle nullable Description
	descriptionTf := types.StringNull()
	if evaluation.Description != nil {
		descriptionTf = types.StringValue(*evaluation.Description)
	}
	plan.Description = descriptionTf

	// Convert prompts slice to Terraform nested objects
	var promptObjects []basetypes.ObjectValue
	for _, prompt := range evaluation.Prompts {
		obj, diags := basetypes.NewObjectValue(
			map[string]attr.Type{
				"prompt":           types.StringType,
				"eval_prompt_uuid": types.StringType,
				"type":             types.StringType,
				"created_at":       types.StringType,
			},
			map[string]attr.Value{
				"prompt":           types.StringValue(prompt.Prompt),
				"eval_prompt_uuid": types.StringValue(prompt.EvalPromptUUID),
				"type":             types.StringValue(prompt.Type),
				"created_at":       types.StringValue(prompt.CreatedAt),
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		promptObjects = append(promptObjects, obj)
	}
	promptsList, diags := basetypes.NewListValueFrom(ctx, basetypes.ObjectType{
		AttrTypes: map[string]attr.Type{
			"prompt":           types.StringType,
			"eval_prompt_uuid": types.StringType,
			"type":             types.StringType,
			"created_at":       types.StringType,
		},
	}, promptObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Prompts = promptsList

	plan.UpdatedAt = types.StringValue(evaluation.UpdatedAt)
	plan.CreatedAt = types.StringValue(evaluation.CreatedAt)

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectAgentEvaluationsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Declare variables to import from state
	var organizationUuid string
	var projectUuid string
	var agentUuid string
	var evaluationUuid string

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("organization_uuid"), &organizationUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project_uuid"), &projectUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("agent_uuid"), &agentUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("evaluation_uuid"), &evaluationUuid)...)

	// Get current state. This is needed to preserve Terraform-managed attributes.
	var currentState projectAgentEvaluationsResourceModel
	diags := req.State.Get(ctx, &currentState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get evaluation from service
	evaluationService := services.NewAgentEvaluationsService(r.client)
	evaluation, err := evaluationService.GetEvaluations(ctx, projectUuid, agentUuid, evaluationUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lightdash project agent evaluation",
			fmt.Sprintf("Unable to read evaluation with Project UUID %q, Agent UUID %q, and Evaluation UUID %q: %s", projectUuid, agentUuid, evaluationUuid, err.Error()),
		)
		return
	}

	// Update state from controller response
	var newState projectAgentEvaluationsResourceModel

	// Preserve fields that are computed once or managed by Terraform
	newState.ID = currentState.ID
	newState.OrganizationUUID = currentState.OrganizationUUID
	newState.ProjectUUID = currentState.ProjectUUID
	newState.AgentUUID = currentState.AgentUUID
	newState.EvaluationUUID = currentState.EvaluationUUID
	newState.DeleteProtection = currentState.DeleteProtection

	// Update fields from API response
	newState.Title = types.StringValue(evaluation.Title)

	// Handle nullable Description
	descriptionTf := types.StringNull()
	if evaluation.Description != nil {
		descriptionTf = types.StringValue(*evaluation.Description)
	}
	newState.Description = descriptionTf

	// Convert prompts slice to Terraform nested objects
	var promptObjects []basetypes.ObjectValue
	for _, prompt := range evaluation.Prompts {
		obj, diags := basetypes.NewObjectValue(
			map[string]attr.Type{
				"prompt":           types.StringType,
				"eval_prompt_uuid": types.StringType,
				"type":             types.StringType,
				"created_at":       types.StringType,
			},
			map[string]attr.Value{
				"prompt":           types.StringValue(prompt.Prompt),
				"eval_prompt_uuid": types.StringValue(prompt.EvalPromptUUID),
				"type":             types.StringValue(prompt.Type),
				"created_at":       types.StringValue(prompt.CreatedAt),
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		promptObjects = append(promptObjects, obj)
	}
	promptsList, diags := basetypes.NewListValueFrom(ctx, basetypes.ObjectType{
		AttrTypes: map[string]attr.Type{
			"prompt":           types.StringType,
			"eval_prompt_uuid": types.StringType,
			"type":             types.StringType,
			"created_at":       types.StringType,
		},
	}, promptObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	newState.Prompts = promptsList

	newState.UpdatedAt = types.StringValue(evaluation.UpdatedAt)
	newState.CreatedAt = types.StringValue(evaluation.CreatedAt)

	// Set refreshed state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectAgentEvaluationsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read plan and state data
	var plan, state projectAgentEvaluationsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract values from state
	projectUuid := state.ProjectUUID.ValueString()
	agentUuid := state.AgentUUID.ValueString()
	evaluationUuid := state.EvaluationUUID.ValueString()

	// Prepare update request with only changed fields
	var title *string
	if !plan.Title.Equal(state.Title) {
		titleVal := plan.Title.ValueString()
		title = &titleVal
	}

	var description *string
	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			description = nil
		} else {
			descriptionVal := plan.Description.ValueString()
			description = &descriptionVal
		}
	}

	var prompts []string
	if !plan.Prompts.Equal(state.Prompts) {
		if plan.Prompts.IsNull() {
			prompts = []string{}
		} else {
			var promptObjects []types.Object
			diags := plan.Prompts.ElementsAs(ctx, &promptObjects, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			for _, obj := range promptObjects {
				// Extract the prompt field from the object
				promptAttr := obj.Attributes()["prompt"]
				if promptAttr != nil {
					promptValue := promptAttr.(types.String)
					prompts = append(prompts, promptValue.ValueString())
				}
			}
		}
	}

	// Update evaluation via service
	evaluationService := services.NewAgentEvaluationsService(r.client)
	evaluation, err := evaluationService.UpdateEvaluations(ctx, projectUuid, agentUuid, evaluationUuid, title, description, prompts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Lightdash project agent evaluation",
			fmt.Sprintf("Could not update evaluation %q for agent %q in project %q: %s", evaluationUuid, agentUuid, projectUuid, err.Error()),
		)
		return
	}

	// Update the plan with the response data
	plan.ID = types.StringValue(fmt.Sprintf("organizations/%s/projects/%s/agents/%s/evaluations/%s",
		plan.OrganizationUUID.ValueString(), projectUuid, evaluation.AgentUUID, evaluation.EvalUUID))
	plan.ProjectUUID = types.StringValue(projectUuid)
	plan.AgentUUID = types.StringValue(agentUuid)
	plan.EvaluationUUID = types.StringValue(evaluation.EvalUUID)
	plan.Title = types.StringValue(evaluation.Title)

	// Handle nullable Description
	descriptionTf := types.StringNull()
	if evaluation.Description != nil {
		descriptionTf = types.StringValue(*evaluation.Description)
	}
	plan.Description = descriptionTf

	// Convert prompts slice to Terraform nested objects
	var promptObjects []basetypes.ObjectValue
	for _, prompt := range evaluation.Prompts {
		obj, diags := basetypes.NewObjectValue(
			map[string]attr.Type{
				"prompt":           types.StringType,
				"eval_prompt_uuid": types.StringType,
				"type":             types.StringType,
				"created_at":       types.StringType,
			},
			map[string]attr.Value{
				"prompt":           types.StringValue(prompt.Prompt),
				"eval_prompt_uuid": types.StringValue(prompt.EvalPromptUUID),
				"type":             types.StringValue(prompt.Type),
				"created_at":       types.StringValue(prompt.CreatedAt),
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		promptObjects = append(promptObjects, obj)
	}
	promptsList, diags := basetypes.NewListValueFrom(ctx, basetypes.ObjectType{
		AttrTypes: map[string]attr.Type{
			"prompt":           types.StringType,
			"eval_prompt_uuid": types.StringType,
			"type":             types.StringType,
			"created_at":       types.StringType,
		},
	}, promptObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Prompts = promptsList

	plan.UpdatedAt = types.StringValue(evaluation.UpdatedAt)
	plan.CreatedAt = types.StringValue(evaluation.CreatedAt)

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectAgentEvaluationsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state projectAgentEvaluationsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract values from state
	projectUuid := state.ProjectUUID.ValueString()
	agentUuid := state.AgentUUID.ValueString()
	evaluationUuid := state.EvaluationUUID.ValueString()
	deletionProtection := state.DeleteProtection.ValueBool()

	// Check deletion protection
	if deletionProtection {
		resp.Diagnostics.AddError(
			"Deletion protection is enabled",
			fmt.Sprintf("Cannot delete evaluation %q for agent %q in project %q because deletion protection is enabled. Set deletion_protection to false to allow deletion.", evaluationUuid, agentUuid, projectUuid),
		)
		return
	}

	// Delete evaluation via service
	evaluationService := services.NewAgentEvaluationsService(r.client)
	err := evaluationService.DeleteEvaluations(ctx, projectUuid, agentUuid, evaluationUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Lightdash project agent evaluation",
			fmt.Sprintf("Could not delete evaluation %q for agent %q in project %q: %s", evaluationUuid, agentUuid, projectUuid, err.Error()),
		)
		return
	}
}

func (r *projectAgentEvaluationsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Extract the resource ID
	extracted_strings, err := extractProjectAgentEvaluationResourceId(req.ID)
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
	evaluation_uuid := extracted_strings[3]

	tflog.Info(ctx, fmt.Sprintf("Importing evaluation with Organization UUID %s, Project UUID %s, Agent UUID %s, Evaluation UUID %s", organization_uuid, project_uuid, agent_uuid, evaluation_uuid))

	// Fetch the evaluation data from the API
	evaluationService := services.NewAgentEvaluationsService(r.client)
	evaluation, err := evaluationService.GetEvaluations(ctx, project_uuid, agent_uuid, evaluation_uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing Lightdash project agent evaluation",
			fmt.Sprintf("Could not retrieve evaluation %q for agent %q in project %q: %s", evaluation_uuid, agent_uuid, project_uuid, err.Error()),
		)
		return
	}

	// Set all the resource attributes from the fetched evaluation data
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_uuid"), organization_uuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), project_uuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("agent_uuid"), evaluation.AgentUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("evaluation_uuid"), evaluation.EvalUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("title"), evaluation.Title)...)

	// Handle nullable Description
	if evaluation.Description != nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), *evaluation.Description)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), types.StringNull())...)
	}

	// Convert prompts slice to Terraform nested objects
	var promptObjects []basetypes.ObjectValue
	for _, prompt := range evaluation.Prompts {
		obj, diags := basetypes.NewObjectValue(
			map[string]attr.Type{
				"prompt":           types.StringType,
				"eval_prompt_uuid": types.StringType,
				"type":             types.StringType,
				"created_at":       types.StringType,
			},
			map[string]attr.Value{
				"prompt":           types.StringValue(prompt.Prompt),
				"eval_prompt_uuid": types.StringValue(prompt.EvalPromptUUID),
				"type":             types.StringValue(prompt.Type),
				"created_at":       types.StringValue(prompt.CreatedAt),
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		promptObjects = append(promptObjects, obj)
	}
	promptsList, diags := basetypes.NewListValueFrom(ctx, basetypes.ObjectType{
		AttrTypes: map[string]attr.Type{
			"prompt":           types.StringType,
			"eval_prompt_uuid": types.StringType,
			"type":             types.StringType,
			"created_at":       types.StringType,
		},
	}, promptObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("prompts"), promptsList)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("updated_at"), evaluation.UpdatedAt)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), evaluation.CreatedAt)...)

	// Set deletion protection to false by default for imported resources (matches schema default)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("deletion_protection"), false)...)
}

func extractProjectAgentEvaluationResourceId(input string) ([]string, error) {
	// Extract the captured groups
	pattern := `^organizations/([^/]+)/projects/([^/]+)/agents/([^/]+)/evaluations/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	// Return the captured strings
	organization_uuid := groups[0]
	project_uuid := groups[1]
	agent_uuid := groups[2]
	evaluation_uuid := groups[3]
	return []string{organization_uuid, project_uuid, agent_uuid, evaluation_uuid}, nil
}
