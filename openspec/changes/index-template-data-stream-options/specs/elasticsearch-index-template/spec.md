# `elasticstack_elasticsearch_index_template` — `data_stream_options` Delta Spec

Capability: `elasticsearch-index-template`
Base spec: `openspec/specs/elasticsearch-index-template/spec.md`

This delta spec documents the requirements added by the `index-template-data-stream-options` change. All base-spec requirements remain in force.

## Schema addition

The `template` block gains a new optional nested block:

```hcl
resource "elasticstack_elasticsearch_index_template" "example" {
  name           = "my-datastream-template"
  index_patterns = ["my-datastream-*"]

  data_stream {}

  template {
    data_stream_options {
      failure_store {
        enabled = true
      }
    }
  }
}
```

Full schema delta for `template`:

```hcl
template {
  # ... existing attributes unchanged ...

  data_stream_options {                    # optional, list, max 1
    failure_store {                        # optional, list, max 1
      enabled = <required, bool>
    }
  }
}
```

## ADDED Requirements

### Requirement: `data_stream_options` schema (REQ-032)

The `template` block SHALL expose an optional `data_stream_options` block (`TypeList`, `MaxItems: 1`). Inside it, an optional `failure_store` block (`TypeList`, `MaxItems: 1`) SHALL contain a required boolean `enabled` attribute. Omitting `data_stream_options` entirely SHALL leave the field absent from the API payload (`omitempty`).

#### Scenario: Schema presence — omitted

- GIVEN `data_stream_options` is omitted from configuration
- WHEN Terraform plans
- THEN no `data_stream_options` key SHALL appear in the API request body

### Requirement: Version gate for `data_stream_options` (REQ-033)

When `data_stream_options` is configured (non-empty), the resource SHALL check the Elasticsearch server version. If the version is below 9.0.0, the resource SHALL return an error diagnostic and SHALL NOT call the Put index template API.

#### Scenario: Version gate blocks write

- GIVEN `data_stream_options` is set and the Elasticsearch server reports version < 9.0.0
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic referencing the minimum required version (9.0.0)
- AND the Put index template API SHALL NOT be called

### Requirement: Create/update with `data_stream_options` (REQ-034)

When `data_stream_options` is configured and the version gate passes, the resource SHALL include `data_stream_options` in the API request body with the exact shape `{"failure_store": {"enabled": <bool>}}`. `enabled = false` SHALL be serialized as `false` (not omitted), so that an explicit opt-out is transmitted to Elasticsearch.

#### Scenario: enabled true serialized

- GIVEN `data_stream_options.failure_store.enabled = true`
- WHEN create or update runs
- THEN the request body SHALL contain `"data_stream_options": {"failure_store": {"enabled": true}}`

#### Scenario: enabled false serialized

- GIVEN `data_stream_options.failure_store.enabled = false`
- WHEN create or update runs
- THEN the request body SHALL contain `"data_stream_options": {"failure_store": {"enabled": false}}`

### Requirement: Read/state refresh for `data_stream_options` (REQ-035)

On read, when the API response `template.data_stream_options` is present and non-null, the resource SHALL populate the `data_stream_options` block in Terraform state. When the field is absent from the API response, the block SHALL be absent from state (no phantom empty block).

#### Scenario: API returns data_stream_options

- GIVEN the stored template has `data_stream_options.failure_store.enabled = true`
- WHEN read refreshes state
- THEN `template[0].data_stream_options[0].failure_store[0].enabled` SHALL be `true` in state

#### Scenario: API does not return data_stream_options

- GIVEN the stored template has no `data_stream_options`
- WHEN read refreshes state
- THEN `data_stream_options` SHALL be absent (empty list) in state

### Requirement: Removal of `data_stream_options` (REQ-036)

When `data_stream_options` is removed from configuration (previously set, now absent), the resource SHALL send a full-replace update to Elasticsearch with `data_stream_options` absent from the request body. The Elasticsearch Put API replaces the template definition entirely, so omitting the field removes it.

#### Scenario: Remove block from config

- GIVEN `data_stream_options` was previously set
- WHEN the user removes the block and update runs
- THEN the Put API SHALL be called with no `data_stream_options` in the request body

### Requirement: Data source read symmetry (REQ-037)

The `elasticstack_elasticsearch_index_template` data source SHALL expose the same `data_stream_options` block structure in its `template` attribute (read-only). On read, when the API response includes `data_stream_options`, the data source SHALL populate it in state using the same flatten logic as the resource.

#### Scenario: Data source surfaces data_stream_options

- GIVEN an index template with `data_stream_options.failure_store.enabled = true` exists in Elasticsearch
- WHEN the data source reads the template
- THEN `template[0].data_stream_options[0].failure_store[0].enabled` SHALL be `true` in data source attributes
