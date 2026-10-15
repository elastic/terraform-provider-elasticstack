## Context

`elasticstack_kibana_dashboard` models XY reference-line thresholds as a typed
`ListNestedAttribute` (`internal/kibana/dashboard/panel/lensxy/schema_attrs.go`).
The `axis` attribute's validator (`stringvalidator.OneOf("bottom", "left",
"right")`) predates the Dashboard API: it was carried over from a pre-9.4 Lens
visualization schema. Checking
`kbn-lens-embeddable-utils/config_builder/schema/charts/xy.ts` at release tags
confirms the enum changed during the 9.4 development cycle (PRs "[Lens as Code]
Improve axes schemas" #261254 and "[Lens as code] Fix Axis definitions"
#262752):

| Kibana | Reference-line `axis` enum | Default |
|---|---|---|
| v9.3.0, v9.3.4 | `bottom \| left \| right` | `left` |
| v9.4.0, v9.4.1, v9.5.0, 9.4/9.5 branches, `main` | `x \| y \| y2` | `y` |

The 9.3 enum belongs to the Lens config-builder schema that predates the provider's
Dashboard API support (9.4+). Whether 9.3 exposes that schema over an HTTP endpoint
was not verified; it does not matter here, because no Kibana version the resource
supports accepts the provider's current schema values. The generated client (`generated/kbapi`,
`KibanaHTTPAPIsVisXyReferenceLineLayerESQLStaticThresholdsAxis`) already models
`x|y|y2`.

Issue #4902 / PR #4948 found and fixed the read-back side of this for Kibana
9.6 (server-injected `axis = "y"` when omitted), landed in
`lens-kibana-9-6-readback-defaults`, but explicitly deferred the schema/validator
rewrite — tracked by this issue — because it is a breaking contract change
distinct from omit-default normalization.

### Live evidence (disposable 9.4.0, 9.5.5, 9.6.0-SNAPSHOT and 9.6.0 stacks)

`POST /api/dashboards` with an XY panel containing a `reference_lines` layer and
one `static_value` threshold, varying only `axis`:

| `axis` sent | 9.4.0 read-back | 9.5.5 read-back | 9.6.0 read-back |
|---|---|---|---|
| omitted | absent | absent | `"y"` (injected) |
| `left` / `right` / `bottom` | HTTP 400 | HTTP 400 | HTTP 400 |
| `x` | echoed | echoed | echoed |
| `y` | **absent (stripped)** | **absent (stripped)** | echoed |
| `y2` | echoed | echoed | echoed |

9.6.0-SNAPSHOT behaved identically to 9.6.0. Only 9.4.0, 9.5.5 and 9.6.x were probed;
other patch releases are assumed to match their minor series. The behavior change
therefore sits between 9.5.5 and 9.6.0.

The only 9.4-vs-9.6 difference is read-back normalization of the default axis,
not the accepted enum:

- 9.6 injects `"y"` when `axis` is omitted — already handled by `#4948` via
  `PreserveNullIfStateEquals(planThreshold.Axis, &stateThreshold.Axis,
  types.StringValue("y"))` at `alignment_panels.go:208`.
- 9.4.0 strips an explicit `"y"` from the read-back. Once `y` is a valid
  schema value, a practitioner who sets `axis = "y"` on a 9.4/9.5 stack would hit
  "Provider produced inconsistent result after apply" (planned `"y"`, state
  null) unless the alignment logic also restores the plan value in that
  direction. This is a new case introduced by this change, not something
  `#4948` needed to handle (the old schema couldn't express `"y"` at all).

Other findings from the same investigation, out of scope for this change but
recorded for anyone picking up related work:

- Y-metric `config_json` `axis` also moved from `left|right` (9.3) to `y|y2`
  on the same timeline. It is free-form JSON in the provider (no schema
  enum), so no schema change is needed there — only the stale
  `converter_test.go:48` fixture (`"axis":"left"`) needs updating.
- Reference-line `position` (`auto|left|right`) is a different field and is
  unchanged on Kibana's side; it is not touched by this change.

## Goals / Non-Goals

**Goals:**
- Make `thresholds[].axis` accept exactly what Kibana's Dashboard API accepts
  on every supported version: `x`, `y`, `y2`.
- Round-trip without plan/state drift on both Kibana 9.4/9.5 (omitted and
  explicit `"y"` are not echoed) and Kibana 9.6+ (omitted injects `"y"`,
  explicit `"y"` is echoed).
- Remove the now-obsolete "omit this attribute" workaround language from the
  schema description, spec, and generated docs.

**Non-Goals:**
- Version gating (`EnforceMinVersion` / `panelkit/version_support.go`). The
  accepted enum is identical across every Kibana version with the Dashboard
  API; only response normalization differs, and that is handled by inspecting
  the response, not the stack version.
- A `StateUpgrader` or deprecation shim for `bottom`/`left`/`right`. No valid
  persisted state can contain these values — they were rejected by Kibana at
  apply time on every version, so no working configuration depends on them.
  Existing configs using the old values will now fail at `terraform plan`
  (clear validator error) instead of `terraform apply` (opaque 400); this is
  strictly earlier, clearer feedback for an already-broken configuration.
- Changing Y-metric `config_json`'s free-form `axis` handling, or
  reference-line `position` — neither has a provider-side enum/validator tied
  to this issue.
- Re-exploring alternatives to a validator/enum rewrite (e.g., keeping the old
  enum and silently remapping `left`→`y`, `right`→`y2`). Silent remapping
  would misrepresent the practitioner's configuration (`left` and `right` are
  not semantically `y`/`y2`) and was rejected in favor of a direct breaking
  enum change, consistent with how `#4948` treated this as a schema-contract
  mismatch rather than an omit-default.

