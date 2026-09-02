## Context

`internal/elasticsearch/security/role/schema.go` defines `indices` and `remote_indices` as `schema.SetNestedBlock`. Each block's `allow_restricted_indices` attribute (`attrAllowRestrictedIndices`) is `Optional: true, Computed: true`, with `PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}` (lines ~126-133 for `indices`, ~191-198 for `remote_indices`).

`boolplanmodifier.UseStateForUnknown()` only substitutes the prior state value when the **planned** value is unknown. On `Update`, when a `dynamic "indices"` block appends a brand-new set element, Terraform Core's Set-element correlation does not reliably mark that new element's omitted `allow_restricted_indices` as unknown — it can remain a concrete `null` through planning. A plan modifier never runs against an already-concrete `null`, so the plan keeps `null` while Elasticsearch's Put Role API normalizes the omitted field to `false` on write. The resulting Set comparison (`null` planned vs. `false` actual) fails Terraform's post-apply correlation check.

On `Create`, every element is new, so Core marks the omitted computed attribute `Unknown`, and reconciliation against Elasticsearch's `false` succeeds. The bug is therefore specific to appending elements to a `Set` that already has state.

A `resource.UnitTest` against a mock Elasticsearch that echoes omitted `allow_restricted_indices` as `false` reproduced the exact #4759 error on the unfixed schema (`planned set element` with `allow_restricted_indices: cty.NullVal(cty.Bool)`). The same test passed after replacing `UseStateForUnknown` with `Default: booldefault.StaticBool(false)`. That is the evidence this change relies on — not a live-cluster re-run of the update-append path.

## Goals / Non-Goals

**Goals:**
- Eliminate the "does not correlate with any element in actual" failure when a new `indices` or `remote_indices` set element omitting `allow_restricted_indices` is added during `Update`.
- Keep the existing HCL-facing contract: `allow_restricted_indices` remains optional; omitting it continues to mean `false`. Explicit `true` / `false` are unchanged. Create with the field omitted continues to succeed.
- Apply the same fix uniformly to `indices` and `remote_indices`.

**Non-Goals:**
- Do not address `field_security.except` or other `Optional+Computed` attributes using the same `UseStateForUnknown` pattern elsewhere in this resource.
- Do not migrate `indices`/`remote_indices` from `SetNestedBlock` to `ListNestedBlock` or `SetNestedAttribute`.
- Do not perform a provider-wide audit for the same pattern in other resources.
- Do not bump `CurrentSchemaVersion` (still `1`). The attribute type remains bool; `Default` is plan-time only and does not change stored state shape.
- Do not introduce a custom `null ≈ false` type (Approach 2). The mock apply confirmed Approach 1 is sufficient.

## Decisions

- **Mechanism**: replace `PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}` with `Default: booldefault.StaticBool(false)` on `attrAllowRestrictedIndices` for both `indices` and `remote_indices`.
  - `schema.Default` resolves whenever the practitioner's **config** value is null, regardless of what Core's plan-time Set correlation does to the intermediate planned value.
  - Do not keep `UseStateForUnknown` alongside `Default`. If config is unknown (interpolation) rather than null, Default does not fire; the plan stays unknown; `toAPIModel` omits the field; Elasticsearch returns `false`; unknown becoming `false` is a legal apply.
  - This is the same pattern already used widely in the provider for `Optional+Computed` booleans with a fixed default (e.g. `internal/elasticsearch/index/ilm/schema_actions.go`, `internal/fleet/serverhost/schema.go`, `internal/kibana/security_detection_rule/schema.go`).
- **Payload**: after Default resolves, an omitted HCL field is a known `false` at write time, so `toAPIModel` includes `"allow_restricted_indices": false` instead of omitting the key. Elasticsearch treats omit and `false` as the same. `fromAPIModel` is unchanged. REQ-038 must describe the new write behavior.
- **Scope**: apply to both `indices` and `remote_indices` in the same change. They share the schema shape and failure mode.
- **Verification**: the primary regression is a `resource.UnitTest` with a mock Elasticsearch (no `TF_ACC`), matching the #4759 create-then-append sequence. Live acceptance tests cover the same HCL against a real cluster; `remote_indices` steps use the existing 8.10 `SkipFunc`.
- **No model/read-path changes**: do not change `toAPIModel`/`fromAPIModel`. Do not rewrite `models_test.go` fixtures that construct `types.BoolNull()` by hand (e.g. version-requirement tests); those values never go through Framework planning.
- **Docs**: `make docs-generate` picks up the schema Default. `descriptions/allow_restricted_indices.md` does not currently claim a prior-state default, so leave it unless review finds a contradiction.
- **Spec shape**: keep REQ-023–REQ-024 as `field_security.except` `UseStateForUnknown` only. Put Default semantics in a new REQ-039 (and keep REQ-037 aligned). Do not fold Default into the “unknown values in plan” requirement.

## Risks / Trade-offs

- [Risk] PUT bodies for omitted `allow_restricted_indices` gain an explicit `false`. → Mitigation: Elasticsearch already normalizes omit to `false`; no practitioner-visible privilege change. Document as a bug fix, not a breaking change.
- [Risk] An omitted field after an explicit `true` in state now plans `false` for pre-existing set elements, not only newly appended ones, and the next apply revokes restricted-index access. → Mitigation: this matches REQ-039 and the Elasticsearch omit-means-false contract; call it out in the PR changelog so practitioners in that narrow situation are not surprised.
- [Risk] Practitioners inspecting state after Create-with-omit already see `false` today (EmptySets acc test). After this change they also see `false` in the *plan* before apply. → Mitigation: intended; removes the Create/Update inconsistency.
- [Risk] A live-only test would hide the Framework correlation bug when no stack is available. → Mitigation: the mock UnitTest is the merge gate for the correlation failure; live acc tests are additional coverage.

## Open questions

- Does `Default: booldefault.StaticBool(false)` alone clear the update-append repro? — **Resolved.** A mock `resource.UnitTest` reproduced #4759 on the unfixed schema and passed after Default. Approach 2 is not needed.
- Should the same fix apply to `remote_indices.allow_restricted_indices`? — **Resolved: yes.**
- Are other `Optional+Computed` attributes nested in `indices`/`remote_indices` (e.g. `field_security.except`) in scope? — **Resolved: out of scope.**
