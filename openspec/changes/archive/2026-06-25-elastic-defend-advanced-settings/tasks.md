## 1. Schema and models

- [x] 1.1 Add optional `advanced_settings` `MapAttribute` (`ElementType: String`) to `schema.go` with description linking to Elastic Defend advanced settings docs
- [x] 1.2 Add `AdvancedSettings types.Map` field to `elasticDefendIntegrationPolicyModel` in `models.go`
- [x] 1.3 Add key-prefix validation (must match `^(linux|mac|windows)\.advanced\.`) via map key validator or apply-time check

## 2. Advanced settings mapping helpers

- [x] 2.1 Implement `advancedSettingsToNested(os, keys)` to convert dot-notation keys into nested `advanced` maps
- [x] 2.2 Implement `nestedAdvancedToSettings(os, advanced)` to flatten API `advanced` objects back to dot-notation keys
- [x] 2.3 Add unit tests in `advanced_settings_test.go` for round-trip, multi-OS merge, invalid keys, and deep paths (e.g. `artifacts.global.base_url`)

## 3. Request and response integration

- [x] 3.1 Extend `buildPolicyPayload` to merge `advanced_settings` into each OS block's `advanced` subtree when attribute is set
- [x] 3.2 Extend `populateModelFromAPI` / `mapPolicyFromAPI` to populate `advanced_settings` when originally configured in state
- [x] 3.3 Ensure finalize and update paths (`buildFinalizeInputConfig`) include advanced settings without affecting `artifact_manifest` or `version` handling
- [x] 3.4 Update `mapping_test.go` for request bodies containing advanced settings under `policy.{os}.advanced`

## 4. Documentation

- [x] 4.1 Regenerate `docs/resources/fleet_elastic_defend_integration_policy.md` with `advanced_settings` attribute and air-gapped example
- [x] 4.2 Add HCL example showing `linux.advanced.artifacts.global.base_url` (and optionally mac/windows counterparts)

## 5. Acceptance tests

- [x] 5.1 Add `TestAccResourceElasticDefendIntegrationPolicy_advancedSettings` with create asserting artifact base URL, update changing value, and read round-trip
- [x] 5.2 Add testdata fixtures under `testdata/TestAccResourceElasticDefendIntegrationPolicy_advancedSettings/`

## 6. Validation

- [x] 6.1 Run `make build` and targeted unit tests for `internal/fleet/elastic_defend_integration_policy`
- [x] 6.2 Run acceptance test when Elastic Stack is available
- [x] 6.3 Run `make check-openspec` (or `openspec validate`) for this change
