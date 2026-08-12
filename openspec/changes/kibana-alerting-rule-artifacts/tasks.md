> **Implementation scope note:** The **investigation_guide** portion shipped in PR #4489. This follow-up completes the **`artifacts.dashboards`** portion (REQ-045 dashboards, REQ-047 dashboards mapping): a `dashboards` list nested attribute, request/response mapping, read-back preservation, at-least-one validation, and tests. All deferred items below are now implemented.

## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-alerting-rule-artifacts --type change` (or `make check-openspec` after sync).
- [x] 1.2 Resolve the open question on minimum Kibana version for `artifacts` support: **confirmed 9.1.0** — `investigation_guide` first appears in the alerting rule create/update bodies of the Kibana OpenAPI at v9.1.0 (absent in v9.0.0). A `GetVersionRequirements` gate enforces `>= 9.1.0` with a clear diagnostic.
- [ ] 1.3 On completion of implementation, **sync** delta into `openspec/specs/kibana-alerting-rule/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [x] 2.1 Extend `internal/models.AlertingRule` with an `Artifacts *AlertingRuleArtifacts` field; added `AlertingRuleArtifacts` and `AlertingRuleInvestigationGuide` types. (Dashboard type deferred with the dashboards feature.)
- [x] 2.2 Map `artifacts` in `internal/clients/kibanaoapi/alerting_rule_builders.go`: `artifacts.investigation_guide.blob` is included in `buildCreateRequestBody`/`buildUpdateRequestBody` (via the shared `ruleBodyOptionalFields`/`artifactsWire`) when configured and omitted otherwise; `ConvertResponseToModel` unmarshals `artifacts.investigation_guide.blob`. (Dashboards mapping deferred.)
- [x] 2.3 Added `artifacts` to `internal/kibana/alertingrule/schema.go` as a `SingleNestedAttribute` (not `SingleNestedBlock`) containing an `investigation_guide` `SingleNestedAttribute` (optional `content`, optional `content_path`, computed `checksum`). Mutual exclusion of `content`/`content_path` enforced via `stringvalidator.ExactlyOneOf`. **Design deviation:** the design specified single nested *blocks*, but an omitted `SingleNestedBlock` still materialises as a non-null object, which makes the nested `ExactlyOneOf` validator fire even when `artifacts` is absent (breaking every rule without artifacts). `SingleNestedAttribute` is genuinely null when omitted (matching the existing optional `flapping` attribute) and fixes this. HCL therefore uses attribute-assignment syntax (`artifacts = { investigation_guide = { ... } }`). When dashboards are added, `artifacts` may need to become a block again with `dashboards` as a list nested block; investigation_guide can remain a nested attribute. (Dashboards deferred.)
- [x] 2.4 Added `artifactsModel` and `investigationGuideModel` to `internal/kibana/alertingrule/models.go`; wired `alertingRuleModel.Artifacts types.Object`; implemented `populateArtifactsFromAPI` and the `artifacts` portion of `toAPIModel`. (`dashboardModel` deferred.)
- [x] 2.5 Implemented `ModifyPlan` in `internal/kibana/alertingrule/plan_modifier.go` to read the file at `content_path` during plan, compute SHA-256, and mark `checksum` unknown when the digest changes (or on create) — mirroring `elasticstack_fleet_custom_integration/plan_modifier.go`. The concrete checksum is recorded in Create/Update via `applyInvestigationGuideChecksum`.
- [x] 2.6 Read path: `content` is populated from API `blob` when prior state used `content`; when prior state used `content_path`, the path and checksum are preserved and the blob is not surfaced as `content`.
- [x] 2.7 Added embedded descriptions (`descriptions/artifacts.md`, `descriptions/investigation_guide.md`) and schema `MarkdownDescription` for the `artifacts` and `investigation_guide` blocks; regenerated `docs/resources/kibana_alerting_rule.md`.
- [x] 2.8 Added a version gate in `GetVersionRequirements` requiring Elastic Stack `>= 9.1.0` when `artifacts` is set, with a clear diagnostic.

### Dashboards (this follow-up)

- [x] 2.1d Added `AlertingRuleArtifactDashboard` model type and dashboards mapping in the client builders (`artifactsWire.Dashboards` on the request; `ConvertResponseToModel` reads `artifacts.dashboards[].id`).
- [x] 2.3d Added the `dashboards` list nested **attribute** (required `id` string) to the schema, `Optional`+`Computed` with `listplanmodifier.UseStateForUnknown()`. Consistent with the attribute-based `artifacts` (no revert to blocks needed). The `artifacts` object validator was relaxed from `AlsoRequires(investigation_guide)` to a `ValidateConfig` "at least one of investigation_guide or dashboards" check.
- [x] 2.4d Added `dashboardModel` and dashboards conversion in `models.go`; refactored the artifacts rebuild sites (`populateArtifactsFromAPI`, `artifactsToAPI`, `applyInvestigationGuideChecksum`, `setInvestigationGuideChecksumUnknown`) onto a shared `buildArtifactsObject` helper so touching one field never drops the other. Read path preserves prior dashboards when the API omits artifacts (pre-9.5.0).

## 3. Testing

- [x] 3.1 Added `TestAccResourceAlertingRuleArtifactsDashboards`: create with two linked dashboards, update the list (add/remove), and a 9.5.0-gated import step that proves the GET round-trip. Plus unit tests for dashboards request/response mapping and model conversion, and `validateArtifactsNotEmpty`.
- [x] 3.2 Added `TestAccResourceAlertingRuleInvestigationGuide` inline-`content` steps (create + update text; asserts state stores/updates the text). Gated at Stack `>= 9.1.0`.
- [x] 3.3 Added `TestAccResourceAlertingRuleInvestigationGuide` `content_path` steps: writes a temp file, creates the rule, asserts `checksum` is set, mutates the file via `PreConfig`, and asserts a non-empty (update) plan via `plancheck.ExpectResourceAction`.
- [x] 3.4 The dashboards acceptance test's update step exercises replacing the dashboards list; unit tests cover the empty-`artifacts` rejection and dashboards preservation when the API omits artifacts.
- [x] 3.5 Added unit tests for `content` vs `content_path` request-body construction in `internal/clients/kibanaoapi/alerting_rule_artifacts_body_test.go`. (Version gating verified via `GetVersionRequirements`.)
- [x] 3.6 Added unit tests for read-path mapping in `internal/kibana/alertingrule/models_artifacts_test.go`: blob → `content` when prior used `content`; no overwrite when prior used `content_path`.
