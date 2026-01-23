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

package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type AddSpaceShareToUserV1Request struct {
	UserUUID  string `json:"userUuid"`
	SpaceRole string `json:"spaceRole"`
}

func AddSpaceShareToUserV1(c *api.Client,
	projectUuid string,
	spaceUuid string,
	userUuid string,
	spaceRole models.SpaceMemberRole) error {
	// Create the request body
	data := AddSpaceShareToUserV1Request{
		UserUUID:  userUuid,
		SpaceRole: spaceRole.String(),
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("impossible to marshall data: %s", err)
	}
	// Create the request
	path := fmt.Sprintf("%s/api/v1/projects/%s/spaces/%s/share", c.HostUrl, projectUuid, spaceUuid)
	req, err := http.NewRequest("POST", path, bytes.NewReader(marshalled))
	if err != nil {
		return fmt.Errorf("failed to create new request for space share: %w", err)
	}
	// Do request
	_, err = c.DoRequest(req)
	if err != nil {
		return fmt.Errorf("request to share space failed: %w", err)
	}

	return nil
}
