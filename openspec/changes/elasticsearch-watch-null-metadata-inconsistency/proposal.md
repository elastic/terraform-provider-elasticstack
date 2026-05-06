## Why

`elasticstack_elasticsearch_watch` produces an "inconsistent result after apply" error when
`metadata = jsonencode(null)` is set in configuration, as reported in issue
[#2706](https://github.com/elastic/terraform-provider-elasticstack/issues/2706). The
error appeared after upgrading from v0.14.3 (SDK-based) to v0.14.5 (Plugin Framework-based):

```
Provider produced inconsistent result after apply
.metadata: was cty.StringVal("null"), but now cty.StringVal("{}")
```

During the SDK→Plugin Framework migration the `fromAPIModel` function was changed to return `"{}"` when
Elasticsearch returns `null` metadata. The SDK implementation faithfully stored `"null"`. When
`metadata = jsonencode(null)` (the literal JSON string `"null"`) is in the Terraform config,
the plan holds `"null"` while the provider returns `"{}"` after apply, causing Terraform to
reject the result as inconsistent.

## What Changes

- Fix `fromAPIModel` in `internal/elasticsearch/watcher/watch/models.go` to return `"null"` when
  the API response has nil metadata, restoring SDK-era behavior.
- Add an acceptance test that creates a watch with `metadata = jsonencode(null)`, verifies that
  create succeeds without an inconsistency error, and confirms the subsequent plan is empty.
- Update the `elasticsearch-watch` delta spec to add a requirement covering nil-metadata
  round-trip behavior.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `elasticsearch-watch`: fix null-metadata round-trip so `metadata = jsonencode(null)` no longer
  produces an inconsistency error.

## Impact

- Specs: delta spec under
  `openspec/changes/elasticsearch-watch-null-metadata-inconsistency/specs/elasticsearch-watch/spec.md`.
- Provider behavior: one-line fix in `internal/elasticsearch/watcher/watch/models.go`.
- Acceptance tests: new test case in `internal/elasticsearch/watcher/watch/acc_test.go` and a
  matching Terraform config in the `testdata/` directory for the watch resource.
