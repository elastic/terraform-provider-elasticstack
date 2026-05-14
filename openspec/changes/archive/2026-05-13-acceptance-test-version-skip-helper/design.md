## Context

The codebase currently uses three coexisting patterns for version-based acceptance-test skipping:

1. **Per-step `SkipFunc`** (284 test cases): `SkipFunc: versionutils.CheckIfVersionIsUnsupported(v)` repeated identically on every `resource.TestStep`.
2. **Manual top-level skip** (4 test cases): `skipFunc := versionutils.CheckIfVersionIsUnsupported(v); if skip, err := skipFunc(); ...; t.Skip(...)` before `resource.Test()`.
3. **Skip inside `PreCheck`** (2 test cases): custom helpers that call `t.Skip()` directly.

`CheckIfVersionIsUnsupported` and `CheckIfVersionMeetsConstraints` return `func() (bool, error)` factories designed for `TestStep.SkipFunc`. When the same skip applies to every step, the factory is invoked N times per test, creating a fresh Elasticsearch client each time and producing verbose `"Skipping step X/Y"` log output.

`TestCase` itself has no `SkipFunc` field — only `TestStep` does — so the framework offers no built-in top-level skip mechanism. A helper in `versionutils` is the cleanest place to fill this gap.

## Goals / Non-Goals

**Goals:**
- Provide a single top-level call site that replaces N identical per-step `SkipFunc` assignments.
- Reduce redundant Elasticsearch version API calls from ~850 to ~284 per full acceptance run.
- Make the skip reason explicit in Go test output (`--- SKIP: TestXxx`) rather than buried in per-step framework logs.
- Combine version and flavor checks into one call so tests that need both do not split intent across two helpers.
- Keep the helper in `versionutils` alongside the existing `CheckIf...` factory functions.

**Non-Goals:**
- Removing or replacing the existing `CheckIfVersionIsUnsupported` / `CheckIfVersionMeetsConstraints` / `CheckIfNotServerless` factory functions. Per-step progressive gating and custom skip logic still need them.
- Changing any provider resource schema, API client, or Terraform behavior.
- Eliminating non-version `SkipFunc` uses (e.g. `skipIfUsingFakeAPIKey`).
- Touching the `TestCase` struct in the upstream Terraform plugin SDK.

## Decisions

### Two explicit helpers instead of one type-switched helper

**Decision:** Provide `SkipIfUnsupported(t, *version.Version, Flavor)` and `SkipIfUnsupportedConstraints(t, version.Constraints, Flavor)`.

**Rationale:**
- `*version.Version` for a minimum version is the ergonomically dominant call site (250 uses vs 74 constraint-range uses). A single `any`-typed helper using type switches would sacrifice compile-time safety for the 91% case.
- Go has no function overloading. Two explicit functions is the idiomatic way to serve two distinct parameter shapes.
- Both functions delegate to the same unexported internal logic.

### Serverless bypasses version checks

**Decision:** When the connected cluster reports `build_flavor == "serverless"`, version constraints are treated as satisfied (skip = false).

**Rationale:**
- The existing `ElasticsearchScopedClient.EnforceMinVersion` already implements this behavior: serverless always passes version checks because features are either present or not, not gated by a linear version number.
- Test authors who write `SkipIfUnsupported(t, v8_11_0, FlavorAny)` mean "this feature needs 8.11.0 on stateful; on serverless, if the feature exists it's supported."

### `Flavor` enum with three values

**Decision:** Define `const Flavor int` with `FlavorAny = iota`, `FlavorStateful`, `FlavorServerless`.

**Rationale:**
- `FlavorAny` covers the 95% of tests that don't care about deployment mode.
- `FlavorStateful` replaces the standalone `CheckIfNotServerless()` pattern.
- `FlavorServerless` has zero callers today but completes the logical matrix and prevents future one-off helpers.

### Error on version/flavor check failure → `t.Fatal`

**Decision:** Infrastructure errors (cannot create ES client, cannot parse version) call `t.Fatal`.

**Rationale:**
- A failure to determine the server version or flavor means the test cannot make an informed skip decision. Continuing to `resource.Test()` would compound the error.
- This matches the existing manual pattern currently in `security_enable_rule/acc_test.go`.

### Retain existing factory functions unchanged

**Decision:** `CheckIfVersionIsUnsupported`, `CheckIfVersionMeetsConstraints`, and `CheckIfNotServerless` remain as-is.

**Rationale:**
- 36 test cases use per-step `SkipFunc` with non-identical or non-version logic. Those continue to work.
- Any external consumers of this public API are unaffected.
- The new helpers build on top of the same `clients.NewAcceptanceTestingElasticsearchScopedClient()` call already used by the factories.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| **Mechanical migration is large** (~284 call sites, ~57 files) | Split into per-package batches; each package is an independent task. Existing tests continue to pass with or without migration. |
| **Some test authors may prefer per-step SkipFunc for explicitness** | The helper is additive, not mandatory. Per-step SkipFunc still works. Documentation and code review nudge toward the helper when appropriate. |
| **Skip message no longer names the offending step** | With top-level `t.Skip()`, the whole test is skipped as one unit — this is the desired behavior for uniform skips. Per-step skips remain available for progressive gating. |
| **Flavour strings may change in future ES/Kibana versions** | The helper delegates to `ServerFlavor()` which already normalises `"serverless"` vs `"default"` (stateful). If new flavours appear, only the enum and a single switch case change. |

## Migration Plan

1. Add helpers and `Flavor` enum to `internal/versionutils/testutils.go`.
2. Migrate one representative package (e.g. `internal/kibana/dashboard/`) to validate the pattern.
3. Roll out to remaining packages in batches: `internal/elasticsearch/`, `internal/fleet/`, `internal/kibana/`, `provider/`.
4. At each step, run `make check-lint` and targeted acceptance tests to confirm no regressions.

No rollback strategy is needed — the change is purely additive to test code. Removing the helpers would simply revert the affected tests to using per-step `SkipFunc`.

## Open Questions

- Should the helper accept `*testing.B` for benchmark skip parity? (No current benchmark uses version skipping; defer until needed.)
