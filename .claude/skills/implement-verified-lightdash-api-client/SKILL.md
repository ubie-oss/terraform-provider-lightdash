---
name: implement-verified-lightdash-api-client
description: Implement verified Lightdash API clients with documentation research and live schema verification.
---

# Implement Verified Lightdash API Client

## Description

This skill implements a new Lightdash API operation in the `internal/lightdash/api` directory, along with necessary models in `internal/lightdash/models`. It relies on the [research-lightdash-api](../research-lightdash-api/SKILL.md) skill for researching and verifying the API schema before implementation.

## Input

The user should provide:

1. **HTTP Method**: (GET, POST, PUT, DELETE, PATCH)
2. **API Endpoint**: (e.g., `/api/v1/projects/:projectUuid/spaces`)
3. **Function Name**: (e.g., `GetSpaceV1`, `CreateDashboardV1`)
4. **Documentation URL**: (Optional) Link to the official Lightdash API docs.

## Workflow

### 1. Research & Verification

- **Invoke Research**: Use the [research-lightdash-api](../research-lightdash-api/SKILL.md) skill to research the endpoint and verify its schema against the live API (if `LIGHTDASH_API_KEY` is available).
- **Outcome**: Ensure you have a **verified** JSON response or documentation-based schema before proceeding to model generation.

### 2. Model Analysis & Generation

- **Check existing models**: Look in `internal/lightdash/models` for existing structs that match the resource.
- **Generate/Update models**: If no suitable model exists, create/update the file `internal/lightdash/models/<resource>.go`.
  - Use strict typing based on the **verified** JSON.
  - Add `json` tags.
  - Handle nullable fields with pointers where appropriate.
  - Follow Go naming conventions (PascalCase for exported fields).

### 3. API Client Implementation

- **Create File**: Create a new file `internal/lightdash/api/<verb>_<resource>_v1.go` (snake_case).
- **Implement Method**: Add the method to the `Client` struct: `func (c *Client) <FunctionName>(...) (*<ReturnType>, error)`.
- **Request Construction**:
  - Use `http.NewRequest`.
  - Construct the path using `fmt.Sprintf` and `c.HostUrl`.
  - Validate input parameters (check for empty strings for UUIDs).
- **Execution**:
  - Call `c.doRequest(req)`.
  - Unmarshal the response body into the typed model.
  - Return the `Results` field if the API wraps the response in a [Results envelope](references/response_envelope.md).

### 4. Verification

- **Compilation Check**: Ensure the code compiles.
- **Unit Tests**: Create/Update `internal/lightdash/api/<verb>_<resource>_v1_test.go` to test unmarshaling logic with sample JSON.

## Example Pattern

Ref: `internal/lightdash/api/get_space_v1.go`

```go
package api

import (
 "encoding/json"
 "fmt"
 "net/http"
 "strings"

 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type GetResourceV1Response struct {
 Results models.Resource `json:"results"`
 Status  string          `json:"status"`
}

func (c *Client) GetResourceV1(id string) (*models.Resource, error) {
    if len(strings.TrimSpace(id)) == 0 {
        return nil, fmt.Errorf("id is empty")
    }

    path := fmt.Sprintf("%s/api/v1/resource/%s", c.HostUrl, id)
    req, err := http.NewRequest("GET", path, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %w", err)
    }

    body, err := c.doRequest(req)
    if err != nil {
        return nil, fmt.Errorf("error performing request: %w", err)
    }

    var response GetResourceV1Response
    if err := json.Unmarshal(body, &response); err != nil {
        return nil, fmt.Errorf("error unmarshaling response: %w", err)
    }

    return &response.Results, nil
}
```

## Reference

- [Lightdash API Documentation](https://docs.lightdash.com/api-reference/v1/introduction)
- [Lightdash Client Rules](.cursor/rules/lightdash-client.mdc)
- [API Client Structure](references/api_client_structure.md)
- [API Response Envelope](references/response_envelope.md)

## Assets

- [API Operation Boilerplate](assets/api_boilerplate.go)
- [Model Boilerplate](assets/model_boilerplate.go)
- [Unit Test Boilerplate](assets/api_test_boilerplate.go)
- [Schema Verification Script](../research-lightdash-api/assets/verify_schema.go)
