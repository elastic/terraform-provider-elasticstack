## 1. Tool scaffolding

- [x] 1.1 Create `scripts/targeted-testacc/` directory with `main.go` (package main, flag parsing, orchestration, stdout output)
- [x] 1.2 Create `scripts/targeted-testacc/classifier.go` — maps changed file paths to Go package paths; detects force-all prefixes; attributes every changed file (testdata fixtures, go:embed'ed description files, fixture directories not named exactly testdata) to its nearest ancestor Go package
- [x] 1.3 Create `scripts/targeted-testacc/depgraph.go` — builds forward import graph via `go list -f '{{.ImportPath}} {{join .Imports " "}}'`; exposes `BuildReverseDepGraph` and `WalkReverseDeps`
- [x] 1.4 Create `scripts/targeted-testacc/entityname.go` — scans `.go` files in a directory for calls to every entity-declaring constructor exported by `entitycore` (`NewResourceBase`, `NewDataSourceBase`, `NewEphemeralBase`, `NewActionBase`, `NewElasticsearchResource`, `NewElasticsearchDataSource`, `NewElasticsearchEphemeralResource`, `NewElasticsearchAction`, `NewKibanaResource`, `NewKibanaDataSource`, `NewKibanaEphemeralResource`, `NewKibanaAction`) via regex; returns `[]EntityRef{Component, Name}` → full type string `elasticstack_<component>_<name>`
- [x] 1.5 Create `scripts/targeted-testacc/testconsumers.go` — greps the enumeration roots (`internal/`, `provider/`) recursively for an entity name string in `*.tf` and `*_test.go` files; maps matching file paths to their owning Go package paths
- [x] 1.6 Create `scripts/targeted-testacc/acctestpackages.go` — walks the enumeration roots (`internal/`, `provider/`) to enumerate all Go packages containing at least one acceptance-test file (`*_test.go` declaring `func TestAcc*` or invoking `resource.Test`/`resource.ParallelTest`, honouring import aliases); returns `[]string` of import paths
- [x] 1.7 Create `scripts/targeted-testacc/selector.go` — accepts force-all check result, phase 1 packages, phase 2 packages, full acc-test package list, and thresholds; returns final sorted package list (or all packages if run-all triggered); exposes `ApplyShard` for shard-aware output
- [x] 1.8 Create `scripts/targeted-testacc/gitdiff.go` — resolves diff baseline (flag → env → merge-base → HEAD~1 fallback); returns changed file paths via `git diff --name-only`

## 2. Tool unit tests

- [x] 2.1 `classifier_test.go` — test file-to-package mapping for `.go` files, `testdata/*.tf` files, embedded non-Go package files, and non-Go files outside any Go package; test force-all prefix detection for all force-all prefixes
- [x] 2.2 `entityname_test.go` — test regex extraction for every entity-declaring constructor from source snippets; test component string mapping to `elasticstack_<component>_<name>`; guard test that every exported `entitycore` constructor is classified so future envelope types fail the test
- [x] 2.3 `depgraph_test.go` — test reverse dep walk on a synthetic graph (A imports B, B imports C → change C → reverse walk finds B and A)
- [x] 2.4 `selector_test.go` — test force-all short-circuit; test run-all threshold (70%); test union/dedup of phase 1 + phase 2; test `ApplyShard` for all cases: `index >= total` → empty, `count < 30 && index > 0` → empty, `count < 30 && index == 0` → all, `count >= 30` → round-robin split
- [x] 2.5 `acctestpackages_test.go` — test enumeration using a minimal synthetic directory tree with `*_test.go` files
- [x] 2.6 `testonlyimports_test.go` — guard test that fails when a package under `internal/` is imported only from test files and is neither force-all nor entity-declaring

## 3. Makefile targets

- [x] 3.1 Add `TARGETED_TESTACC_BASE ?=` and `TARGETED_TESTACC_VERBOSE ?= 0` variable declarations
- [x] 3.2 Add `targeted-testacc` target: invoke tool via `$(shell go run ./scripts/targeted-testacc/... ...)` capturing output; if empty print notice and exit 0; otherwise invoke `go tool gotestsum` with same flags as `testacc` and `--packages="$(TARGETED_PKGS)"`
- [x] 3.3 Add `targeted-testacc-dry-run` target: invoke tool with `--dry-run`; no `TF_ACC` required; does not invoke `gotestsum`
- [x] 3.4 Add `.PHONY` declarations for both new targets and add help comments

## 4. CI workflow update (`provider.yml`)

- [x] 4.0 Add `merge_group:` trigger to the `on:` section of `provider.yml` (enables future merge-queue support)
- [x] 4.1 Add `compute-packages` step to the `test` job, positioned after `make vendor` and before `Pre-pull fleet image`; implement non-PR fast-path (`has_packages=true`, `targeted_pkgs=`) and PR path (git fetch + tool invocation + output setting)
- [x] 4.2 Add `if: matrix.fleetImage && steps.targeted.outputs.has_packages == 'true'` to the `Pre-pull fleet image` step
- [x] 4.3 Add `if: steps.targeted.outputs.has_packages == 'true'` to the `Start stack with docker compose` step
- [x] 4.4 Add `if: steps.targeted.outputs.has_packages == 'true'` to the `Wait for stack readiness` step
- [x] 4.5 Add `if: steps.targeted.outputs.has_packages == 'true'` to the `Get ES API key` step
- [x] 4.6 Add `steps.targeted.outputs.has_packages == 'true'` condition to the `Force install synthetics` step (AND with existing version condition)
- [x] 4.7 Update the `TF acceptance tests` step: add `if: steps.targeted.outputs.has_packages == 'true'`; route between `make targeted-testacc TARGETED_PKGS=...` (when `targeted_pkgs` non-empty; no re-sharding — the tool already applied `--total-shards`/`--shard-index` during package selection, so the `TARGETED_PKGS` list is final for this shard) and `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}` (when empty)
- [x] 4.8 Verify `Tear down docker compose stack` step retains `if: always()` and runs `make docker-clean` (no-op when stack was never started)

## 5. Validation

- [x] 5.1 Run `make build` to confirm the tool compiles cleanly within the module
- [x] 5.2 Run `go test ./scripts/targeted-testacc/...` to confirm all unit tests pass
- [x] 5.3 Run `make targeted-testacc-dry-run` on a branch with at least one changed resource file; confirm output lists expected packages with rationale
- [x] 5.4 Run `make targeted-testacc-dry-run ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=1` on a change with fewer than 30 selected packages; confirm empty output
- [x] 5.5 Run `npx openspec validate --specs` to confirm specs are structurally valid

## 6. Unit-test job and gate wiring

- [x] 6.1 Add a dedicated `unit-test` job to `provider.yml` running `go test ./... -skip '^TestAcc'` stackless, gated on `provider_changes` (not `has_packages`)
- [x] 6.2 Add `unit-test` to the `gate` job `needs` and consume the `PROVIDER_GATE_UNIT_TEST_RESULT` workflow output in the gate decision
- [x] 6.3 Thread `unitTestResult` through `.github/scripts/workflows/lib/gate-provider.js` and `runners/gate.js`
- [x] 6.4 Extend `gate-provider.test.mjs` with tests covering the new `unitTestResult` dimension (failure gates, unexpected skip gates, success passes)
