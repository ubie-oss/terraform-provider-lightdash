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

	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

var (
	_ resource.Resource                = &oauthApplicationResource{}
	_ resource.ResourceWithConfigure   = &oauthApplicationResource{}
	_ resource.ResourceWithImportState = &oauthApplicationResource{}
)

func NewOAuthApplicationResource() resource.Resource {
	return &oauthApplicationResource{}
}

type oauthApplicationResource struct {
	client *api.Client
}

type oauthApplicationResourceModel struct {
	ID                types.String `tfsdk:"id"`
	OrganizationUUID  types.String `tfsdk:"organization_uuid"`
	ClientID          types.String `tfsdk:"client_id"`
	ClientName        types.String `tfsdk:"client_name"`
	RedirectURIs      types.Set    `tfsdk:"redirect_uris"`
	ClientSecret      types.String `tfsdk:"client_secret"`
	Scopes            types.List   `tfsdk:"scopes"`
	CreatedAt         types.String `tfsdk:"created_at"`
	CreatedByUserUUID types.String `tfsdk:"created_by_user_uuid"`
	DeleteProtection  types.Bool   `tfsdk:"deletion_protection"`
}

func (r *oauthApplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_application"
}

func (r *oauthApplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_lightdash_oauth_application.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Manages a Lightdash OAuth application",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource identifier. It is computed as `organizations/<organization_uuid>/oauth_applications/<client_id>`.",
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
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The OAuth client ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the OAuth application.",
				Required:            true,
				Validators: []validator.String{
					ValidateNonEmptyString{},
				},
			},
			"redirect_uris": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Allowed OAuth redirect URIs. Updates replace the entire list in Lightdash.",
				Required:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The OAuth client secret. Returned only at creation; not refreshed on read.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"scopes": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "OAuth scopes stored on the client.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "ISO 8601 timestamp when the client was created.",
				Computed:            true,
			},
			"created_by_user_uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the user who created the client, if known.",
				Computed:            true,
			},
			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: "When set to `true`, prevents the destruction of the OAuth application resource by Terraform.",
				Required:            true,
			},
		},
	}
}

func (r *oauthApplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *oauthApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan oauthApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateOrganizationUUID(r.client, plan.OrganizationUUID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	redirectURIs, diags := stringSetToStringSlice(ctx, plan.RedirectURIs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := apiv1.CreateOAuthClientV1(r.client, plan.ClientName.ValueString(), redirectURIs)
	if err != nil {
		resp.Diagnostics.AddError("Error creating OAuth application", err.Error())
		return
	}

	plan.ClientID = types.StringValue(created.ClientID)
	if created.ClientSecret == "" {
		resp.Diagnostics.AddError(
			"Error creating OAuth application",
			"The Lightdash API did not return a client secret. The application may have been created but cannot be managed without rotating the secret in Lightdash.",
		)
		return
	}
	plan.ClientSecret = types.StringValue(created.ClientSecret)
	plan.ID = types.StringValue(getOAuthApplicationResourceID(plan.OrganizationUUID.ValueString(), created.ClientID))
	resp.Diagnostics.Append(setOAuthApplicationResourceFromClient(ctx, &plan, *created)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oauthApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state oauthApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := services.NewOAuthApplicationsService(r.client)
	client, err := service.GetByClientID(ctx, state.ClientID.ValueString())
	if err != nil {
		if errors.Is(err, services.ErrOAuthApplicationNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("OAuth application %s not found during Read, removing from state", state.ClientID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading OAuth application", err.Error())
		return
	}

	resp.Diagnostics.Append(setOAuthApplicationResourceFromClient(ctx, &state, *client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *oauthApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state oauthApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	redirectURIs, diags := stringSetToStringSlice(ctx, plan.RedirectURIs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := apiv1.UpdateOAuthClientV1(
		r.client,
		state.ClientID.ValueString(),
		plan.ClientName.ValueString(),
		redirectURIs,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating OAuth application", err.Error())
		return
	}

	plan.ID = state.ID
	plan.OrganizationUUID = state.OrganizationUUID
	plan.ClientID = state.ClientID
	plan.ClientSecret = state.ClientSecret
	resp.Diagnostics.Append(setOAuthApplicationResourceFromClient(ctx, &plan, *updated)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oauthApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state oauthApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Deleting OAuth application %s", state.ClientID.ValueString()))

	if state.DeleteProtection.ValueBool() {
		resp.Diagnostics.AddError(
			"Deletion Protection Enabled",
			"Cannot delete OAuth application because deletion_protection is set to true. Set deletion_protection to false to allow deletion.",
		)
		return
	}

	if err := apiv1.DeleteOAuthClientV1(r.client, state.ClientID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting OAuth application", err.Error())
		return
	}
}

func (r *oauthApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	extracted, err := extractOAuthApplicationResourceID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error extracting resource ID", err.Error())
		return
	}
	organizationUUID := extracted[0]
	clientID := extracted[1]

	resp.Diagnostics.Append(validateOrganizationUUID(r.client, organizationUUID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := services.NewOAuthApplicationsService(r.client)
	client, err := service.GetByClientID(ctx, clientID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading OAuth application for import", err.Error())
		return
	}

	state := oauthApplicationResourceModel{
		ID:               types.StringValue(req.ID),
		OrganizationUUID: types.StringValue(organizationUUID),
		ClientID:         types.StringValue(client.ClientID),
		DeleteProtection: types.BoolValue(true),
	}
	resp.Diagnostics.Append(setOAuthApplicationResourceFromClient(ctx, &state, *client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func validateOrganizationUUID(client *api.Client, organizationUUID string) diag.Diagnostics {
	organization, err := apiv1.GetMyOrganizationV1(client)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to read organization", err.Error()),
		}
	}
	if organization.OrganizationUUID != organizationUUID {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Organization UUID mismatch",
				fmt.Sprintf(
					"organization_uuid %q does not match the organization for the configured API token (%q)",
					organizationUUID,
					organization.OrganizationUUID,
				),
			),
		}
	}
	return nil
}

func setOAuthApplicationResourceFromClient(ctx context.Context, model *oauthApplicationResourceModel, client apiv1.OAuthClientV1) diag.Diagnostics {
	var diags diag.Diagnostics

	redirectURIs, redirectDiags := stringSliceToStringSet(ctx, client.RedirectURIs)
	diags.Append(redirectDiags...)

	scopes, scopesDiags := stringSliceToStringList(ctx, client.Scopes)
	diags.Append(scopesDiags...)

	createdBy := types.StringNull()
	if client.CreatedByUserUUID != nil {
		createdBy = types.StringValue(*client.CreatedByUserUUID)
	}

	model.ClientName = types.StringValue(client.ClientName)
	model.RedirectURIs = redirectURIs
	model.Scopes = scopes
	model.CreatedAt = types.StringValue(client.CreatedAt)
	model.CreatedByUserUUID = createdBy

	return diags
}

func getOAuthApplicationResourceID(organizationUUID string, clientID string) string {
	return fmt.Sprintf("organizations/%s/oauth_applications/%s", organizationUUID, clientID)
}

func extractOAuthApplicationResourceID(input string) ([]string, error) {
	pattern := `^organizations/([^/]+)/oauth_applications/([^/]+)$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}
	return []string{groups[0], groups[1]}, nil
}
