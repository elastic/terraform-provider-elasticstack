## Context

The Kibana 9.5 dashboard API changed: POST/PUT without a `description` now returns
`description: ""` rather than omitting the field. The provider's current read path calls
`typeutils.StringishPointerValue(data.Data.Description)`, which maps `""` to `types.StringValue("")`.
When the practitioner omitted `description` (planned as `null`), Terraform sees a `null → ""`
diff and raises "Provider produced inconsistent result after apply".

Affected code path:
- `internal/kibana/dashboard/models.go`, line 52:
  `m.Description = typeutils.StringishPointerValue(data.Data.Description)`

Relevant helpers already in `internal/utils/typeutils/string.go`:
- `StringishPointerValue` — maps nil → null, ptr → string (does NOT strip `""`)
- `NonEmptyStringishPointerValue` — maps nil/`""` → null (blanket; ignores user intent)
- `TrimmedStringishPointerValue` — maps nil/whitespace-only → null (blanket; ignores user intent)

The `NonEmptyStringishPointerValue` helper almost solves the problem but silently drops an
explicit `description = ""` set by the practitioner. Intent preservation requires a plan-aware check.

## Goals / Non-Goals

**Goals:**
- Fix the `null → ""` inconsistency when `description` is omitted (planned as null) and Kibana 9.5 returns `""`.
- Preserve `description = ""` when the practitioner explicitly set it.
- Preserve existing behavior for non-empty `description` values.
- Fix all ~14 failing acceptance tests on 9.5 without changing their configs.

**Non-Goals:**
- Normalizing `description` on other Kibana resources (scope is dashboard root only).
- Schema changes or migration.
- Changing write-path behavior (empty-string descriptions already handled correctly).

## Decisions

**Normalization strategy**: Intent-preserving plan-aware check in `models.go`. The `FromAPI`
function already receives the prior model `m` as the receiver. Before calling `FromAPI`, the
caller sets `m` from prior state/plan. The check is:

```go
// If Kibana returns "" for description and prior intent was null, preserve null.
apiDesc := typeutils.StringishPointerValue(data.Data.Description)
if apiDesc.ValueString() == "" && m.Description.IsNull() {
    m.Description = types.StringNull()
} else {
    m.Description = apiDesc
}
```

This is the minimal, localized change. It does NOT touch other fields or resources.

**Why not `NonEmptyStringishPointerValue`?** It would silently coerce an explicit `description = ""`
(a legitimate practitioner choice) to null. The intent-preserving check keeps `""` when the user
set it.

**Why not a plan modifier?** A plan modifier on the schema attribute would work but adds schema
boilerplate. The read-path fix in `models.go` is simpler, consistent with how other optional
fields in the same function are handled (e.g., `time_range.mode`), and does not require a new
schema.PlanModifier.

**Acceptance test coverage**: No new test configs are needed — the existing ~14 tests that omit
`description` should pass after the fix. One targeted test should be added to assert that an
explicit `description = ""` round-trips correctly on 9.5+.

## Risks / Trade-offs

- [Low risk] If a practitioner currently has `description = ""` in their config, the fix preserves
  it correctly. No state migration needed.
- [Low risk] The fix is localized to one line in `models.go`; no cascading side effects.
- [Low risk] The fix applies uniformly to 8.x and 9.x: on 8.x, Kibana returned nil/omitted for
  `description` → `StringishPointerValue` returned null anyway. On 9.x, Kibana returns `""`;
  the new check restores null when intent was null.

## Open questions

None — root cause, fix strategy, and scope are all clear from the issue.
