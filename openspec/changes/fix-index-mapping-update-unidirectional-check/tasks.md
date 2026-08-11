## 1. Slice A — Add helper and unit tests

- [x] 1.1 Add `RequiresMappingsUpdate(ctx context.Context, state MappingsValue) (bool, diag.Diagnostics)` as a method on `MappingsValue` in `internal/elasticsearch/index/mappings_value.go`, mirroring `StringSemanticEquals`'s null/unknown/decode handling but returning `!MappingsSemanticallyEqual(planMap, stateMap)` (receiver is plan, argument is state).
- [x] 1.2 Document the argument-order contrast with `StringSemanticEquals` in the method's doc comment.
- [x] 1.3 Add `TestMappingsValue_RequiresMappingsUpdate` in `internal/elasticsearch/index/mappings_value_test.go` covering: plan adds a field not in state (`true`), plan is a strict superset of state (`true`), state is a superset of plan / template-injected extras only (`false`), plan equals state (`false`), plan null/unknown (`false`), state null/unknown with non-null plan (`true`).
- [x] 1.4 Run `go test ./internal/elasticsearch/index/...` and confirm all unit tests pass.

## 2. Slice B — Wire helper and acceptance tests

- [x] 2.1 In `internal/elasticsearch/index/index/update.go`'s `updateMappings`, replace `planMappings.StringSemanticEquals(ctx, stateMappings)` with `planMappings.RequiresMappingsUpdate(ctx, stateMappings)` and invert the early-return branch.
- [x] 2.2 Confirm `internal/elasticsearch/index/index/create.go`'s `adoptExistingIndexOnCreate` (which calls the same `updateMappings` helper) picks up the fix with no code change; no separate call site to update.
- [x] 2.3 Add an acceptance test in `internal/elasticsearch/index/index/acc_test.go` that creates an index, then in a second step adds a field to `mappings` and applies, asserting via a direct Elasticsearch read that the new field is present in the live cluster after apply — the regression test for `elastic/protections-cloud#19769`. Do not rely only on the refreshed Terraform `mappings` state, which can preserve the planned field even when the API update was skipped.
- [x] 2.4 Extend `use_existing` adoption acceptance coverage (near `TestAccResourceIndexUseExistingAdopt`) with a config that both adds a field absent from the live mapping and omits a field present only in the live mapping, asserting the added field is written and the omitted field is retained.
- [ ] 2.5 Run `TF_ACC=1 go test ./internal/elasticsearch/index/index/...` against a live Elastic Stack (see `dev-docs/high-level/testing.md`) and confirm the regression test and adoption test pass.

## 3. Spec sync

- [x] 3.1 Update `openspec/specs/elasticsearch-index/spec.md`'s Update flow (REQ-015–REQ-018) requirement text and add a scenario for the unidirectional Put Mapping decision.
- [x] 3.2 Update the `Opt-in adoption of existing indices via use_existing` requirement to describe the same unidirectional decision during adoption, with a scenario covering the mixed add/omit case.
- [x] 3.3 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fix-index-mapping-update-unidirectional-check --type change` and resolve any reported problems.

## 4. Validate

- [x] 4.1 Run `go build ./...` and `go vet ./...`.
