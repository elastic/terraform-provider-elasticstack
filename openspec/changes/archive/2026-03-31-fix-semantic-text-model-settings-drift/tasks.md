## 1. Plan Modifier Fix

- [x] 1.1 Update `modifyMappings` in `mapping_modifier.go` to copy `model_settings` from state into the plan for `semantic_text` fields when absent from config
- [x] 1.2 Remove the overly-broad general key carry-forward loop and replace with the targeted `semantic_text` + `model_settings` check

## 2. Unit Tests

- [x] 2.1 Update test name in `mapping_modifier_test.go` to reflect `semantic_text`-specific behaviour
- [x] 2.2 Add test case: `semantic_text` field where config explicitly specifies `model_settings` (should not be overwritten by state)
- [x] 2.3 Add test case: nested `semantic_text` field (inside `properties` of another field) without `model_settings` in config

## 3. Spec Sync

- [ ] 3.1 Sync delta spec into `openspec/specs/elasticsearch-index/spec.md` (run `/opsx:archive` or apply delta manually)
