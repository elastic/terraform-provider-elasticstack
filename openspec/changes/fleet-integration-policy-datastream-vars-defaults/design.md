## Context

Fleet 9.5 introduced two enrichment behaviors (via Kibana PRs #214216, #258143, #119739) that cause post-apply inconsistency errors for input-type packages in `elasticstack_fleet_integration_policy`. The provider's existing semantic-equality machinery partially addresses server-injected defaults (`applyDefaultsToVars`) but has two incomplete gaps:

1. `data_stream.type` is injected by Fleet into stream `vars` but is absent from `defaults.vars`, so `applyDefaultsToVars` does not back-fill it. The downstream `StringSemanticEquals` call in `compareStreams` is key-order-insensitive but not key-presence-tolerant, so the plan-side `vars` (missing `data_stream.type`) does not equal the API-side `vars` (containing it), resulting in `InputValue.ObjectSemanticEquals` returning false and Terraform reporting an inconsistent result.

2. The `defaults` attribute is a plain `types.Object` with no custom semantic-equality type. When Fleet 9.5 returns a populated `defaults` block that was `null` in the plan, the raw `null` ⇄ object transition fails the inconsistency check at the `defaults` leaf level. The `InputValue.ObjectSemanticEquals` logic only absorbs this transition if `ObjectSemanticEquals` itself returns true — but since gap #1 is also present, it returns false before the `defaults` leaf is even compared.

The issue notes that fixing gap #1 alone would allow `ObjectSemanticEquals` to return true, which would absorb gap #2 (since `defaults` inequality is a leaf-level comparison swallowed by the parent's semantic equality). However, the design adopted here fixes both gaps independently for robustness against future Fleet changes.

## Goals / Non-Goals

**Goals:**

- Eliminate `"Provider produced inconsistent result after apply"` errors for input-type packages on Elastic Stack 9.5+ for both identified gaps.
- Maintain backward compatibility with Elastic Stack 9.4 and earlier (fixes are no-ops when `data_stream.*` keys are absent and `defaults` remains null).
- Pass `TestAccResourceIntegrationPolicyGCPPubSub` and `TestAccResourceIntegrationPolicySecrets` (including `multi-valued_secrets`) against 9.5.0-SNAPSHOT without changing the test configurations.
- Unit test coverage for both new behaviors.

**Non-Goals:**

- Surfacing `data_stream.type` / `data_stream.dataset` as user-configurable attributes (they are server-managed; stripping at comparison is the right level).
- Changing the read-back behavior to strip `data_stream.*` from state (they should remain in state as returned by Fleet; only the comparison step normalizes them).
- Fixing acceptance tests by updating test fixture configurations to include the injected keys (brittle and doesn't help real users — issue option E2, rejected).
- Applying changes to any other Fleet resource or the `vars_json` top-level attribute.

## Decisions

### Decision 1: Strip `data_stream.*` keys at the comparison step, not at the read-back step

The `data_stream.type` and `data_stream.dataset` keys injected by Fleet into stream `vars` are server-managed: they reflect Fleet's internal enrichment and should not be treated as user configuration. The provider should strip them from the API-returned `vars` before comparison but leave them in state as returned (so Terraform shows the true API state to the user).

**Alternative (E3 from the issue):** Strip from the read-back path (never write them to state). This is simpler but discards information Terraform would otherwise surface to users and may cause re-diffs if Fleet changes its injection behavior.

**Why Decision 1:** Stripping at the comparison step is consistent with how the existing `applyDefaultsToVars` logic normalizes the two sides before comparing. It makes the semantic equality logic the canonical place for server-managed-key handling, which is easier to extend in the future.

### Decision 2: Fix the `data_stream.*` stripping in `compareStreams`, not in `applyDefaultsToVars`

`applyDefaultsToVars` fills missing keys from `defaults.vars` but does not strip extra keys from the API side. Adding a stripping step there would conflate two different operations (filling defaults vs. normalizing server-managed keys). Instead, add a helper `stripServerManagedVarsKeys(vars jsontypes.Normalized) jsontypes.Normalized` and call it in `compareStreams` on both sides before the `StringSemanticEquals` call.

**Server-managed key list:** `["data_stream.type", "data_stream.dataset"]`. These are the only keys documented (in the issue and Kibana PRs) as server-injected by Fleet 9.5 for input-type packages. If Fleet injects additional keys in future, this list can be extended.

### Decision 3: Fix the `defaults` null ⇄ object transition via `InputValue.ObjectSemanticEquals`

The cleanest fix is to make `InputValue.ObjectSemanticEquals` treat a `null` planned `defaults` as semantically equal to any API-returned `defaults` value. This is logically correct: the user did not configure `defaults` (it is purely computed); Fleet populates it from package info. Any API-returned value is valid and should not trigger a diff when the planned value was `null`.

**How:** In `ObjectSemanticEquals`, after extracting `oldInput` and `newInput`, compare their `Defaults` fields with a helper that treats `null` ⇄ any-value as semantically equal. The comparison is: if `oldInput.Defaults` is null, return true for the defaults component (proceed to compare the rest). Symmetrically, if `newInput.Defaults` is null, return true for that component.

**Alternative (plan modifier / `UseStateForUnknown`):** Attach a plan modifier to the `defaults` schema attribute that marks the value as unknown during planning when it is null in config. This lets Terraform treat the post-apply API value as an expected unknown resolution. More invasive (requires schema change and possibly a state upgrader if the null-in-state cases proliferate). Not required given that `InputValue.ObjectSemanticEquals` is the primary equality gate.

### Decision 4: No schema version bump

Both fixes are within the semantic-equality and plan-comparison layer. The state format for `defaults` (already `Computed: true`, populated from package info) does not change. No state upgrader is required.

### Decision 5: Unit tests are required for both fixes

The existing test files (`input_value_test.go`, `vars_json_value_test.go`, `models_defaults_test.go`) SHALL be extended with targeted test cases for:

- `compareStreams` with `data_stream.*` keys on one side only → semantically equal.
- `InputValue.ObjectSemanticEquals` with a null `defaults` on one side and a populated `defaults` on the other → semantically equal.

## Open questions

1. **Should `data_stream.dataset` also be stripped in `compareStreams`?** The issue states that `data_stream.dataset` *does* appear in `defaults.vars` (so `applyDefaultsToVars` already back-fills it on the plan side). However, if the plan-side `vars` does not include it (e.g. the user explicitly omits it), the existing back-fill may not fire correctly if `defaults` itself is null at comparison time. The safe approach is to include `data_stream.dataset` in the strip list alongside `data_stream.type`. Implementation should verify empirically whether stripping `data_stream.dataset` is redundant or necessary.

2. **Are there other server-managed keys beyond `data_stream.type` and `data_stream.dataset`?** The issue documents these two. If additional Kibana PRs inject further keys, the strip list will need maintenance. Consider making the list a named constant or a package-level variable for easier extension.

3. **`TestAccResourceIntegrationPolicySecrets` / `multi-valued_secrets` sub-test** — does the `defaults` fix also resolve the secrets-related inconsistency, or does the secrets test fail for a different reason? The issue groups both tests together, but the root cause for each should be confirmed during implementation.
