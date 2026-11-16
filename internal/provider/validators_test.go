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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateOrganizationMemberRole(t *testing.T) {
	tests := []struct {
		name    string
		value   types.String
		wantErr bool
	}{
		{name: "member", value: types.StringValue("member"), wantErr: false},
		{name: "interactive_viewer", value: types.StringValue("interactive_viewer"), wantErr: false},
		{name: "mixed case is rejected", value: types.StringValue("Viewer"), wantErr: true},
		{name: "space is rejected", value: types.StringValue("interactive viewer"), wantErr: true},
		{name: "unknown role", value: types.StringValue("superadmin"), wantErr: true},
		{name: "null is skipped", value: types.StringNull(), wantErr: false},
		{name: "unknown is skipped", value: types.StringUnknown(), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{Path: path.Root("role"), ConfigValue: tt.value}
			resp := &validator.StringResponse{}
			ValidateOrganizationMemberRole{}.ValidateString(context.Background(), req, resp)
			if got := resp.Diagnostics.HasError(); got != tt.wantErr {
				t.Errorf("HasError() = %v, want %v (diagnostics: %v)", got, tt.wantErr, resp.Diagnostics)
			}
		})
	}
}
