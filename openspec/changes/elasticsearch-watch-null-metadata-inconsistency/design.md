## Context

The `elasticstack_elasticsearch_watch` resource was migrated from the Terraform Plugin SDK to the
Plugin Framework in v0.14.5. During the migration, `fromAPIModel` in
`internal/elasticsearch/watcher/watch/models.go` was written to normalize a nil API metadata field
to the JSON string `"{}"`:

```go
if watch.Body.Metadata == nil {
    d.Metadata = jsontypes.NewNormalizedValue(`{}`)
}
```

The old SDK implementation stored `"null"` when the API returned nil metadata. When a practitioner
sets `metadata = jsonencode(null)` in their Terraform configuration — producing the plan value
`"null"` — Elasticsearch receives and stores null metadata. On the next read the API returns nil
metadata, and the PF `fromAPIModel` translates it to `"{}"`. Terraform detects the drift between
the planned `"null"` and the read-back `"{}"` and aborts with the "inconsistent result after apply"
error.

Cases that are NOT affected:
- `metadata` omitted from config: the schema default `"{}"` is used, ES receives `{}` and returns a
  non-nil empty map, so the nil branch is never reached.
- `metadata = null` as HCL null: the schema default `"{}"` is applied at plan time; same flow as
  omitted.
- `metadata = jsonencode({})` (empty object string): ES returns a non-nil empty map; nil branch not
  reached.

Only `metadata = jsonencode(null)` (sends JSON `null` to ES, ES returns nil) is broken.

## Goals / Non-Goals

**Goals:**

- Restore round-trip consistency for `metadata = jsonencode(null)`: after apply and read-back the
  provider stores `"null"`, matching the plan.
- Add a regression acceptance test that reproduces the exact reported scenario.
- Update the `elasticsearch-watch` requirements spec to cover the nil-metadata case explicitly.

**Non-Goals:**

- Changing the schema default for `metadata` (remains `"{}"`).
- Changing behavior for omitted, HCL-null, or empty-object metadata configurations.
- Migrating state from existing resources in the field (no state upgrade is required).

## Decisions

| Topic | Decision | Alternatives considered |
|-------|----------|------------------------|
| Fix location | Change the nil branch in `fromAPIModel` to return `"null"` instead of `"{}"` | Returning `"{}"` and treating `"null"` as equivalent during plan were rejected because they change schema semantics and could affect users who expect `jsonencode(null)` to round-trip cleanly. |
| Scope of nil-metadata behavior | Only `metadata` is affected; `input`, `condition`, and `actions` have their own explicit defaults that match SDK behavior and are not impacted. | N/A |
| Acceptance test approach | Add a new test step using `metadata = jsonencode(null)`, assert no inconsistency error on create, and assert empty plan on the second step. | Unit-level test alone was considered insufficient because the inconsistency manifests as a Terraform framework error during the apply cycle. |

## Risks / Trade-offs

- [Risk] Practitioners who previously relied on `metadata = jsonencode(null)` mapping to `"{}"` in
  state would see a change in stored value. Mitigation: the inconsistency error means no such
  practitioner can currently apply successfully; this is a net fix.
- [Risk] Changing the nil branch could affect import behavior for watches that have nil metadata
  in Elasticsearch. Mitigation: after this fix, import will correctly store `"null"` for such
  watches, matching what `metadata = jsonencode(null)` would produce, which is the correct and
  self-consistent result.

## Migration Plan

- No practitioner migration is needed. The fix restores behavior that was working in v0.14.3.
- No state schema version bump is required; the `metadata` attribute type is unchanged.
- Rollback is a normal code revert with no data migration.

## Open Questions

- None blocking implementation.
