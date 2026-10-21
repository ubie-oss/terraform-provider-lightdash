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
	"errors"
	"strings"
	"testing"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/services"
)

// All acceptance-test credentials use this label prefix so we can purge them via the DELETE API.
const accAiProviderCredentialLabelPrefix = "test (Acceptance Test -"

func testAccPreCheckAiProviderCredentials(t *testing.T) {
	testAccPreCheck(t)
	purgeAccAiTestCredentials(t)
}

// registerAccAiCredentialPurge deletes acc-test credentials after the test finishes (pass or fail).
func registerAccAiCredentialPurge(t *testing.T) {
	t.Cleanup(func() { purgeAccAiTestCredentials(t) })
}

func purgeAccAiTestCredentials(t *testing.T) {
	client, err := newAccTestAPIClient()
	if err != nil {
		t.Errorf("acc AI credential purge: %v", err)
		return
	}

	ctx := context.Background()
	service := services.NewAiProviderCredentialsService(client)
	if err := resetAccTestProjectBindingToAccCredential(ctx, client, service); err != nil {
		t.Errorf("acc AI credential purge: reset project binding: %v", err)
	}

	for round := 0; round < 15; round++ {
		list, err := service.List(ctx)
		if err != nil {
			t.Errorf("acc AI credential purge: list: %v", err)
			return
		}

		acc := accCredentialsWithLabelPrefix(list)
		if len(acc) == 0 {
			return
		}

		progress := false
		for _, cred := range acc {
			if cred.IsDefault {
				continue
			}
			if err := apiv1.DeleteAiProviderCredentialV1(client, cred.UUID); err == nil {
				progress = true
			}
		}
		if progress {
			continue
		}

		list, err = service.List(ctx)
		if err != nil {
			t.Errorf("acc AI credential purge: list: %v", err)
			return
		}
		acc = accCredentialsWithLabelPrefix(list)
		if len(acc) == 0 {
			return
		}

		for _, cred := range acc {
			if !cred.IsDefault {
				continue
			}
			if replacement := pickNonDefaultCredentialUUID(list.Credentials, cred.UUID); replacement != "" {
				if err := apiv1.SetDefaultAiProviderCredentialV1(client, replacement); err != nil {
					t.Errorf("acc AI credential purge: move default off %s: %v", cred.Label, err)
					return
				}
				progress = true
				break
			}
			if err := apiv1.DeleteAiProviderCredentialV1(client, cred.UUID); err == nil {
				progress = true
				break
			}
		}
		if !progress {
			t.Errorf("acc AI credential purge: could not remove: %s", formatCredentialLabels(acc))
			return
		}
	}
}

func resetAccTestProjectBindingToAccCredential(ctx context.Context, client *api.Client, service *services.AiProviderCredentialsService) error {
	projectUUID, err := getLightdashProjectUuid()
	if err != nil {
		return err
	}

	selection, err := apiv1.GetProjectAiProviderCredentialV1(client, *projectUUID)
	if err != nil {
		return err
	}
	if selection.CredentialUUID == nil {
		return nil
	}

	cred, err := service.GetByUUID(ctx, *selection.CredentialUUID)
	if err != nil {
		if errors.Is(err, services.ErrAiProviderCredentialNotFound) {
			return nil
		}
		return err
	}
	if !isAccAiTestCredentialLabel(cred.Label) {
		return nil
	}

	return apiv1.SetProjectAiProviderCredentialV1(client, *projectUUID, models.ProjectAiCredentialSelection{
		CredentialUUID: nil,
	})
}

func accCredentialsWithLabelPrefix(list *models.AiProviderCredentialsList) []models.AiProviderCredential {
	if list == nil {
		return nil
	}
	var out []models.AiProviderCredential
	for _, cred := range list.Credentials {
		if isAccAiTestCredentialLabel(cred.Label) {
			out = append(out, cred)
		}
	}
	return out
}

// Prefer a non-acc credential as org default so acc credentials can be deleted.
func pickNonDefaultCredentialUUID(credentials []models.AiProviderCredential, currentDefaultUUID string) string {
	for _, cred := range credentials {
		if cred.UUID == currentDefaultUUID {
			continue
		}
		if !isAccAiTestCredentialLabel(cred.Label) {
			return cred.UUID
		}
	}
	for _, cred := range credentials {
		if cred.UUID != currentDefaultUUID {
			return cred.UUID
		}
	}
	return ""
}

func formatCredentialLabels(creds []models.AiProviderCredential) string {
	labels := make([]string, 0, len(creds))
	for _, cred := range creds {
		labels = append(labels, cred.Label)
	}
	return strings.Join(labels, ", ")
}

func isAccAiTestCredentialLabel(label string) bool {
	return strings.HasPrefix(label, accAiProviderCredentialLabelPrefix) ||
		label == "terraform-provider-lightdash acc AI bootstrap"
}

func newAccTestAPIClient() (*api.Client, error) {
	host, err := getLightdashUrl()
	if err != nil {
		return nil, err
	}
	token, err := getLightdashApiKey()
	if err != nil {
		return nil, err
	}
	return api.NewClient(host, token, nil)
}
