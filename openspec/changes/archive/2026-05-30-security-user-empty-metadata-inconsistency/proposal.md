## Why

When a user sets `metadata = jsonencode(each.value.metadata)` and the variable defaults to an empty
map (`{}`), `jsonencode({})` evaluates to the string `"{}"`. Starting with provider v0.16.0 (Plugin
Framework migration with entitycore read-after-write), the provider oscillates between `"{}"` and
`null` every apply, producing a **"Provider produced inconsistent result after apply"** error:

```
Error: Provider produced inconsistent result after apply
  .metadata: was cty.StringVal("{}"), but now null.
```

**Root cause**: `readUser` in `internal/elasticsearch/security/user/read.go` (lines 62–71)
unconditionally sets `state.Metadata = jsontypes.NewNormalizedNull()` when the Elasticsearch API
returns an empty metadata map. But the write path (`writeUser` in `update.go`) treats the plan
value `"{}"` as non-null and sends `metadata: {}` to the API. After the write, the entitycore
envelope calls `readUser` for read-after-write, which sets metadata to `null`. The plan carried
`"{}"` (non-null), but the post-write state carries `null` — a consistency error.

This regression was introduced by the SDK → Plugin Framework migration: the SDK did not perform a
mandatory read-after-write by default, so the `readUser`/`writeUser` asymmetry was never surfaced
before v0.16.0.

## What Changes

- **Read-side fix** (`internal/elasticsearch/security/user/read.go`): When the Elasticsearch API
  returns empty metadata, treat `null` and `"{}"` as equivalent. If the incoming state already
  holds an empty JSON object, preserve it rather than overwriting with `null`. Add a small helper
  `isEmptyJSONObject(v jsontypes.Normalized) bool` (~8 lines). Drift detection is preserved: if
  the incoming state holds non-empty metadata and the API returns empty, override to `null` as
  before.

- **Acceptance test** (`internal/elasticsearch/security/user/acc_test.go`): Add a test step (or
  new test function) that creates a user with `metadata = jsonencode({})` and asserts the apply
  completes without a consistency error, and that `metadata` in state equals `"{}"`.

- **Delta spec** (`openspec/changes/security-user-empty-metadata-inconsistency/specs/elasticsearch-security-user/spec.md`):
  Amend REQ-016/REQ-017 to document that `null` and `"{}"` are semantically equivalent at the
  Elasticsearch API level, and that the read path must preserve an empty-object state value when
  the API returns empty metadata.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- **`elasticsearch-security-user`**: Read path treats `null` and `"{}"` as equivalent for the
  metadata attribute, eliminating the "Provider produced inconsistent result after apply" error
  when `metadata = jsonencode({})` is used. Acceptance test coverage for the empty-metadata case.

## Impact

- **Users**: `metadata = jsonencode({})` no longer causes a perpetual provider inconsistency error.
  Users do not need to guard the metadata field or use a workaround with an older provider version.
- **Code**: `internal/elasticsearch/security/user/read.go` — ~10-line change plus a small helper.
  No schema changes, no plan modifier, no state upgrade needed.
- **Non-goals**: Changing the metadata attribute type; fixing the data-source/resource metadata
  asymmetry; adding a plan modifier to normalize `"{}"` to `null` at plan time (Approach B was
  evaluated and rejected — see design.md).
