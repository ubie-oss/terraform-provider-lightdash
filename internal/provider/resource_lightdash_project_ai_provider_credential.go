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

	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

var (
	_ resource.Resource                = &projectAiProviderCredentialResource{}
	_ resource.ResourceWithConfigure   = &projectAiProviderCredentialResource{}
	_ resource.ResourceWithImportState = &projectAiProviderCredentialResource{}
)

func NewProjectAiProviderCredentialResource() resource.Resource {
	return &projectAiProviderCredentialResource{}
}

type projectAiProviderCredentialResource struct {
	client *api.Client
}

type projectAiProviderCredentialResourceModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationUUID types.String `tfsdk:"organization_uuid"`
	ProjectUUID      types.String `tfsdk:"project_uuid"`
	CredentialUUID   types.String `tfsdk:"credential_uuid"`
	DeleteProtection types.Bool   `tfsdk:"deletion_protection"`
}

func (r *projectAiProviderCredentialResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_ai_provider_credential"
}

func (r *projectAiProviderCredentialResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_lightdash_project_ai_provider_credential.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Manages project-level AI provider credential selection",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier: `organizations/<organization_uuid>/projects/<project_uuid>/ai_provider_credential`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash project.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"credential_uuid": schema.StringAttribute{
				MarkdownDescription: "Organization AI provider credential UUID for this project. Omit or leave unset to use the organization default.",
				Optional:            true,
				Computed:            true,
			},
			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: "When set to `true`, prevents Terraform from resetting the project binding on destroy.",
				Required:            true,
			},
		},
	}
}

func (r *projectAiProviderCredentialResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectAiProviderCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectAiProviderCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateOrganizationUUID(r.client, plan.OrganizationUUID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	selection := projectAiCredentialSelectionFromModel(plan.CredentialUUID)
	if err := apiv1.SetProjectAiProviderCredentialV1(r.client, plan.ProjectUUID.ValueString(), selection); err != nil {
		resp.Diagnostics.AddError("Error setting project AI provider credential", err.Error())
		return
	}

	plan.ID = types.StringValue(getProjectAiProviderCredentialResourceID(
		plan.OrganizationUUID.ValueString(),
		plan.ProjectUUID.ValueString(),
	))
	resp.Diagnostics.Append(setProjectAiProviderCredentialFromAPI(ctx, r.client, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectAiProviderCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectAiProviderCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setProjectAiProviderCredentialFromAPI(ctx, r.client, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectAiProviderCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state projectAiProviderCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	selection := projectAiCredentialSelectionFromModel(plan.CredentialUUID)
	if err := apiv1.SetProjectAiProviderCredentialV1(r.client, plan.ProjectUUID.ValueString(), selection); err != nil {
		resp.Diagnostics.AddError("Error updating project AI provider credential", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(setProjectAiProviderCredentialFromAPI(ctx, r.client, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectAiProviderCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectAiProviderCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DeleteProtection.ValueBool() {
		resp.Diagnostics.AddError(
			"Deletion Protection Enabled",
			"Cannot delete project AI provider credential because deletion_protection is set to true. Set deletion_protection to false to allow deletion.",
		)
		return
	}

	if err := apiv1.SetProjectAiProviderCredentialV1(r.client, state.ProjectUUID.ValueString(), models.ProjectAiCredentialSelection{
		CredentialUUID: nil,
	}); err != nil {
		resp.Diagnostics.AddError("Error resetting project AI provider credential", err.Error())
		return
	}
}

func (r *projectAiProviderCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	extracted, err := extractProjectAiProviderCredentialResourceID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error extracting resource ID", err.Error())
		return
	}
	organizationUUID := extracted[0]
	projectUUID := extracted[1]

	resp.Diagnostics.Append(validateOrganizationUUID(r.client, organizationUUID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := projectAiProviderCredentialResourceModel{
		ID:               types.StringValue(req.ID),
		OrganizationUUID: types.StringValue(organizationUUID),
		ProjectUUID:      types.StringValue(projectUUID),
		DeleteProtection: types.BoolValue(true),
	}
	resp.Diagnostics.Append(setProjectAiProviderCredentialFromAPI(ctx, r.client, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func projectAiCredentialSelectionFromModel(credentialUUID types.String) models.ProjectAiCredentialSelection {
	if credentialUUID.IsNull() || credentialUUID.IsUnknown() || credentialUUID.ValueString() == "" {
		return models.ProjectAiCredentialSelection{CredentialUUID: nil}
	}
	uuid := credentialUUID.ValueString()
	return models.ProjectAiCredentialSelection{CredentialUUID: &uuid}
}

func setProjectAiProviderCredentialFromAPI(_ context.Context, client *api.Client, model *projectAiProviderCredentialResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	selection, err := apiv1.GetProjectAiProviderCredentialV1(client, model.ProjectUUID.ValueString())
	if err != nil {
		diags.AddError("Error reading project AI provider credential", err.Error())
		return diags
	}

	if selection.CredentialUUID == nil {
		model.CredentialUUID = types.StringNull()
	} else {
		model.CredentialUUID = types.StringValue(*selection.CredentialUUID)
	}

	model.ID = types.StringValue(getProjectAiProviderCredentialResourceID(
		model.OrganizationUUID.ValueString(),
		model.ProjectUUID.ValueString(),
	))

	return diags
}

func getProjectAiProviderCredentialResourceID(organizationUUID, projectUUID string) string {
	return fmt.Sprintf("organizations/%s/projects/%s/ai_provider_credential", organizationUUID, projectUUID)
}

func extractProjectAiProviderCredentialResourceID(input string) ([]string, error) {
	pattern := `^organizations/([^/]+)/projects/([^/]+)/ai_provider_credential$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}
	return []string{groups[0], groups[1]}, nil
}
