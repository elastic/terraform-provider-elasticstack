## Context

Canonical requirements for this resource live in
[`openspec/specs/elasticsearch-security-user/spec.md`](../../specs/elasticsearch-security-user/spec.md).
The read path lives in
[`internal/elasticsearch/security/user/read.go`](../../../internal/elasticsearch/security/user/read.go).
The write path lives in
[`internal/elasticsearch/security/user/update.go`](../../../internal/elasticsearch/security/user/update.go).
The entitycore read-after-write trigger is at
[`internal/entitycore/resource_envelope.go:413`](../../../internal/entitycore/resource_envelope.go).

### Root cause trace

1. **Plan phase**: config has `metadata = "{}"` (non-null JSON string) — plan carries this value.
2. **Create/Update**: `writeUser` (`update.go:89-106`) unmarshals `"{}"` into an empty map and
   sends `metadata: {}` to the Elasticsearch PUT user API. Returns `plan` unchanged, so
   `written.Model.Metadata = "{}"` (non-null).
3. **Read-after-write** (entitycore envelope at `resource_envelope.go:413`): calls `readFunc`
   immediately after the write, passing `written.Model` as the initial state.
4. **`readUser`** (`read.go:62-71`): checks `len(user.Metadata) > 0`. The Elasticsearch API
   returns an empty map for a user with no/empty metadata — `len == 0` — so the code
   unconditionally sets `state.Metadata = jsontypes.NewNormalizedNull()`.
5. **Framework consistency check**: plan carried `"{}"` (known non-null), new state carries
   `null` → "Provider produced inconsistent result after apply".

The root asymmetry: write treats `"{}"` as a valid non-null value to send, but read treats an
empty API response as `null`. They are semantically equivalent at the Elasticsearch API level
(no metadata set ≈ empty metadata object), but the provider exposes them as different Terraform
values.

### Fix strategy

In `read.go`, replace the unconditional `state.Metadata = jsontypes.NewNormalizedNull()` (the
`else` branch of `if len(user.Metadata) > 0`) with a conditional check:

```go
} else {
    // API returned no metadata. Treat "{}" and null as equivalent:
    // preserve state if it holds an empty JSON object so that
    // `metadata = jsonencode({})` round-trips correctly after write.
    if !isEmptyJSONObject(state.Metadata) {
        state.Metadata = jsontypes.NewNormalizedNull()
    }
}

func isEmptyJSONObject(v jsontypes.Normalized) bool {
    if v.IsNull() || v.IsUnknown() {
        return false
    }
    var m map[string]any
    return json.Unmarshal([]byte(v.ValueString()), &m) == nil && m != nil && len(m) == 0
}
```

**How it handles each case**:

| Incoming state (plan/state) | API response | New state (fixed) | Correct? |
|---|---|---|---|
| `null` | empty | `null` | ✓ unchanged |
| `"{}"` | empty | `"{}"` preserved | ✓ **fixes bug** |
| `{"k":"v"}` | empty (server drift) | `null` | ✓ drift detected |
| any | non-empty map | JSON-serialized map | ✓ unchanged path |

**Required change scope**: `internal/elasticsearch/security/user/read.go` only — ~10 lines plus
a small helper function. No schema changes, no plan modifier, no state upgrade needed.

## Goals / Non-Goals

**Goals:**

- Eliminate the "Provider produced inconsistent result after apply" error when `metadata = jsonencode({})`.
- Preserve drift detection: state `{"k":"v"}` + API empty → override to null.
- Add acceptance test coverage for the empty-metadata case.
- Update the delta spec to document the amended metadata read-side invariant.

**Non-goals:**

- Adding a plan modifier to normalize `"{}"` → `null` at plan time (Approach B — see Decisions).
- Changing the `metadata` attribute type from `jsontypes.Normalized` to `types.Map` or
  `types.Object` — a larger refactor with state migration implications.
- Fixing the data-source/resource metadata asymmetry as part of this change.
- Addressing the caller's module design (always encoding metadata even when absent).
- Fixing the same pattern for other `Optional+Computed` JSON metadata attributes (e.g.,
  `elasticstack_elasticsearch_security_role_mapping`) — out of scope; tracked separately.

## Decisions

- **Approach A (targeted read-side fix) is adopted.** The fix is localized to a single file,
  semantically correct (Elasticsearch treats no-metadata and empty-metadata identically at the
  API level), and does not alter what users see in `terraform plan`. The drift-detection path
  (non-empty → empty) is preserved.

- **Approach B (plan modifier) was evaluated and rejected.** A plan modifier would resolve the
  inconsistency at plan time, but it silently merges "user set empty object" with "user did not
  set metadata", changes user-visible plan output (showing `null` when the config says `"{}"`),
  and does not fix the conceptual read/write asymmetry. See research comment on issue #3437
  for full comparison.

## Open questions

- Does the Elasticsearch PUT user API always include `"metadata": {}` in the GET response when an
  empty object was explicitly PUT, vs. omitting the field when metadata was never set? A three-way
  distinction (omitted / `{}` / non-empty) would require testing against an 8.x cluster; the fix
  in Approach A handles both the omitted and empty-object cases correctly regardless.

- Should `readUserDataSource` in `internal/elasticsearch/security/user/data_source.go` also be
  updated to return `null` for empty metadata instead of always marshalling to `"{}"`? Currently
  the data source and resource are asymmetric on this boundary.

- Are other `Optional+Computed` JSON metadata attributes in the codebase (e.g.,
  `elasticstack_elasticsearch_security_role_mapping`) exposed to the same issue once they are
  migrated to the entitycore read-after-write envelope?

## Risks / Trade-offs

- **Read behavior becomes state-dependent at the null/empty boundary**: A mild departure from
  "read entirely fresh from API." Acceptable because the two values represent identical server
  state. The non-empty-metadata drift-detection path is fully preserved.
- **No schema version bump or state migration**: Existing state with `null` metadata continues to
  work correctly; existing state with `"{}"` metadata (produced if the bug was worked around) also
  continues to work correctly.
