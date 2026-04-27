## Context

The `elasticstack_elasticsearch_index_template` resource's `template` block currently supports `alias`, `mappings`, `settings`, and `lifecycle`. The Elasticsearch 9.x index template API added `data_stream_options` inside the `template` object, which controls how a new data stream behaves when it matches the template, including whether its failure store is enabled.

The `data_stream_options` object in the API has the following shape:

```json
{
  "data_stream_options": {
    "failure_store": {
      "enabled": true,
      "lifecycle": {
        "data_retention": "30d"
      }
    }
  }
}
```

Key facts from the API documentation:
- `data_stream_options` is applied to a data stream only once, at creation time. Changes to the template do not propagate to existing data streams; use the `PUT /_data_stream/{name}/_options` API for that.
- The `failure_store.enabled` boolean activates or deactivates document redirection to the failure store.
- The optional `failure_store.lifecycle.data_retention` string sets how long failure store documents are retained (e.g. `"10d"`, `"30d"`). Defaults to 30 days cluster-wide if unset.
- The feature is available from Elasticsearch 9.1 (preview) and 9.2 (GA).

## Goals / Non-Goals

**Goals:**
- Expose `data_stream_options.failure_store.enabled` and `data_stream_options.failure_store.lifecycle.data_retention` in the `template` block.
- Version-gate the new attributes to Elasticsearch >= 9.1.0.
- Extend the internal `models.Template` struct to carry the new field.
- Update the expand and flatten helpers to handle the new field end-to-end.
- Add acceptance test coverage.

**Non-Goals:**
- Manage existing data stream options via the `PUT /_data_stream/{name}/_options` API (that is a separate resource concern, not an index template concern).
- Surface any other `data_stream_options` fields that are not yet documented or confirmed in the API (this change covers the full documented schema).
- Apply `data_stream_options` to the `elasticstack_elasticsearch_component_template` resource (it is not part of the component template API shape).

## Decisions

### 1. Block shape: nested `data_stream_options > failure_store > lifecycle`

The HCL schema mirrors the API structure with typed blocks rather than a raw JSON string, consistent with how the existing `lifecycle` block and `data_stream` block are handled. This gives practitioners plan-time validation and avoids arbitrary JSON manipulation.

```hcl
template {
  data_stream_options {
    failure_store {
      enabled          = true
      lifecycle {
        data_retention = "30d"
      }
    }
  }
}
```

Using nested blocks instead of a JSON string attribute is the preferred pattern in this provider for structured objects that have a small, stable, well-defined shape.

### 2. Version gating at >= 9.1.0

The provider already gates `ignore_missing_component_templates` to >= 8.7.0. The same pattern will be applied: check `serverVersion` in the create/update path and return an error diagnostic if the version is below the minimum. The check is applied only when `data_stream_options` is configured. Version 9.1.0 is used as the minimum rather than 9.2.0 to allow preview users to benefit from the feature.

### 3. Shared `models.Template` struct extension

The shared `models.Template` struct (used by both index templates and component templates) will gain a `DataStreamOptions *DataStreamOptions` field tagged `json:"data_stream_options,omitempty"`. Because of `omitempty`, component templates that never populate this field continue to serialize cleanly without the key appearing in API payloads.

A new `DataStreamOptions` struct and nested `FailureStoreOptions`/`FailureStoreLifecycle` types will be added to `models.go`:

```go
type DataStreamOptions struct {
    FailureStore *FailureStoreOptions `json:"failure_store,omitempty"`
}

type FailureStoreOptions struct {
    Enabled   bool                    `json:"enabled"`
    Lifecycle *FailureStoreLifecycle  `json:"lifecycle,omitempty"`
}

type FailureStoreLifecycle struct {
    DataRetention string `json:"data_retention,omitempty"`
}
```

### 4. `failure_store.enabled` is required within the block

When a practitioner declares `failure_store`, they must configure `enabled`. Omitting it would default to the Go zero value (`false`), which would silently send `enabled: false`. Requiring the attribute makes intent explicit. The Terraform schema will use `Required: true` for `enabled`.

### 5. Read path: preserve `data_stream_options` from API response

On read, `flattenTemplateData` will check whether `template.DataStreamOptions` is non-nil and, if so, build the nested HCL block representation in state. This mirrors how `lifecycle` is flattened.

### 6. No plan diff suppression needed

Unlike `mappings` and `settings` (raw JSON with arbitrary key ordering), `data_stream_options` is a small structured object with deterministic field names. No custom diff suppression is needed.

## Risks / Trade-offs

- **Shared struct side-effect**: Adding `DataStreamOptions` to `models.Template` means the field will appear in deserialized GET responses for index templates that contain `data_stream_options`. Component template GET responses will not include this key, so deserialization is safe.
- **One-time application semantics**: Practitioners who update `data_stream_options` in their Terraform configuration after their data stream already exists will see the plan show a change, and the put-template API will be called successfully — but the existing data stream will not be affected. This is an API-level constraint, not something the provider can or should work around. It will be documented clearly in the resource and attribute descriptions.
- **Preview version support**: Accepting 9.1.0 as the minimum means the provider may need to handle any minor differences between the preview and GA API shapes, though none are currently known.

## Migration Plan

1. Add Go model types to `internal/models/models.go`.
2. Extend the `template` schema in `internal/elasticsearch/index/template.go` with the `data_stream_options` block and version gate.
3. Update `expandTemplate` to populate `DataStreamOptions` from Terraform config.
4. Update `flattenTemplateData` to populate Terraform state from `DataStreamOptions` in API response.
5. Add acceptance test fixtures and test cases covering:
   - Create with `data_stream_options.failure_store.enabled = true`.
   - Update to change `enabled` value.
   - Create with `lifecycle.data_retention` set.
   - Read-back correctness.
6. Regenerate `docs/resources/elasticsearch_index_template.md`.

## Open Questions

- Whether the `failure_store.lifecycle.data_retention` field accepts the same duration string format as `template.lifecycle.data_retention` (e.g., `"30d"`, `"1h"`). Based on doc examples (`"10d"`, `"30d"`), this appears consistent. No format validation beyond non-empty string is planned unless further API contract evidence appears during implementation.
