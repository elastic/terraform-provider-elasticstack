# kibana-case-field-definition Specification

## Purpose
Defines the `elasticstack_kibana_case_field_definition` resource, which manages a single field in the Kibana Cases field library (Technical Preview API, Stack 9.6+) so that case custom-field definitions can be declared as code, reviewed, and drift-detected instead of configured by hand in the Kibana UI.

## Requirements

### Requirement: Field library CRUD API

The resource SHALL manage a field definition through `POST /api/cases/field_definitions` and `GET`/`PUT`/`DELETE /api/cases/field_definitions/{field_definition_id}`, routed through the space given by `space_id`. After a successful create or update, the resource SHALL read the field and derive final state from that read.
Reference: https://www.elastic.co/docs/api/doc/kibana/operation/operation-createcasefielddefinitiondefaultspace

#### Scenario: Create writes the configured field

- GIVEN a configuration with `owner = "securitySolution"` and a `definition` YAML whose internal `name` key is `triage-priority`
- WHEN apply creates the resource
- THEN the resource SHALL send a create request with that `owner` and `definition`
- AND the resulting state SHALL contain `name = "triage-priority"`, a non-empty `field_definition_id`, and `id = "default/<field_definition_id>"`

#### Scenario: Update preserves field identity

- GIVEN the resource exists with a given `field_definition_id`
- WHEN `definition`, `description`, or `is_global` changes and apply runs
- THEN the resource SHALL send an update request identifying the field by `field_definition_id` and including the required `owner` and `definition` fields
- AND the resource SHALL perform a read after the update and SHALL NOT change `field_definition_id` in the resulting state

### Requirement: Schema and validation

`owner` SHALL be required, restricted to `cases`, `securitySolution`, or `observability`, and force replacement. `definition` SHALL be required, valid YAML of at most 30000 characters, using a semantic-equality type. `name` SHALL be optional, computed, and force replacement. `description` and `is_global` SHALL be optional and computed and keep their prior state value when omitted. `space_id` SHALL be optional, computed, default `"default"`, and force replacement.

#### Scenario: Invalid YAML is rejected

- GIVEN `definition` that is not valid YAML
- WHEN Terraform validates the resource configuration
- THEN the resource SHALL return an `Invalid YAML` diagnostic on `definition`
- AND no API call SHALL be made

#### Scenario: Equivalent YAML formatting

- GIVEN two `definition` values that decode to the same YAML structure but differ in whitespace or map key order
- WHEN Terraform reconciles state against configuration
- THEN the resource SHALL treat them as semantically equal when comparing state values
- AND a reformat-only edit MAY still plan an in-place update (known v1 limitation, documented in the schema description), which SHALL apply without replacement

#### Scenario: Omitting description or is_global preserves the stored value

- GIVEN the resource exists with `description` and `is_global` set
- WHEN configuration removes those attributes and apply runs
- THEN state SHALL keep the prior values (the provider carries them forward, and the API also preserves them on omission, verified live on 9.6)
- AND setting `description = ""` SHALL clear the description

#### Scenario: Invalid owner is rejected at validation

- GIVEN a configuration with an `owner` other than `cases`, `securitySolution`, or `observability`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return a validation diagnostic and make no API call

#### Scenario: Omitting name defers to the server

- GIVEN a configuration that omits `name` and sets `definition`'s YAML `name` key to `triage-priority`
- WHEN apply creates the resource
- THEN the resulting state SHALL contain `name = "triage-priority"`

#### Scenario: Changing owner replaces the resource

- GIVEN the resource exists with `owner = "securitySolution"`
- WHEN the configuration changes `owner` to `observability`
- THEN the plan SHALL show the resource being replaced

#### Scenario: Changing name replaces the resource

- GIVEN the resource exists with `name = "triage-priority"`
- WHEN the configuration changes `name` to a different, YAML-consistent value
- THEN the plan SHALL show the resource being replaced, not an in-place update

### Requirement: Identity mismatch errors surface from the API

The resource SHALL NOT parse `definition`'s YAML to pre-empt identity constraints. If an in-place update changes the YAML `name` key while top-level `name` is unchanged, or changes the YAML `type` key, the resource SHALL surface the API's error diagnostic (expected HTTP 400, or HTTP 409 with `attributes.code = "field_identity_immutable"`, respectively) without a special-cased message.

#### Scenario: YAML-internal name no longer matches the top-level name

- GIVEN the resource exists with top-level `name = "triage-priority"` and a consistent YAML `name` key
- WHEN configuration changes only the YAML-internal `name` key to a different value, leaving the top-level `name` attribute unchanged, and apply runs
- THEN apply SHALL fail with a diagnostic derived from the API's error response

#### Scenario: YAML-internal type changes

