## Purpose

Defines the `elasticstack_kibana_case_template` resource, which manages a single Kibana Cases template (standalone templates API, Technical Preview, Stack 9.5.2+) so that case defaults and custom-field pre-fills can be declared as code, reviewed, and drift-detected instead of configured by hand in the Kibana UI.

## Schema

```hcl
resource "elasticstack_kibana_case_template" "example" {
  owner           = string                        # required, one of cases|securitySolution|observability, forces replacement
  name            = string                         # required, 1-50 chars, unique per owner (case-insensitive)
  definition_yaml = <string, normalized YAML>      # required, at most 30000 characters; case defaults and fields/$ref entries
  description     = <optional, string>             # at most 1000 chars; omitting on update clears it
  tags            = <optional, list of string>     # at most 200 items, each at most 256 chars; order and duplicates preserved
  is_enabled      = <optional, computed, bool>     # server default true
  space_id        = <optional, computed, string>   # default "default"; forces replacement

  # Read-only
  id               = <computed, string>            # "<space_id>/<template_id>"
  template_id      = <computed, string>            # server-assigned, stable across versions
  template_version = <computed, int64>             # increments on every successful update
  field_count      = <computed, nullable int64>
  usage_count      = <computed, nullable int64>
  is_default       = <computed, nullable bool>

  kibana_connection {                              # optional scoped-connection override
    # ...
  }
}
```

## ADDED Requirements

### Requirement: Case template CRUD API

The resource SHALL manage a case template through `POST /api/cases/templates` and `GET`/`PUT`/`DELETE /api/cases/templates/{template_id}`, routed through the space given by `space_id`. After a successful create or update, the resource SHALL read the template and derive final state from that read.
Reference: https://www.elastic.co/docs/explore-analyze/cases/create-case-templates

#### Scenario: Create writes the configured template

- GIVEN a configuration with `owner = "securitySolution"`, `name = "triage-template"`, and a `definition_yaml` declaring case defaults and a `fields` array
- WHEN apply creates the resource
- THEN the resource SHALL send a create request with that `owner`, `name`, and `definition_yaml`
- AND the resulting state SHALL contain a non-empty `template_id`, `template_version = 1`, and `id = "default/<template_id>"`

#### Scenario: Update sends the full replacement body

- GIVEN the resource exists with a given `template_id`
- WHEN `name`, `definition_yaml`, `description`, `tags`, or `is_enabled` changes and apply runs
- THEN the resource SHALL send an update request identifying the template by `template_id` and including the required `owner`, `name`, and `definition_yaml` fields
- AND the resource SHALL perform a read after the update and SHALL NOT change `template_id` in the resulting state
- AND the resulting state SHALL contain the configured `name`
- AND `template_version` SHALL increment

### Requirement: Schema and validation

`owner` SHALL be required, restricted to `cases`, `securitySolution`, or `observability`, and force replacement. `name` SHALL be required, 1-50 characters, and unique per `owner` (case-insensitive). `definition_yaml` SHALL be required, valid YAML of at most 30000 characters, using a semantic-equality type. `space_id` SHALL be optional, computed, default `"default"`, and force replacement. See the dedicated schema-attribute scenarios below for `description`, `tags`, and `is_enabled`.

#### Scenario: Optional attribute constraints

- GIVEN a configuration setting `description`, `tags`, and `is_enabled`
- WHEN Terraform validates the configuration
- THEN `description` SHALL be accepted up to 1000 characters
- AND `tags` SHALL be accepted as a list of at most 200 strings, each at most 256 characters, preserving order and duplicates
- AND `is_enabled` SHALL be optional and computed, defaulting to `true` server-side when omitted

#### Scenario: Invalid YAML is rejected

- GIVEN `definition_yaml` that is not valid YAML
- WHEN Terraform validates the resource configuration
- THEN the resource SHALL return an `Invalid YAML` diagnostic on `definition_yaml`
- AND no API call SHALL be made

#### Scenario: Invalid owner is rejected at validation

