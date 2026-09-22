## Why

`TestAccSpacesDataSource_withImageURL` (`internal/kibana/spaces/data_source_test.go`) failed against Elastic Stack `9.6.0-SNAPSHOT` on [PR #4890](https://github.com/elastic/terraform-provider-elasticstack/pull/4890) (CI run [34672008631](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/34672008631), shard 1):

```
data_source_test.go:213: Step 1/1 error: Check failed: Check 1/2 error:
    data.elasticstack_kibana_spaces.all_spaces.spaces.0.id:
    expected default, got acc_test_space
```

This is a test-assertion bug, not a product bug. The `elasticstack_kibana_spaces` data source has never guaranteed a particular ordering of the returned `spaces` list — nothing in `openspec/specs/kibana-spaces/spec.md` documents an ordering contract, and the same test file already has an order-independent helper, `testCheckSpaceAttrByID`, used for the custom space's `image_url` check. On Kibana 9.6.0-SNAPSHOT the Get All Spaces API (or the provider's flatten) returns the built-in `default` space at a different list position than on earlier versions, so every hard-coded `spaces.0.id == "default"` assertion is fragile and version-dependent.

The same `spaces.0.id == "default"` pattern (or an equivalent index-0 assumption) appears in every test in the file: `TestAccSpacesDataSource`, `TestAccSpacesDataSource_multipleSpaces` (whose code comment additionally claims "Default space is always the first element."), `TestAccSpacesDataSource_noDescription`, `TestAccSpacesDataSource_withImageURL`, and `TestAccSpacesDataSource_withKibanaConnection`. Only `TestAccSpacesDataSource_withImageURL` has been observed failing so far (snapshot jobs are `continue-on-error`, so the run stayed green), but every other occurrence shares the identical, order-dependent assumption and can fail the same way on any Kibana version — snapshot or GA — that changes list ordering.

## What Changes

- Replace every hard-coded `spaces.0.*` (or other fixed-index) assertion against the built-in `default` space in `internal/kibana/spaces/data_source_test.go` with an id-based lookup, reusing the existing `testCheckSpaceAttrByID` helper (or an equivalent order-independent check) already used for the custom space's attributes.
- Remove or correct the stale `// Default space is always the first element.` comment in `TestAccSpacesDataSource_multipleSpaces` once its assertion no longer depends on list position.
- No production code changes: the `elasticstack_kibana_spaces` data source's read/mapping behavior is unaffected. This is a test-only fix.
- Document, in the `kibana-spaces` capability spec, that the data source's `spaces` list has no ordering contract and that acceptance tests for this data source MUST assert space attributes by `id`, not by list index, so this class of regression cannot silently reappear.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-spaces`: add a requirement documenting that the `spaces` list attribute has no guaranteed ordering and that acceptance tests verifying entries in that list MUST look up entries by `id` rather than by fixed index.

## Impact

- `internal/kibana/spaces/data_source_test.go` — update the five affected tests (`TestAccSpacesDataSource`, `TestAccSpacesDataSource_multipleSpaces`, `TestAccSpacesDataSource_noDescription`, `TestAccSpacesDataSource_withImageURL`, `TestAccSpacesDataSource_withKibanaConnection`) to assert the `default` space's attributes by id instead of by `spaces.0`.
- `openspec/specs/kibana-spaces/spec.md` — add the no-ordering-contract / id-based-assertion requirement.
- No changes to `internal/kibana/spaces` non-test source files, generated clients, or documentation.
