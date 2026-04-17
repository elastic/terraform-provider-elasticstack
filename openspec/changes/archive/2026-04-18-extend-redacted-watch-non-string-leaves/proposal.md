## Why

The redacted-leaf preservation added in PR #2296 only substitutes the `::es_redacted::` sentinel when the prior value at the same path is a non-redacted **string**. Watch authors who set webhook headers — such as `Authorization` — to a stored-script reference object (`{"id": "<script-id>"}`) or an inline-script object (`{"source": "...", "lang": "painless"}`) still see perpetual Terraform drift, because Elasticsearch returns the sentinel as a string and the prior value at that path is a `map[string]any`.

## What Changes

- Broaden `mergePreserveRedactedLeaves` so the redacted-string sentinel is replaced by the prior value at the same path whenever the prior is non-nil and is not itself the sentinel — regardless of its JSON type (string, object, array, number, bool).
- Update the existing prior-type-mismatch unit test to assert the new substitution behavior, and add tests covering script-reference, inline-script, array, and round-trip JSON cases mirroring the user-reported watcher payload.
- Add a Plugin Framework acceptance test that uses an inline-script `Authorization` header and asserts no drift after read-after-write, with the script object preserved in state.
- Refresh the `elasticsearch-watch` requirements so refresh / read-after-write behavior covers prior values of any JSON type, not just strings.

## Capabilities

### New Capabilities

### Modified Capabilities

- `elasticsearch-watch`: extend the redacted-action-leaf preservation rules so prior known values of any JSON type — not just strings — are preserved when the API returns the redacted string sentinel.

## Impact

- `internal/elasticsearch/watcher/watch/actions_merge.go`
- `internal/elasticsearch/watcher/watch/actions_merge_test.go`
- `internal/elasticsearch/watcher/watch/acc_test.go`
- `internal/elasticsearch/watcher/watch/testdata/TestAccResourceWatch_redactedScriptHeaderPreserved/`
- `openspec/specs/elasticsearch-watch/spec.md`
