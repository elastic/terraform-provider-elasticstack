## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `npx openspec validate kibana-alerting-rule-flapping --type change` (or `make check-openspec` after sync).
- [ ] 1.2 On completion of implementation, **sync** delta into `openspec/specs/kibana-alerting-rule/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [ ] 2.1 Extend `internal/models.AlertingRule` (and related types if needed) with rule-level flapping fields matching API shape.
- [ ] 2.2 Map `flapping` in `internal/clients/kibanaoapi`: `buildCreateRequestBody`, `buildUpdateRequestBody`, and `ConvertResponseToModel` (omit `flapping` on update when not configured).
- [ ] 2.3 Add `flapping` single nested block to `internal/kibana/alertingrule/schema.go` with **int64** attributes; require both count attributes when the block is present; optional `enabled`.
- [ ] 2.4 Wire `alertingRuleModel` / `toAPIModel` / `populateFromAPI` with version check **≥ 8.16.0** when flapping is configured (mirror `alert_delay` / `alerts_filter` patterns in `models.go`).
- [ ] 2.5 Update embedded descriptions (`descriptions/*.md` or `resource-description.md`) and generated docs if applicable.

## 3. Testing

- [ ] 3.1 Add acceptance test steps (fixtures under `internal/kibana/alertingrule/testdata` or project convention) with `SkipFunc` for stack **&lt; 8.16.0**: create with `flapping`, assert state; update values; assert API round-trip if existing helpers allow.
- [ ] 3.2 Add or extend unit tests for version gating and request body omission behavior where practical.
