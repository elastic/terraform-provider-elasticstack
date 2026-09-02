## Context

`internal/elasticsearch/security/role/schema.go` defines `indices` and `remote_indices` as `schema.SetNestedBlock`. Each block's `allow_restricted_indices` attribute (`attrAllowRestrictedIndices`) is `Optional: true, Computed: true`, with `PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}` (lines ~126-133 for `indices`, ~191-198 for `remote_indices`).

`boolplanmodifier.UseStateForUnknown()` only substitutes the prior state value when the **planned** value is unknown. On `Update`, when a `dynamic "indices"` block appends a brand-new set element, Terraform Core's Set-element correlation does not reliably mark that new element's omitted `allow_restricted_indices` as unknown — it can remain a concrete `null` through planning. A plan modifier never runs against an already-concrete `null`, so the plan keeps `null` while Elasticsearch's Put Role API normalizes the omitted field to `false` on write. The resulting Set comparison (`null` planned vs. `false` actual) fails Terraform's post-apply correlation check.

On `Create`, every element is new, so Core marks the whole object (including omitted computed attributes) `Unknown`, and the plan modifier — or, after this change, the schema default — resolves cleanly. The bug is therefore specific to appending elements to a `Set` that already has state.

## Goals / Non-Goals

**Goals:**
- Eliminate the "does not correlate with any element in actual" failure when a new `indices` or `remote_indices` set element omitting `allow_restricted_indices` is added during `Update`.
- Keep the existing HCL-facing contract: `allow_restricted_indices` remains optional; omitting it continues to mean "let Elasticsearch/the provider decide," which resolves to `false`.
- Apply the same fix uniformly to `indices` and `remote_indices` (human direction on the issue confirmed this scope).

**Non-Goals:**
- Do not address `field_security.except` or other `Optional+Computed` attributes using the same `UseStateForUnknown` pattern elsewhere in this resource — confirmed out of scope by human direction on the issue.
- Do not migrate `indices`/`remote_indices` from `SetNestedBlock` to `ListNestedBlock` or `SetNestedAttribute` to change Core's correlation semantics — that is a larger, separately-scoped change.
- Do not perform a provider-wide audit for the same pattern in other resources.

## Decisions

- **Mechanism**: replace `PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}` with `Default: booldefault.StaticBool(false)` on `attrAllowRestrictedIndices` for both `indices` and `remote_indices`.
  - `schema.Default` resolves whenever the practitioner's **config** value is null, regardless of what Core's plan-time Set correlation does to the intermediate planned value. This sidesteps the specific gap that defeats `UseStateForUnknown` for newly-appended Set elements, because it does not depend on Core first marking the value `Unknown`.
  - This is the same pattern already used ~60 times elsewhere in the provider for `Optional+Computed` booleans with a fixed default (e.g. `internal/elasticsearch/index/ilm/schema_actions.go`, `internal/fleet/serverhost/schema.go`, `internal/kibana/security_detection_rule/schema.go`), so it is idiomatic for this codebase.
  - Matches the actual Elasticsearch behavior: `PUT /_security/role/{name}` normalizes an omitted `allow_restricted_indices` to `false`, so hard-coding `false` as the schema default introduces no drift between plan and API.
- **Scope**: apply to both `indices.allow_restricted_indices` and `remote_indices.allow_restricted_indices` in the same change, since they share the identical schema shape and failure mode, and human direction on the issue confirmed both should be fixed together.
- **Verification-first**: implementation must add an acceptance test that reproduces the issue's exact repro steps (create with one `indices` element via `dynamic`, then a second `apply` that appends a second element through the same `dynamic` block, omitting `allow_restricted_indices`) before considering the fix complete, per the research comment's blocking open question — human direction on the issue already confirmed `Default` clears the repro against a live cluster, but the regression test still needs to exist in-tree.
- **No model/read-path changes**: `toAPIModel`/`fromAPIModel` in `models.go` already build the correct Put Role payload and parse the correct read-back value. This change is confined to the plan-time default mechanism in `schema.go`; no changes to `models.go` mapping logic are needed.

## Risks / Trade-offs

- [Risk] Existing unit tests assert `allow_restricted_indices` remains `types.BoolNull()` in scenarios where the schema now resolves a default (e.g. `models_test.go`). → Mitigation: update those test expectations to `types.BoolValue(false)` where they represent a config-null case that the new default now resolves at plan time.
- [Risk] Practitioners relying on `allow_restricted_indices` remaining literally `null` in state after a `Create` with the field omitted will now see `false` in state instead. → Mitigation: this matches the value Elasticsearch already returns and stores server-side; `false` and omitted are semantically identical to Elasticsearch, and this is the intended bug fix, not a new incompatibility. Document as a bug fix, not a breaking change, since `null`/`false` were already inconsistent in practice depending on Create vs. Update.

## Open questions

- Does `Default: booldefault.StaticBool(false)` alone clear the repro against a live ES cluster? (blocking) — **Resolved by human direction on the issue: confirmed, continue with this approach.**
- Should the same fix apply to `remote_indices.allow_restricted_indices` (same shape, not reported but presumably susceptible)? — **Resolved by human direction on the issue: yes.**
- Are other `Optional+Computed` attributes nested in `indices`/`remote_indices` elements (e.g. `field_security.except`) susceptible to the same failure mode? — **Resolved by human direction on the issue: out of scope for this change.**
