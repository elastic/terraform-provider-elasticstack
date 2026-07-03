## 1. Tool scaffolding

- [ ] 1.1 Create `scripts/targeted-testacc/` directory with `main.go` (package main, flag parsing, orchestration, stdout output)
- [ ] 1.2 Create `scripts/targeted-testacc/classifier.go` — maps changed file paths to Go package paths; detects force-all prefixes; treats testdata/* files as belonging to their nearest ancestor Go package
- [ ] 1.3 Create `scripts/targeted-testacc/depgraph.go` — builds forward import graph via `go list -f '{{.ImportPath}} {{join .Imports " "}}'`; exposes `BuildReverseDepGraph` and `WalkReverseDeps`
- [ ] 1.4 Create `scripts/targeted-testacc/entityname.go` — scans `.go` files in a directory for `NewResourceBase`, `NewElasticsearchResource`, `NewKibanaResource`, `NewKibanaDataSource` calls via regex; returns `[]EntityRef{Component, Name}` → full type string `elasticstack_<component>_<name>`
- [ ] 1.5 Create `scripts/targeted-testacc/testconsumers.go` — greps `internal/` recursively for an entity name string in `*.tf` and `*_test.go` files; maps matching file paths to their owning Go package paths
- [ ] 1.6 Create `scripts/targeted-testacc/acctestpackages.go` — walks `internal/` to enumerate all Go packages containing at least one `func TestAcc` in a `*_test.go` file; returns `[]string` of import paths
- [ ] 1.7 Create `scripts/targeted-testacc/selector.go` — accepts force-all check result, phase 1 packages, phase 2 packages, full acc-test package list, and thresholds; returns final sorted package list (or all packages if run-all triggered); exposes `ApplyShard` for shard-aware output
- [ ] 1.8 Create `scripts/targeted-testacc/gitdiff.go` — resolves diff baseline (flag → env → merge-base → HEAD~1 fallback); returns changed file paths via `git diff --name-only`

## 2. Tool unit tests

- [ ] 2.1 `classifier_test.go` — test file-to-package mapping for `.go` files, `testdata/*.tf` files, and non-Go files; test force-all prefix detection for all five prefixes
- [ ] 2.2 `entityname_test.go` — test regex extraction for all four call patterns (`NewResourceBase`, `NewElasticsearchResource`, `NewKibanaResource`, `NewKibanaDataSource`) from source snippets; test component string mapping to `elasticstack_<component>_<name>`
- [ ] 2.3 `depgraph_test.go` — test reverse dep walk on a synthetic graph (A imports B, B imports C → change C → reverse walk finds B and A)
- [ ] 2.4 `selector_test.go` — test force-all short-circuit; test run-all threshold (70%); test union/dedup of phase 1 + phase 2; test `ApplyShard` for all cases: `index >= total` → empty, `count < 30 && index > 0` → empty, `count < 30 && index == 0` → all, `count >= 30` → round-robin split
- [ ] 2.5 `acctestpackages_test.go` — test enumeration using a minimal synthetic directory tree with `*_test.go` files

## 3. Makefile targets

- [ ] 3.1 Add `TARGETED_TESTACC_BASE ?=` variable declaration
- [ ] 3.2 Add `targeted-testacc` target: invoke tool via `$(shell go run ./scripts/targeted-testacc/... ...)` capturing output; if empty print notice and exit 0; otherwise invoke `go tool gotestsum` with same flags as `testacc` and `--packages="$(TARGETED_PKGS)"`
- [ ] 3.3 Add `targeted-testacc-dry-run` target: invoke tool with `--dry-run`; no `TF_ACC` required; does not invoke `gotestsum`
- [ ] 3.4 Add `.PHONY` declarations for both new targets and add help comments

## 4. CI workflow update (`provider.yml`)

- [ ] 4.1 Add `compute-packages` step to the `test` job, positioned after `make vendor` and before `Pre-pull fleet image`; implement non-PR fast-path (`has_packages=true`, `targeted_pkgs=`) and PR path (git fetch + tool invocation + output setting)
- [ ] 4.2 Add `if: matrix.fleetImage && steps.targeted.outputs.has_packages == 'true'` to the `Pre-pull fleet image` step
- [ ] 4.3 Add `if: steps.targeted.outputs.has_packages == 'true'` to the `Start stack with docker compose` step
- [ ] 4.4 Add `if: steps.targeted.outputs.has_packages == 'true'` to the `Wait for stack readiness` step
- [ ] 4.5 Add `if: steps.targeted.outputs.has_packages == 'true'` to the `Get ES API key` step
- [ ] 4.6 Add `steps.targeted.outputs.has_packages == 'true'` condition to the `Setup Fleet` step (AND with existing version condition)
- [ ] 4.7 Add `steps.targeted.outputs.has_packages == 'true'` condition to the `Force install synthetics` step (AND with existing version condition)
- [ ] 4.8 Update the `TF acceptance tests` step: add `if: steps.targeted.outputs.has_packages == 'true'`; route between `make targeted-testacc TARGETED_PKGS=...` (when `targeted_pkgs` non-empty) and `make testacc` (when empty), both with `ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}`
- [ ] 4.9 Verify `Tear down docker compose stack` step retains `if: always()` and runs `make docker-clean` (no-op when stack was never started)

## 5. Validation

- [ ] 5.1 Run `make build` to confirm the tool compiles cleanly within the module
- [ ] 5.2 Run `go test ./scripts/targeted-testacc/...` to confirm all unit tests pass
- [ ] 5.3 Run `make targeted-testacc-dry-run` on a branch with at least one changed resource file; confirm output lists expected packages with rationale
- [ ] 5.4 Run `make targeted-testacc-dry-run ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=1` on a change with fewer than 30 selected packages; confirm empty output
- [ ] 5.5 Run `npx openspec validate --specs` to confirm specs are structurally valid
