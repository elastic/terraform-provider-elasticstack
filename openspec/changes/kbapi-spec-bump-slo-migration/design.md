## Context

`generated/kbapi` is produced by `make -C generated/kbapi all`: download Kibana's published OpenAPI
bundle at a pinned `github_ref` (`generated/kbapi/Makefile`), run it through a fixed list of
`TransformFunc`s (`generated/kbapi/transform_schema.go`'s `transformers` slice), then run
`oapi-codegen`. Several existing transforms already hardcode component paths by name — e.g.
`fixSloFloatFormats` sets `schemas.SLOs_objective.properties.target.format` directly — because the
transform helper type (`Map`) creates missing paths on `Set` rather than failing when a path does not
exist. That means a transform whose target path moved or was renamed upstream does not produce a build
error or a test failure; it just quietly stops having any effect.

Kibana's SLO OpenAPI spec is no longer hand-written; it is now generated from route schemas
(elastic/kibana#300516). That regeneration adds the `sli.synthetics.availability` indicator (not
previously documented) and replaces every `kbapi.SLOs*` Go type the provider depends on today with a
`kbapi.KibanaHTTPAPIsSLO*`-prefixed type. Roughly 330 occurrences of `kbapi.SLOs*` exist across 19 files
under `internal/` (`internal/models/slo.go`, `internal/kibana/slo/*.go`,
`internal/clients/kibanaoapi/slo.go`), so a straight rename touches the whole SLO domain/resource layer.

Separately, the same spec bump introduces a `Cases_attachment_type_filter` query parameter whose inline
`oneOf` (string | string array) `oapi-codegen` cannot name, emitting `AsN0()`/`FromN0()` that reference
undefined `N0`/`N1` types and breaking compilation of `generated/kbapi/kibana.gen.go` outright — this is
unrelated to SLO and blocks the build regardless of the SLO migration's scope.

## Human direction (authoritative for this proposal)

> Regenerating at `cca2b0f` confirms both breakages: the `fixCasesAttachmentTypeFilterParameter`
> transform clears the N0/N1 errors, but the SLO migration is a rewrite rather than a rename
> (TimeWindow and KqlWithFilters change shape, APM params become union wrappers, and
> fixSloFloatFormats silently no-ops on the new schema names, turning targets into float32).

This is a primary-source confirmation — an actual `make -C generated/kbapi all` run against a pinned
commit — of exactly the claims the research comment below flagged as unverified. It resolves the
research's "Grounded: ❌ Fail" gap for the N0/N1 failure and the rename's mechanics, and triggers the
"feasibility gate" the research itself proposed (see Reconciliation below) rather than contradicting
the research's recommended approach.

## Reconciliation with the research comment

The research comment's `### Recommendation` (Approach B, phased: fix the build first, defer
`sli.synthetics.availability` to a follow-up) is adopted unchanged — the human direction does not
dispute the phasing decision.

What the human direction changes is the research's working assumption that Phase 1 is "well-scoped...
rename only, subject to feasibility gate" — see the research's own "Feasibility gate before treating
Phase 1 as rename-only" paragraph, which explicitly anticipated this exact outcome:

> If that inspection finds no breaking difference, the Terraform schema version and state
> representation stay unchanged; if it does, scope those differences as additional Phase 1 tasks before
> merging.

The human direction is the result of running that inspection, and it found breaking differences. This
proposal therefore treats Phase 1 as confirmed to include, in addition to the mechanical parts of the
rename:

- A reshape of the `TimeWindow` and `KqlWithFilters` mapping code (not just a type-name substitution).
- A rewrite of APM latency/availability indicator construction and decoding to go through a `oneOf`
  union wrapper's `From*`/`As*` accessors, matching the pattern `models_kql_custom_indicator.go`
  already uses for `KqlWithFilters`, instead of a plain struct literal.
- A fix to `fixSloFloatFormats` and `fixSloResponseArtifacts` so they target the regenerated spec's
  actual component paths — without this, `objective.target` silently reverts to `float32`, regressing
  the fix shipped for issue #2396, with no compiler or test signal under the transform's current
  `Map.Set`-creates-missing-paths behavior.

This is not treated as a contradiction of the research's recommendation; it is the feasibility gate's
answer, folded into Phase 1's task list (see `tasks.md`) rather than deferred to a second PR, because
the float-precision regression risk is a correctness issue in the same build-unblocking change, not an
independent feature.

## Goals / Non-Goals

