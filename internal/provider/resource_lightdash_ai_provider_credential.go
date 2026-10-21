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
	"slices"
	"strings"

	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

var (
	_ resource.Resource                = &aiProviderCredentialResource{}
	_ resource.ResourceWithConfigure   = &aiProviderCredentialResource{}
	_ resource.ResourceWithImportState = &aiProviderCredentialResource{}
)

func NewAiProviderCredentialResource() resource.Resource {
	return &aiProviderCredentialResource{}
}

type aiProviderCredentialResource struct {
	client *api.Client
}

type aiProviderCredentialResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	OrganizationUUID   types.String `tfsdk:"organization_uuid"`
	CredentialUUID     types.String `tfsdk:"credential_uuid"`
	AIProvider         types.String `tfsdk:"ai_provider"`
	Label              types.String `tfsdk:"label"`
	Region             types.String `tfsdk:"region"`
	AllowedModels      types.List   `tfsdk:"allowed_models"`
	InferenceGeography types.String `tfsdk:"inference_geography"`
	APIKey             types.String `tfsdk:"api_key"`
	APIKeyHint         types.String `tfsdk:"api_key_hint"`
	IsDefault          types.Bool   `tfsdk:"is_default"`
	DeleteProtection   types.Bool   `tfsdk:"deletion_protection"`
}

func (r *aiProviderCredentialResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ai_provider_credential"
}

func (r *aiProviderCredentialResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_lightdash_ai_provider_credential.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Manages a Lightdash organization AI provider credential",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Credential UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization. Must match the organization for the configured API token.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"credential_uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the credential in Lightdash.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ai_provider": schema.StringAttribute{
				MarkdownDescription: "AI provider type. Currently only `bedrock` is supported.",
				Required:            true,
				Validators: []validator.String{
					ValidateNonEmptyString{},
				},
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Display label for the credential.",
				Required:            true,
				Validators: []validator.String{
					ValidateNonEmptyString{},
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "AWS region for Bedrock.",
				Required:            true,
				Validators: []validator.String{
					ValidateNonEmptyString{},
				},
			},
			"allowed_models": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Model identifiers allowed for this credential.",
				Required:            true,
			},
			"inference_geography": schema.StringAttribute{
				MarkdownDescription: "Bedrock inference geography (`jp`, `us`, `eu`, `apac`, `global`).",
				Optional:            true,
				Computed:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Bedrock API key or gateway bearer token. Required on create; optional on update to rotate.",
				Optional:            true,
				Sensitive:           true,
			},
			"api_key_hint": schema.StringAttribute{
				MarkdownDescription: "Hint of the stored API key from Lightdash.",
				Computed:            true,
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "When true, sets this credential as the organization default after create or update.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: "When set to `true`, prevents Terraform from destroying the credential.",
				Required:            true,
			},
		},
	}
}

