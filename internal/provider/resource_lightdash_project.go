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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

// projectResource defines the resource implementation.
type projectResource struct {
	client *api.Client
}

// projectResourceModel describes the resource data model.
type projectResourceModel struct {
	ID                                                 types.String         `tfsdk:"id"`
	OrganizationUUID                                   types.String         `tfsdk:"organization_uuid"`
	ProjectUUID                                        types.String         `tfsdk:"project_uuid"`
	Name                                               types.String         `tfsdk:"name"`
	Type                                               models.ProjectType   `tfsdk:"type"`
	DbtConnectionType                                  types.String         `tfsdk:"dbt_connection_type"`
	DbtConnectionRepository                            types.String         `tfsdk:"dbt_connection_repository"`
	DbtConnectionBranch                                types.String         `tfsdk:"dbt_connection_branch"`
	DbtConnectionProjectSubPath                        types.String         `tfsdk:"dbt_connection_project_sub_path"`
	DbtConnectionHostDomain                            types.String         `tfsdk:"dbt_connection_host_domain"`
	WarehouseConnectionType                            models.WarehouseType `tfsdk:"warehouse_connection_type"`
	DatabricksConnectionServerHostName                 types.String         `tfsdk:"databricks_connection_server_host_name"`
	DatabricksConnectionHTTPPath                       types.String         `tfsdk:"databricks_connection_http_path"`
	DatabricksConnectionPersonalAccessToken            types.String         `tfsdk:"databricks_connection_personal_access_token"`
	DatabricksConnectionCatalog                        types.String         `tfsdk:"databricks_connection_catalog"`
	SnowflakeWarehouseConnectionAccount                types.String         `tfsdk:"snowflake_warehouse_connection_account"`
	SnowflakeWarehouseConnectionRole                   types.String         `tfsdk:"snowflake_warehouse_connection_role"`
	SnowflakeWarehouseConnectionDatabase               types.String         `tfsdk:"snowflake_warehouse_connection_database"`
	SnowflakeWarehouseConnectionSchema                 types.String         `tfsdk:"snowflake_warehouse_connection_schema"`
	SnowflakeWarehouseConnectionWarehouse              types.String         `tfsdk:"snowflake_warehouse_connection_warehouse"`
	SnowflakeWarehouseConnectionThreads                types.Int32          `tfsdk:"snowflake_warehouse_connection_threads"`
	SnowflakeWarehouseConnectionClientSessionKeepAlive types.Bool           `tfsdk:"snowflake_warehouse_connection_client_session_keep_alive"`
}

type projectMemberModelForProject struct {
	UserUUID types.String `tfsdk:"user_uuid"`
}

func (r *projectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for managing Lightdash projects, " +
			"their members and groups can be managed via the corresponding resources. ",
		Description: "Manages a Lightdash project",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier for the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization to which the project belongs.",
				Required:            true,
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash project.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Lightdash project.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of project to create, either DEFAULT or DEVELOPMENT",
				Required:            true,
			},
			"dbt_connection_type": schema.StringAttribute{
				MarkdownDescription: "dbt project connection type, currently only support 'github', which is the default",
				Required:            false,
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("github"),
			},
			"dbt_connection_repository": schema.StringAttribute{
				MarkdownDescription: "Repository name in <org>/<repo> format",
				Required:            true,
			},
			"dbt_connection_branch": schema.StringAttribute{
				MarkdownDescription: "Branch to use, default 'main'",
				Required:            false,
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("main"),
			},
			"dbt_connection_project_sub_path": schema.StringAttribute{
				MarkdownDescription: "Sub path to find the project in the repo, default '/'",
				Required:            false,
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("/"),
			},
			"dbt_connection_host_domain": schema.StringAttribute{
				MarkdownDescription: "Host domain of the repo, default 'github.com'",
				Required:            false,
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("github.com"),
			},
			// TODO: Convert warehouse connection to nested attribute
			"warehouse_connection_type": schema.StringAttribute{
				MarkdownDescription: "Type of warehouse to connect to, must be one of 'snowflake' or 'databricks'",
				Optional:            true,
			},
			"databricks_connection_server_host_name": schema.StringAttribute{
				MarkdownDescription: "Databricks - Server host name for connection",
				Optional:            true,
			},
			"databricks_connection_http_path": schema.StringAttribute{
				MarkdownDescription: "Databricks - HTTP path for connection",
				Optional:            true,
			},
			"databricks_connection_personal_access_token": schema.StringAttribute{
				MarkdownDescription: "Databricks - Personal access token for connection (warning: will store token in state file!)",
				Optional:            true,
			},
			"databricks_connection_catalog": schema.StringAttribute{
				MarkdownDescription: "Databricks - Catalog name for connection",
				Optional:            true,
			},
			"snowflake_warehouse_connection_account": schema.StringAttribute{
				MarkdownDescription: "Snowflake - Account identifier, including region/ cloud path",
				Optional:            true,
			},
			"snowflake_warehouse_connection_role": schema.StringAttribute{
				MarkdownDescription: "Snowflake - Role to connect to the warehouse with",
				Optional:            true,
			},
			"snowflake_warehouse_connection_database": schema.StringAttribute{
				MarkdownDescription: "Snowflake - Database to connect to",
				Optional:            true,
			},
			"snowflake_warehouse_connection_schema": schema.StringAttribute{
				MarkdownDescription: "Snowflake - Schema to connect to",
				Optional:            true,
			},
			"snowflake_warehouse_connection_warehouse": schema.StringAttribute{
				MarkdownDescription: "Snowflake - Warehouse to use",
				Optional:            true,
			},
			"snowflake_warehouse_connection_client_session_keep_alive": schema.BoolAttribute{
				MarkdownDescription: "Snowflake - Client session keep alive param",
				Optional:            true,
			},
			"snowflake_warehouse_connection_threads": schema.Int32Attribute{
				MarkdownDescription: "Snowflake - Number of threads to use",
				Optional:            true,
			},
		},
	}
}

