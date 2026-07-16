## 1. State upgrader fix

- [ ] 1.1 In `internal/elasticsearch/index/ilm/state_upgrade.go`, inside `migrateILMStateV0ToV1`, add `stateutil.NullifyEmptyString(stateMap, "metadata")` after the phase-unwrapping loop and before `stateutil.MarshalStateMap`.
- [ ] 1.2 In `internal/elasticsearch/index/ilm/state_upgrade.go`, inside the phase-unwrapping loop, after `unwrapPhaseActionLists(phaseObj)`, add a block that checks for an `allocate` action object and calls `stateutil.NullifyEmptyString(allocateObj, attrInclude, attrExclude, attrRequire)`.

## 2. Spec update

- [ ] 2.1 In `openspec/specs/elasticsearch-index-lifecycle/spec.md`, extend REQ-030–REQ-031 to add a requirement and scenario specifying that the v0→v1 state upgrader SHALL call `stateutil.NullifyEmptyString` for `metadata` (top-level) and for `allocate.include`, `allocate.exclude`, `allocate.require` within each phase block that contains an `allocate` action.

## 3. Unit tests

- [ ] 3.1 In `internal/elasticsearch/index/ilm/state_upgrade_test.go`, add a `metadata_empty_string` test case: top-level `"metadata": ""`, no phases; assert `metadata == nil` in upgraded state and that upgraded state decodes against v1 schema without error.
- [ ] 3.2 In `internal/elasticsearch/index/ilm/state_upgrade_test.go`, add an `allocate_include_empty_string` test case: a warm phase with an `allocate` block where `"include": ""`; assert `warm.allocate.include == nil` after upgrade.
- [ ] 3.3 In `internal/elasticsearch/index/ilm/state_upgrade_test.go`, add an `allocate_all_json_attrs_empty_string` test case: warm phase `allocate` with `"include": ""`, `"exclude": ""`, `"require": ""`; assert all three become `nil` after upgrade.
- [ ] 3.4 In `internal/elasticsearch/index/ilm/state_upgrade_test.go`, add a `metadata_and_allocate_empty_strings` test case combining `"metadata": ""` at top level and a cold phase `allocate` with all three JSON attributes empty; assert all four fields become `nil`.

## 4. Acceptance test

- [ ] 4.1 In `internal/elasticsearch/index/ilm/` (or the appropriate test file), add `TestAccResourceILMFromSDKNoMetadata` covering: SDK 0.14.5 creation of an ILM policy with a warm phase `allocate` block and no `metadata`/`include`/`exclude`/`require`; provider upgrade to current Plugin Framework provider; assertion that `terraform plan` produces no diff.
- [ ] 4.2 Add the Terraform configuration for `TestAccResourceILMFromSDKNoMetadata` under `internal/elasticsearch/index/ilm/testdata/TestAccResourceILMFromSDKNoMetadata/`, mirroring the layout of existing ILM testdata directories.

## 5. Validation

- [ ] 5.1 Run focused unit tests: `go test ./internal/elasticsearch/index/ilm/ -run TestMigrateILMState` (adjust pattern to match the actual test function name).
- [ ] 5.2 Run `make build` and confirm the provider compiles without errors.
- [ ] 5.3 Run `make lint` and fix any issues.
- [ ] 5.4 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate ilm-state-upgrade-nullify-empty-json-fields --type change` and fix any reported problems.
- [ ] 5.5 Run the acceptance test `TestAccResourceILMFromSDKNoMetadata` with `TF_ACC=1` against a live Elastic Stack if available.
