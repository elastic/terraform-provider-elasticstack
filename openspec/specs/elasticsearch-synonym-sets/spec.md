# elasticsearch-synonym-sets Specification

## Purpose
TBD - created by archiving change elasticsearch-synonym-sets. Update Purpose after archive.
## Requirements
### Requirement: Synonym set APIs (REQ-001)

The `elasticstack_elasticsearch_synonym_set` resource SHALL use `PUT /_synonyms/{id}` to create or replace a synonym set, `GET /_synonyms/{id}` to read it, and `DELETE /_synonyms/{id}` to remove it. Non-success API responses (other than 404 on read) SHALL be surfaced as Terraform diagnostics. The data source SHALL use `GET /_synonyms/{id}` (with full pagination) to look up a synonym set by `synonym_set_id`.

#### Scenario: API errors surfaced

- GIVEN a failing Elasticsearch API response (other than 404 on read)
- WHEN the provider processes the response
- THEN diagnostics SHALL include the API error message

#### Scenario: Data source synonym set not found

- GIVEN a `synonym_set_id` that does not exist in Elasticsearch
- WHEN the data source reads
- THEN diagnostics SHALL include an error indicating the synonym set was not found

### Requirement: Identity (REQ-002)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<synonym_set_id>`. The resource SHALL compute `id` from the current cluster UUID and the configured `synonym_set_id` after a successful PUT. The data source SHALL expose the same computed `id` format.

#### Scenario: ID set after create

- GIVEN a successful PUT synonym set call
- WHEN create completes
- THEN `id` in state SHALL be `<cluster_uuid>/<synonym_set_id>`

#### Scenario: Data source ID set

- GIVEN a successful GET synonym set call from the data source
- WHEN the data source reads
- THEN `id` in state SHALL be `<cluster_uuid>/<synonym_set_id>`

### Requirement: synonym_set_id requires replacement (REQ-003)

Changing `synonym_set_id` SHALL require resource replacement, because synonym set identifiers are immutable once created in Elasticsearch.

#### Scenario: synonym_set_id change triggers replace

- GIVEN `synonym_set_id` changes in configuration
- WHEN Terraform plans
- THEN replacement SHALL be required

### Requirement: Synonym rule ID — Optional+Computed (REQ-004)

Within each `synonyms_set` block, `id` SHALL be `Optional` and `Computed`. When the user omits `id` from configuration, the provider SHALL generate a stable UUID and store it in state on create. On subsequent applies the provider SHALL use the stored ID in the PUT request body, keeping the set fully deterministic. When the user provides an explicit `id`, that value SHALL be used as-is.

#### Scenario: Rule ID auto-generated when omitted

- GIVEN a `synonyms_set` block with no `id` set
- WHEN create runs
- THEN the provider SHALL generate a UUID and store it as `id` in state

#### Scenario: Stored rule ID reused on re-apply

- GIVEN a resource was created with a provider-generated rule `id` in state
- WHEN apply runs again with the same configuration (no `id` specified)
- THEN the PUT request SHALL include the stored rule `id` and Terraform SHALL plan no changes

#### Scenario: Explicit rule ID preserved

- GIVEN a `synonyms_set` block with `id = "rule-car"`
- WHEN create and read run
- THEN `id` in state SHALL be `"rule-car"`

### Requirement: Synonym rules list ordering (REQ-005)

`synonyms_set` SHALL be a list (ordered), not a set. The resource SHALL preserve rule ordering on create and update. A subsequent plan against the same configuration SHALL show no changes as long as the API returns rules in the same order.

#### Scenario: Rule order preserved round-trip

- GIVEN a synonym set created with rules in a specific order
- WHEN the resource reads state after create
- THEN rules in state SHALL appear in the same order as configured and a subsequent plan SHALL show no changes

### Requirement: Paginated read — all rules retrieved (REQ-006)

The `GET /_synonyms/{id}` API supports `from`/`size` pagination. The resource SHALL loop through all pages (using `size=500` per request) until all rules are retrieved, so that synonym sets of any size are fully represented in Terraform state.

#### Scenario: All rules retrieved from a large set

- GIVEN a synonym set with more than 500 rules
- WHEN read runs
- THEN all rules SHALL be present in state

