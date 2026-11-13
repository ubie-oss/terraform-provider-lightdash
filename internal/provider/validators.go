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

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// ValidateNonEmptyString validates that a string attribute is not empty or null.
type ValidateNonEmptyString struct{}

// Description returns a plain text description of the validator's behavior.
func (v ValidateNonEmptyString) Description(ctx context.Context) string {
	return "string must not be empty or null"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (v ValidateNonEmptyString) MarkdownDescription(ctx context.Context) string {
	return "string must not be empty or null"
}

// ValidateString performs the validation.
func (v ValidateNonEmptyString) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid String Value",
			"String cannot be null or unknown",
		)
		return
	}

	value := strings.TrimSpace(req.ConfigValue.ValueString())
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid String Value",
			fmt.Sprintf("String cannot be empty. Got: %q", req.ConfigValue.ValueString()),
		)
		return
	}
}

// ValidateProjectMemberRole accepts only the lowercase project role slugs. The provider
// stores the slug it gets back from the API, so a spelling such as "Editor" would fail
// the apply with an inconsistent result.
type ValidateProjectMemberRole struct{}

const projectMemberRoleDescription = "must be one of: viewer, interactive_viewer, editor, developer, admin"

// Description returns a plain text description of the validator's behavior.
func (v ValidateProjectMemberRole) Description(ctx context.Context) string {
	return projectMemberRoleDescription
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (v ValidateProjectMemberRole) MarkdownDescription(ctx context.Context) string {
	return projectMemberRoleDescription
}

// ValidateString performs the validation.
func (v ValidateProjectMemberRole) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !models.ProjectMemberRole(req.ConfigValue.ValueString()).IsValid() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Project Role",
			fmt.Sprintf("%s. Got: %q", projectMemberRoleDescription, req.ConfigValue.ValueString()),
		)
	}
}
