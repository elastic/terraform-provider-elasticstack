## Context

`elasticstack_elasticsearch_ml_datafeed` exposes `indices_options.expand_wildcards` as a `schema.ListAttribute`. The Elasticsearch ML Datafeed API accepts the shorthand token `"all"` (meaning open + closed + hidden indices) but always returns the expanded form `["open", "closed", "hidden"]`. The provider performs exact list equality during plan, so `["all"]` vs `["open", "closed", "hidden"]` never matches. This causes spurious perpetual plan diffs and unnecessary API updates.

The repo already has a precedent for this pattern: `internal/kibana/slo/group_by_type.go` implements `GroupByType` / `GroupByValue` as a custom list type with `ListSemanticEquals`. We follow the same pattern, using the set variant because `expand_wildcards` is an unordered collection.

## Goals / Non-Goals

**Goals:**

- Fix the perpetual plan diff caused by `"all"` normalization.
- Model `expand_wildcards` as a set so element order does not matter.
- Keep the change scoped to `internal/elasticsearch/ml/datafeed` and its spec.
- Preserve the user's written token (`"all"`) in write requests — no normalization on send.

**Non-Goals:**

- State upgrader (add only if acceptance testing confirms it is necessary — the raw JSON array shape is the same whether stored as list or set).
- Changing any other datafeed attribute or resource.
- Adding a Kibana-side analogue or a shared utility across resources.

## Decisions

### D1. `set(string)` rather than `list(string)` with a custom list type

**Choice:** Change the schema attribute to `schema.SetAttribute` and implement `ExpandWildcardsType` / `ExpandWildcardsValue` wrapping `basetypes.SetType` / `basetypes.SetValue`.

**Rationale:** The `expand_wildcards` values have no meaningful order. Using a set removes the ordering concern at the schema level, which is the correct semantic. A custom list type with semantic equals would suppress the plan diff but still expose ordering to users and state files, adding confusion.

**Alternatives considered:**

- *Custom list type with `ListSemanticEquals` only.* Would fix the diff but keep misleading ordering semantics. Rejected.
- *Plain `schema.SetAttribute` without semantic equals.* Would fix ordering but not the `"all"` vs `["open","closed","hidden"]` equivalence issue. Rejected.

### D2. Semantic equality rule

**Choice:** `ExpandWildcardsValue.SetSemanticEquals` returns true when:
  1. Both values are null ↔ null or unknown ↔ unknown (identity).
  2. The normalized token set of the new value equals the normalized token set of the prior value, where normalization expands `"all"` to `{"open", "closed", "hidden"}`.

Normalization is applied to both sides independently before comparison. `"none"` is not expanded and compares literally.

**Rationale:** Elasticsearch's documented expansion of `"all"` is stable and well-known. Bidirectional normalization means `{"all"}` == `{"open","closed","hidden"}` == `{"closed","open","hidden"}` (order-insensitive via set). `"none"` has no documented expansion so it stays literal. Null and unknown are handled conservatively.

**Alternatives considered:**

- *Normalize only the API-returned side.* Fragile — requires knowing which side is "plan" vs "state" inside the hook.
- *Normalize away `"all"` on write so both sides always hold expanded values.* Would strip intent from the stored config and break round-trip fidelity for the `"all"` token.

### D3. Custom type is internal to the datafeed package

**Choice:** Define `ExpandWildcardsType` and `ExpandWildcardsValue` in `internal/elasticsearch/ml/datafeed/expand_wildcards_type.go`. No export outside the package.

**Rationale:** No other resource currently has this normalization need. Inlining keeps the change small and avoids premature abstraction in a shared utils package.

**Alternatives considered:**

- *`internal/utils/customtypes/expand_wildcards_type.go`.* Premature generalization.

### D4. State upgrader deferred

**Choice:** Do not add a state upgrader in this PR. Verify during acceptance testing whether existing list-shaped state can be decoded by the new set-typed schema.

**Rationale:** The Terraform Plugin Framework deserializes both lists and sets from JSON arrays. The raw state representation is `["open","closed"]` regardless of whether the schema type is list or set; the framework should reconstruct the set value without a state version bump. If testing proves otherwise, add a `StateUpgraders` entry with schema version 1.

**Alternatives considered:**

- *Add a state upgrader preemptively.* Unnecessary complexity if the framework handles it transparently.

## Risks / Trade-offs

- **Risk:** State upgrade requirement missed during implementation. **Mitigation:** Task checklist includes a dedicated acceptance test step with existing state to confirm no state migration error.
- **Risk:** `"none"` expansion ambiguity (does Elasticsearch normalize it?). **Mitigation:** Treat `"none"` as literal and not subject to expansion; if the API normalizes it, the literal comparison is still safe (no false positives).
- **Trade-off:** Users who previously wrote `expand_wildcards = ["open", "closed"]` as an ordered list will now see it treated as a set in configs and diffs. This is semantically correct but is a minor UX change.
