## Why

`indices_options.expand_wildcards` in `elasticstack_elasticsearch_ml_datafeed` is declared as a `list(string)`. Elasticsearch normalizes the shorthand token `"all"` into its constituent values `["open", "closed", "hidden"]` when it returns a datafeed configuration. Because the provider performs exact list equality, a user who writes `expand_wildcards = ["all"]` sees a perpetual plan diff after the first apply: Terraform compares the user's `["all"]` against the API's `["open", "closed", "hidden"]` and marks the attribute as changed. The resource is then updated unnecessarily on every subsequent plan.

## What Changes

- Change `indices_options.expand_wildcards` from `schema.ListAttribute` to `schema.SetAttribute` so element ordering is irrelevant.
- Add a datafeed-local custom set type (`ExpandWildcardsType` / `ExpandWildcardsValue`) that implements `basetypes.SetValuableWithSemanticEquals`. The semantic equality rule: `{"all"}` is considered equal to `{"open", "closed", "hidden"}`; all other comparisons are unordered set equality.
- Update `IndicesOptions.ExpandWildcards` in models.go from `types.List` to the custom set value and adjust the object attribute-type map accordingly.
- Update acceptance test assertions from index-based list checks (`expand_wildcards.0`) to set-element checks (`TestCheckTypeSetElemAttr`).
- Add unit tests for the custom type: `all == {open, closed, hidden}`, order insensitivity, partial-expansion inequality, `none`, null, and unknown cases.
- Update the `openspec/specs/elasticsearch-ml-datafeed/spec.md` delta to document `set(string)` and the semantic equality plan behavior.

## Capabilities

### Modified Capabilities
- `elasticsearch-ml-datafeed`: `indices_options.expand_wildcards` changes from `list(string)` to `set(string)` with semantic equality for the `all` shorthand.

### New Capabilities
- None (custom type is internal to the datafeed package).

## Impact

- New Go file: `internal/elasticsearch/ml/datafeed/expand_wildcards_type.go` (custom type and unit tests in `expand_wildcards_type_test.go`).
- Modified Go files: `schema.go`, `models.go` in `internal/elasticsearch/ml/datafeed/`.
- Modified acceptance test: `internal/elasticsearch/ml/datafeed/acc_test.go`.
- Modified spec: `openspec/specs/elasticsearch-ml-datafeed/spec.md` (via delta spec in this change).
- No Terraform schema breaking change for users whose state was previously stored as a list: JSON array shape is preserved; a state upgrader is unlikely to be needed but should be verified during implementation.
- No provider config changes, no other resources affected.

## Non-Goals

- Normalizing `"all"` away before sending to the API (user token is preserved in the write path).
- Adding a state upgrader unless testing confirms existing list-shaped state cannot be decoded.
- Changing any other datafeed attributes.
