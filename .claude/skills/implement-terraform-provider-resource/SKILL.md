---
name: implement-terraform-provider-resource
description: Implement Terraform Resources and Data Sources in `internal/provider` with comprehensive testing.
---

# Implement Terraform Provider Resource

## Description

This skill implements a new Terraform Resource or Data Source in the `internal/provider` directory using the `hashicorp/terraform-plugin-framework`. It includes comprehensive implementation steps, unit tests, and acceptance tests following the project's established patterns.

## Input

The user should provide:

1. **Type**: Resource or Data Source.
2. **Name**: (e.g., `lightdash_project_agent`).
3. **Schema**: List of attributes (name, type, required/optional/computed).
4. **API Client Method**: Which `api.Client` method(s) to use for CRUD/Read operations.

## Workflow

### 1. Implementation

- **File Creation**: Create `internal/provider/resource_<name>.go` or `data_source_<name>.go`.
- **Define Model**: Create a Go struct with `tfsdk` tags. Use framework types (`types.String`, `types.Bool`, etc.).
- **Implement Interface**: Implement `resource.Resource` or `datasource.DataSource`.
- **Schema**: Define the schema in the `Schema` method. Use descriptions from documentation.
- **Configure**: In the `Configure` method, retrieve the `*api.Client` from `req.ProviderData`.
- **CRUD/Read**: Implement `Create`, `Read`, `Update`, `Delete` (for resources) or `Read` (for data sources).
  - Call the appropriate `api.Client` or `services` layer methods.
  - Handle diagnostics (`resp.Diagnostics`) for errors.

### 2. Unit Testing

- **Test File**: Create `internal/provider/resource_<name>_test.go` or `data_source_<name>_test.go`.
- **Focus**: Test schema validation, custom validators, or helper functions that don't require a live API.

### 3. Acceptance Testing

- **Setup**: Use `isIntegrationTestMode()` and `testAccPreCheck(t)`.
- **Test Configurations**:
  - Create directory `internal/provider/acc_tests/resources/<name>/` or `internal/provider/acc_tests/data_sources/<name>/`.
  - Add `.tf` files for different test scenarios (e.g., `010_create.tf`, `020_update.tf`).
- **Test Case**: Implement `TestAcc...` using `resource.Test` from `github.com/hashicorp/terraform-plugin-testing/helper/resource`.
  - Include `ImportState: true` for resource tests.
  - Verify attributes using `resource.TestCheckResourceAttr`.

## Example Patterns

### Resource Structure

```go
package provider

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

var _ resource.Resource = &exampleResource{}

type exampleResource struct {
    client *api.Client
}

type exampleResourceModel struct {
    ID   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
}

func (r *exampleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    // ... Schema definition ...
}

func (r *exampleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // ... Create logic ...
}
```

### Acceptance Test

```go
func TestAccExampleResource(t *testing.T) {
    if !isIntegrationTestMode() {
        t.Skip("Skipping acceptance test")
    }

    resource.Test(t, resource.TestCase{
        PreCheck: func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: providerConfig + readAccTestResource("resources/example/010_create.tf"),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("lightdash_example.test", "name", "value"),
                ),
            },
        },
    })
}
```

## Reference

- [Terraform Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework)
- [Provider Implementation Rules](.cursor/rules/provider-implementation.mdc)
- [Project Structure Rules](.cursor/rules/project-structure.mdc)
- [Testing Setup Reference](references/testing_setup.md)

## Assets

- [Resource Boilerplate](assets/resource_boilerplate.go)
- [Data Source Boilerplate](assets/data_source_boilerplate.go)
- [Acceptance Test Boilerplate](assets/test_boilerplate.go)