- GIVEN a configuration with an `owner` other than `cases`, `securitySolution`, or `observability`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return a validation diagnostic and make no API call

#### Scenario: Tags preserve order and duplicates

- GIVEN a configuration with `tags = ["z", "a", "z"]`
- WHEN apply creates or updates the resource
- THEN the resulting state SHALL contain `tags = ["z", "a", "z"]` in that exact order

#### Scenario: Changing owner replaces the resource

- GIVEN the resource exists with `owner = "securitySolution"`
- WHEN the configuration changes `owner` to `observability`
- THEN the plan SHALL show the resource being replaced

### Requirement: `definition_yaml` diff suppression for semantically equivalent edits

A config edit to `definition_yaml` that changes only whitespace, comments, or mapping-key order, without changing the parsed YAML structure, SHALL produce an empty plan and SHALL NOT result in an update call. A config edit that changes the parsed YAML structure SHALL produce a non-empty plan and SHALL result in an update call that increments `template_version`.

#### Scenario: Reformat-only edit produces an empty plan

- GIVEN the resource exists with a given `definition_yaml` and `template_version`
- WHEN the configuration's `definition_yaml` is edited only in whitespace, comments, or mapping-key order
- THEN `terraform plan` SHALL report no changes
- AND no update request SHALL be sent
- AND `template_version` SHALL remain unchanged after a subsequent apply

#### Scenario: Structural edit produces a non-empty plan

- GIVEN the resource exists with a given `definition_yaml` and `template_version`
- WHEN the configuration's `definition_yaml` is edited so the parsed YAML structure changes (for example, a field value changes)
- THEN `terraform plan` SHALL report an in-place update
- AND apply SHALL send an update request
- AND `template_version` SHALL increment

### Requirement: Null-vs-empty handling for `description` and `tags`

Omitting `description` or `tags` from configuration on create SHALL result in a null value in state. Omitting a previously set `description` or `tags` from configuration on a later update SHALL clear the stored value, and the resulting state SHALL be null. Setting `description = ""` or `tags = []` explicitly SHALL be accepted and SHALL remain an empty value in state, distinct from omission; the resource SHALL NOT merge or coalesce the two forms.

#### Scenario: Omitted description and tags on create

- GIVEN a configuration that omits `description` and `tags`
- WHEN apply creates the resource
- THEN the resulting state SHALL contain `description = null` and `tags = null`

#### Scenario: Removing description and tags on update clears them

- GIVEN the resource exists with `description` and `tags` set to non-empty values
- WHEN the configuration removes both attributes and apply runs
- THEN the resulting state SHALL contain `description = null` and `tags = null`

#### Scenario: Explicit empty description and tags remain distinct from omission

- GIVEN a configuration with `description = ""` and `tags = []`
- WHEN apply creates or updates the resource
- THEN the resulting state SHALL contain `description = ""` and `tags = []`
- AND neither value SHALL be coalesced to null

### Requirement: Identity and import

The resource `id` SHALL be the composite `<space_id>/<template_id>`. The resource SHALL support import by that composite id, and after import a refresh SHALL populate `space_id`, `template_id`, `owner`, `name`, `definition_yaml`, `description`, `tags`, `is_enabled`, `template_version`, `field_count`, `usage_count`, and `is_default` from the API.

#### Scenario: Import by composite id

- GIVEN a case template exists in space `default` with `template_id = "abc-123"`
- WHEN `terraform import elasticstack_kibana_case_template.example default/abc-123` runs
- THEN the state SHALL contain `id = "default/abc-123"`, `space_id = "default"`, and `template_id = "abc-123"`

#### Scenario: Import in a named space

- GIVEN a case template exists in space `security-team` with `template_id = "abc-123"`
- WHEN `terraform import elasticstack_kibana_case_template.example security-team/abc-123` runs
- THEN the state SHALL contain `space_id = "security-team"`

### Requirement: Read and not-found handling

On read, the resource SHALL populate state from the API's `GET` response. When the template cannot be found (HTTP 404), the resource SHALL remove itself from state. Any other API error SHALL be surfaced as a Terraform diagnostic and the resource SHALL NOT be removed from state.

