## 1. Flatten-path fix

- [x] 1.1 Generalize `priorHasDeclaredToggle` (or add a sibling `priorHasDeclaredAction` helper) in `internal/elasticsearch/index/ilm/flatten.go` so the same prior-state-declaration check can be applied to the `rollover` action name.
- [x] 1.2 Add an explicit `case ilmActionRollover:` in `flattenPhase`'s action switch: when the returned action is empty (`len(action) == 0`) and the prior state does not have a declared non-null `rollover` object for this phase, skip writing `rollover` into `phase` (treat it as absent). Otherwise fall through to writing `phase[actionName] = []any{action}` as today.
- [x] 1.3 Confirm no change is needed in `expand.go`'s `ilmActionRollover` case (write path already omits rollover correctly when unconfigured). `objectToExpandMap` skips null/unknown attributes, so an undeclared `rollover` never reaches `expandPhase`; no expand.go change.

## 2. Unit tests

- [x] 2.1 Add a unit test asserting that a hot phase whose actions map contains `"rollover": {}` and whose prior state has `rollover = null` flattens to a null `rollover` in the resulting object.
- [x] 2.2 Add a unit test asserting that a hot phase whose actions map contains `"rollover": {}` but whose prior state already had a non-null `rollover` object preserves that object (round-trip stability for explicitly-empty user-declared rollover).
- [x] 2.3 Add/confirm a unit test that a hot phase with a populated `rollover` action (real conditions) continues to flatten those conditions unchanged.

## 3. Acceptance test

- [x] 3.1 Add an acceptance test (e.g. `TestAccResourceILMHotPhaseWithoutRollover`) that creates an `elasticstack_elasticsearch_index_lifecycle` policy with a `hot` phase containing no `rollover` block (only, e.g., `set_priority`), verifies that the live Elasticsearch GET response contains an empty `"rollover": {}` action, and asserts `terraform plan` after apply/refresh shows no diff. Use a supported test setup/version that produces this response so the normalization branch is exercised.
- [x] 3.2 Confirm the test also covers a subsequent `terraform plan` with no configuration changes to catch regressions of the perpetual-diff bug.

## 4. Spec and validation

- [ ] 4.1 Add the new requirement to `openspec/specs/elasticsearch-index-lifecycle/spec.md` describing the rollover-omission normalization (done as part of this change's delta spec; apply on archive).
- [ ] 4.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate ilm-hot-phase-rollover-omission --type change` and resolve any reported issues.
- [ ] 4.3 During implementation, run `make build`, `go vet ./internal/elasticsearch/index/ilm/...`, and `go test ./internal/elasticsearch/index/ilm/...` (unit tests); run the new acceptance test with `TF_ACC=1` against a running Elastic Stack per `dev-docs/high-level/testing.md`.