**Goals:**
- Unblock the Kibana OpenAPI spec bump (Renovate PR #5464) by fixing both independent build breakages
  at one pinned target commit.
- Migrate all `kbapi.SLOs*` references in the SLO domain/resource/client layers to their regenerated
  `kbapi.KibanaHTTPAPIsSLO*` equivalents, including the confirmed shape changes in `TimeWindow`,
  `KqlWithFilters`, and APM indicator params.
- Preserve the existing `elasticstack_kibana_slo` Terraform schema, state shape, and wire JSON for every
  indicator type, `group_by` variant, and `settings`/`artifacts` field exactly as today — this is a
  compile-time/internal-type migration, not a resource behavior change.
- Re-target `fixSloFloatFormats` and `fixSloResponseArtifacts` at the new component paths so float
  precision and the `artifacts` response field keep working, and add a regeneration-time guard so a
  future rename of these paths fails loudly instead of silently no-oping again.
- Add the `fixCasesAttachmentTypeFilterParameter` transform, reusing `Cases_string`/`Cases_string_array`
  only after confirming their shape (in particular `maxItems: 100`) still matches the target ref's
  inline schema.

**Non-Goals:**
- Adding `sli.synthetics.availability` indicator support. Its minimum Kibana version and exact
  request/response field shape are unverified by any source available to this proposal (see Open
  Questions). Tracked as a separate follow-up issue per the research's Approach B and the issue's own
  stated escape hatch.
- Any other new SLO indicator type beyond the seven covered by the regenerated spec.
- Changing Renovate configuration beyond closing/superseding #5464 once this change lands and CI is
  green.
- Any Terraform-schema-visible change to `elasticstack_kibana_slo` (new attributes, version gates, or
  state-shape changes). Existing version gates (`GetVersionRequirements` in
  `internal/kibana/slo/models.go`, constants in `internal/kibana/slo/constants.go`) are left unchanged.

## Decisions

### 1. Pin one immutable target commit SHA before doing any of the work

Record a single `github_ref` value in `generated/kbapi/Makefile` for this change (replacing
`23436cdb1b7104ea2973e3eee33e3d9c481ce15a`) rather than treating "latest Kibana main" as a moving
target. Every transform and type-rename decision in this proposal is verified against that one pinned
ref's actual downloaded `oas.yaml`, not against "whatever main is" at some other point in time.

Why: the issue names two different candidate refs (`cd914df` from the Renovate PR, `cca2b0f` from the
human direction's confirmation run) and asks generally to bump to "the latest main," which drifts. A
pinned SHA makes the regeneration and its test coverage reproducible.

### 2. Add `fixCasesAttachmentTypeFilterParameter` as a straight port of the existing sibling-filter pattern, after a shape check

`Cases_attachment_type_filter`'s inline `oneOf` (string | string array) matches the shape already
handled for five sibling filter parameters (e.g. `Cases_assignees_filter`) that `$ref` the
`Cases_string`/`Cases_string_array` components. Add the new transform to point
`Cases_attachment_type_filter` at the same two components, but first diff the target ref's actual
inline array schema against `Cases_string_array`'s current definition (item type, `maxItems`/`minItems`,
uniqueness, `style`/`explode`). If the target's inline array has no `maxItems` or a different bound,
mint a dedicated array component instead of reusing `Cases_string_array` as-is, so the fix only repairs
codegen naming and does not silently tighten or loosen the parameter's accepted values.

Why: minimizes new schema surface by reusing an existing, already-used pattern, while the explicit
shape check avoids silently changing the API contract as a side effect of a naming fix.

Alternatives considered:
- Minting a brand-new `oneOf` component just for this parameter regardless of shape: rejected as
  unnecessary duplication when the existing components already cover the same shape, unless the shape
  check finds a real mismatch.

### 3. Add a regeneration-time guard that fails when a schema-name-dependent transform's source node is missing

Add a guard function, run first in `transformers` (or as a pre-transform validation step invoked from
`transform_schema.go`'s `main`), that checks the just-downloaded spec actually contains the nodes that
`fixCasesAttachmentTypeFilterParameter`, `fixSloFloatFormats`, and `fixSloResponseArtifacts` target, and
fails the `make` target (not just a Go unit test) if any are absent.

Why: `Map.Set` creates missing paths rather than erroring, so these transforms can silently no-op
against a renamed or moved schema node with no build failure — exactly what happened to
`fixSloFloatFormats` in the regeneration that produced the human direction's confirmation. A fixture
test alone cannot catch this, because a fixture built to match a transform's hardcoded path keeps
passing even if the real upstream node moved; only a check against the actual downloaded spec can.

Alternatives considered:
- Rely solely on fixture tests in `transform_schema_test.go`: rejected as insufficient on its own,
  since fixtures validate a transform's logic against a fixed hand-built input, not whether that input
  still matches the real spec.

### 4. Migrate SLO code to `kbapi.KibanaHTTPAPIsSLO*` with explicit reshaping for `TimeWindow`, `KqlWithFilters`, and APM params

Rename straightforward `kbapi.SLOs*` → `kbapi.KibanaHTTPAPIsSLO*` references across
`internal/models/slo.go`, `internal/kibana/slo/*.go`, and `internal/clients/kibanaoapi/slo.go` where the
underlying shape is unchanged. For the three confirmed shape changes:

- **`TimeWindow`**: rebuild the conversion in `internal/kibana/slo/models.go` (`toAPIModel`, today a
  direct two-field struct literal at `models.go:145`) against whatever field/variant structure the
  renamed type actually has, verified against the pinned ref's regenerated `kibana.gen.go`.
- **`KqlWithFilters`**: rebuild `kqlTFFormToKqlWithFiltersUnion` /
  `kqlWithFiltersAPIToTFFormFilter` (`internal/kibana/slo/models_kql_custom_indicator.go`) against the
  renamed union's actual `From*`/`As*` accessor set, preserving the existing string-or-object JSON wire
  behavior.
- **APM params**: rewrite `apmLatencyIndicatorToAPI`/`populateFromApmLatencyIndicator`
  (`models_apm_latency_indicator.go`) and the availability equivalent
  (`models_apm_availability_indicator.go`) to construct the params through a union wrapper's `From*`
  method and decode through its `As*`/discriminator accessor, replacing today's direct anonymous-struct
  literal assignment (`models_apm_latency_indicator.go:47-66`).

Every rewrite must produce the same JSON request/response shape as the pre-migration implementation —
verified by the updated unit tests (Decision 6) and the existing acceptance suite — since this
migration is scoped as internal-type-only, not a resource behavior change.

Why: the human direction confirms these three are not mechanical renames; treating them as such would
produce a compile fix that either doesn't build or silently changes wire behavior.

Alternatives considered:
- Wrapping old-shaped helper types around the new kbapi types to minimize the diff in `models.go`:
  rejected because it would hide the real shape change behind an indirection layer, making future kbapi
  bumps harder to reason about, for no behavior benefit.

### 5. Re-target the two schema-name-dependent SLO transforms instead of leaving them silently inert

Update `fixSloFloatFormats` and `fixSloResponseArtifacts` to set `format: double` / the `artifacts` ref
on the regenerated spec's actual component paths (confirmed against the pinned ref's downloaded
`oas.yaml`, not assumed from a prefix-swap heuristic). Add the regeneration-time guard from Decision 3
specifically covering these two transforms' target paths so a future rename cannot silently reintroduce
this regression.

Why: per the human direction, this is exactly what broke in the `cca2b0f` regeneration — the transform
ran without error, but `objective.target` lost its `format: double` override and reverted to the
spec's default `type: number`, which `oapi-codegen` maps to `float32`. That is a direct regression of
issue #2396 if shipped unfixed.

### 6. Update and extend unit tests to lock in the reshaped mapping code

Update the compile-time-broken existing unit tests (`models_api_mapping_test.go`, each indicator's
`*_test.go`, `internal/clients/kibanaoapi/slo_test.go`) to reference the renamed types, and add new
table-driven fixture tests in `generated/kbapi/transform_schema_test.go` for
`TestFixCasesAttachmentTypeFilterParameter`, `TestFixSloFloatFormats`, and `TestFixSloResponseArtifacts`
(the latter two currently untested), following the existing style (e.g. `TestFixControlValuesSourceSchemas`).
Add or extend indicator-mapping unit tests specifically covering the reshaped `TimeWindow`,
`KqlWithFilters`, and APM-union code paths, and a float-precision round-trip test asserting
`objective.target` stays `float64` end to end after the transform fix.

Why: matches the research's "Generator-level tests" plan (isolated fixture tests plus the
regeneration-time drift check from Decision 3) and gives the reshaped mapping code explicit coverage
instead of relying only on the acceptance suite to catch a shape regression.

