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

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// lastUpdatedOnChange marks a Computed "last_updated" timestamp attribute as
// Unknown during plan whenever any other attribute of the resource is planned
// to change. Without this, the framework preserves the prior state value in
// the plan while Update rewrites the timestamp, causing
// "Provider produced inconsistent result after apply".
type lastUpdatedOnChange struct{}

// LastUpdatedOnChange returns the plan modifier.
func LastUpdatedOnChange() planmodifier.String {
	return lastUpdatedOnChange{}
}

func (lastUpdatedOnChange) Description(_ context.Context) string {
	return "Marks last_updated as unknown during plan when any other attribute is planned to change."
}

func (m lastUpdatedOnChange) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (lastUpdatedOnChange) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Create: no prior state, framework handles unknown itself.
	if req.State.Raw.IsNull() {
		return
	}
	// Destroy: nothing to plan.
	if req.Plan.Raw.IsNull() {
		return
	}
	// If the plan differs from state anywhere in the resource, Update will
	// run and rewrite last_updated. Mark it Unknown so the post-apply value
	// is accepted as consistent.
	if !req.State.Raw.Equal(req.Plan.Raw) {
		resp.PlanValue = types.StringUnknown()
	}
}
