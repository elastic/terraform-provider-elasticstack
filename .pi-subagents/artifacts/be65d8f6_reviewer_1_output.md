# Review: openspec-verify-change — `dashboard-description-null-normalization`

## Summary scorecard

| Dimension        | Status | Notes |
|------------------|--------|-------|
| Completeness     | ✅     | All 4 task sections complete (0 unchecked); all artifacts present (proposal/design/specs/tasks). |
| Correctness      | ✅     | Implementation matches design snippet exactly; verified prior-intent seeding across Create/Update/Read paths. |
| Coherence        | ✅     | Spec delta aligns with main spec REQ-008/REQ-009; design decisions followed. |
| Tests            | ✅     | Unit tests pass; acceptance test + fixtures present (CI-gated). |
| Validation        | ✅     | `openspec validate dashboard-description-null-normalization` and `--changes`/`--specs` (`make check-openspec`) all pass; `go build`/`go vet` clean. |

## Correct (evidence)

- **Fix matches design exactly.** `internal/kibana/dashboard/models.go:56-61` implements the intent-preserving check verbatim from `design.md`:
  ```go
  apiDescription := typeutils.StringishPointerValue(data.Data.Description)
  if apiDescription.ValueString() == "" && m.Description.IsNull() {
      m.Description = types.StringNull()
  } else {
      m.Description = apiDescription
  }
  ```
- **Prior-intent seeding verified across all code paths.** The check depends on `m.Description` reflecting prior plan/state before `dashboardPopulateFromAPI` runs:
  - **Create/Update:** `internal/entitycore/kibana_resource_envelope.go:378` calls `r.readFunc(ctx, client, readResourceID, readSpaceID, written.Model)` where `written.Model` is `planModel` (decoded from `req.Plan` at line ~302). So `m.Description` carries the planned intent (null when omitted). ✅
  - **Read/refresh:** `internal/entitycore/base_envelope.go:76` does `req.State.Get(ctx, &model)` before calling `b.read(...)`. So `m.Description` reflects prior state. ✅
  This is the crux of the fix and it is correct.
- **Write path untouched** (task 1.2). `models.go:162-165` and `221-224` still use `typeutils.IsKnown(m.Description)` to send description on create/update; null is omitted. No regression.
- **Spec delta is valid.** `specs/kibana-dashboard/spec.md` extends REQ-008 (read behavior) and REQ-009 (state preservation) with three scenarios (omitted→null, explicit ""→"", non-empty→value) plus a no-drift scenario. Consistent with the main spec's existing REQ-009 null-preservation pattern for `time_range.mode`, panel fields, etc.
- **Unit tests cover all spec scenarios.** `internal/kibana/dashboard/models_dashboard_description_test.go` exercises: API "" + prior null → null; API "" + prior "" → ""; API non-empty → value; API nil → null. All 4 subtests pass.
- **Acceptance test + fixtures present.** `TestAccResourceDashboardDescriptionNormalization` (`acc_test.go`) uses `ConfigDirectory: NamedTestCaseDirectory("omitted")` and `("empty")` with plan-only follow-up steps asserting no drift. Fixtures verified: `testdata/TestAccResourceDashboardDescriptionNormalization/omitted/main.tf` (no `description`) and `empty/main.tf` (`description = ""`).
- **OpenSpec validation clean:** `openspec validate dashboard-description-null-normalization` → "Change is valid"; `openspec validate --changes` → 13/13 passed; `make check-openspec` → 235/235 passed.
- **Build/vet clean:** `go build ./...` and `go vet ./internal/kibana/dashboard/` exit 0.

## Fixed
- None. Review-only; no edits made.

## Blocker
- None.

## Note

- **SUGGESTION — Import edge case (minor, pre-existing pattern).** On import, `resource.ImportStatePassthroughID` (`resource.go:62`) sets only `id` in state, so the subsequent Read sees `m.Description` as null (zero value). If an imported dashboard genuinely had `description = ""`, the normalization would coerce it to null. This matches the existing REQ-009 behavior for `time_range.mode` and panel fields (import cannot know prior intent), and the design's Risks section acknowledges uniform application. No action required; flagged for completeness.
- **WARNING — Acceptance tests not executed locally.** `TestAccResourceDashboardDescriptionNormalization` requires a running 9.5 stack (`TF_ACC=1`); the commit message explicitly defers execution to CI. Unit tests pass and prove the core logic, but the end-to-end "inconsistent-result-after-apply" fix on a live 9.5 stack is unverified in this review. Recommend ensuring CI runs this test against a 9.5 stack before archive.
- **SUGGESTION — Scenario coverage gap vs. spec.** The spec's REQ-009 delta scenario "Empty-string description treated as null for null-intent practitioners" asserts "no drift SHALL be reported on the next plan." The acceptance test's plan-only steps cover this for both omitted and explicit-"" cases, so this is satisfied. No gap.

## Final Assessment

No critical issues. 1 warning (CI-gated acceptance test) and 2 suggestions to consider. **Ready for archive** once CI confirms the acceptance test passes against a 9.5 stack.