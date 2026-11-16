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
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"
)

var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
	_ resource.ResourceWithModifyPlan  = &projectResource{}
)

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResource struct {
	client *api.Client
}

type projectResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ProjectUUID      types.String `tfsdk:"project_uuid"`
	Slug             types.String `tfsdk:"slug"`
	Name             types.String `tfsdk:"name"`
	ProjectType      types.String `tfsdk:"type"`
	DbtVersion       types.String `tfsdk:"dbt_version"`
	DeleteProtection types.Bool   `tfsdk:"deletion_protection"`
}

// projectDeleteDecision is the Terraform-only result of destroying a project.
// The Lightdash API has no project delete endpoint, so Delete never calls one.
type projectDeleteDecision struct {
	Block   bool
	Summary string
	Detail  string
}

// deleteDecision reports whether Delete must fail. When protection is off, Delete
// still does not remove the project in Lightdash; it only drops Terraform state.
func deleteDecision(protected bool) projectDeleteDecision {
	if protected {
		return projectDeleteDecision{
			Block:   true,
			Summary: "Deletion Protection Enabled",
			Detail:  "Cannot delete Lightdash project because deletion_protection is set to true. Set deletion_protection to false to allow Terraform to remove the project from state. The project itself must still be deleted manually in the Lightdash UI, because the API has no project delete endpoint.",
		}
	}
	return projectDeleteDecision{
		Block:   false,
		Summary: "Lightdash project was not deleted",
		Detail:  "Terraform removed the project from state only. The Lightdash API has no project delete endpoint, so delete the project manually in the Lightdash UI.",
	}
}

func (r *projectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_lightdash_project.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Manages a Lightdash project",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The project UUID. Import an existing project with this value.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the Lightdash project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL slug of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The project name. Changing it replaces the Terraform resource. The previous Lightdash project is left in place.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					ValidateNonEmptyString{},
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The project type. One of `DEFAULT`, `PREVIEW`, or `TRAINING`. Changing it replaces the Terraform resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					ValidateProjectType{},
				},
			},
			"dbt_version": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The dbt version sent to Lightdash unchanged. Lightdash rejects a version it does not support. Changing it replaces the Terraform resource. The dbt connection is always type `none`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					ValidateNonEmptyString{},
				},
			},
			"deletion_protection": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "When set to `true`, Terraform will not remove the project from state. Imported projects default to `true`. Destroy never deletes the project in Lightdash.",
			},
		},
	}
}

func (r *projectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := apiv1.CreateProjectV1(r.client, plan.Name.ValueString(), plan.ProjectType.ValueString(), plan.DbtVersion.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			fmt.Sprintf("Could not create project %q: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	if detail, unsupported := unsupportedDbtConnection(created); unsupported {
		resp.Diagnostics.AddError("Unsupported dbt connection", detail)
		return
	}
	applyProjectToModel(&plan, created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := apiv1.GetProjectV1(r.client, state.ProjectUUID.ValueString())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("Project %s not found during Read, removing from state", state.ProjectUUID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading project",
			fmt.Sprintf("Could not read project %s: %s", state.ProjectUUID.ValueString(), err.Error()),
		)
		return
	}

	if detail, unsupported := unsupportedDbtConnection(project); unsupported {
		resp.Diagnostics.AddWarning(
			"Unsupported dbt connection",
			detail+" The project stays in state so it can be removed from Terraform.",
		)
	}

	applyProjectToModel(&state, project)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var plan, state projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !projectReplacementBlocked(state, plan) {
		return
	}

	resp.Diagnostics.AddError(
		"Deletion Protection Enabled",
		"Turn deletion_protection off and apply that change before replacing the project. Replacement uses the protection value already stored in state.",
	)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// name, type, and dbt_version require replacement. The only in-place change is deletion_protection.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	decision := deleteDecision(state.DeleteProtection.ValueBool())
	detail := fmt.Sprintf("%s Project %q (%s).", decision.Detail, state.Name.ValueString(), state.ProjectUUID.ValueString())
	if decision.Block {
		resp.Diagnostics.AddError(decision.Summary, detail)
		return
	}

	tflog.Warn(ctx, detail)
	resp.Diagnostics.AddWarning(decision.Summary, detail)
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectUUID := strings.TrimSpace(req.ID)
	if projectUUID == "" {
		resp.Diagnostics.AddError("Error importing project", "The import ID is empty. Pass the project UUID.")
		return
	}

	project, err := apiv1.GetProjectV1(r.client, projectUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing project",
			fmt.Sprintf("Could not read project %s: %s", projectUUID, err.Error()),
		)
		return
	}
	if detail, unsupported := unsupportedDbtConnection(project); unsupported {
		resp.Diagnostics.AddError("Unsupported dbt connection", detail)
		return
	}

	state := projectResourceModel{DeleteProtection: types.BoolValue(true)}
	applyProjectToModel(&state, project)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func applyProjectToModel(model *projectResourceModel, project *apiv1.GetProjectV1Results) {
	model.ID = types.StringValue(project.ProjectUUID)
	model.ProjectUUID = types.StringValue(project.ProjectUUID)
	model.Slug = types.StringValue(project.Slug)
	model.Name = types.StringValue(project.ProjectName)
	model.ProjectType = types.StringValue(project.ProjectType)
	model.DbtVersion = types.StringValue(project.DbtVersion)
}

func projectReplacementBlocked(state, plan projectResourceModel) bool {
	if !state.DeleteProtection.ValueBool() {
		return false
	}
	return !plan.Name.Equal(state.Name) || !plan.ProjectType.Equal(state.ProjectType) || !plan.DbtVersion.Equal(state.DbtVersion)
}

func unsupportedDbtConnection(project *apiv1.GetProjectV1Results) (string, bool) {
	if project.DbtConnection != nil && project.DbtConnection.Type == apiv1.DbtConnectionTypeNone {
		return "", false
	}
	got := "<missing>"
	if project.DbtConnection != nil {
		got = project.DbtConnection.Type
	}
	return fmt.Sprintf("Project %s has dbt connection type %q. This resource only manages projects whose dbt connection type is %q.", project.ProjectUUID, got, apiv1.DbtConnectionTypeNone), true
}
