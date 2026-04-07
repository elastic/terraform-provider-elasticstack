# Delta Spec: Require `searchable_snapshot` in the ILM `frozen` phase

Base spec: `openspec/specs/elasticsearch-index-lifecycle/spec.md`
Last requirement in base spec: REQ-032
This delta introduces: REQ-033

---

## Schema refinements

The `frozen` phase continues to support only the `searchable_snapshot` action, but unlike the `hot` and `cold` phases, the `searchable_snapshot` block is required when `frozen` is declared.

```hcl
frozen {
  min_age = <optional + computed, string>

  searchable_snapshot {              # required when frozen is declared
    snapshot_repository = <optional, string>          # required when block is present
    force_merge_index   = <optional + computed, bool> # default true
  }
}
```

---

## MODIFIED Requirements

### Requirement: Validation and connection selection (REQ-008–REQ-010)

The existing REQ-008–REQ-010 text:

> The resource SHALL reject configuration that omits all five phase blocks `hot`, `warm`, `cold`, `frozen`, and `delete`. The resource SHALL accept `metadata` and allocation filters only when they are valid JSON objects. By default, the resource SHALL use the provider-level Elasticsearch client; when `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for create, read, update, and delete.

is extended with:

> When the user declares the `frozen` phase, the configuration SHALL include a `searchable_snapshot` block inside `frozen`; omission SHALL be rejected during Terraform validation before any lifecycle API call.

#### Scenario: Frozen phase without searchable snapshot is rejected (ADDED)

- GIVEN a resource configuration with `frozen { min_age = "30d" }` and no `searchable_snapshot`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error before any Elasticsearch ILM API call

---

## ADDED Requirements

### Requirement: Frozen phase requires searchable snapshot (REQ-033)

When the `frozen` phase is configured, the resource SHALL require the `frozen.searchable_snapshot` nested block in the Terraform schema rather than treating it as optional.

Within that required block, `snapshot_repository` SHALL remain required when the `searchable_snapshot` block is present, consistent with REQ-032.

The generated Terraform documentation for the resource SHALL reflect this schema shape by describing `frozen.searchable_snapshot` as required within the `frozen` phase.

#### Scenario: Valid frozen phase includes searchable snapshot

- GIVEN a resource configuration with:
  - `frozen.min_age = "30d"`
  - `frozen.searchable_snapshot.snapshot_repository = "repo-a"`
- WHEN Terraform plans or applies the resource
- THEN the provider SHALL accept the `frozen` phase schema shape
- AND the lifecycle policy expansion SHALL include the `searchable_snapshot` action for the `frozen` phase

#### Scenario: Required nested field within frozen searchable snapshot

- GIVEN a resource configuration with `frozen.searchable_snapshot { force_merge_index = false }`
- WHEN Terraform validates the configuration
- THEN validation SHALL fail because `snapshot_repository` is required when the `searchable_snapshot` block is present

#### Scenario: Generated docs match frozen schema requirement

- GIVEN the provider documentation is generated from the resource schema
- WHEN the `elasticstack_elasticsearch_index_lifecycle` docs are refreshed
- THEN the `frozen` section SHALL describe `searchable_snapshot` as required within `frozen`
