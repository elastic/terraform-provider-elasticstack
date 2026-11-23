## 1. Fix

- [x] 1.1 Pass an `isImport` flag to `populateFromAPI` and map `output_id` from the API only when it is configured or on import
- [x] 1.2 Pass `isImport` from `read.go` and `false` from `create.go` and `update.go`
- [x] 1.3 Document on the `output_id` attribute that removing it from configuration does not clear it on the server

## 2. Tests

- [x] 2.1 Unit tests for `populateFromAPI`: configured value is refreshed, unconfigured stays null for an output ID and for `""`, import populates the value
- [x] 2.2 Shared acceptance-test helper in `internal/acctest` that changes Fleet objects through the configured Fleet client
- [x] 2.3 Acceptance test: output set outside Terraform while `output_id` is unset; apply succeeds, state stays null, server keeps the output
- [x] 2.4 Acceptance test: `output_id` set to `""` outside Terraform; apply succeeds and state stays null
- [x] 2.5 Acceptance test: configured `output_id` removed from configuration; apply succeeds, state is null, server keeps the output
- [x] 2.6 Acceptance test: import of a policy with an output populates `output_id`
- [x] 2.7 Acceptance test for `elasticstack_fleet_agent_policy` outputs set outside Terraform

## 3. Validation

- [x] 3.1 Run unit tests and golangci-lint
- [x] 3.2 Run `openspec validate --all`
