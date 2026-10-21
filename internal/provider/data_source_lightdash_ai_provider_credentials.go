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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

var (
	_ datasource.DataSource              = &aiProviderCredentialsDataSource{}
	_ datasource.DataSourceWithConfigure = &aiProviderCredentialsDataSource{}
)

func NewAiProviderCredentialsDataSource() datasource.DataSource {
	return &aiProviderCredentialsDataSource{}
}

type aiProviderCredentialsDataSource struct {
	client *api.Client
}

type aiProviderCredentialListItemModel struct {
	CredentialUUID     types.String `tfsdk:"credential_uuid"`
	AIProvider         types.String `tfsdk:"ai_provider"`
	Label              types.String `tfsdk:"label"`
	Region             types.String `tfsdk:"region"`
	AllowedModels      types.List   `tfsdk:"allowed_models"`
	APIKeyHint         types.String `tfsdk:"api_key_hint"`
	IsDefault          types.Bool   `tfsdk:"is_default"`
	InferenceGeography types.String `tfsdk:"inference_geography"`
}

type aiProviderCredentialsDataSourceModel struct {
	ID               types.String                        `tfsdk:"id"`
	OrganizationUUID types.String                        `tfsdk:"organization_uuid"`
	Credentials      []aiProviderCredentialListItemModel `tfsdk:"credentials"`
	LegacyBedrock    types.Object                        `tfsdk:"legacy_bedrock"`
}

func (d *aiProviderCredentialsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ai_provider_credentials"
}

func (d *aiProviderCredentialsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_lightdash_ai_provider_credentials.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Lightdash organization AI provider credentials data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier: `organizations/<organization_uuid>/ai_provider_credentials`.",
				Computed:            true,
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
				Required:            true,
			},
			"credentials": schema.ListNestedAttribute{
				MarkdownDescription: "Named AI provider credentials in the organization.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"credential_uuid": schema.StringAttribute{
							Computed: true,
						},
						"ai_provider": schema.StringAttribute{
							Computed: true,
						},
						"label": schema.StringAttribute{
							Computed: true,
						},
						"region": schema.StringAttribute{
							Computed: true,
						},
						"allowed_models": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"api_key_hint": schema.StringAttribute{
							Computed: true,
						},
						"is_default": schema.BoolAttribute{
							Computed: true,
						},
						"inference_geography": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"legacy_bedrock": schema.SingleNestedAttribute{
				MarkdownDescription: "Legacy Bedrock configuration when present.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"api_key_hint": schema.StringAttribute{
						Computed: true,
					},
					"region": schema.StringAttribute{
						Computed: true,
					},
					"allowed_models": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
		},
	}
}

func (d *aiProviderCredentialsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *aiProviderCredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config aiProviderCredentialsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateOrganizationUUID(d.client, config.OrganizationUUID.ValueString())...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := services.NewAiProviderCredentialsService(d.client)
	list, err := service.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list AI provider credentials", err.Error())
		return
	}

	credentials := make([]aiProviderCredentialListItemModel, 0, len(list.Credentials))
	for _, cred := range list.Credentials {
		item, diags := aiProviderCredentialToListItemModel(ctx, cred)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		credentials = append(credentials, item)
	}

	sort.Slice(credentials, func(i, j int) bool {
		return credentials[i].CredentialUUID.ValueString() < credentials[j].CredentialUUID.ValueString()
	})

	state := aiProviderCredentialsDataSourceModel{
		ID:               types.StringValue(fmt.Sprintf("organizations/%s/ai_provider_credentials", config.OrganizationUUID.ValueString())),
		OrganizationUUID: config.OrganizationUUID,
		Credentials:      credentials,
	}

	if list.LegacyBedrock != nil {
		legacyObj, diags := legacyBedrockToObject(ctx, *list.LegacyBedrock)
		resp.Diagnostics.Append(diags...)
		state.LegacyBedrock = legacyObj
	} else {
		state.LegacyBedrock = types.ObjectNull(legacyBedrockAttrTypes())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func aiProviderCredentialToListItemModel(ctx context.Context, cred models.AiProviderCredential) (aiProviderCredentialListItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	allowedModels, listDiags := stringSliceToStringList(ctx, cred.AllowedModels)
	diags.Append(listDiags...)

	inferenceGeography := types.StringNull()
	if cred.InferenceGeography != "" {
		inferenceGeography = types.StringValue(cred.InferenceGeography)
	}

	return aiProviderCredentialListItemModel{
		CredentialUUID:     types.StringValue(cred.UUID),
		AIProvider:         types.StringValue(cred.Provider),
		Label:              types.StringValue(cred.Label),
		Region:             types.StringValue(cred.Region),
		AllowedModels:      allowedModels,
		APIKeyHint:         types.StringValue(cred.APIKeyHint),
		IsDefault:          types.BoolValue(cred.IsDefault),
		InferenceGeography: inferenceGeography,
	}, diags
}

func legacyBedrockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"api_key_hint":   types.StringType,
		"region":         types.StringType,
		"allowed_models": types.ListType{ElemType: types.StringType},
	}
}

func legacyBedrockToObject(ctx context.Context, legacy models.LegacyBedrockCredential) (types.Object, diag.Diagnostics) {
	allowedModels, diags := stringSliceToStringList(ctx, legacy.AllowedModels)
	if diags.HasError() {
		return types.ObjectNull(legacyBedrockAttrTypes()), diags
	}

	obj, objDiags := types.ObjectValue(legacyBedrockAttrTypes(), map[string]attr.Value{
		"api_key_hint":   types.StringValue(legacy.APIKeyHint),
		"region":         types.StringValue(legacy.Region),
		"allowed_models": allowedModels,
	})
	diags.Append(objDiags...)
	return obj, diags
}
