## 1. kibana/security_exception_item

- [x] 1.1 Add `GetID()`, `GetResourceID()` (→ `ItemID`), `GetSpaceID()` (→ `SpaceID`), `GetKibanaConnection()` value-receiver methods to the exception item model struct
- [x] 1.2 Add `GetVersionRequirements()` to the model: emit `MinVersionExpireTime` (8.7.2) with the existing error message when `ExpireTime` is known and non-null
- [x] 1.3 Delete inline `EnforceMinVersion(ctx, MinVersionExpireTime)` checks inside `models.go`; drop the `client clients.MinVersionEnforceable` parameter from `toCreateRequest`/`toUpdateRequest` (and any other helper) once the inline check is removed
- [x] 1.4 Remove `kibana_connection` block from `schema.go` schema function
- [x] 1.5 Extract `Create` method body to `func createExceptionItem(ctx, *KibanaScopedClient, KibanaWriteRequest[ExceptionItemModel]) (KibanaWriteResult[ExceptionItemModel], diag.Diagnostics)`
- [x] 1.6 Extract `Read` method body to `func readExceptionItem(ctx, *KibanaScopedClient, resourceID, spaceID string, model ExceptionItemModel) (ExceptionItemModel, bool, diag.Diagnostics)`
- [x] 1.7 Extract `Update` method body to `func updateExceptionItem(ctx, *KibanaScopedClient, KibanaWriteRequest[ExceptionItemModel]) (KibanaWriteResult[ExceptionItemModel], diag.Diagnostics)`
- [x] 1.8 Extract `Delete` method body to `func deleteExceptionItem(ctx, *KibanaScopedClient, resourceID, spaceID string, model ExceptionItemModel) diag.Diagnostics`
- [x] 1.9 Swap `*entitycore.ResourceBase` for `*entitycore.KibanaResource[ExceptionItemModel]` in resource struct and constructor; retain `ValidateConfig` on the wrapper struct
- [x] 1.10 Add `entitycore_contract_test.go` asserting the resource embeds `KibanaResource[ExceptionItemModel]`
- [x] 1.11 Run `make build` and `go test ./internal/kibana/security_exception_item/...`

## 2. kibana/security_detection_rule

- [x] 2.1 Add `GetID()`, `GetResourceID()` (→ `RuleID`), `GetSpaceID()` (→ `SpaceID`), `GetKibanaConnection()` value-receiver methods to the `Data` model struct
- [x] 2.2 Add `GetVersionRequirements()` to `Data`: emit `MinVersionResponseActions` (8.16.0) when the rule configures response actions; emit `MinVersionAlertsFilter` (8.9.0) when alerts_filter is configured (mirror the existing conditions inside `models.go`/processor `ToCreateProps`/`ToUpdateProps`)
- [x] 2.3 Delete inline `EnforceMinVersion` calls within model conversion paths; drop `clients.MinVersionEnforceable` parameters from `ToCreateProps`/`ToUpdateProps` on each rule-type processor (`models_eql.go`, `models_esql.go`, `models_query.go`, `models_saved_query.go`, `models_machine_learning.go`, `models_new_terms.go`, `models_threat_match.go`, `models_threshold.go`) and from any shared helpers in `models.go`/`models_to_api_type_utils.go`/`rule_processor.go`
- [x] 2.4 Remove `kibana_connection` block from `schema.go` schema function
- [x] 2.5 Promote the private `read` helper to a package-level `readDetectionRule(ctx, *KibanaScopedClient, resourceID, spaceID string, model Data) (Data, bool, diag.Diagnostics)` callback function
- [x] 2.6 Extract `Create` method body to `func createDetectionRule(ctx, *KibanaScopedClient, KibanaWriteRequest[Data]) (KibanaWriteResult[Data], diag.Diagnostics)` — remove the manual read-after-write call (envelope handles it via `readDetectionRule`)
- [x] 2.7 Extract `Update` method body to `func updateDetectionRule(ctx, *KibanaScopedClient, KibanaWriteRequest[Data]) (KibanaWriteResult[Data], diag.Diagnostics)`
- [x] 2.8 Extract `Delete` method body to `func deleteDetectionRule(ctx, *KibanaScopedClient, resourceID, spaceID string, model Data) diag.Diagnostics`
- [x] 2.9 Swap `*entitycore.ResourceBase` for `*entitycore.KibanaResource[Data]`; retain `UpgradeState` and passthrough `ImportState` on the wrapper struct
- [x] 2.10 Add `entitycore_contract_test.go` asserting the resource embeds `KibanaResource[Data]`
- [x] 2.11 Run `make build` and `go test ./internal/kibana/security_detection_rule/...`

