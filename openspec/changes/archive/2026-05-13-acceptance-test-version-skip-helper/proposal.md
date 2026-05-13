## Why

284 of 320 acceptance test cases that use `SkipFunc` define the identical version-based skip on every single step. This forces ~850 redundant Elasticsearch version API calls per full test run, clutters step definitions with repetitive boilerplate, and fragments the codebase across three coexisting skip patterns (per-step `SkipFunc`, manual top-level `t.Skip()`, and `PreCheck` helpers). A single top-level test helper eliminates all three problems.

## What Changes

- Add `versionutils.SkipIfUnsupported(t, minVersion, flavor)` and `versionutils.SkipIfUnsupportedConstraints(t, constraints, flavor)` helpers that perform a single version/flavor check and call `t.Skip()` or `t.Fatal()` before `resource.Test()` is entered.
- Introduce a `Flavor` enum (`Any`, `Stateful`, `Serverless`) so server-flavor gating is expressed at the same call site as version gating, replacing the standalone `CheckIfNotServerless()` pattern.
- Retain existing `CheckIfVersionIsUnsupported` and `CheckIfVersionMeetsConstraints` factory functions for per-step `SkipFunc` use (progressive gating, partial skipping, and custom non-version skips still need them).
- Migrate all 284 test cases with identical version-based `SkipFunc` on every step to the new top-level helper.

## Capabilities

### New Capabilities
- `acceptance-test-version-skip`: Top-level test helper for skipping acceptance tests based on Elasticsearch server version and deployment flavor. Provides `SkipIfUnsupported` (min-version) and `SkipIfUnsupportedConstraints` (constraint-range) variants with a `Flavor` enum.

### Modified Capabilities
- *(none — no Terraform resource or provider behavior changes)*

## Impact

- `internal/versionutils/testutils.go` — new exported helpers and `Flavor` type.
- ~284 acceptance test files across `internal/elasticsearch/`, `internal/fleet/`, `internal/kibana/`, and `provider/` — remove per-step `SkipFunc` boilerplate, add single top-level skip call.
- No provider schema, API client, or resource logic changes.