#### Scenario: External deletion is detected on refresh

- GIVEN the resource is in Terraform state
- AND the template has been deleted directly via the API, outside Terraform
- WHEN refresh runs
- THEN the resource SHALL be removed from state
- AND the next plan SHALL propose re-creating the resource

#### Scenario: API error on read

- GIVEN the templates API returns a non-404 error on read
- WHEN refresh runs
- THEN the resource SHALL return the error as a diagnostic
- AND SHALL NOT remove the resource from state

### Requirement: Destroy

On destroy, the resource SHALL send `DELETE /api/cases/templates/{template_id}` and then remove itself from state. An HTTP 404 on delete SHALL be treated as success. Destroy soft-deletes all versions of the template; existing cases created from it keep their pinned template reference and are unaffected.

#### Scenario: Normal destroy

- GIVEN the resource exists
- WHEN `terraform destroy` runs and the API returns HTTP 204
- THEN the resource SHALL be deleted via the API
- AND the resource SHALL be removed from state

#### Scenario: Destroy when already deleted

- GIVEN the template has been deleted outside Terraform
- WHEN `terraform destroy` runs
- THEN destroy SHALL succeed and the resource SHALL be removed from state

### Requirement: Create and update error surfacing

A duplicate template name within the same owner (case-insensitive) SHALL surface as an apply-time diagnostic derived from the API's 409 response. A malformed template definition (for example, an invalid field `control` value) SHALL surface as an apply-time diagnostic derived from the API's 400 response. Neither case SHALL be pre-validated client-side beyond `definition_yaml`'s plan-time YAML syntax check.

#### Scenario: Duplicate name within the same owner

- GIVEN a template named `"triage-template"` already exists for `owner = "securitySolution"`
- WHEN a second resource configuration with the same `owner` and `name` is applied
- THEN apply SHALL fail with a diagnostic derived from the API's 409 response

#### Scenario: Malformed template schema

- GIVEN `definition_yaml` is syntactically valid YAML but declares a field with an invalid `control` value
- WHEN apply creates or updates the resource
- THEN apply SHALL fail with a diagnostic derived from the API's 400 response

### Requirement: Minimum stack version

The resource SHALL require Elastic Stack 9.5.2 or newer. When the target server is older, any lifecycle operation SHALL fail with an unsupported-version error stating the minimum version, and no templates API call SHALL be made.

#### Scenario: Server older than 9.5.2

- GIVEN the target stack version is 9.5.1
- WHEN apply creates the resource
- THEN the provider SHALL return an unsupported-version error mentioning `9.5.2`
- AND no templates API call SHALL be made

#### Scenario: Refresh against a server older than 9.5.2

- GIVEN the resource is in state and the target stack version is 9.5.1
- WHEN refresh runs
- THEN the provider SHALL return an unsupported-version error mentioning `9.5.2`

#### Scenario: Server at 9.5.2 or newer

- GIVEN the target stack version is 9.5.2 or newer
- WHEN apply creates the resource
- THEN the resource SHALL proceed with the create request

### Requirement: Provider-level Kibana client by default with scoped override

The resource SHALL use the provider's configured Kibana OpenAPI client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use it for all template operations.

#### Scenario: Standard provider configuration

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any template API call runs
- THEN it SHALL use the provider-level Kibana OpenAPI client

#### Scenario: Scoped Kibana connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any template API call runs
- THEN it SHALL use the scoped Kibana OpenAPI client derived from that block

### Requirement: API and client error surfacing

For create, read, update, and delete, when the provider cannot obtain a Kibana OpenAPI client, the operation SHALL return an error diagnostic. Transport errors and unexpected HTTP statuses from the templates API SHALL also be surfaced as error diagnostics, except where this spec defines a specific not-found, already-deleted, duplicate-name, or malformed-definition outcome.

#### Scenario: Missing Kibana client

- GIVEN the resource cannot obtain a Kibana OpenAPI client from provider configuration
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic
