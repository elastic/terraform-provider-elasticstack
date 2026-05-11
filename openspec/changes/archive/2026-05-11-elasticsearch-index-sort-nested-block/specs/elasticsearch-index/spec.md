## ADDED Requirements

### Requirement: Nested `sort` attribute for per-field sort configuration (REQ-SORT-01)

The `elasticstack_elasticsearch_index` resource SHALL expose a new optional `sort` attribute as a `ListNestedAttribute`. Each element of the list SHALL represent one sort entry with the following nested attributes:

- `field` (required, string): The index field to sort by. Must have `doc_values` enabled (e.g. boolean, numeric, date, keyword).
- `order` (optional, string, allowed: `"asc"`, `"desc"`): The sort direction. Defaults to `"asc"` at the Elasticsearch level when not specified.
- `missing` (optional, string, allowed: `"_last"`, `"_first"`): How to treat documents that are missing the sort field. Defaults to `"_last"` at the Elasticsearch level when not specified.
- `mode` (optional, string, allowed: `"min"`, `"max"`): Which value to use when a sort field has multiple values. Defaults to `"min"` when order is `asc` and `"max"` when order is `desc` at the Elasticsearch level when not specified.

The `sort` attribute maps to the Elasticsearch static settings `index.sort.field`, `index.sort.order`, `index.sort.missing`, and `index.sort.mode`. Because these are immutable static settings, any change to the configured `sort` list SHALL require resource replacement, subject to the migration suppression rules in REQ-SORT-03.

The `sort` attribute and the deprecated `sort_field`/`sort_order` attributes SHALL be mutually exclusive. The schema SHALL enforce this with a `ConflictsWith` validator that produces a plan-time error when both `sort` and either `sort_field` or `sort_order` are set in the same configuration.

#### Scenario: Index created with nested sort attribute

- **GIVEN** a configuration with `sort = [{ field = "date", order = "desc", missing = "_last" }]`
- **WHEN** the resource is created
- **THEN** the Elasticsearch index SHALL be created with `index.sort.field = ["date"]`, `index.sort.order = ["desc"]`, and `index.sort.missing = ["_last"]`

#### Scenario: Multi-field sort preserves order

- **GIVEN** a configuration with `sort = [{ field = "date", order = "desc" }, { field = "username", order = "asc" }]`
- **WHEN** the resource is created
- **THEN** the Elasticsearch index SHALL be created with `index.sort.field = ["date", "username"]` and `index.sort.order = ["desc", "asc"]` in that order

#### Scenario: Mixing `sort` and `sort_field` is rejected at plan time

- **GIVEN** a configuration that sets both `sort` and `sort_field`
- **WHEN** Terraform validates the configuration
- **THEN** validation SHALL fail with a diagnostic before any API call is made

#### Scenario: Changing `sort` requires replace

- **GIVEN** an existing index managed with the `sort` attribute
- **WHEN** a configuration change modifies any entry in `sort` (e.g. changes `order`)
- **THEN** Terraform SHALL plan to destroy and recreate the resource

---

### Requirement: Deprecate `sort_field` and `sort_order` attributes (REQ-SORT-02)

The existing `sort_field` and `sort_order` attributes SHALL remain in the schema as deprecated optional attributes, still functioning as before. Both SHALL carry a `DeprecationMessage` directing users to use the `sort` attribute instead. Both SHALL carry `ConflictsWith` validators that produce a plan-time error when used alongside the new `sort` attribute.

#### Scenario: Deprecated attributes still work after provider upgrade

- **GIVEN** an existing configuration that uses `sort_field` and `sort_order`
- **WHEN** the provider is upgraded to the version containing this change
- **THEN** Terraform SHALL plan no changes and the resource SHALL remain under management without requiring migration

#### Scenario: Deprecation warning is surfaced

- **GIVEN** a configuration that uses `sort_field` or `sort_order`
- **WHEN** Terraform plans or applies
- **THEN** Terraform SHALL surface a deprecation warning for the attribute

---

### Requirement: Private-state-backed migration path from legacy to new `sort` attribute (REQ-SORT-03)

The resource SHALL store the ordered sort configuration from Elasticsearch in private state during every `Read` operation. This ordered sort configuration SHALL be stored under a private state key `"sort_config"` as a JSON-marshaled object containing ordered arrays for `fields`, `orders`, and optional per-position `missing`/`mode` values (as reported by Elasticsearch static settings).

