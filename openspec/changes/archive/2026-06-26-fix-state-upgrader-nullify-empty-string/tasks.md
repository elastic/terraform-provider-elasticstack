## 1. State upgrader fixes

- [x] 1.1 In `internal/elasticsearch/index/componenttemplate/state_upgrade.go`, update `migrateComponentTemplateStateV0ToV1` so the `if tmpl, ok := stateMap[attrTemplate].(map[string]any); ok` block calls `stateutil.NullifyEmptyString(tmpl, attrMappings, attrSettings)` immediately after `stateutil.EnsureMapKeys(tmpl, attrAlias, attrMappings, attrSettings, attrDataStreamOptions)` and before `aliasutil.NormalizeTemplateAliasesInV1State(tmpl)`.
- [x] 1.2 In `internal/elasticsearch/index/componenttemplate/state_upgrade.go`, call `stateutil.NullifyEmptyString(stateMap, "metadata")` after the `template` block and before the following `version` cleanup block.
- [x] 1.3 In `internal/elasticsearch/index/template/state_upgrade.go`, update `migrateIndexTemplateStateV0ToV1` so the second `if tmpl, ok := stateMap[attrTemplate].(map[string]any); ok` block (after all `CollapseListPath` calls) calls `stateutil.NullifyEmptyString(tmpl, attrMappings, attrSettings)` immediately after `stateutil.EnsureMapKeys`.
- [x] 1.4 In `internal/elasticsearch/index/template/state_upgrade.go`, call `stateutil.NullifyEmptyString(stateMap, "metadata")` after the `template` block.

## 2. Unit tests

- [x] 2.1 In `internal/elasticsearch/index/componenttemplate/state_upgrade_test.go`, add a `settings_only_empty_string_mappings` upgrade test case where `template` contains `{"mappings": "", "settings": "{\"index\":{\"number_of_replicas\":\"1\"}}"}`; assert `template.mappings == nil`, `template.settings` is preserved, and `requireUpgradedStateDecodes` passes.
- [x] 2.2 In `internal/elasticsearch/index/componenttemplate/state_upgrade_test.go`, add a `mappings_only_empty_string_settings` upgrade test case where `template` contains `{"mappings": "{\"properties\":{\"field\":{\"type\":\"keyword\"}}}", "settings": ""}`; assert `template.settings == nil`, `template.mappings` is preserved, and `requireUpgradedStateDecodes` passes.
- [x] 2.3 In `internal/elasticsearch/index/componenttemplate/state_upgrade_test.go`, add a `metadata_empty_string` upgrade test case with top-level `"metadata": ""` and no `template` block; assert `metadata == nil` and `requireUpgradedStateDecodes` passes.
- [x] 2.4 In `internal/elasticsearch/index/template/state_upgrade_test.go`, mirror the three component template cases for `migrateIndexTemplateStateV0ToV1`, using the index template state shape and including `composed_of` and `index_patterns` in base state where required.

## 3. Acceptance tests

- [x] 3.1 In `internal/elasticsearch/index/componenttemplate/acc_from_sdk_test.go`, add `TestAccResourceComponentTemplateFromSDKSettingsOnly` covering SDK 0.14.5 creation of a component template with only `settings`, Plugin Framework re-apply with `template.settings` set and `template.mappings` empty/null, and a final no-op plan (`PlanOnly: true`, `ExpectNonEmptyPlan: false`).
- [x] 3.2 Add Terraform configuration for `TestAccResourceComponentTemplateFromSDKSettingsOnly` under `internal/elasticsearch/index/componenttemplate/testdata/TestAccResourceComponentTemplateFromSDKSettingsOnly/config/`, mirroring the existing `TestAccResourceComponentTemplateFromSDK` test data layout.
- [x] 3.3 In `internal/elasticsearch/index/template/acc_from_sdk_test.go`, add `TestAccResourceIndexTemplateFromSDKSettingsOnly` covering SDK 0.14.5 creation of an index template with `index_patterns` and only `settings`, Plugin Framework re-apply without error, and a final no-op plan.
- [x] 3.4 Add Terraform configuration for `TestAccResourceIndexTemplateFromSDKSettingsOnly` under `internal/elasticsearch/index/template/testdata/TestAccResourceIndexTemplateFromSDKSettingsOnly/`.

## 4. Validation

- [x] 4.1 Run focused unit tests for `internal/elasticsearch/index/componenttemplate` and `internal/elasticsearch/index/template`.
- [ ] 4.2 Run targeted acceptance tests for `TestAccResourceComponentTemplateFromSDKSettingsOnly` and `TestAccResourceIndexTemplateFromSDKSettingsOnly` with `TF_ACC=1` against an available Elastic Stack.
- [x] 4.3 Run `make lint`, `make build`, and `make check-openspec`; fix any issues.
