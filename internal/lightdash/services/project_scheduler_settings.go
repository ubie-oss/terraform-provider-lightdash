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
	"fmt"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type ProjectSchedulerSettingsService struct {
	client      *api.Client
	projectUuid string
}

func NewProjectSchedulerSettingsService(client *api.Client, projectUuid string) *ProjectSchedulerSettingsService {
	return &ProjectSchedulerSettingsService{
		client:      client,
		projectUuid: projectUuid,
	}
}

func (s *ProjectSchedulerSettingsService) GetProjectSchedulerSettings(projectUuid string) (*models.ProjectSchedulerSettings, error) {
	// Get the project
	project, err := s.client.GetProjectV1(projectUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get project (%s): %w", projectUuid, err)
	}

	// Get the project scheduler settings
	schedulerSettings := &models.ProjectSchedulerSettings{
		SchedulerTimezone: project.SchedulerTimezone,
	}

	return schedulerSettings, nil
}

func (s *ProjectSchedulerSettingsService) UpdateProjectSchedulerSettings(
	projectSchedulerSettings *models.ProjectSchedulerSettings) error {

	// Update the project scheduler settings
	var schedulerTimezone = projectSchedulerSettings.SchedulerTimezone
	_, err := s.client.UpdateSchedulerSettingsV1(s.projectUuid, schedulerTimezone)
	if err != nil {
		return fmt.Errorf("failed to update project scheduler settings in project (%s) with timezone (%s): %w", s.projectUuid, schedulerTimezone, err)
	}
	return nil
}