When Terraform plans a configuration that:
1. Has `sort` as null in state (the resource was created using the deprecated attributes),
2. Has a non-null `sort` in the plan (the user is migrating to the new attribute),
3. And private state contains the ordered sort config from Elasticsearch,

the plan modifier on `sort` SHALL compare the plan's `sort[*].field` and `sort[*].order` (treating null `order` as `"asc"`) against the private state. The modifier SHALL also compare `sort[*].missing` and `sort[*].mode` using semantic normalization against existing index settings:

- Treat explicit defaults as equivalent to absent settings in both plan and Elasticsearch (`missing`: `"_last"`; `mode`: `"min"` when order is `asc`, `"max"` when order is `desc`).
- Compare values per position after order normalization.

When fields and orders match exactly (in order), and all planned `missing`/`mode` values are semantically equivalent to the existing index settings, the modifier SHALL suppress replace so users can migrate representations without destroying the index.

If any planned `sort[*].missing` or `sort[*].mode` value is not semantically equivalent to the existing index setting at the same position, replace SHALL be required.

If private state is absent (first `terraform apply` after provider upgrade before a `Read` has populated it), the modifier SHALL default to requiring replace. Users can avoid this by running `terraform refresh` before `terraform apply` after upgrading.

The deprecated `sort_field` and `sort_order` plan modifiers SHALL suppress replace for those attributes when `sort` is non-null in the plan (the new attribute's plan modifier owns the replace decision in that case).

#### Scenario: Migrating from legacy to new sort attribute does not replace the index

- **GIVEN** an existing index managed with `sort_field = ["date"]` and `sort_order = ["desc"]`
- **AND** the resource has been read at least once (private state is populated)
- **WHEN** the configuration is changed to `sort = [{ field = "date", order = "desc" }]`
- **THEN** Terraform SHALL NOT plan a destroy+recreate
- **AND** Terraform SHALL plan an in-place update (or no-change if no other attributes differ)

#### Scenario: Explicit default `missing`/`mode` values during migration do not require replace

- **GIVEN** an existing index managed with `sort_field = ["date"]` and `sort_order = ["desc"]`
- **AND** the existing index has no explicit `index.sort.missing` or `index.sort.mode` settings (Elasticsearch defaults apply)
- **WHEN** the configuration is changed to `sort = [{ field = "date", order = "desc", missing = "_last", mode = "max" }]`
- **THEN** Terraform SHALL NOT plan a destroy+recreate

#### Scenario: Non-equivalent `missing` or `mode` during migration requires replace

- **GIVEN** an existing index managed with `sort_field = ["date"]` and `sort_order = ["desc"]`
- **WHEN** the configuration is changed to `sort = [{ field = "date", order = "desc", missing = "_first" }]`
- **THEN** Terraform SHALL plan to destroy and recreate the resource

#### Scenario: First apply after upgrade forces replace when private state is absent

- **GIVEN** an existing index managed with `sort_field`/`sort_order` and no prior read since the provider was upgraded
- **WHEN** the configuration is changed to use `sort`
- **THEN** Terraform SHALL plan to destroy and recreate the resource (private state is not yet populated)

---

### Requirement: `sort.missing` and `sort.mode` in static settings and adoption (REQ-SORT-04)

The `sort.missing` and `sort.mode` Elasticsearch settings keys SHALL be added to `staticSettingsKeys`. The `use_existing` index adoption flow SHALL compare these settings against the existing index's static settings when they are explicitly set in the plan.

When the plan originates from the new `sort` `ListNestedAttribute`, the `compareStaticPlanAndES` function in `use_existing.go` SHALL compare `sort.field`, `sort.order`, `sort.missing`, and `sort.mode` as ordered string slices (using `stringSliceOrderedFromAny`). This preserves the order-significant semantics of the nested `sort` list defined in REQ-SORT-01.

When the plan originates from the deprecated `sort_field`/`sort_order` attributes, the `compareStaticPlanAndES` function in `use_existing.go` SHALL preserve the existing legacy behavior for `sort.field`: compare `sort.field` as an unordered set, while continuing to compare ordered per-position settings only where the plan shape preserves positional meaning.

#### Scenario: Adoption fails when nested `sort.field` order differs

- **GIVEN** `use_existing = true` and an existing index with `index.sort.field = ["date", "id"]`
- **AND** the plan specifies `sort = [{ field = "id" }, { field = "date" }]`
- **WHEN** the resource is created
- **THEN** the adoption flow SHALL return an error diagnostic naming the mismatched `sort.field` setting
- **AND** SHALL NOT call any mutating API

#### Scenario: Adoption preserves legacy unordered comparison for `sort_field`

- **GIVEN** `use_existing = true` and an existing index with `index.sort.field = ["date", "id"]`
- **AND** the plan specifies `sort_field = ["id", "date"]`
- **WHEN** the resource is created
- **THEN** the adoption flow SHALL allow adoption without a `sort.field` mismatch based only on element order

#### Scenario: Adoption fails when `sort.missing` differs

- **GIVEN** `use_existing = true` and an existing index with `index.sort.missing = ["_last"]`
- **AND** the plan specifies `sort = [{ field = "date", missing = "_first" }]`
- **WHEN** the resource is created
- **THEN** the adoption flow SHALL return an error diagnostic naming the mismatched `sort.missing` setting
- **AND** SHALL NOT call any mutating API

---

### Requirement: Schema — sort attribute and deprecated sort_field/sort_order (REQ-SORT-05)

The `sort_field` and `sort_order` attributes SHALL remain in the schema as optional attributes but SHALL carry a `DeprecationMessage` directing users to the new `sort` attribute. The schema SHALL additionally expose the `sort` attribute as a `ListNestedAttribute` with nested `field` (required string), `order` (optional string), `missing` (optional string), and `mode` (optional string) attributes. The `sort` attribute and `sort_field`/`sort_order` SHALL be mutually exclusive; the schema SHALL enforce this with `ConflictsWith` validators.

```hcl
resource "elasticstack_elasticsearch_index" "example" {
  # Static settings (force new on change)
  sort_field = <optional, set(string), DEPRECATED — use sort>  # force new
  sort_order = <optional, list(string), DEPRECATED — use sort> # force new

  # Replaces sort_field and sort_order
  sort = <optional, list(object)> {   # force new (with migration suppression — see REQ-SORT-03)
    field   = <required, string>
    order   = <optional, string>  # allowed: "asc", "desc"
    missing = <optional, string>  # allowed: "_last", "_first"
    mode    = <optional, string>  # allowed: "min", "max"
  }
}
```

#### Scenario: Deprecated attributes emit deprecation warnings

- **GIVEN** a Terraform configuration that uses `sort_field` or `sort_order`
- **WHEN** Terraform validates or applies the configuration
- **THEN** a deprecation warning SHALL be surfaced for each deprecated attribute

---

## MODIFIED Requirements

### Requirement: Lifecycle — static settings require replacement (REQ-009)

In addition to the existing static settings listed in REQ-009, entries in the new `sort` `ListNestedAttribute` SHALL also trigger resource replacement when changed, subject to the migration suppression defined in REQ-SORT-03. Specifically, changing any of `sort[*].field`, `sort[*].order`, `sort[*].missing`, or `sort[*].mode` SHALL require replacement. The deprecated `sort_field` attribute SHALL continue to require replace when changed, except when `sort` is simultaneously being introduced in the plan (REQ-SORT-03 governs in that case).

#### Scenario: Changing an existing `sort` entry's `order` requires replace

- **GIVEN** an existing index managed with `sort = [{ field = "date", order = "asc" }]`
- **WHEN** the configuration changes to `sort = [{ field = "date", order = "desc" }]`
- **THEN** Terraform SHALL plan to destroy and recreate the resource

---

### Requirement: Opt-in adoption of existing indices via `use_existing`

The set of static settings compared during `use_existing` adoption SHALL be extended to include `sort.missing` and `sort.mode`. When these settings are explicitly set in the plan, the adoption flow SHALL compare them against the existing index's static settings and SHALL return an error diagnostic when they differ, consistent with the behavior for `sort.field` and `sort.order`.

#### Scenario: Adoption compares `sort.missing` against existing index

- **GIVEN** `use_existing = true` and an existing index where `index.sort.missing` is `["_last"]`
- **AND** the plan specifies `sort = [{ field = "date", missing = "_first" }]`
- **WHEN** create runs
- **THEN** the adoption flow SHALL return an error diagnostic naming the mismatched `sort.missing` value
- **AND** SHALL NOT call any mutating API on the index