- GIVEN the resource exists with a given YAML `type` key
- WHEN configuration changes only the YAML-internal `type` key, with `name` unchanged and consistent, and apply runs
- THEN apply SHALL fail with a diagnostic derived from the API's `field_identity_immutable` error response

### Requirement: Identity and import

The resource `id` SHALL be the composite `<space_id>/<field_definition_id>`. The resource SHALL support import by that composite id, and after import a refresh SHALL populate `space_id`, `field_definition_id`, `owner`, `definition`, `name`, `description`, `is_global`, and `display_order` from the API.

#### Scenario: Import by composite id

- GIVEN a field definition exists in space `default` with `field_definition_id = "abc-123"`
- WHEN `terraform import elasticstack_kibana_case_field_definition.example default/abc-123` runs
- THEN the state SHALL contain `id = "default/abc-123"`, `space_id = "default"`, and `field_definition_id = "abc-123"`

#### Scenario: Import in a named space

- GIVEN a field definition exists in space `security-team` with `field_definition_id = "abc-123"`
- WHEN `terraform import elasticstack_kibana_case_field_definition.example security-team/abc-123` runs
- THEN the state SHALL contain `space_id = "security-team"`

### Requirement: Read and not-found handling

On read, the resource SHALL populate state from the API's `GET` response. When the field definition cannot be found (HTTP 404), the resource SHALL remove itself from state. Any other API error SHALL be surfaced as a Terraform diagnostic and the resource SHALL NOT be removed from state.

#### Scenario: External deletion is detected on refresh

- GIVEN the resource is in Terraform state
- AND the field definition has been deleted directly via the API, outside Terraform
- WHEN refresh runs
- THEN the resource SHALL be removed from state
- AND the next plan SHALL propose re-creating the resource

#### Scenario: API error on read

- GIVEN the field library API returns a non-404 error on read
- WHEN refresh runs
- THEN the resource SHALL return the error as a diagnostic
- AND SHALL NOT remove the resource from state

### Requirement: Destroy

On destroy, the resource SHALL send `DELETE /api/cases/field_definitions/{field_definition_id}` and then remove itself from state. An HTTP 404 on delete SHALL be treated as success. A 409 response because the field is referenced by an active template or custom field SHALL be surfaced as a destroy-time diagnostic, not a silent success.

#### Scenario: Normal destroy

- GIVEN the resource exists and is not referenced by an active template or custom field
- WHEN `terraform destroy` runs and the API returns HTTP 204
- THEN the resource SHALL be deleted via the API
- AND the resource SHALL be removed from state

#### Scenario: Destroy when already deleted

- GIVEN the field definition has been deleted outside Terraform
- WHEN `terraform destroy` runs
- THEN destroy SHALL succeed and the resource SHALL be removed from state

#### Scenario: Destroy blocked by an active reference

- GIVEN the field definition is referenced by an active template or custom field
- WHEN `terraform destroy` runs and the API returns a 409
- THEN destroy SHALL fail with a diagnostic containing the API error
- AND the resource SHALL NOT be removed from state

### Requirement: Minimum stack version

The resource SHALL require Elastic Stack 9.6.0 or newer. When the target server is older, any lifecycle operation SHALL fail with an unsupported-version error stating the minimum version, and no field-definition API call SHALL be made.

#### Scenario: Server older than 9.6.0

- GIVEN the target stack version is 9.5.5
- WHEN apply creates the resource
- THEN the provider SHALL return an unsupported-version error mentioning `9.6.0`
- AND no field-definition API call SHALL be made

#### Scenario: Refresh against a server older than 9.6.0

- GIVEN the resource is in state and the target stack version is 9.5.5
- WHEN refresh runs
- THEN the provider SHALL return an unsupported-version error mentioning `9.6.0`

#### Scenario: Server at 9.6.0 or newer

- GIVEN the target stack version is 9.6.0 or newer
- WHEN apply creates the resource
- THEN the resource SHALL proceed with the create request

### Requirement: Provider-level Kibana client by default with scoped override

The resource SHALL use the provider's configured Kibana OpenAPI client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use it for all field-definition operations.

#### Scenario: Standard provider configuration

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any field-definition API call runs
- THEN it SHALL use the provider-level Kibana OpenAPI client

#### Scenario: Scoped Kibana connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any field-definition API call runs
- THEN it SHALL use the scoped Kibana OpenAPI client derived from that block

### Requirement: API and client error surfacing

For create, read, update, and delete, when the provider cannot obtain a Kibana OpenAPI client, the operation SHALL return an error diagnostic. Transport errors and unexpected HTTP statuses from the field-definition API SHALL also be surfaced as error diagnostics, except where this spec defines a specific not-found or already-deleted outcome.

#### Scenario: Missing Kibana client

- GIVEN the resource cannot obtain a Kibana OpenAPI client from provider configuration
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic
