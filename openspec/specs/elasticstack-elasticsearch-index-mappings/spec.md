# elasticstack-elasticsearch-index-mappings Specification

## Purpose
TBD - created by archiving change elasticsearch-index-mappings. Update Purpose after archive.
## Requirements
### Requirement: Schema (REQ-001)

The resource SHALL expose the following top-level attributes:

```hcl
resource "elasticstack_elasticsearch_index_mappings" "example" {
  id    = <computed, string>                        # "<cluster_uuid>/<index_name>"
  index = <required, string, forces replacement>    # name of the target Elasticsearch index

  mappings = <required, JSON string>                # all top-level mapping keys accepted:
                                                    # properties, dynamic, _source,
                                                    # dynamic_templates, runtime, etc.
                                                    # MUST be a non-empty JSON object

  elasticsearch_connection { ... }                  # standard provider connection block
}
```

- `id` SHALL be computed and unknown until create completes; it SHALL use `stringplanmodifier.UseStateForUnknown()`.
- `index` SHALL be required and SHALL force resource replacement when changed.
- `mappings` SHALL be a required JSON string using `index.MappingsType{}` as the custom type for semantic equality normalization. It SHALL accept any combination of valid top-level Elasticsearch mapping parameters (not limited to `properties`). It SHALL additionally be validated as a non-empty JSON object — an empty object (`{}`) SHALL be rejected at plan time.
- `elasticsearch_connection` is injected by the provider scaffold and SHALL NOT be declared manually in the schema factory.

#### Scenario: Schema validation — index is required

- GIVEN a configuration that omits `index`
- WHEN `terraform validate` runs
- THEN Terraform SHALL emit a required-attribute error

#### Scenario: Schema validation — mappings is required

- GIVEN a configuration that omits `mappings`
- WHEN `terraform validate` runs
- THEN Terraform SHALL emit a required-attribute error

#### Scenario: Schema validation — mappings must be non-empty JSON object

- GIVEN a configuration where `mappings = jsonencode({})`
- WHEN `terraform validate` runs
- THEN Terraform SHALL emit an attribute validation error stating the value must be a non-empty JSON object

### Requirement: Create — index must exist (REQ-002)

On create, the resource SHALL verify that the target index exists before issuing `PUT /{index}/_mapping`. If the index does not exist, the resource SHALL return an error diagnostic and SHALL NOT create any Elasticsearch resource.

#### Scenario: Create fails when index is absent

- GIVEN the target `index` does not exist in Elasticsearch
- WHEN `terraform apply` runs the create operation
- THEN Terraform diagnostics SHALL include an error stating the index was not found
- AND no API mapping update call SHALL be issued

#### Scenario: Create succeeds when index exists

- GIVEN the target `index` exists in Elasticsearch
- WHEN `terraform apply` runs the create operation
- THEN `PUT /{index}/_mapping` SHALL be called with the `mappings` JSON as the request body
- AND the resource SHALL be added to state with `id = "<cluster_uuid>/<index_name>"`

---

### Requirement: Update (REQ-003)

On update, the resource SHALL call `PUT /{index}/_mapping` with the new `mappings` JSON. No existence check is required on update (the index is expected to persist between plan and apply).

#### Scenario: Field added on update

- GIVEN a managed index with one declared field in `properties`
- WHEN the user adds a second field to `mappings` and runs `terraform apply`
- THEN `PUT /{index}/_mapping` SHALL be called with the updated JSON
- AND both fields SHALL appear in the next `terraform plan` with no diff

---

### Requirement: Read — user-declared subset only (REQ-004)

On read, the resource SHALL retrieve the index metadata via the existing `GetIndex` helper (which returns the `Mappings` payload from the index state) and reconstruct the stored `mappings` by retaining **only the top-level keys that are present in the previously stored state**. Top-level keys that exist in the API response but are absent from the stored state SHALL be silently discarded.

For the `properties` top-level key, the filtering SHALL be **recursive**: only field names that appear in the previously stored `properties` tree SHALL be retained at every nesting level. Dynamically-added fields within `properties` that are absent from the stored state SHALL be silently discarded.

If the previously stored `mappings` is empty (e.g. immediately after `terraform import` via `ImportStatePassthroughID`), the resource SHALL store the full API response as the initial mask. This allows users to narrow the declaration in subsequent configuration changes.

The resource SHALL use `index.MappingsType{}` semantic equality so that equivalent JSON representations (different key ordering, different whitespace) do not produce a spurious diff.

#### Scenario: Dynamic extras do not cause drift

- GIVEN a resource that declares only `properties` with two explicit fields (`title` and `body`)
- AND Elasticsearch adds a dynamic field `tags` to the index (e.g. via auto-mapping a document)
- WHEN `terraform plan` runs after the dynamic field is added
- THEN the plan SHALL show no diff for the `mappings` attribute
- AND the stored state SHALL continue to contain only `title` and `body` under `properties`

#### Scenario: Not found on read removes from state

- GIVEN the target index is deleted outside Terraform
- WHEN `terraform refresh` or `terraform plan` runs
- THEN the resource SHALL be removed from state
- AND Terraform SHALL propose recreating it on the next apply

---

### Requirement: Delete — no-op (REQ-005)

On `terraform destroy`, the resource SHALL remove itself from Terraform state without issuing any API call. Elasticsearch does not support removing field mappings without a full reindex; the resource acknowledges this constraint by design.

The resource description and documentation SHALL clearly state that `destroy` does not revert or remove field mappings from the index.

#### Scenario: Destroy leaves index mappings intact

- GIVEN a managed resource with one or more declared fields
- WHEN `terraform destroy` runs
- THEN no Elasticsearch API call SHALL be issued
- AND the resource SHALL be removed from Terraform state
- AND the field mappings SHALL remain on the Elasticsearch index unchanged

---

### Requirement: Identity and import (REQ-006)

The resource `id` SHALL follow the format `<cluster_uuid>/<index_name>`. The resource SHALL support `terraform import` using the same ID format via `resource.ImportStatePassthroughID`.

Because `ImportStatePassthroughID` sets only the resource ID, the first Read after import encounters an empty `mappings` state. The resource SHALL handle this by storing the **full API mappings** as the initial state. Users are expected to adjust their Terraform configuration to the desired subset; the first apply after import will converge to that subset and subsequent reads will filter dynamically-added fields.

#### Scenario: Import

- GIVEN an existing index with known mappings including fields `title` and `body`
- WHEN the user runs `terraform import elasticstack_elasticsearch_index_mappings.example <cluster_uuid>/<index_name>`
- THEN the resource SHALL be added to state with the full `mappings` from the API
- AND a subsequent `terraform plan` with a narrowed config (e.g. only `title`) SHALL show a diff
- AND `terraform apply` SHALL converge state to the declared subset
- AND a later `terraform plan` SHALL show no diff if no new dynamic fields are added

---

### Requirement: API errors surface as diagnostics (REQ-007)

When the Elasticsearch API returns a non-success response (other than 404 on read), the resource SHALL surface the API error to Terraform diagnostics rather than silently ignoring it.

#### Scenario: API failure on create

- GIVEN the Elasticsearch `PUT /{index}/_mapping` API returns a non-success response
- WHEN create runs
- THEN Terraform diagnostics SHALL include the API error

#### Scenario: API failure on update

- GIVEN the Elasticsearch `PUT /{index}/_mapping` API returns a non-success response during update
- WHEN update runs
- THEN Terraform diagnostics SHALL include the API error

