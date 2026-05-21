## Why

The `elasticstack_elasticsearch_index_mappings` resource allows `mappings = jsonencode({})`, which defeats the resource's purpose: the resource exists to manage a declared subset of index mappings, and an empty declaration causes the read-after-write flow to store the full API mappings into state (intersection logic treats `{}` as "no prior mask"). This creates a confusing user experience where a seemingly empty declaration silently captures everything.

## What Changes

- Parameterize the `StringIsJSONObject` validator with an optional `NonEmpty` boolean field. Zero-value `NonEmpty: false` preserves existing behavior across all current call sites.
- Add the `StringIsJSONObject{NonEmpty: true}` validator to the `mappings` attribute on the `elasticstack_elasticsearch_index_mappings` resource schema.
- Add unit tests for the validator covering empty `{}` rejection, non-object rejection, and the existing pass-through behavior.
- Add an acceptance test step with `ExpectError` asserting `mappings = jsonencode({})` is rejected at plan time.

## Capabilities

### New Capabilities
- *(none)*

### Modified Capabilities
- `elasticstack-elasticsearch-index-mappings`: Schema requirement REQ-001 is updated to require that `mappings` is a **non-empty** JSON object. The requirement wording and the "mappings is required" scenario are updated; a new scenario covers the non-empty validation.

## Impact

- **Code**: `internal/elasticsearch/index/validation.go`, `internal/elasticsearch/index/validation_test.go` (new), `internal/elasticsearch/index/indexmappings/schema.go`, `internal/elasticsearch/index/indexmappings/acc_test.go` (new test step).
- **No breaking changes** to existing resources: all other `StringIsJSONObject{}` call sites use the zero-value which preserves current behavior.
- **No API or dependency changes**.
