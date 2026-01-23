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
	"fmt"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

func RemoveProjectAccessFromGroupV1(c *api.Client, projectUuid string, groupUuid string) error {
	// Create the request
	path := fmt.Sprintf("%s/api/v1/groups/%s/projects/%s", c.HostUrl, groupUuid, projectUuid)
	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to create new request for removing project access from group: %w", err)
	}
	// Do request
	_, err = c.DoRequest(req)
	if err != nil {
		return fmt.Errorf("request to remove project access from group failed: %w", err)
	}

	return nil
}
