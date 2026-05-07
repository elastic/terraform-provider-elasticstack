## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-alerting-rule-artifacts --type change` (or `make check-openspec` after sync).
- [ ] 1.2 Resolve the open question on minimum Kibana version for `artifacts` support (see `design.md`); update delta spec with a version compatibility requirement if confirmed.
- [ ] 1.3 On completion of implementation, **sync** delta into `openspec/specs/kibana-alerting-rule/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [ ] 2.1 Extend `internal/models.AlertingRule` with an `Artifacts *AlertingRuleArtifacts` field; add `AlertingRuleArtifacts`, `AlertingRuleArtifactDashboard`, and `AlertingRuleArtifactInvestigationGuide` types matching the API shape.
- [ ] 2.2 Map `artifacts` in `internal/clients/kibanaoapi/alerting_rule.go`: include `artifacts` in `buildCreateRequestBody` and `buildUpdateRequestBody` when configured (omit the key when not configured); unmarshal `artifacts` from intermediate struct in `ConvertResponseToModel`.
- [ ] 2.3 Add `artifacts` single nested block to `internal/kibana/alertingrule/schema.go`, containing a `dashboards` list nested block (with required `id` string) and an `investigation_guide` single nested block (with optional `content`, optional `content_path`, and computed `checksum`). Enforce mutual exclusion of `content` and `content_path` via `objectvalidator.ExactlyOneOf` or equivalent.
- [ ] 2.4 Add `artifactsModel`, `dashboardModel`, and `investigationGuideModel` to `internal/kibana/alertingrule/models.go`; wire `alertingRuleModel` with `Artifacts types.Object`; implement `populateArtifactsFromAPI` and the `artifacts` portion of `toAPIModel`.
- [ ] 2.5 Implement `ModifyPlan` (or extend the existing one if present) to read the file at `content_path` during plan, compute SHA-256, and mark `checksum` (and the resource `id`) as unknown when the digest changes — mirroring `elasticstack_fleet_custom_integration/plan_modifier.go`.
- [ ] 2.6 On the read path: if prior state had `content` set, populate `content` from API `blob`; if prior state had `content_path` set, preserve `content_path` in state and do not overwrite from API (checksum is managed by the plan modifier).
- [ ] 2.7 Update embedded descriptions (`descriptions/*.md` or `resource-description.md`) and schema `MarkdownDescription` for the `artifacts`, `dashboards`, and `investigation_guide` blocks.
- [ ] 2.8 If a minimum Kibana version is confirmed for `artifacts` (task 1.2), add a version gate in `toAPIModel` (mirroring the `alert_delay` / `flapping` pattern) with a clear diagnostic.

## 3. Testing

- [ ] 3.1 Add acceptance test(s) for `artifacts.dashboards`: create a rule with one or more dashboard IDs; assert state matches; update the list; assert round-trip.
- [ ] 3.2 Add acceptance test for `artifacts.investigation_guide` with inline `content`: create rule with guide text; assert state stores the text; update text; assert change is applied.
- [ ] 3.3 Add acceptance test for `artifacts.investigation_guide` with `content_path`: write a temp file; create rule; assert `checksum` is set; modify file; run `terraform plan`; assert a non-empty plan is produced; apply; assert `checksum` reflects the new file content.
- [ ] 3.4 Add acceptance test for clearing `artifacts`: create rule with artifacts; remove the `artifacts` block; assert Kibana's stored artifacts remain (provider omits key from PUT, API leaves value unchanged); then add `artifacts` back with empty `dashboards` list to clear dashboards explicitly; assert state is consistent.
- [ ] 3.5 Add unit tests for version gating (if a minimum version is confirmed) and for `content` vs `content_path` request body construction, mirroring existing unit test patterns in `models_flapping_test.go`.
- [ ] 3.6 Add unit tests for the read-path mapping: blob → `content` when prior state used `content`; no overwrite of `content_path` when prior state used `content_path`.
