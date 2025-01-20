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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

// Ensure LightdashProvider satisfies various provider interfaces.
var _ provider.Provider = &lightdashProvider{}

// lightdashProvider defines the provider implementation.
type lightdashProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// lightdashProviderModel describes the provider data model.
type lightdashProviderModel struct {
	HostURL               types.String `tfsdk:"host"`
	Token                 types.String `tfsdk:"token"`
	MaxConcurrentRequests types.Int64  `tfsdk:"max_concurrent_requests"`
	RequestTimeout        types.Int64  `tfsdk:"request_timeout"`
	RetryTimes            types.Int64  `tfsdk:"retry_times"`
}

func (p *lightdashProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "lightdash"
	resp.Version = p.version
}

func (p *lightdashProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Terraform provider for Lightdash",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Lightdash Host",
				Required:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Personal access token for Lightdash",
				Required:            true,
				Sensitive:           true,
			},
			"max_concurrent_requests": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of concurrent requests to the Lightdash API",
				Optional:            true,
			},
			"request_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for requests to the Lightdash API in seconds",
				Optional:            true,
			},
			"retry_times": schema.Int64Attribute{
				MarkdownDescription: "Number of times to retry requests to the Lightdash API",
				Optional:            true,
			},
		},
	}
}

func (p *lightdashProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config lightdashProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate configuration
	if config.HostURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Lightdash API Host",
			"Please set the `host` attribute to the Lightdash API Host.",
		)
		return
	}
	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Lightdash API Token",
			"Please set the `token` attribute to the Lightdash API Token.",
		)
		return
	}

	// Configuration values are now available.
	host := config.HostURL.ValueString()
	token := config.Token.ValueString()

	// Set default values if not provided in the configuration
	maxConcurrentRequests := int64(10) // Default value
	if !config.MaxConcurrentRequests.IsNull() {
		maxConcurrentRequests = config.MaxConcurrentRequests.ValueInt64()
	}

	requestTimeout := int64(180) // Default value
	if !config.RequestTimeout.IsNull() {
		requestTimeout = config.RequestTimeout.ValueInt64()
	}

	retryTimes := int64(3) // Default value
	if !config.RetryTimes.IsNull() {
		retryTimes = config.RetryTimes.ValueInt64()
	}

	client, _ := api.NewClient(&host, &token, maxConcurrentRequests, requestTimeout, retryTimes)

	// Check if the token is valid as long as the test mode is not disabled
	if !isIntegrationTestMode() {
		_, err := client.GetMyOrganizationV1()
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("token"),
				"Invalid Lightdash API Token",
				"Please set the valid `token` attribute to the Lightdash API Token.",
			)
			return
		}
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *lightdashProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganizationRoleMemberResource,
		NewProjectRoleMemberResource,
		NewSpaceResource,
		NewGroupResource,
		NewProjectRoleGroupResource,
		NewProjectSchedulerSettingsResource,
	}
}

func (p *lightdashProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAuthenticatedUserDataSource,
		NewOrganizationDataSource,
		NewOrganizationMembersDataSource,
		NewOrganizationMemberDataSource,
		NewProjectsDataSource,
		NewProjectDataSource,
		NewProjectMembersDataSource,
		NewSpacesDataSource,
		NewOrganizationGroupsDataSource,
		NewProjectGroupAccessesDataSource,
		NewGroupDataSource,
		NewGroupMembersDataSource,
		NewProjectSchedulerSettingsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &lightdashProvider{
			version: version,
		}
	}
}
