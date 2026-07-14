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
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &projectUpstreamResource{}
	_ resource.ResourceWithConfigure   = &projectUpstreamResource{}
	_ resource.ResourceWithImportState = &projectUpstreamResource{}
)

func NewProjectUpstreamResource() resource.Resource {
	return &projectUpstreamResource{}
}

// projectUpstreamResource defines the resource implementation.
type projectUpstreamResource struct {
	client *api.Client
}

// projectUpstreamResourceModel describes the resource data model.
type projectUpstreamResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	OrganizationUUID    types.String `tfsdk:"organization_uuid"`
	ProjectUUID         types.String `tfsdk:"project_uuid"`
	UpstreamProjectUUID types.String `tfsdk:"upstream_project_uuid"`
}

func (r *projectUpstreamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_upstream"
}

func (r *projectUpstreamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_project_upstream.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Manages Lightdash project upstream (Data Ops) relationship",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource identifier. It is computed as `organizations/<organization_uuid>/projects/<project_uuid>/upstream`.",
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
				MarkdownDescription: "The UUID of the source Lightdash project (for example a development or preview project).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"upstream_project_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the upstream (destination) project used for content promotion.",
				Required:            true,
			},
		},
	}
}

func (r *projectUpstreamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectUpstreamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectUpstreamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyUpstream(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Error creating project upstream", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectUpstreamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectUpstreamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectUUID := state.ProjectUUID.ValueString()
	upstreamService := services.NewProjectUpstreamService(r.client, projectUUID)
	upstreamUUID, err := upstreamService.GetProjectUpstream(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading project upstream",
			"Could not read project upstream for ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	if upstreamUUID == nil {
		// Upstream link was cleared outside Terraform; remove from state.
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(getProjectUpstreamResourceId(state.OrganizationUUID.ValueString(), projectUUID))
	state.UpstreamProjectUUID = types.StringValue(*upstreamUUID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectUpstreamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectUpstreamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyUpstream(ctx, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating project upstream",
			fmt.Sprintf("Could not update upstream for project UUID '%s', unexpected error: %s", plan.ProjectUUID.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectUpstreamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectUpstreamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	upstreamService := services.NewProjectUpstreamService(r.client, state.ProjectUUID.ValueString())
	if err := upstreamService.UpdateProjectUpstream(ctx, nil); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting project upstream",
			fmt.Sprintf("Could not clear upstream for project UUID '%s', unexpected error: %s", state.ProjectUUID.ValueString(), err.Error()),
		)
		return
	}
}

func (r *projectUpstreamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	extractedStrings, err := extractProjectUpstreamResourceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting resource ID",
			"Could not extract resource ID, unexpected error: "+err.Error(),
		)
		return
	}
	organizationUUID := extractedStrings[0]
	projectUUID := extractedStrings[1]

	upstreamService := services.NewProjectUpstreamService(r.client, projectUUID)
	upstreamUUID, err := upstreamService.GetProjectUpstream(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting project upstream",
			fmt.Sprintf("Could not get upstream for organization UUID %s and project UUID %s, unexpected error: %s", organizationUUID, projectUUID, err.Error()),
		)
		return
	}

	if upstreamUUID == nil {
		resp.Diagnostics.AddError(
			"Project upstream not set",
			fmt.Sprintf("Project %s has no upstream project configured; nothing to import.", projectUUID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), getProjectUpstreamResourceId(organizationUUID, projectUUID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_uuid"), organizationUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), projectUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("upstream_project_uuid"), *upstreamUUID)...)
}

func (r *projectUpstreamResource) applyUpstream(ctx context.Context, plan *projectUpstreamResourceModel) error {
	upstreamUUID := strings.TrimSpace(plan.UpstreamProjectUUID.ValueString())
	if upstreamUUID == "" {
		return fmt.Errorf("upstream_project_uuid must be a non-empty project UUID")
	}

	projectUUID := plan.ProjectUUID.ValueString()
	upstreamService := services.NewProjectUpstreamService(r.client, projectUUID)
	if err := upstreamService.UpdateProjectUpstream(ctx, &upstreamUUID); err != nil {
		return err
	}

	plan.UpstreamProjectUUID = types.StringValue(upstreamUUID)
	plan.ID = types.StringValue(getProjectUpstreamResourceId(plan.OrganizationUUID.ValueString(), projectUUID))
	return nil
}

func getProjectUpstreamResourceId(organizationUUID string, projectUUID string) string {
	return fmt.Sprintf("organizations/%s/projects/%s/upstream", organizationUUID, projectUUID)
}

func extractProjectUpstreamResourceId(input string) ([]string, error) {
	pattern := `^organizations/([^/]+)/projects/([^/]+)/upstream$`
	groups, err := extractStrings(input, pattern)
	if err != nil {
		return nil, fmt.Errorf("could not extract resource ID: %w", err)
	}
	return groups, nil
}
