## Context

`elasticstack_elasticsearch_watch` resources fail `terraform apply` when the watch uses an HTTP
`input` with basic authentication and the password is supplied via a sensitive Terraform variable.
The failure is `.input: inconsistent values for sensitive attribute` — a Terraform Plugin Framework
error produced when the plan's concrete password differs from the `::es_redacted::` sentinel that
Elasticsearch returns in the Get Watch response.

The identical bug was fixed for `actions` in PR #2296 (via `mergeActionsPreservingRedactedLeaves`)
and extended in a follow-up for non-string prior values. The `input` field was explicitly deferred
as out of scope at the time; the same fix pattern applies directly.

Trace logs from the issue confirm that `SemanticEquals` is invoked on the `input` attribute with
`::es_redacted::` during both `ReadResource` and `ApplyResourceChange` RPCs, meaning the framework
has already detected the mismatch before any lifecycle rule could suppress it.

## Goals / Non-Goals

**Goals:**

- Preserve prior concrete `input` values in Terraform state when Watcher read responses redact
  nested secret leaves (e.g. `input.http.request.auth.basic.password`).
- Keep refresh authoritative for non-secret `input` fields so real drift outside redacted leaves
  still surfaces in state.
- Scope the fix to the watch resource without introducing a provider-wide JSON comparison type.
- Add focused regression coverage for the redacted-input apply and refresh paths.
- Update the canonical `elasticsearch-watch` spec to make `input` redaction preservation an
  explicit traced requirement.

**Non-Goals:**

- Introduce a new generic custom Terraform type for redaction-aware JSON equality.
- Change the `input` schema shape or split auth credentials into separate Terraform attributes.
- Recover original `input` secret values during import or any first read where Terraform has no
  prior concrete `input` value.
- Broaden the change to `trigger`, `condition`, `metadata`, or `transform` — none carry user-
  supplied secrets that Elasticsearch redacts.

## Decisions

### 1. Extend `fromAPIModel` to accept and apply `priorInput`

In `internal/elasticsearch/watcher/watch/models.go`, the `fromAPIModel` function already has the
`priorActions jsontypes.Normalized` parameter and applies `mergeActionsPreservingRedactedLeaves`
to the `actions` field (lines 163–179). The fix adds a parallel `priorInput jsontypes.Normalized`
parameter and applies `mergePreserveRedactedLeaves` directly (the `input` root is already a JSON
object, not an actions-keyed map, so there is no actions-wrapper shim needed) to the `input`
field.

The call site in `read.go` currently passes `state.Actions` as `priorActions`; it will also pass
`state.Input` as `priorInput`.

Why:

- The root cause is read-time state replacement, so fixing read-time behavior addresses it directly.
- Reusing the already-tested `mergePreserveRedactedLeaves` helper means no new logic is required.
- The change is small (~15–20 lines of production Go) and consistent with existing practice.
- Covers all nested paths under `input` without hard-coding field names.

### 2. Delegate merge to `mergePreserveRedactedLeaves` (not the actions wrapper)

`mergeActionsPreservingRedactedLeaves` is an actions-specific shim that iterates the top-level
action keys before recursing. For `input`, the top-level structure is the input type key (e.g.
`http`, `search`, `chain`) rather than action names, but semantically it is the same recursive
walk. Calling `mergePreserveRedactedLeaves(apiInputMap, priorInputMap)` directly — with both
unmarshalled as `map[string]any` — handles all nested paths correctly without requiring a new
wrapper or modifying the existing `mergeActionsPreservingRedactedLeaves` signature.

Why:

- The recursive helper is already general-purpose (not actions-specific); it handles maps, arrays,
  strings, and all other JSON types.
- No new helper is needed; the only production code change is wiring the parameter and call.

### 3. Keep imports and first reads explicit about their limitation

When Terraform has no prior concrete `input` value, the resource stores the API response as
returned, which may contain `::es_redacted::`. This is the same documented limitation that exists
for `actions`.

Why:

- There is no trustworthy prior secret value to restore on import or the first refresh after state
  loss.
- Making the limitation explicit keeps the design honest and avoids inventing hidden storage.

## Risks / Trade-offs

- **[Risk] Out-of-band changes to `input` secrets will not appear in Terraform state once a prior
  concrete value is preserved** → Mitigation: limit preservation to redacted leaves only.
- **[Risk] `chain` inputs with nested HTTP steps may carry additional redacted paths** → Mitigation:
  `mergePreserveRedactedLeaves` walks the full JSON tree recursively, so nested chain inputs are
  handled without field-name coupling.

## Open Questions

1. Does Elasticsearch also redact `input.http.request.auth.basic.username` as well as `password`?
   The recursive merge handles both automatically; an acceptance test covering both would be useful.
2. Are there `input` sub-types besides `http` (e.g. `chain` inputs with nested HTTP steps) that can
   carry ES-redacted secrets? The recursive merge handles them generically; targeted acceptance
   coverage for `chain` inputs may be worth a follow-up issue.
3. Should the acceptance test scope cover only the post-apply state consistency, or should it also
   verify that a subsequent unrelated update (e.g. changing `condition`) succeeds without resending
   the sentinel?

## Migration Plan

1. Extend `fromAPIModel` signature and body to accept `priorInput` and apply
   `mergePreserveRedactedLeaves` to the `input` field.
2. Update `readWatch` in `read.go` to pass `state.Input` as `priorInput`.
3. Add unit tests for the `input` redaction-preservation path (nested HTTP basic auth password,
   no-prior-value case, non-string prior value case).
4. Add an acceptance test verifying that `terraform apply` succeeds for a watch with HTTP input
   basic auth and a sensitive password variable, and that a subsequent plan produces no diff.
5. Update `openspec/specs/elasticsearch-watch/spec.md` to add REQ-030 and parallel scenarios for
   `input` redaction preservation.
6. Validate and verify the OpenSpec change artifacts.
