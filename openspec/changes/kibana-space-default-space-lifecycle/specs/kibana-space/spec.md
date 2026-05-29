## MODIFIED Requirements

### Requirement: Delete lifecycle for the default Kibana space (REQ-DELETE-DEFAULT)

When the resource is configured with `space_id = "default"` and a delete operation runs, the provider SHALL NOT call `DELETE /api/spaces/space/default`. Instead, the provider SHALL remove the resource from Terraform state only and SHALL emit a warning-level log message to surface the skip to operators.

The Kibana API permanently rejects `DELETE /api/spaces/space/default` with HTTP 400 Bad Request. This is a hard platform invariant on all supported Kibana versions; encoding it directly in the provider is the correct approach.

#### Scenario: Destroy default space — skip API call and remove from state

- **GIVEN** a `elasticstack_kibana_space` resource with `space_id = "default"` is in Terraform state
- **WHEN** `terraform destroy` runs
- **THEN** the provider SHALL NOT call `DELETE /api/spaces/space/default`
- **AND** the provider SHALL remove the resource from Terraform state
- **AND** the provider SHALL emit a `tflog.Warn` with the message: `"default Kibana space cannot be deleted; removing from Terraform state only"`

#### Scenario: Destroy non-default space — normal API delete

- **GIVEN** a `elasticstack_kibana_space` resource with `space_id` set to any value other than `"default"`
- **WHEN** `terraform destroy` runs
- **THEN** the provider SHALL call `DELETE /api/spaces/space/{space_id}` as before

### Requirement: Create 409 Conflict diagnostic (REQ-CREATE-409)

When `POST /api/spaces/space` returns HTTP 409 Conflict, the provider SHALL return an error diagnostic that:
- names the space id from the request
- instructs the practitioner to import the existing space using `terraform import elasticstack_kibana_space.<NAME> <space_id>`

The provider SHALL NOT attempt to auto-fallback to `PUT` (auto-import) on 409. Explicit import is the required workflow.

#### Scenario: Create fails with 409 — actionable diagnostic returned

- **GIVEN** a `elasticstack_kibana_space` resource with `space_id = "default"` (or any id of an existing space)
- **WHEN** `terraform apply` runs and Kibana returns HTTP 409 Conflict for `POST /api/spaces/space`
- **THEN** the provider SHALL return an error diagnostic naming the conflicting space id
- **AND** the diagnostic SHALL include an import command of the form `terraform import elasticstack_kibana_space.<NAME> <space_id>`
- **AND** the provider SHALL NOT attempt a `PUT /api/spaces/space/{id}` auto-fallback

#### Scenario: Create fails with other errors — existing behavior unchanged

- **GIVEN** a `elasticstack_kibana_space` create request
- **WHEN** Kibana returns any HTTP status other than 200 or 409
- **THEN** the provider SHALL surface the error via the existing `HandleMutateTypedResponse` path (unchanged)

## ADDED Requirements

### Requirement: Default-space acceptance test coverage (REQ-TEST-DEFAULT-SPACE)

The acceptance test suite SHALL include a test `TestAccResourceSpace_DefaultSpace` that verifies the complete import-update-destroy lifecycle for the default Kibana space without gating on a minimum stack version.

#### Scenario: Import default space, update, then destroy without error

- **GIVEN** a live Kibana instance with a default space
- **WHEN** `TestAccResourceSpace_DefaultSpace` runs
- **THEN** step 1 SHALL import the default space (`ImportState: true`, `ImportStateId: "default"`)
- **AND** step 2 SHALL apply a fixture config with only `space_id` and `name` (no `solution`) and assert `space_id == "default"` and `name == "Default"`
- **AND** the destroy step at the end of the test SHALL complete without error
- **AND** the test SHALL use no `CheckDestroy` (the default space persists after Terraform destroy)
