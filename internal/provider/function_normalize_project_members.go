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
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Custom function to normalize the members of a project by role.
// If a member belongs to multiple roles, they will only appear once in the returned list.
// The member will be assigned the highest role in the list of roles.
//
// provider::lightdash::function::normalize_project_members(:
//   admins: list(string),
//   developers: list(string),
//   editors: list(string),
//   interactive_viewers: list(string),
//   viewers: list(string),
// )
//
// Returns a list of normalize members by role
// {
//   admins: list(string),
//   developers: list(string),
//   editors: list(string),
//   interactive_viewers: list(string),
//   viewers: list(string),
// }

// Ensure NormalizeProjectMembersFunction satisfies the function.Function interface.
var _ function.Function = &NormalizeProjectMembersFunction{}

// NormalizeProjectMembersFunction defines the function implementation.
type NormalizeProjectMembersFunction struct{}

func NewNormalizeProjectMembersFunction() function.Function {
	return &NormalizeProjectMembersFunction{}
}

// Metadata returns the function type name.
func (f *NormalizeProjectMembersFunction) Metadata(
	_ context.Context,
	req function.MetadataRequest,
	resp *function.MetadataResponse,
) {
	// The TypeName is inferred from the function name in the provider's Functions method.
	resp.Name = "normalize_project_members"
}

// Definition defines the function schema including parameters and return type.
func (f *NormalizeProjectMembersFunction) Definition(
	ctx context.Context,
	req function.DefinitionRequest,
	resp *function.DefinitionResponse,
) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/functions/function_normalize_project_members.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Definition = function.Definition{
		Summary:             "Normalize the members of a project by role.",
		MarkdownDescription: markdownDescription,

		Parameters: []function.Parameter{
			function.ListParameter{
				Name:                "admins",
				MarkdownDescription: "List of admin member UUIDs.",
				ElementType:         types.StringType,
			},
			function.ListParameter{
				Name:                "developers",
				MarkdownDescription: "List of developer member UUIDs.",
				ElementType:         types.StringType,
			},
			function.ListParameter{
				Name:                "editors",
				MarkdownDescription: "List of editor member UUIDs.",
				ElementType:         types.StringType,
			},
			function.ListParameter{
				Name:                "interactive_viewers",
				MarkdownDescription: "List of interactive viewer member UUIDs.",
				ElementType:         types.StringType,
			},
			function.ListParameter{
				Name:                "viewers",
				MarkdownDescription: "List of viewer member UUIDs.",
				ElementType:         types.StringType,
			},
		},
		Return: function.ObjectReturn{
			AttributeTypes: map[string]attr.Type{
				"admins":              types.ListType{ElemType: types.StringType},
				"developers":          types.ListType{ElemType: types.StringType},
				"editors":             types.ListType{ElemType: types.StringType},
				"interactive_viewers": types.ListType{ElemType: types.StringType},
				"viewers":             types.ListType{ElemType: types.StringType},
			},
		},
	}
}

// NormalizeProjectMembersParametersModel maps the function parameters and return.
type NormalizeProjectMembersParametersModel struct {
	Admins             types.List `tfsdk:"admins"`
	Developers         types.List `tfsdk:"developers"`
	Editors            types.List `tfsdk:"editors"`
	InteractiveViewers types.List `tfsdk:"interactive_viewers"`
	Viewers            types.List `tfsdk:"viewers"`
}

// NormalizeProjectMembersResponseModel maps the function value.
type NormalizeProjectMembersResponseModel struct {
	Admins             types.List `tfsdk:"admins"`
	Developers         types.List `tfsdk:"developers"`
	Editors            types.List `tfsdk:"editors"`
	InteractiveViewers types.List `tfsdk:"interactive_viewers"`
	Viewers            types.List `tfsdk:"viewers"`
}

