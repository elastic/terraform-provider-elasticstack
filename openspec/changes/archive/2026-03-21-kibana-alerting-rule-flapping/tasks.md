## 1. Spec

- [x] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `npx openspec validate kibana-alerting-rule-flapping --type change` (or `make check-openspec` after sync).
- [x] 1.2 On completion of implementation, **sync** delta into `openspec/specs/kibana-alerting-rule/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [x] 2.1 Extend `internal/models.AlertingRule` (and related types if needed) with rule-level flapping fields matching API shape.
- [x] 2.2 Map `flapping` in `internal/clients/kibanaoapi`: `buildCreateRequestBody`, `buildUpdateRequestBody`, and `ConvertResponseToModel` (omit `flapping` on update when not configured).
- [x] 2.3 Add `flapping` single nested block to `internal/kibana/alertingrule/schema.go` with **int64** attributes; require both count attributes when the block is present; optional `enabled`.
- [x] 2.4 Wire `alertingRuleModel` / `toAPIModel` / `populateFromAPI` with version check **≥ 8.16.0** when flapping is configured (mirror `alert_delay` / `alerts_filter` patterns in `models.go`).
- [x] 2.5 Update embedded descriptions (`descriptions/*.md` or `resource-description.md`) and generated docs if applicable.
- [x] 2.6 When **`flapping.enabled`** is set in configuration, enforce stack **≥ 9.3.0** on create/update with a clear diagnostic if the version is lower (leave integer-only `flapping` at the **8.16.0** gate).
- [x] 2.7 Document **`flapping.enabled`** minimum version (**9.3**) in schema / embedded resource descriptions so practitioners see it in registry docs.

## 3. Testing

- [x] 3.1 Add acceptance test steps (fixtures under `internal/kibana/alertingrule/testdata` or project convention) with `SkipFunc` for stack **&lt; 8.16.0**: create with `flapping`, assert state; update values; assert API round-trip if existing helpers allow.
- [x] 3.2 Add or extend unit tests for version gating and request body omission behavior where practical.
- [x] 3.3 Gate any acceptance test that configures **`flapping.enabled`** on stack **≥ 9.3.0** (for example a dedicated `minSupportedVersion` / `SkipFunc` for those steps).
- [x] 3.4 Ensure acceptance coverage for **`flapping`** with **only** `look_back_window` and **`status_change_threshold`** (no `enabled`) remains available on **8.16.0+** stacks—add fixtures and test steps if the suite would otherwise only cover `enabled`.
- [x] 3.5 Extend unit tests for the **`enabled`** vs **8.16-only** flapping version matrix (e.g. `enabled` set at 8.16 → error; integers only at 8.16 → OK).
