// Copyright 2023 Ubie, inc.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type ExampleResponse struct {
	Results models.Example `json:"results"`
	Status  string         `json:"status"`
}

func (c *Client) GetExampleV1(id string) (*models.Example, error) {
	if len(strings.TrimSpace(id)) == 0 {
		return nil, fmt.Errorf("id is empty")
	}

	path := fmt.Sprintf("%s/api/v1/example/%s", c.HostUrl, id)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}

	var response ExampleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &response.Results, nil
}
