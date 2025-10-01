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