## Decisions

1. **Replace the enum with `x | y | y2`.** Breaking, but justified: the
   removed values are rejected by every supported Kibana, so no working
   configuration can depend on them.
2. **Expose `x`.** Kibana accepts it and the generated client already models
   it; omitting it from the Terraform schema while Kibana allows it would be
   an unnecessary gap.
3. **Preserve an explicit `y` in state, including on 9.4/9.5 where Kibana drops
   it**, so the practitioner's configured value round-trips regardless of
   stack version. Implementation: extend the existing
   `alignXYLayerStateFromPlan` threshold loop
   (`internal/kibana/dashboard/panel/lensxy/alignment_panels.go`) with the
   inverse of today's `PreserveNullIfStateEquals` call — when the plan value
   is known and equals the default `"y"` but the state (API read-back) is
   null, copy the plan value into state. Combined behavior:

   | plan | state (from API) | result |
   |---|---|---|
   | null | `"y"` (9.6 injected) | null (existing behavior, unchanged) |
   | null | null (9.4/9.5 omitted) | null (existing behavior, unchanged) |
   | `"y"` | null (9.4/9.5 stripped) | `"y"` (new) |
   | `"y"` | `"y"` (9.6 echoed) | `"y"` (unchanged; already equal) |
   | `"x"` / `"y2"` | echoed | unchanged (Kibana never strips these) |

   Only the literal default value `"y"` is restored this way; any other value
   (`"x"`, `"y2"`) being stripped by the API would indicate a genuine
   discrepancy and must still surface as a plan-after-apply inconsistency
   rather than being silently papered over.
4. **No version gating** — see Non-Goals.
5. **Keep `axis` `Optional`, not `Computed`**, with the existing
   null-restoration of the 9.6-injected default unchanged. Making it
   `Computed` would be a materially larger schema change than this fix
   requires and was already rejected for the related `#4948` change for the
   same reason (see that change's design doc, "Non-Goals").

The write path (`thresholdToAPI`) and read path (`thresholdFromAPIJSON`) in
`internal/kibana/dashboard/panel/lensxy/layer_conv.go` need no change — they
already pass the `axis` string through verbatim; only the schema validator and
the alignment logic change.

## Risks / Trade-offs

- **Breaking change risk**: any existing configuration using `bottom`,
  `left`, or `right` will fail at `terraform plan` after this change ships,
  where it previously failed at `terraform apply`. This is treated as a
  strict improvement (earlier, clearer error) rather than a regression,
  because no such configuration could ever reach a successful `apply`.
  Mitigate by flagging it as breaking in the implementation PR title/body with
  the rationale (the changelog is generated from merged PR history).
- **New plan-known/state-null restoration direction for `axis = "y"`
  specifically** (Decision 3) is new code, not a port of an existing pattern
  in the other direction. Mitigate with unit tests covering every row of the
  table above, plus an acceptance test run against 9.4 and 9.5 stacks to confirm
  the stripped-`"y"` case is observed and handled (the acceptance matrix in
  `lens-kibana-9-6-readback-defaults` only ran against 9.6.0-SNAPSHOT).
- **Import has no plan to align against.** An imported threshold shows
  `axis = "y"` on Kibana 9.6+ (injected) and `axis = null` on Kibana 9.4/9.5
  (stripped), matching the existing documented import behavior for other
  omit-defaults in REQ-011. For the omitted-axis case this is pre-existing
  behavior. One case is new: because `y` is now configurable, a practitioner on
  9.4/9.5 who sets `axis = "y"` and imports will see `null` in state and a plan
  diff to `"y"` until the next apply. This is accepted (it is consistent with the
  other import omit-defaults and harmless) and is covered by an import test step
  (tasks 4.4) so the behavior is documented rather than accidental.
