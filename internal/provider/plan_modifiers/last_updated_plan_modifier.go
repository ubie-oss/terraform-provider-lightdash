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

package plan_modifiers

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type lastUpdatedPlanModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m lastUpdatedPlanModifier) Description(_ context.Context) string {
	return "Timestamp of the last Terraform update of the space."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m lastUpdatedPlanModifier) MarkdownDescription(_ context.Context) string {
	return "Timestamp of the last Terraform update of the space."
}

// PlanModifyString implements the plan modification logic.
func (m lastUpdatedPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the plan value is unknown and there is a state value, it means the resource is being updated
	// and the last_updated field is not explicitly set in config.
	// In this case, we want to set it to the current time to reflect the update.
	// We also need to check that the resource isn't being created (state is null) or destroyed (plan is null).

	if !req.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.Plan.Raw.IsNull() {
		return // Do nothing if planned value is known, creating, or destroying
	}

	// Set the planned value to the current time
	resp.PlanValue = types.StringValue(time.Now().Format(time.RFC850))
	tflog.Debug(ctx, "Set last_updated plan value to current time", map[string]any{"value": resp.PlanValue.ValueString()})
}

// SetLastUpdatedOnUpdate returns a plan modifier that sets the last_updated attribute to the current time on update.
func SetLastUpdatedOnUpdate() planmodifier.String {
	return lastUpdatedPlanModifier{}
}
