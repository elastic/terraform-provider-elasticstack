## 1. Implement the unidirectional update-decision helper

- [ ] 1.1 Add `RequiresMappingsUpdate(ctx context.Context, state MappingsValue) (bool, diag.Diagnostics)` as a method on `MappingsValue` in `internal/elasticsearch/index/mappings_value.go`, mirroring `StringSemanticEquals`'s null/unknown/decode handling but returning `!MappingsSemanticallyEqual(planMap, stateMap)` (receiver is plan, argument is state).
- [ ] 1.2 Document the argument-order contrast with `StringSemanticEquals` in the method's doc comment.

## 2. Wire the helper into both call sites

- [ ] 2.1 In `internal/elasticsearch/index/index/update.go`'s `updateMappings`, replace `planMappings.StringSemanticEquals(ctx, stateMappings)` with `planMappings.RequiresMappingsUpdate(ctx, stateMappings)` and invert the early-return branch.
- [ ] 2.2 Confirm `internal/elasticsearch/index/index/create.go`'s `adoptExistingIndexOnCreate` (which calls the same `updateMappings` helper) picks up the fix with no code change; no separate call site to update.

## 3. Unit tests

- [ ] 3.1 Add `TestMappingsValue_RequiresMappingsUpdate` in `internal/elasticsearch/index/mappings_value_test.go` covering: plan adds a field not in state (`true`), plan is a strict superset of state (`true`), state is a superset of plan / template-injected extras only (`false`), plan equals state (`false`), plan null/unknown (`false`), state null/unknown with non-null plan (`true`).

## 4. Acceptance tests

- [ ] 4.1 Add an acceptance test in `internal/elasticsearch/index/index/acc_test.go` that creates an index, then in a second step adds a field to `mappings` and applies, asserting via a direct read (or `TestCheckResourceAttrWith` against the refreshed `mappings` state) that the new field is present after apply — the regression test for `elastic/protections-cloud#19769`.
- [ ] 4.2 Extend `use_existing` adoption acceptance coverage (near `TestAccResourceIndexUseExistingAdopt`) with a config that both adds a field absent from the live mapping and omits a field present only in the live mapping, asserting the added field is written and the omitted field is retained.

## 5. Spec sync

- [ ] 5.1 Update `openspec/specs/elasticsearch-index/spec.md`'s Update flow (REQ-015–REQ-018) requirement text and add a scenario for the unidirectional Put Mapping decision.
- [ ] 5.2 Update the `Opt-in adoption of existing indices via use_existing` requirement to describe the same unidirectional decision during adoption, with a scenario covering the mixed add/omit case.
- [ ] 5.3 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fix-index-mapping-update-unidirectional-check --type change` and resolve any reported problems.

## 6. Validate

- [ ] 6.1 Run `go build ./...` and `go vet ./...`.
- [ ] 6.2 Run `go test ./internal/elasticsearch/index/...` (unit tests only; acceptance tests require `TF_ACC=1` and a running Elastic Stack — see `dev-docs/high-level/testing.md`).
