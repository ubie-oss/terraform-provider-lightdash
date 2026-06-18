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

package services

import (
	"context"
	"fmt"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"
)

// GetOrganizationUUID returns the organization UUID for the configured API token.
func GetOrganizationUUID(ctx context.Context, client *api.Client) (string, error) {
	org, err := apiv1.GetMyOrganizationV1(client)
	if err != nil {
		return "", fmt.Errorf("failed to get organization: %w", err)
	}
	return org.OrganizationUUID, nil
}