func (r *projectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new project
	organization_uuid := plan.OrganizationUUID.ValueString()
	project_name := plan.Name.ValueString()
	project_type := plan.Type
	dbt_connection_type := plan.DbtConnectionType.ValueString()
	dbt_connection_repository := plan.DbtConnectionRepository.ValueString()
	dbt_connection_branch := plan.DbtConnectionBranch.ValueString()
	dbt_connection_project_sub_path := plan.DbtConnectionProjectSubPath.ValueString()
	dbt_connection_host_domain := plan.DbtConnectionHostDomain.ValueString()
	warehouse_connection_type := plan.WarehouseConnectionType
	databricks_connection_server_host_name := plan.DatabricksConnectionServerHostName.ValueString()
	databricks_connection_http_path := plan.DatabricksConnectionHTTPPath.ValueString()
	databricks_connection_personal_access_token := plan.DatabricksConnectionPersonalAccessToken.ValueString()
	databricks_connection_catalog := plan.DatabricksConnectionCatalog.ValueString()
	snowflake_warehouse_connection_account := plan.SnowflakeWarehouseConnectionAccount.ValueString()
	snowflake_warehouse_connection_role := plan.SnowflakeWarehouseConnectionRole.ValueString()
	snowflake_warehouse_connection_database := plan.SnowflakeWarehouseConnectionDatabase.ValueString()
	snowflake_warehouse_connection_schema := plan.SnowflakeWarehouseConnectionSchema.ValueString()
	snowflake_warehouse_connection_warehouse := plan.SnowflakeWarehouseConnectionWarehouse.ValueString()
	snowflake_warehouse_connection_threads := plan.SnowflakeWarehouseConnectionThreads.ValueInt32()
	snowflake_warehouse_connection_client_session_keep_alive := plan.SnowflakeWarehouseConnectionClientSessionKeepAlive.ValueBool()

	if !models.ProjectType.IsValid(project_type) {
		resp.Diagnostics.AddError(
			"invalid project type: %s", string(project_type),
		)
		return
	}
	dbtConnection := api.DbtConnection{
		Type:           dbt_connection_type,
		Repository:     dbt_connection_repository,
		Branch:         dbt_connection_branch,
		ProjectSubPath: dbt_connection_project_sub_path,
		HostDomain:     dbt_connection_host_domain,
	}

	if !models.WarehouseType.IsValidWarehouseType(warehouse_connection_type) {
		resp.Diagnostics.AddError(
			"invalid warehouse type: %s", string(warehouse_connection_type),
		)
		return
	}
	warehouseConnection := api.WarehouseConnection{
		Type: warehouse_connection_type,
	}
	if warehouseConnection.Type == models.SNOWFLAKE {
		warehouseConnection.Account = snowflake_warehouse_connection_account
		warehouseConnection.Role = snowflake_warehouse_connection_role
		warehouseConnection.Database = snowflake_warehouse_connection_database
		warehouseConnection.Warehouse = snowflake_warehouse_connection_warehouse
		warehouseConnection.Schema = snowflake_warehouse_connection_schema
		warehouseConnection.ClientSessionKeepAlive = snowflake_warehouse_connection_client_session_keep_alive
		warehouseConnection.Threads = snowflake_warehouse_connection_threads
	}
	if warehouseConnection.Type == models.DATABRICKS {
		warehouseConnection.ServerHostName = databricks_connection_server_host_name
		warehouseConnection.HTTPPath = databricks_connection_http_path
		warehouseConnection.PersonalAccessToken = databricks_connection_personal_access_token
		warehouseConnection.Catalog = databricks_connection_catalog
	}

	createdProject, err := r.client.CreateProjectV1(organization_uuid, project_name, project_type, dbtConnection, warehouseConnection)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			"Could not create project, unexpected error: "+err.Error(),
		)
		return
	}

	// Assign the plan values to the state
	stateId := getProjectResourceId(organization_uuid, createdProject.ProjectUUID)
	plan.ID = types.StringValue(stateId)
	plan.ProjectUUID = types.StringValue(createdProject.ProjectUUID)
	plan.Name = types.StringValue(createdProject.Name)
	plan.Type = createdProject.Type
	plan.DbtConnectionType = types.StringValue(createdProject.DbtConnection.Type)
	plan.DbtConnectionRepository = types.StringValue(createdProject.DbtConnection.Repository)
	plan.DbtConnectionBranch = types.StringValue(createdProject.DbtConnection.Branch)
	plan.DbtConnectionProjectSubPath = types.StringValue(createdProject.DbtConnection.ProjectSubPath)
	plan.DbtConnectionHostDomain = types.StringValue(createdProject.DbtConnection.HostDomain)
	plan.WarehouseConnectionType = createdProject.WarehouseConnection.Type
	if warehouseConnection.Type == models.SNOWFLAKE {
		plan.SnowflakeWarehouseConnectionAccount = types.StringValue(createdProject.WarehouseConnection.Account)
		plan.SnowflakeWarehouseConnectionRole = types.StringValue(createdProject.WarehouseConnection.Role)
		plan.SnowflakeWarehouseConnectionDatabase = types.StringValue(createdProject.WarehouseConnection.Database)
		plan.SnowflakeWarehouseConnectionWarehouse = types.StringValue(createdProject.WarehouseConnection.Warehouse)
		plan.SnowflakeWarehouseConnectionSchema = types.StringValue(createdProject.WarehouseConnection.Schema)
		plan.SnowflakeWarehouseConnectionClientSessionKeepAlive = types.BoolValue(createdProject.WarehouseConnection.ClientSessionKeepAlive)
		plan.SnowflakeWarehouseConnectionThreads = types.Int32Value(createdProject.WarehouseConnection.Threads)
	}
	if warehouseConnection.Type == models.DATABRICKS {
		plan.DatabricksConnectionServerHostName = types.StringValue(createdProject.WarehouseConnection.ServerHostName)
		plan.DatabricksConnectionHTTPPath = types.StringValue(createdProject.WarehouseConnection.HTTPPath)
		plan.DatabricksConnectionPersonalAccessToken = types.StringValue(databricks_connection_personal_access_token)
		plan.DatabricksConnectionCatalog = types.StringValue(createdProject.WarehouseConnection.Catalog)
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Declare variables to import from state
	var organizationUuid string
	var projectUuid string
	var projectName string
	var projectType string

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("organization_uuid"), &organizationUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project_uuid"), &projectUuid)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("name"), &projectName)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("type"), &projectType)...)

	// Get current state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project
	project, err := r.client.GetProjectV1(projectUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading project",
			"Could not read project ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Set the state values
	state.OrganizationUUID = types.StringValue(project.OrganizationUUID)
	state.ProjectUUID = types.StringValue(project.ProjectUUID)
	state.Name = types.StringValue(project.ProjectName)
	if project.ProjectType == string(models.DEFAULT_PROJECT_TYPE) {
		state.Type = models.DEFAULT_PROJECT_TYPE
	}
	if project.ProjectType == string(models.PREVIEW_PROJECT_TYPE) {
		state.Type = models.PREVIEW_PROJECT_TYPE
	}
	state.DbtConnectionType = types.StringValue(project.DbtConnection.Type)
	state.DbtConnectionRepository = types.StringValue(project.DbtConnection.Repository)
	state.DbtConnectionBranch = types.StringValue(project.DbtConnection.Branch)
	state.DbtConnectionProjectSubPath = types.StringValue(project.DbtConnection.ProjectSubPath)
	state.DbtConnectionHostDomain = types.StringValue(project.DbtConnection.HostDomain)
	state.WarehouseConnectionType = project.WarehouseConnection.Type

	state.SnowflakeWarehouseConnectionAccount = types.StringValue(project.WarehouseConnection.Account)
	state.SnowflakeWarehouseConnectionRole = types.StringValue(project.WarehouseConnection.Role)
	state.SnowflakeWarehouseConnectionDatabase = types.StringValue(project.WarehouseConnection.Database)
	state.SnowflakeWarehouseConnectionWarehouse = types.StringValue(project.WarehouseConnection.Warehouse)
	state.SnowflakeWarehouseConnectionSchema = types.StringValue(project.WarehouseConnection.Schema)
	state.SnowflakeWarehouseConnectionClientSessionKeepAlive = types.BoolValue(project.WarehouseConnection.ClientSessionKeepAlive)
	state.SnowflakeWarehouseConnectionThreads = types.Int32Value(project.WarehouseConnection.Threads)

	state.DatabricksConnectionServerHostName = types.StringValue(project.WarehouseConnection.ServerHostName)
	state.DatabricksConnectionHTTPPath = types.StringValue(project.WarehouseConnection.HTTPPath)
	state.DatabricksConnectionCatalog = types.StringValue(project.WarehouseConnection.Catalog)
	if project.WarehouseConnection.Type == "databricks" {
		var databricksConnectionPersonalAccessToken string
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("databricks_connection_personal_access_token"), &databricksConnectionPersonalAccessToken)...)
		state.DatabricksConnectionPersonalAccessToken = types.StringValue(databricksConnectionPersonalAccessToken)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan, state projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Project updates

	// Set state
	diags := resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing project
	projectUuid := state.ProjectUUID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting project %s", projectUuid))
	err := r.client.DeleteProjectV1(projectUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting project",
			"Could not delete project, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Extract the resource ID
	extracted_strings, err := extractProjectResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	organization_uuid := extracted_strings[0]
	projectUuid := extracted_strings[1]

	// Get the imported project
	importedProject, err := r.client.GetProjectV1(projectUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting project",
			fmt.Sprintf("Could not get project with organization UUID %s and project UUID %s, unexpected error: %s", organization_uuid, projectUuid, err.Error()),
		)
		return
	}

	// Set the resource attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), projectUuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_uuid"), importedProject.OrganizationUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), importedProject.ProjectUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), importedProject.ProjectName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), importedProject.ProjectType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dbt_connection_type"), importedProject.DbtConnection.Type)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dbt_connection_repository"), importedProject.DbtConnection.Repository)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dbt_connection_branch"), importedProject.DbtConnection.Branch)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dbt_connection_project_sub_path"), importedProject.DbtConnection.ProjectSubPath)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dbt_connection_host_domain"), importedProject.DbtConnection.HostDomain)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("warehouse_connection_type"), importedProject.WarehouseConnection.Type)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("snowflake_warehouse_connection_account"), importedProject.WarehouseConnection.Account)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("snowflake_warehouse_connection_role"), importedProject.WarehouseConnection.Role)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("snowflake_warehouse_connection_database"), importedProject.WarehouseConnection.Database)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("snowflake_warehouse_connection_schema"), importedProject.WarehouseConnection.Schema)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("snowflake_warehouse_connection_warehouse"), importedProject.WarehouseConnection.Warehouse)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("snowflake_warehouse_connection_threads"), importedProject.WarehouseConnection.Threads)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("snowflake_warehouse_connection_client_session_keep_alive"), importedProject.WarehouseConnection.ClientSessionKeepAlive)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("databricks_connection_server_host_name"), importedProject.WarehouseConnection.ServerHostName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("databricks_connection_http_path"), importedProject.WarehouseConnection.HTTPPath)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("databricks_connection_personal_access_token"), importedProject.WarehouseConnection.PersonalAccessToken)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("databricks_connection_catalog"), importedProject.WarehouseConnection.Catalog)...)
}

func getProjectResourceId(organization_uuid string, project_uuid string) string {
	// Return the resource ID
	return fmt.Sprintf("organizations/%s/projects/%s", organization_uuid, project_uuid)
}

func extractProjectResourceId(input string) ([]string, error) {
	// Extract the captured projects
	pattern := `^organizations/([^/]+)/projects/([^/]+)$`
	projects, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}

	// Return the captured strings
	organization_uuid := projects[0]
	project_uuid := projects[1]
	return []string{organization_uuid, project_uuid}, nil
}