### 7. Defer `sli.synthetics.availability` to a follow-up issue; no version-gate decision needed in this change

This change does not add a new indicator block to `internal/kibana/slo/schema.go`, so no new
`VersionRequirement` or `constants.go` entry is needed here. The follow-up issue inherits the research's
interim version-gating strategy (gate behind a constant set to the first Stack-version-matrix entry
where a manual probe confirms the indicator works, lowered only as earlier versions are confirmed).

## Risks / Trade-offs

- **A reshaped `TimeWindow`/`KqlWithFilters`/APM mapping could subtly change wire JSON despite "no
  behavior change" intent** → Unit-test every reshaped mapping function against the exact pre-migration
  JSON shape (Decision 6), and keep the existing acceptance suite's per-indicator coverage as an
  end-to-end check against a real Kibana instance.
- **`fixSloFloatFormats`/`fixSloResponseArtifacts` could be re-targeted at paths that still don't match
  the real spec, silently no-oping again** → The regeneration-time guard (Decision 3) fails the `make`
  target itself when the targeted node is absent, rather than relying solely on a fixture test that
  cannot observe the real downloaded spec.
- **The `Cases_attachment_type_filter` fix could silently tighten or loosen the parameter's accepted
  values if `Cases_string_array`'s `maxItems` doesn't match the target's inline array** → Perform the
  explicit shape diff described in Decision 2 before reusing the existing component.
