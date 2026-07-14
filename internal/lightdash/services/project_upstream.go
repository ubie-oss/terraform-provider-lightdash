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

	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

type ProjectUpstreamService struct {
	client      *api.Client
	projectUuid string
}

func NewProjectUpstreamService(client *api.Client, projectUuid string) *ProjectUpstreamService {
	return &ProjectUpstreamService{
		client:      client,
		projectUuid: projectUuid,
	}
}

// GetProjectUpstream returns the upstream project UUID, or nil when unset/empty.
func (s *ProjectUpstreamService) GetProjectUpstream(ctx context.Context) (*string, error) {
	project, err := apiv1.GetProjectV1(s.client, s.projectUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get project (%s): %w", s.projectUuid, err)
	}
	return normalizeUpstreamProjectUUID(project.UpstreamProjectUUID), nil
}

// UpdateProjectUpstream sets upstreamProjectUuid, or clears it when upstreamProjectUuid is nil.
func (s *ProjectUpstreamService) UpdateProjectUpstream(ctx context.Context, upstreamProjectUuid *string) error {
	_, err := apiv1.UpdateProjectMetadataV1(s.client, s.projectUuid, upstreamProjectUuid)
	if err != nil {
		if upstreamProjectUuid == nil {
			return fmt.Errorf("failed to clear upstream project for project (%s): %w", s.projectUuid, err)
		}
		return fmt.Errorf("failed to set upstream project (%s) for project (%s): %w", *upstreamProjectUuid, s.projectUuid, err)
	}
	return nil
}

func normalizeUpstreamProjectUUID(upstreamUUID *string) *string {
	if upstreamUUID == nil || *upstreamUUID == "" {
		return nil
	}
	return upstreamUUID
}
