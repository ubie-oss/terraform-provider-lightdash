// Copyright 2023 Ubie, inc.
package api

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

func TestGetExampleV1ResponseJSONUnmarshal(t *testing.T) {
	jsonStr := `{
		"status": "ok",
		"results": {
			"uuid": "example-uuid",
			"name": "Example Name"
		}
	}`

	var response ExampleResponse
	err := json.Unmarshal([]byte(jsonStr), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	expectedResults := models.Example{
		UUID: "example-uuid",
		Name: "Example Name",
	}

	if response.Status != "ok" {
		t.Errorf("expected Status to be 'ok', got '%s'", response.Status)
	}

	if !reflect.DeepEqual(response.Results, expectedResults) {
		t.Errorf("expected Results to be %+v, got %+v", expectedResults, response.Results)
	}
}
