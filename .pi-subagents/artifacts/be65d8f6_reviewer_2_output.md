# Schema coverage report: `elasticstack_kibana_dashboard` — `description` attribute

## Scope
- **Schema**: `internal/kibana/dashboard/schema.go:105-107` (`"description"`: `schema.StringAttribute`, Optional, not Computed, no ForceNew, no plan modifiers)
- **Normalization logic (read)**: `internal/kibana/dashboard/models.go:57-66` (`dashboardPopulateFromAPI`)
- **Create/update request building**: `internal/kibana/dashboard/models.go:138-141` (create), `models.go:146-149` (update)
- **Acceptance tests**: `internal/kibana/dashboard/acc_test.go`
- **Unit test (read normalization)**: `internal/kibana/dashboard/models_dashboard_description_test.go`
- **Test fixtures**: `internal/kibana/dashboard/testdata/TestAccResourceDashboardDescriptionNormalization/{omitted,empty}/main.tf`

## What the change does

`dashboardPopulateFromAPI` (models.go:60-66) now conditionally preserves null on read:

```go
apiDescription := typeutils.StringishPointerValue(data.Data.Description)
if apiDescription.ValueString() == "" && m.Description.IsNull() {
    m.Description = types.StringNull()
} else {
    m.Description = apiDescription
}
```

Create/update (`dashboardToAPICreateRequest` / `dashboardToAPIUpdateRequest`) sends description only when `typeutils.IsKnown(m.Description)` is true (i.e. non-null, non-unknown): null → omitted from request body; `""` → sent as `&""`; non-empty → sent as the value.

## Coverage matrix for `description`

| State        | Configured? | Asserted? (value-specific) | Update/transition covered? |
|--------------|-------------|-----------------------------|----------------------------|
| `null` (omitted) | ✅ `normalization/omitted` | ✅ `TestCheckNoResourceAttr` acc_test.go:773 | ❌ no transition |
| `""` (empty)     | ✅ `normalization/empty`   | ✅ `TestCheckResourceAttr(...,"")` acc_test.go:795 | ❌ no transition |
| non-empty        | ✅ many configs            | ✅ acc_test.go:59,77,622,641 | ✅ non-empty→non-empty (EmptyDashboard step1→2 acc_test.go:59→77; UserSuppliedID create→updated acc_test.go:622→641) |

Assertion quality: **good** — every description assertion across the suite is value-specific (`TestCheckResourceAttr`/`TestCheckNoResourceAttr`). There are **no set-only** (`TestCheckResourceAttrSet`) assertions for `description` anywhere.

Unit test (`models_dashboard_description_test.go:54-99`) covers the read-mapping matrix for fixed prior states: API-nil/prior-null→null, API-`""`/prior-null→null, API-`""`/prior-`""`→`""`, API-non-empty/prior-null→value. ✅

## 1) Attributes with no coverage

None. All three description states (null, `""`, non-empty) are at least configured and value-asserted somewhere in the suite.

## 2) Attributes with poor coverage

### P1 — Missing update/transition coverage for description state changes
- **Path**: `description`
- **Observed**: `TestAccResourceDashboardDescriptionNormalization` (acc_test.go:754-810) covers the `null` and `""` states as **two separate resources** (the `""` case uses a distinct `dashboard_title + " (empty)"` suffix, acc_test.go:791). The plan-only steps (acc_test.go:782, 802) verify no-drift for the *same* config only — they do not exercise transitions between states on a single resource.
- **Gaps**:
  - No single-resource transition coverage for:
    - omitted(`null`) → `""`
    - `""` → omitted(`null`)
    - omitted(`null`) → non-empty → omitted(`null`)
    - non-empty → `""`
  - The normalization is **prior-state dependent** (models.go:61-62: `m.Description.IsNull()`), so transitions are precisely where edge cases surface. The unit test covers fixed-prior-state read mapping, but not the full apply→read cycle across a transition.
- **Suggested improvements**:
  - Convert the normalization test to a single-resource lifecycle: create omitted → update to `description=""` → update to a non-empty value → update back to omitted, asserting state after each step and a final plan-only no-drift step.

### P2 — Missing ImportStateVerify in the normalization test
- **Path**: `description`
- **Observed**: `TestAccResourceEmptyDashboard` includes an import step (acc_test.go:108-115), but `TestAccResourceDashboardDescriptionNormalization` (acc_test.go:754-810) does **not**. Importing an omitted-description dashboard on Kibana 9.5 is a distinct code path: there is **no prior state** to guide the `m.Description.IsNull()` branch, so the normalization relies solely on the freshly-built model being null — this path is untested.
- **Gaps**: no import-state round-trip for an omitted (`null`) description; no import for an explicit `description=""`.
- **Suggested improvements**:
  - Add an `ImportState: true, ImportStateVerify: true` step after the `omitted` apply and after the `empty` apply.

### P3 — Non-empty round-trip absent from the normalization-specific test
- **Path**: `description`
- **Observed**: The dedicated normalization test covers only `null` and `""`. The non-empty "preserve unchanged" branch (REQ-008 scenario 3, spec.md) is exercised by `TestAccResourceEmptyDashboard` (acc_test.go:59) and the unit test (models_dashboard_description_test.go:80-84), so it is covered elsewhere — but the normalization test does not exercise the full null/empty/non-empty matrix it implies.
- **Gaps**: non-empty case not in the normalization test's stated matrix.
- **Suggested improvements**: add a non-empty step to the normalization test (lowest priority — already covered elsewhere).