// Run the function to normalize the members.
func (f *NormalizeProjectMembersFunction) Run(
	ctx context.Context,
	req function.RunRequest,
	resp *function.RunResponse,
) {
	var (
		adminMembers             []string
		developerMembers         []string
		editorMembers            []string
		interactiveViewerMembers []string
		viewerMembers            []string
	)

	// Read arguments by position
	res := req.Arguments.GetArgument(ctx, 0, &adminMembers)
	resp.Error = function.ConcatFuncErrors(resp.Error, res)

	res = req.Arguments.GetArgument(ctx, 1, &developerMembers)
	resp.Error = function.ConcatFuncErrors(resp.Error, res)

	res = req.Arguments.GetArgument(ctx, 2, &editorMembers)
	resp.Error = function.ConcatFuncErrors(resp.Error, res)

	res = req.Arguments.GetArgument(ctx, 3, &interactiveViewerMembers)
	resp.Error = function.ConcatFuncErrors(resp.Error, res)

	res = req.Arguments.GetArgument(ctx, 4, &viewerMembers)
	resp.Error = function.ConcatFuncErrors(resp.Error, res)

	if resp.Error != nil {
		return
	}

	// Call the normalizeMembers function
	resultAdmins, resultDevelopers, resultEditors, resultInteractiveViewers, resultViewers := f.normalizeMembers(
		adminMembers,
		developerMembers,
		editorMembers,
		interactiveViewerMembers,
		viewerMembers,
	)

	// Convert Go slices back to Terraform types.List and handle potential diagnostics
	resultModel := NormalizeProjectMembersResponseModel{}
	var listDiags diag.Diagnostics

	resultModel.Admins, listDiags = types.ListValueFrom(ctx, types.StringType, resultAdmins)
	resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, listDiags))

	resultModel.Developers, listDiags = types.ListValueFrom(ctx, types.StringType, resultDevelopers)
	resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, listDiags))

	resultModel.Editors, listDiags = types.ListValueFrom(ctx, types.StringType, resultEditors)
	resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, listDiags))

	resultModel.InteractiveViewers, listDiags = types.ListValueFrom(ctx, types.StringType, resultInteractiveViewers)
	resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, listDiags))

	resultModel.Viewers, listDiags = types.ListValueFrom(ctx, types.StringType, resultViewers)
	resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, listDiags))

	if resp.Error != nil {
		return
	}

	// Set the function execution result.
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, resultModel))
}

func (f *NormalizeProjectMembersFunction) normalizeMembers(
	admins []string,
	developers []string,
	editors []string,
	interactiveViewers []string,
	viewers []string,
) (
	resultAdmins []string,
	resultDevelopers []string,
	resultEditors []string,
	resultInteractiveViewers []string,
	resultViewers []string,
) {
	// Map to track members and their highest role.
	// Higher index means higher role precedence.
	// 0: viewer, 1: interactive_viewer, 2: editor, 3: developer, 4: admin
	memberRoles := make(map[string]int)

	processMembers := func(members []string, role int) {
		for _, member := range members {
			if currentRole, ok := memberRoles[member]; !ok || role > currentRole {
				memberRoles[member] = role
			}
		}
	}

	// Process members in increasing order of role precedence.
	processMembers(viewers, 0)
	processMembers(interactiveViewers, 1)
	processMembers(editors, 2)
	processMembers(developers, 3)
	processMembers(admins, 4)

	// Initialize result slices
	resultAdmins = []string{}
	resultDevelopers = []string{}
	resultEditors = []string{}
	resultInteractiveViewers = []string{}
	resultViewers = []string{}

	// Create the result slices based on the role
	for member, role := range memberRoles {
		switch role {
		case 4:
			resultAdmins = append(resultAdmins, member)
		case 3:
			resultDevelopers = append(resultDevelopers, member)
		case 2:
			resultEditors = append(resultEditors, member)
		case 1:
			resultInteractiveViewers = append(resultInteractiveViewers, member)
		case 0:
			resultViewers = append(resultViewers, member)
		}
	}

	// Sort the lists to be deterministic
	sort.Strings(resultAdmins)
	sort.Strings(resultDevelopers)
	sort.Strings(resultEditors)
	sort.Strings(resultInteractiveViewers)
	sort.Strings(resultViewers)

	// Return the segregated lists
	return resultAdmins, resultDevelopers, resultEditors, resultInteractiveViewers, resultViewers
}
