# Integration Tests

1. Copy `testing.tfvars.template` to `testing.tfvars` and set the variables.
2. Run the following command to apply the changes.

```shell
TF_LOG=DEBUG terraform apply -var-file=testing.tfvars
```