### P4 — `with_options` description configured but not value-asserted
- **Path**: `description`
- **Observed**: `TestAccResourceEmptyDashboard` step 3 (`with_options`) sets `description = "Test dashboard with options"` (testdata/TestAccResourceEmptyDashboard/with_options/main.tf:7) but the Check block (acc_test.go:88-103) does **not** assert `description`. It is only indirectly verified via `ImportStateVerify` in step 4 (acc_test.go:108-115).
- **Gaps**: configured-but-unasserted in that step (weak per skill guidance: "Don't mark an attribute as covered solely because it appears in raw HCL").
- **Suggested improvements**: add `resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "description", "Test dashboard with options")` to step 3.

### P5 — Panel/section tests configure description but never assert it
- **Path**: `description`
- **Observed**: Multiple tests set a non-empty `description` in their config (e.g. `TestAccResourceDashboardPanels_basic` uses `basic` with description set; `with_sections`, `multiple_panels`, `panels_and_sections`, `multi_sections_*`, `with_text`, `with_json`, `lens_metric`, `no_tags`) but their Check blocks assert only title/panels/sections — never `description`.
- **Gaps**: configured-but-unasserted in ~6+ tests.
- **Suggested improvements**: low priority; not a normalization risk. Optionally add a single value assertion in one of these tests if description drift is a concern.

## High-risk untested behaviors

### H1 — `non-empty → omitted(null)` update drift (HIGH risk; may be pre-existing / out of normalization scope)
- **Behavior**: When description transitions from a non-empty value to omitted, `dashboardToAPIUpdateRequest` (models.go:146-149) **omits** `description` from the PUT body because `IsKnown(null)` is false. Kibana's partial-update semantics will likely **retain** the old non-empty description remotely. On read-back, `dashboardPopulateFromAPI` (models.go:60-66) sees a non-empty API value with prior state `null`: the `apiDescription.ValueString() == ""` condition is **false**, so the else branch sets state to the non-empty value — **not** null. Result: config is `null`, state is non-empty → **drift on the next plan**.
- **Why the normalization does not help**: the normalization only maps API `""` → null when prior is null; it does **not** handle a stale non-empty remote value being "cleared" to null. There is no `description = null` clearing sent on update.
- **Evidence**: models.go:146-149 (update omits null description); models.go:60-66 (normalization only handles `""` → null).
- **Untested**: no acceptance test transitions non-empty → omitted and then re-plans.
- **Note**: This is arguably outside the `dashboard-description-null-normalization` change scope (which targets the Kibana 9.5 omitted→`""` echo on read), but it is the highest-risk description-related untested behavior. Worth confirming whether the spec intends update-side clearing or accepts remote retention.

### H2 — `"" → omitted(null)` update: normalization masks remote retention (MEDIUM risk)
- **Behavior**: Transition `description=""` → omitted: PUT omits description; Kibana likely keeps `""` remotely. Read-back: API `""` + prior `null` → normalization sets state to `null` → **no drift shown**, but the remote still holds `""`. The normalization hides the divergence rather than reconciling it.
- **Evidence**: same update/read paths as H1.
- **Untested**: no transition test for `""` → omitted.
- **Note**: behaviorally "correct" from Terraform's drift perspective (state matches config), but the remote is not actually cleared. Confirm whether that is acceptable.

### H3 — Create/update request building has no unit test (LOW risk)
- **Behavior**: `dashboardToAPICreateRequest`/`dashboardToAPIUpdateRequest` (models.go:138-141, 146-149) use `typeutils.IsKnown(m.Description)` to decide whether to send `&desc`. The three branches (null→omit, `""`→send `&""`, non-empty→send value) are exercised only indirectly via acceptance tests, with no unit test for the request-building path.
- **Evidence**: no test references `dashboardToAPICreateRequest`/`dashboardToAPIUpdateRequest` with a description assertion (grep of `internal/kibana/dashboard/*_test.go`).
- **Suggested improvements**: add a small unit test mirroring `models_dashboard_description_test.go` for the create/update request builders.

## Suggested next steps (smallest diffs first)
1. Add value-specific assertion for `description` in `TestAccResourceEmptyDashboard` step 3 (`with_options`) — one line (P4).
2. Add `ImportState`/`ImportStateVerify` steps to `TestAccResourceDashboardDescriptionNormalization` after the `omitted` and `empty` applies (P2).
3. Restructure `TestAccResourceDashboardDescriptionNormalization` into a single-resource lifecycle covering transitions `omitted → "" → non-empty → omitted`, with assertions and a final plan-only no-drift step (P1, also exercises H1/H2).
4. Add a unit test for `dashboardToAPICreateRequest`/`dashboardToAPIUpdateRequest` description handling (H3).
5. (Optional) Confirm with the spec owner whether `non-empty → null` should clear the remote description (H1) or accept remote retention.

## Verdict
The `null`/`empty`/`non-empty` **read** matrix is adequately covered at the unit level and reasonably covered at the acceptance level for the static (single-state) cases. The main weaknesses are: (a) no transition/update coverage between description states on a single resource, (b) no import round-trip for the normalized states, and (c) the highest-risk behavior — non-empty→omitted update drift — is untested and is **not** addressed by the read-only normalization. No blockers for the read-side normalization change itself; the items above are prioritized follow-ups.