## 3. kibana/alertingrule

- [x] 3.1 Add `GetID()`, `GetResourceID()` (→ `RuleID`), `GetSpaceID()` (→ `SpaceID`), `GetKibanaConnection()` value-receiver methods to `alertingRuleModel`
- [x] 3.2 Add `GetVersionRequirements()` to `alertingRuleModel`, emitting the following requirements conditional on model state (preserve the existing error messages verbatim):
  - `8.6.0` (`frequencyMinSupportedVersion`) when any action has `Frequency` set
  - `8.6.0` when `NotifyWhen` is null or empty (notify_when is required below 8.6)
  - `8.9.0` (`alertsFilterMinSupportedVersion`) when any action has `AlertsFilter` set
  - `8.13.0` (`alertDelayMinSupportedVersion`) when `AlertDelay` is set
  - `8.16.0` (`flappingMinSupportedVersion`) when `Flapping` is set
  - `9.3.0` (`flappingEnabledMinSupportedVersion`) when `Flapping.Enabled` is set
- [x] 3.3 Delete `features.go` (the `alertingRuleFeatures` struct, `alertingRuleFeaturesFromVersion`, `alertingRuleFeaturesAllSupported`, `resolveAlertingRuleFeatures`)
- [x] 3.4 Update `toAPIModel` to drop the `features alertingRuleFeatures` parameter; remove all `if !features.Supports*` branches (now enforced by the envelope) — the function body retains the field-mapping logic but loses the version-gating conditionals
- [x] 3.5 Update `create.go`/`update.go` (during their extraction to callbacks) to call `toAPIModel(ctx)` without `features`
- [x] 3.6 Remove `kibana_connection` block from `schema.go` schema function
- [x] 3.7 Extract `Create` method body to `func createAlertingRule(ctx, *KibanaScopedClient, KibanaWriteRequest[alertingRuleModel]) (KibanaWriteResult[alertingRuleModel], diag.Diagnostics)`
- [x] 3.8 Extract `Read` method body to `func readAlertingRule(ctx, *KibanaScopedClient, resourceID, spaceID string, model alertingRuleModel) (alertingRuleModel, bool, diag.Diagnostics)`
- [x] 3.9 Extract `Update` method body to `func updateAlertingRule(ctx, *KibanaScopedClient, KibanaWriteRequest[alertingRuleModel]) (KibanaWriteResult[alertingRuleModel], diag.Diagnostics)`
- [x] 3.10 Extract `Delete` method body to `func deleteAlertingRule(ctx, *KibanaScopedClient, resourceID, spaceID string, model alertingRuleModel) diag.Diagnostics`
- [x] 3.11 Swap `*entitycore.ResourceBase` for `*entitycore.KibanaResource[alertingRuleModel]`; retain `ValidateConfig`, `UpgradeState`, and composite `ImportState` on the wrapper struct
- [x] 3.12 Remove `getRuleIDAndSpaceID()` helper from `models.go` (replaced by envelope's `resolveKibanaResourceIdentity`)
- [x] 3.13 Add `entitycore_contract_test.go` asserting the resource embeds `KibanaResource[alertingRuleModel]`
- [x] 3.14 Run `make build` and `go test ./internal/kibana/alertingrule/...`

## 4. Final validation

- [ ] 4.1 Run `make build` across the full provider
- [ ] 4.2 Run `make lint` (or `make check-lint`) and fix any issues
- [ ] 4.3 Verify no schema changes: confirm `make docs` produces no diff for the three affected resources
