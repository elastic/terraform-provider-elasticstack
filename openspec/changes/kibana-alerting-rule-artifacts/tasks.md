## 1. Spec

- [x] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-alerting-rule-artifacts --type change` (or `make check-openspec` after sync).
- [x] 1.2 Minimum Kibana versions confirmed: write **8.19.0** / **9.1.0**; public GET round-trip **9.5.0** (kibana#247279). Recorded in delta spec as REQ-053–REQ-055.
- [x] 1.3 On completion of implementation, **sync** delta into `openspec/specs/kibana-alerting-rule/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [x] 2.1 Extend `internal/models.AlertingRule` with an `Artifacts *AlertingRuleArtifacts` field; add `AlertingRuleArtifacts`, `AlertingRuleArtifactDashboard`, and `AlertingRuleArtifactInvestigationGuide` types matching the API shape.
- [x] 2.2 Map `artifacts` in `internal/clients/kibanaoapi/alerting_rule.go`: include `artifacts` in `buildCreateRequestBody` and `buildUpdateRequestBody` when configured (omit the key when not configured); unmarshal `artifacts` from intermediate struct in `ConvertResponseToModel`.
- [x] 2.3 Add `artifacts` single nested block to `internal/kibana/alertingrule/schema.go`, containing a `dashboards` list nested block (with required `id` string) and an `investigation_guide` single nested block (with optional `content`, optional `content_path`, and computed `checksum`). Enforce mutual exclusion of `content` and `content_path` via `objectvalidator.ExactlyOneOf` or equivalent.
- [x] 2.4 Add `artifactsModel`, `dashboardModel`, and `investigationGuideModel` to `internal/kibana/alertingrule/models.go`; wire `alertingRuleModel` with `Artifacts types.Object`; implement `populateArtifactsFromAPI` and the `artifacts` portion of `toAPIModel`.
- [x] 2.5 Implement `ModifyPlan` (or extend the existing one if present) to read the file at `content_path` during plan, compute SHA-256, and mark `checksum` (and the resource `id`) as unknown when the digest changes — mirroring `elasticstack_fleet_custom_integration/plan_modifier.go`.
- [x] 2.6 On the read path: if prior state had `content` set, populate `content` from API `blob`; if prior state had `content_path` set, preserve `content_path` in state and do not overwrite from API (checksum is managed by the plan modifier).
- [x] 2.7 Update embedded descriptions (`descriptions/*.md` or `resource-description.md`) and schema `MarkdownDescription` for the `artifacts`, `dashboards`, and `investigation_guide` blocks. Document minimum versions (**8.19** / **9.1** for write) and the public GET limitation on stacks before kibana#247279.
- [x] 2.8 Add version gates in `toAPIModel` when `artifacts` is configured (mirroring `alert_delay` / `flapping`): fail below **8.19.0** on 8.x and below **9.1.0** on 9.x with a diagnostic naming both minimums (REQ-053).

## 3. Testing

- [x] 3.1 Add acceptance test(s) for `artifacts.dashboards`: create a rule with one or more dashboard IDs; assert state matches; update the list. Skip when stack is below **8.19.0** (8.x) or **9.1.0** (9.x). For assertions that depend on GET returning `artifacts`, skip below **9.5.0** unless CI stack includes the kibana#247279 backport.
- [x] 3.2 Add acceptance test for `artifacts.investigation_guide` with inline `content`: create rule with guide text; assert state stores the text; update text. Same version skips as 3.1; read assertions from API gated at **9.5.0** where applicable.
- [x] 3.3 Add acceptance test for `artifacts.investigation_guide` with `content_path`: write a temp file; create rule; assert `checksum` is set; modify file; run `terraform plan`; assert a non-empty plan is produced; apply; assert `checksum` reflects the new file content. Same write-version skip as 3.1.
- [x] 3.4 Add acceptance test for clearing `artifacts`: create rule with artifacts; remove the `artifacts` block; assert Kibana's stored artifacts remain (provider omits key from PUT, API leaves value unchanged). On stacks where GET omits `artifacts`, assert state preserves configured values rather than clearing (REQ-054).
- [x] 3.5 Add unit tests for version gating (8.19 / 9.1 thresholds) and for `content` vs `content_path` request body construction, mirroring existing unit test patterns in `models_flapping_test.go`.
- [x] 3.6 Add unit tests for the read-path mapping: blob → `content` when prior state used `content`; no overwrite of `content_path` when prior state used `content_path`; preserve known `artifacts` when API omits the field (REQ-054).