### Requirement: Read — not found removes resource from state (REQ-007)

When `GET /_synonyms/{id}` returns a 404 response, the resource SHALL remove itself from state without returning an error diagnostic, allowing Terraform to plan a re-create on the next apply.

#### Scenario: Synonym set deleted outside Terraform

- GIVEN the synonym set was deleted in Elasticsearch outside of Terraform
- WHEN the resource read runs
- THEN the resource SHALL be removed from state

### Requirement: Delete — clear diagnostic when set is in use (REQ-008)

When `DELETE /_synonyms/{id}` returns HTTP 400 (because the set is referenced by an index analyzer), the provider SHALL return a descriptive error diagnostic explaining that the synonym set is still referenced by an index analyzer and that the user must remove it from all analyzer configurations before retrying destroy.

#### Scenario: Delete blocked by active analyzer reference

- GIVEN the synonym set is referenced by an active index analyzer
- WHEN destroy runs
- THEN the provider SHALL return an error diagnostic that explains the synonym set is in use (not a generic API error)

### Requirement: Update replaces entire set (REQ-009)

Changing any `synonyms_set` entry (adding, removing, or modifying a rule) SHALL trigger an in-place update via `PUT /_synonyms/{id}`, replacing the entire set atomically. No replacement of the Terraform resource is required for rule-level changes.

#### Scenario: Rule modification triggers PUT without resource replace

- GIVEN a synonym set resource exists in state
- WHEN a `synonyms_set` rule's `synonyms` value changes in configuration
- THEN Terraform SHALL plan an in-place update (not a replacement)
- AND the PUT request SHALL contain the full updated rule list

#### Scenario: Rule addition triggers PUT

- GIVEN a synonym set resource exists with two rules in state
- WHEN a third rule is added to `synonyms_set` in configuration
- THEN Terraform SHALL plan an update and the PUT request SHALL contain all three rules

### Requirement: Connection (REQ-010)

By default, the resource and data source SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource or data source SHALL construct and use a resource-scoped Elasticsearch client for all API calls.

#### Scenario: Resource-level connection override

- GIVEN `elasticsearch_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource-scoped client SHALL be used instead of the provider client

#### Scenario: Data source connection override

- GIVEN `elasticsearch_connection` is configured on the data source
- WHEN the data source reads
- THEN the resource-scoped client SHALL be used instead of the provider client

### Requirement: Import state (REQ-011)

The resource SHALL implement `resource.ResourceWithImportState`. The import ID SHALL be the full resource `id` in the format `<cluster_uuid>/<synonym_set_id>`. On import, the resource SHALL set `id` to the provided import ID and call the Read path to populate all other attributes.

#### Scenario: Import sets state from API

- GIVEN a valid import ID of the form `<cluster_uuid>/<synonym_set_id>`
- WHEN `terraform import` runs
- THEN `id` in state SHALL equal the provided import ID
- AND all `synonyms_set` rules SHALL be populated from the API by the subsequent Read call

#### Scenario: Import followed by plan shows no diff

- GIVEN a resource was successfully imported
- WHEN `terraform plan` runs against a matching configuration
- THEN no attribute differences SHALL be shown and no replacement SHALL be planned

### Requirement: Acceptance test coverage (REQ-012)

The acceptance test suite SHALL cover:
- Basic CRUD: create a synonym set, verify state, update rules (add and modify), verify state, destroy.
- Rule ordering: verify that round-trip preserves the order of synonym rules.
- Optional rule ID: create with `id` omitted from at least one rule; verify the provider generates a stable ID; apply again with the same config and verify no diff.
- Import: create resource, import by composite ID, verify state matches, run plan and confirm no diff.
- Data source: create resource, read via data source, verify all attributes match.

#### Scenario: Acceptance test import step

- GIVEN an existing synonym set created in a prior acceptance test step
- WHEN `ImportState: true, ImportStateVerify: true` runs with the composite ID
- THEN all attributes in the imported state SHALL match the originally configured attributes

#### Scenario: Acceptance test data source

- GIVEN an existing synonym set managed by the resource
- WHEN the data source reads by `synonym_set_id`
- THEN all `synonyms_set` rules returned SHALL match those in the resource state