- **Pinning one commit SHA means this change does not track Kibana `main` going forward** → Expected;
  the provider's existing Renovate-driven bump workflow (which this change unblocks) is the intended
  mechanism for future bumps, not this proposal.
- **Deferring `sli.synthetics.availability` leaves a Kibana-documented indicator type unsupported** →
  Same gap that exists today (not a regression), explicitly scoped out per the issue's own suggested
  escape hatch, and tracked as a separate follow-up issue.

## Open questions

- What is the exact operation/path for `Cases_attachment_type_filter` in the new spec, and do the
  `Cases_string` / `Cases_string_array` component schemas it should reference still exist under those
  names? (They exist in the old spec at `generated/kbapi/oas.yaml:153262-153266` and are already used by
  five sibling filter params with the same `oneOf` shape, e.g. `Cases_assignees_filter` at
  `oas.yaml:150178-150186`, so the fix is very likely a straight port of that existing pattern — but
  this is unverified against the new spec.) **Blocking** for Problem 1's exact transform code, not for
  the overall plan.
- What are the exact new Go type names after the SLO rename? The issue asserts a `KibanaHTTPAPIsSLO*`
  prefix; the old spec's naming convention for other route-generated schemas (e.g.
  `KibanaHTTPAPIsMlAnomalyCharts...`) supports this being a straight prefix swap, but it should be
  confirmed by actually running `make -C generated/kbapi all` against the bumped ref rather than
  assumed. **Blocking** for writing the rename diff.
- Do `internal/kibana/slo/schema.go` and `resource.go` need edits beyond the files already identified by
  grepping `kbapi.SLOs` (19 files: `internal/models/slo.go`, `internal/kibana/slo/*.go`,
  `internal/clients/kibanaoapi/slo.go`)? Not confirmed; verify with a build after regeneration.
- Does the generated shape of `SLOsArtifacts`/dashboards (currently an anonymous struct with an `Id`
  field, hand-converted with `(nolint/redacted):revive` in `internal/kibana/slo/models.go:492-522`)
  survive the rename unchanged, or does the route-schema-generated version differ structurally? The
  issue explicitly warns this may not be a purely mechanical rename, and the human direction confirms
  at least three other SLO types do reshape — this field's shape should be confirmed the same way during
  implementation rather than assumed either way.
- What is the minimum Kibana/Elastic Stack version that supports `sli.synthetics.availability`? Unverified
  anywhere consulted; **blocking** for the deferred follow-up only, not for this change.
- Is there a target Kibana/Elastic Stack version pin for this provider change, and does the project's
  supported-version matrix already include a release containing elastic/kibana#300516? Not addressed in
  the issue.

## References

- `generated/kbapi/Makefile`, `generated/kbapi/transform_schema.go`, `generated/kbapi/transform_schema_test.go`
- `internal/models/slo.go`, `internal/kibana/slo/models.go`, `internal/kibana/slo/constants.go`,
  `internal/kibana/slo/models_apm_latency_indicator.go`, `internal/kibana/slo/models_apm_availability_indicator.go`,
  `internal/kibana/slo/models_kql_custom_indicator.go`, `internal/clients/kibanaoapi/slo.go`,
  `internal/kibana/slo/acc_test.go`
- Issue #5491, elastic/kibana#300516, elastic/kibana#269252, elastic/terraform-provider-elasticstack#2396,
  elastic/terraform-provider-elasticstack#1002, Renovate PR #5464
