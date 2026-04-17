## Context

PR #2296 added [`mergeActionsPreservingRedactedLeaves`](../../../internal/elasticsearch/watcher/watch/actions_merge.go) so that when the Elasticsearch Get Watch API returns the sentinel string `"::es_redacted::"` at a nested action path, the provider replaces the sentinel with the prior known value from Terraform state (on refresh) or plan (on read-after-write). This eliminated drift for the most common case: webhook basic-auth `password` fields, where the prior at the redacted path is a plain string.

Field-team feedback (see the conversation around `Authorization: { id = "service-now-key" }` and the user's reproducer watcher with an inline-script `Authorization` header) shows a second class of redaction: when an action header is set to a **stored-script reference object** (`{"id": "<script-id>"}`) or **inline-script object** (`{"source": "...", "lang": "painless"}`), Elasticsearch still replaces the entire header value with the same sentinel **string** in the Get Watch response. The current substitution branch is gated on the prior also being a `string`, so it falls through:

```
if s, ok := apiVal.(string); ok {
    if s == elasticsearchWatcherRedactedSecret {
        if ps, ok := priorVal.(string); ok && ps != elasticsearchWatcherRedactedSecret {
            return ps
        }
    }
    return apiVal
}
```

The redacted sentinel is then written into state, and the next plan keeps showing drift against the configured object.

## Goals / Non-Goals

**Goals:**

- Eliminate drift for redacted action leaves whose prior value is a non-string JSON node (object, array, number, bool), in addition to the existing string case.
- Keep the merge driven by the API shape so non-redacted sub-trees stay authoritative from the API response.
- Preserve the existing safety net: when there is no prior value or the prior is itself the sentinel, the API value (the sentinel) is stored as-is.
- Add unit and acceptance coverage that mirror the user-reported watcher (script-reference / inline-script `Authorization` header).
- Update the `elasticsearch-watch` requirements to reflect the broader rule.

**Non-Goals:**

- Changing the `read.go` flow or how prior actions are passed in. The plumbing already covers refresh and read-after-write.
- Changing redaction handling for top-level fields other than `actions` (`trigger`, `input`, `condition`, `metadata`, `transform`).
- Adding stored-script-based acceptance coverage that requires creating a Painless stored script via the ES client during `PreCheck` — using an inline-script object is sufficient to exercise the non-string-prior path end-to-end.
- Archiving the change. Per [`.agents/skills/openspec-implementation-loop/SKILL.md`](../../../.agents/skills/openspec-implementation-loop/SKILL.md), implementation only syncs delta specs; the `verify-openspec` workflow archives later.

## Decisions

### Decision: Substitute the prior value at the redacted path regardless of its JSON type

**Rationale**: The prior comes from validated Terraform config or state, not from arbitrary user input at read time. If the user's config legitimately has `headers.Authorization = {"id": "service-now-key"}` at a path that the API now reports as `"::es_redacted::"`, echoing the prior object back into state is exactly what eliminates drift without ever inventing a value. The same is true for inline scripts (object), array headers, or any future shape Elasticsearch chooses to redact at a leaf.

**Alternatives considered**:

- *Restrict substitution to `string` and `map[string]any` only* — covers the immediate report but adds a second special-case (arrays, numbers, bools) that would need follow-up. Worse, it bakes shape assumptions into the merge that don't match how Elasticsearch may evolve the redaction surface.
- *Detect the redaction path by name (`Authorization`, `password`, …)* — fragile and behavior-coupled to specific actions; provides no benefit over the structural rule.

### Decision: Encapsulate the "no usable prior" check in a small helper

Introduce `isRedactedOrAbsent(priorVal any) bool` to cover the two cases where the API sentinel is preserved as-is: `priorVal == nil` and `priorVal` is itself the sentinel string. This keeps `mergePreserveRedactedLeaves` readable and gives the unit tests a single, named pre-condition to reason about.

### Decision: Keep the recursive walk driven by the API shape

The deep walk over `map[string]any` and `[]any` is unchanged. Only leaves where `apiVal` is the sentinel string change behavior. This guarantees that non-redacted fields the API returns (e.g. `host`, `port`, `body`) remain authoritative from the API even when their prior values were equal but stale.

## Risks / Trade-offs

- **[Risk]** A future API quirk could return the sentinel string at a path where the prior in state is a stale object that no longer matches reality. → **Mitigation**: This is the same risk as the original PR #2296. Terraform's prior is the authoritative source for "what the user asked for"; if the user changes their config (e.g. changes `id` from `service-now-key` to a new script), the next plan will diff against the new config and propose a normal update through the Put Watch API. The merge only suppresses the false sentinel→sentinel diff, not real changes.
- **[Risk]** A prior value containing the sentinel string deep inside an object would now be replaced with itself. → **Mitigation**: That is intentional and matches existing behavior for the simple string case; recursing into the prior to scrub nested sentinels is out of scope and not observed in practice.
- **[Trade-off]** The `TestMergeActionsPreservingRedactedLeaves_priorTypeMismatch` test from PR #2296 asserted that a `map` prior at a redacted leaf was *not* substituted in. That assertion is reversed by this change. We rename the test to `_priorObjectReplacesRedacted` so the broader rule is the documented behavior going forward.
