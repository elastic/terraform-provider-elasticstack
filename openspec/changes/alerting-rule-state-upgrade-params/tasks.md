## 1. Implementation

- [ ] 1.1 In `internal/kibana/alertingrule/state_upgrade.go`, inside `migrateV0ToV1`,
  add `stateutil.NullifyEmptyString(stateMap, "params")` immediately after the
  existing `stateutil.NullifyEmptyString(stateMap, "notify_when", "throttle")` call.

- [ ] 1.2 In the same function, inside the `actions` loop (the block that iterates
  `stateMap["actions"].([]any)` and casts each element to `map[string]any`), add
  `stateutil.NullifyEmptyString(action, "params")` as the first statement inside
  the `if action, ok := ...` block, before the `CollapseListPath` calls.

## 2. Testing

- [ ] 2.1 Add unit tests for `migrateV0ToV1` in
  `internal/kibana/alertingrule/state_upgrade_test.go` (create the file if it does
  not exist) covering the following cases:
  - Rule-level `params` is `""` → upgraded state has `params` = `null`.
  - Rule-level `params` is `null` → upgraded state has `params` = `null` (unchanged).
  - Rule-level `params` is a valid JSON string (e.g. `"{}"`) → upgraded state
    preserves the value unchanged.
  - Action-level `params` is `""` → upgraded state has action `params` = `null`.
  - Action-level `params` is a valid JSON string → upgraded state preserves it.
  - Both rule-level and action-level `params` are `""` simultaneously → both are
    nullified.

## 3. Spec

- [ ] 3.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate alerting-rule-state-upgrade-params --type change` and resolve any reported problems.
- [ ] 3.2 When implementation is complete, sync the delta spec into
  `openspec/specs/kibana-alerting-rule/spec.md` or archive the change per the
  project workflow.
