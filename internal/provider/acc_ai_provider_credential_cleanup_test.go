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
	"testing"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

func TestPickNonDefaultCredentialUUIDPrefersProduction(t *testing.T) {
	t.Parallel()

	creds := []models.AiProviderCredential{
		{UUID: "acc-default", Label: "test (Acceptance Test - x)", IsDefault: true},
		{UUID: "prod", Label: "production bedrock", IsDefault: false},
	}

	got := pickNonDefaultCredentialUUID(creds, "acc-default")
	if got != "prod" {
		t.Fatalf("expected prod, got %q", got)
	}
}

func TestAccCredentialsWithLabelPrefix(t *testing.T) {
	t.Parallel()

	list := &models.AiProviderCredentialsList{
		Credentials: []models.AiProviderCredential{
			{Label: "production"},
			{Label: "test (Acceptance Test - lifecycle)"},
		},
	}
	acc := accCredentialsWithLabelPrefix(list)
	if len(acc) != 1 {
		t.Fatalf("expected 1, got %d", len(acc))
	}
}