func (r *aiProviderCredentialResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *aiProviderCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan aiProviderCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateOrganizationUUID(r.client, plan.OrganizationUUID.ValueString())...)
	resp.Diagnostics.Append(validateAiProviderCredentialProvider(plan.AIProvider.ValueString())...)
	resp.Diagnostics.Append(validateInferenceGeography(plan.InferenceGeography)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.APIKey.IsNull() || plan.APIKey.IsUnknown() || plan.APIKey.ValueString() == "" {
		resp.Diagnostics.AddError("Missing API key", "api_key is required when creating an AI provider credential.")
		return
	}

	allowedModels, diags := stringListToStringSlice(ctx, plan.AllowedModels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(allowedModels) == 0 {
		resp.Diagnostics.AddError("Invalid allowed_models", "allowed_models must contain at least one model.")
		return
	}

	createReq := models.CreateAiProviderCredentialRequest{
		Provider:      plan.AIProvider.ValueString(),
		Label:         plan.Label.ValueString(),
		Region:        plan.Region.ValueString(),
		AllowedModels: allowedModels,
		APIKey:        plan.APIKey.ValueString(),
	}
	if !plan.InferenceGeography.IsNull() && !plan.InferenceGeography.IsUnknown() {
		geo := plan.InferenceGeography.ValueString()
		createReq.InferenceGeography = &geo
	}

	credentialUUID, err := apiv1.CreateAiProviderCredentialV1(r.client, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating AI provider credential", err.Error())
		return
	}

	if plan.IsDefault.ValueBool() {
		if err := apiv1.SetDefaultAiProviderCredentialV1(r.client, credentialUUID); err != nil {
			resp.Diagnostics.AddError("Error setting default AI provider credential", err.Error())
			return
		}
	}

	plan.ID = types.StringValue(credentialUUID)
	plan.CredentialUUID = types.StringValue(credentialUUID)
	resp.Diagnostics.Append(setAiProviderCredentialFromAPI(ctx, &plan, credentialUUID, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aiProviderCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state aiProviderCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialUUID := state.CredentialUUID.ValueString()
	if credentialUUID == "" {
		credentialUUID = state.ID.ValueString()
	}

	service := services.NewAiProviderCredentialsService(r.client)
	cred, err := service.GetByUUID(ctx, credentialUUID)
	if err != nil {
		if errors.Is(err, services.ErrAiProviderCredentialNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("AI provider credential %s not found during Read, removing from state", credentialUUID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading AI provider credential", err.Error())
		return
	}

	state.ID = types.StringValue(cred.UUID)
	state.CredentialUUID = types.StringValue(cred.UUID)
	state.AIProvider = types.StringValue(cred.Provider)
	state.Label = types.StringValue(cred.Label)
	state.Region = types.StringValue(cred.Region)
	state.APIKeyHint = types.StringValue(cred.APIKeyHint)
	state.IsDefault = types.BoolValue(cred.IsDefault)
	if cred.InferenceGeography != "" {
		state.InferenceGeography = types.StringValue(cred.InferenceGeography)
	} else {
		state.InferenceGeography = types.StringNull()
	}

	allowedModels, diags := stringSliceToStringList(ctx, cred.AllowedModels)
	resp.Diagnostics.Append(diags...)
	state.AllowedModels = allowedModels

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *aiProviderCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state aiProviderCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateAiProviderCredentialProvider(plan.AIProvider.ValueString())...)
	resp.Diagnostics.Append(validateInferenceGeography(plan.InferenceGeography)...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialUUID := state.CredentialUUID.ValueString()
	allowedModels, diags := stringListToStringSlice(ctx, plan.AllowedModels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(allowedModels) == 0 {
		resp.Diagnostics.AddError("Invalid allowed_models", "allowed_models must contain at least one model.")
		return
	}

	updateReq := models.UpdateAiProviderCredentialRequest{}
	if plan.Label.ValueString() != state.Label.ValueString() {
		label := plan.Label.ValueString()
		updateReq.Label = &label
	}
	if plan.Region.ValueString() != state.Region.ValueString() {
		region := plan.Region.ValueString()
		updateReq.Region = &region
	}
	if !slices.Equal(allowedModels, mustStringListFromState(ctx, state.AllowedModels)) {
		updateReq.AllowedModels = allowedModels
	}
	if !plan.InferenceGeography.Equal(state.InferenceGeography) &&
		!plan.InferenceGeography.IsNull() && !plan.InferenceGeography.IsUnknown() {
		geo := plan.InferenceGeography.ValueString()
		updateReq.InferenceGeography = &geo
	}

	apiKeyChanged := !plan.APIKey.Equal(state.APIKey) && !plan.APIKey.IsNull() && !plan.APIKey.IsUnknown()
	if apiKeyChanged {
		key := plan.APIKey.ValueString()
		updateReq.APIKey = &key
	}

	hasPatch := updateReq.Label != nil || updateReq.Region != nil || updateReq.AllowedModels != nil ||
		updateReq.APIKey != nil || updateReq.InferenceGeography != nil
	if hasPatch {
		if err := apiv1.UpdateAiProviderCredentialV1(r.client, credentialUUID, updateReq); err != nil {
			resp.Diagnostics.AddError("Error updating AI provider credential", err.Error())
			return
		}
	}

	if plan.IsDefault.ValueBool() && !state.IsDefault.ValueBool() {
		if err := apiv1.SetDefaultAiProviderCredentialV1(r.client, credentialUUID); err != nil {
			resp.Diagnostics.AddError("Error setting default AI provider credential", err.Error())
			return
		}
	}

	plan.ID = state.ID
	plan.CredentialUUID = state.CredentialUUID
	if !apiKeyChanged {
		plan.APIKey = state.APIKey
	}

	resp.Diagnostics.Append(setAiProviderCredentialFromAPI(ctx, &plan, credentialUUID, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aiProviderCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state aiProviderCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DeleteProtection.ValueBool() {
		resp.Diagnostics.AddError(
			"Deletion Protection Enabled",
			"Cannot delete AI provider credential because deletion_protection is set to true. Set deletion_protection to false to allow deletion.",
		)
		return
	}

	credentialUUID := state.CredentialUUID.ValueString()
	if credentialUUID == "" {
		credentialUUID = state.ID.ValueString()
	}

	service := services.NewAiProviderCredentialsService(r.client)
	cred, err := service.GetByUUID(ctx, credentialUUID)
	if err != nil {
		if errors.Is(err, services.ErrAiProviderCredentialNotFound) {
			return
		}
		resp.Diagnostics.AddError("Error reading AI provider credential before delete", err.Error())
		return
	}
	if cred.IsDefault {
		list, listErr := service.List(ctx)
		if listErr != nil {
			resp.Diagnostics.AddError("Error listing AI provider credentials before delete", listErr.Error())
			return
		}
		if len(list.Credentials) > 1 {
			resp.Diagnostics.AddError(
				"Cannot Delete Default AI Provider Credential",
				fmt.Sprintf(
					"Credential %s is the organization default. Set is_default = true on another lightdash_ai_provider_credential, apply, then destroy this resource. Lightdash does not support clearing the organization default via the API.",
					credentialUUID,
				),
			)
			return
		}
	}

	if err := apiv1.DeleteAiProviderCredentialV1(r.client, credentialUUID); err != nil {
		summary, detail := aiProviderCredentialDeleteErrorMessage(credentialUUID, err)
		resp.Diagnostics.AddError(summary, detail)
		return
	}
}

func (r *aiProviderCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	credentialUUID := req.ID
	if credentialUUID == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be the credential UUID.")
		return
	}

	organization, err := apiv1.GetMyOrganizationV1(r.client)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read organization", err.Error())
		return
	}

	service := services.NewAiProviderCredentialsService(r.client)
	cred, err := service.GetByUUID(ctx, credentialUUID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading AI provider credential for import", err.Error())
		return
	}

	state := aiProviderCredentialResourceModel{
		ID:               types.StringValue(cred.UUID),
		OrganizationUUID: types.StringValue(organization.OrganizationUUID),
		CredentialUUID:   types.StringValue(cred.UUID),
		AIProvider:       types.StringValue(cred.Provider),
		Label:            types.StringValue(cred.Label),
		Region:           types.StringValue(cred.Region),
		APIKeyHint:       types.StringValue(cred.APIKeyHint),
		IsDefault:        types.BoolValue(cred.IsDefault),
		DeleteProtection: types.BoolValue(true),
		APIKey:           types.StringNull(),
	}
	if cred.InferenceGeography != "" {
		state.InferenceGeography = types.StringValue(cred.InferenceGeography)
	}
	allowedModels, diags := stringSliceToStringList(ctx, cred.AllowedModels)
	resp.Diagnostics.Append(diags...)
	state.AllowedModels = allowedModels

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setAiProviderCredentialFromAPI(ctx context.Context, model *aiProviderCredentialResourceModel, credentialUUID string, client *api.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	service := services.NewAiProviderCredentialsService(client)
	cred, err := service.GetByUUID(ctx, credentialUUID)
	if err != nil {
		diags.AddError("Error reading AI provider credential", err.Error())
		return diags
	}

	model.APIKeyHint = types.StringValue(cred.APIKeyHint)
	model.IsDefault = types.BoolValue(cred.IsDefault)
	model.Label = types.StringValue(cred.Label)
	model.Region = types.StringValue(cred.Region)
	model.AIProvider = types.StringValue(cred.Provider)
	if cred.InferenceGeography != "" {
		model.InferenceGeography = types.StringValue(cred.InferenceGeography)
	}

	allowedModels, listDiags := stringSliceToStringList(ctx, cred.AllowedModels)
	diags.Append(listDiags...)
	model.AllowedModels = allowedModels

	return diags
}

func validateAiProviderCredentialProvider(provider string) diag.Diagnostics {
	if provider != models.AiProviderBedrock {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid provider",
				fmt.Sprintf("provider must be %q, got %q", models.AiProviderBedrock, provider),
			),
		}
	}
	return nil
}

func validateInferenceGeography(value types.String) diag.Diagnostics {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	geo := value.ValueString()
	if geo == "" {
		return nil
	}
	if !slices.Contains(models.ValidBedrockInferenceGeographies, geo) {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid inference_geography",
				fmt.Sprintf("inference_geography must be one of %v, got %q", models.ValidBedrockInferenceGeographies, geo),
			),
		}
	}
	return nil
}

func mustStringListFromState(ctx context.Context, list types.List) []string {
	values, diags := stringListToStringSlice(ctx, list)
	if diags.HasError() {
		return nil
	}
	return values
}

func aiProviderCredentialDeleteErrorMessage(credentialUUID string, err error) (summary string, detail string) {
	if err == nil {
		return "", ""
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "default"):
		return "Cannot Delete Default AI Provider Credential", fmt.Sprintf(
			"Credential %s appears to be the organization default. Set is_default = true on another lightdash_ai_provider_credential, apply, then destroy this resource.",
			credentialUUID,
		)
	case strings.Contains(lower, "in use"), strings.Contains(lower, "project"):
		return "Cannot Delete AI Provider Credential In Use", fmt.Sprintf(
			"Credential %s may still be selected by a project. Reset lightdash_project_ai_provider_credential bindings or choose another credential in the UI, then retry. API error: %s",
			credentialUUID, msg,
		)
	default:
		return "Error deleting AI provider credential", msg
	}
}
