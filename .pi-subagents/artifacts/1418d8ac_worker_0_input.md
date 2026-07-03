# Task for worker

Implement top-level task 1 of OpenSpec change `selective-acceptance-tests`: build the Go tool scaffold under `scripts/targeted-testacc/`.

**Scope (do not implement task 2 tests yet, do not edit Makefile/CI):**
Complete subtasks 1.1 through 1.8:
- 1.1 `scripts/targeted-testacc/main.go` — package main with CLI flags (base, total-shards, shard-index, dry-run, verbose, run-all-threshold, min-shard-packages), orchestration wiring, stdout output.
- 1.2 `scripts/targeted-testacc/classifier.go` — changed file path → Go package path; detect force-all prefixes (`provider/`, `internal/acctest/`, `internal/clients/`, `internal/entitycore/`, `generated/`); treat `testdata/` files as belonging to nearest ancestor Go package; only include paths inside Go packages.
- 1.3 `scripts/targeted-testacc/depgraph.go` — build forward import graph via `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./internal/...` (or `go list ./internal/...` via Go tooling); expose `BuildReverseDepGraph` and `WalkReverseDeps` returning all transitive reverse importers of a set of import paths, using non-test imports only.
- 1.4 `scripts/targeted-testacc/entityname.go` — scan `.go` files in package dirs for calls matching `NewResourceBase`, `NewElasticsearchResource`, `NewKibanaResource`, `NewKibanaDataSource` and return `[]EntityRef{Component, Name}`. Convert to full entity string `elasticstack_<component>_<name>`.
- 1.5 `scripts/targeted-testacc/testconsumers.go` — search `internal/` recursively for an entity name string in `*.tf` and `*_test.go` files; map each matching file path to its owning Go package import path.
- 1.6 `scripts/targeted-testacc/acctestpackages.go` — enumerate all Go packages under `internal/` that contain at least one `func TestAcc` in a `*_test.go` file; return `[]string` of import paths.
- 1.7 `scripts/targeted-testacc/selector.go` — accept force-all result, phase1 packages, phase2 packages, full acc-test package list and thresholds; return final sorted package list, or all packages if force-all or run-all threshold exceeded; expose `ApplyShard` per the shard-aware requirements.
- 1.8 `scripts/targeted-testacc/gitdiff.go` — resolve diff baseline in order: `--base` flag → `TARGETED_TESTACC_BASE` env → `git merge-base origin/main HEAD` → `HEAD~1`; return changed file paths via `git diff --name-only`.

**Requirements to follow:**
- Follow the OpenSpec spec for `selective-acceptance-tests` in `openspec/changes/selective-acceptance-tests/specs/selective-acceptance-tests/spec.md`.
- Use only stdlib + existing module dependencies (golang.org/x/tools v0.45.0 is in go.mod; prefer `golang.org/x/tools/go/packages` or `go list` via exec for import graph as you see fit).
- Keep functions unit-testable; avoid global state; expose pure helpers.
- Do not use a separate Go module; place files directly in `scripts/targeted-testacc/`.
- The tool `go run ./scripts/targeted-testacc/...` must compile.
- Empty diff or only non-Go/non-testdata files should result in zero selected packages (per spec), not "all packages" — verify this matches spec section "Only docs files changed produces no packages".
- For force-all prefixes and empty diff: return all acc-test packages only when a force-all prefix is matched; empty diff returns zero packages (because no changed Go/testdata files). Be careful with spec wording. The design "Empty diff (on main, or when only non-code files changed) → tool emits all acc-test packages (conservative default)" but the spec says "Only docs files changed produces no packages". Since the spec is authoritative, implement so that only when changed files exist and none are Go/testdata → zero packages; when git diff truly empty, also zero packages. (This will be validated with tests/dry-run.)
- Error handling: log errors to stderr, exit nonzero only on unexpected failures (e.g., `go list` failure), not for normal zero-package result.
- Do NOT push to remote. Create small focused git commits as files are completed.
- After implementing, run `go build ./scripts/targeted-testacc/...` to verify compilation.

**Context files to read before editing:**
- `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests/openspec/changes/selective-acceptance-tests/proposal.md`
- `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests/openspec/changes/selective-acceptance-tests/design.md`
- `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests/openspec/changes/selective-acceptance-tests/specs/selective-acceptance-tests/spec.md`
- `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests/openspec/changes/selective-acceptance-tests/tasks.md`

Report back:
- files created and main exported APIs
- git commits created
- result of `go build ./scripts/targeted-testacc/...`
- blockers or questions

## Acceptance Contract
Acceptance level: attested
Completion is not accepted from prose alone. End with a structured acceptance report.

Criteria:
- criterion-1: Return concrete findings with file paths and severity when applicable

Required evidence: review-findings, residual-risks

Finish with a fenced JSON block tagged `acceptance-report` in this shape:
Use empty arrays when no items apply; array fields contain strings unless object entries are shown.
```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "specific proof"
    }
  ],
  "changedFiles": [
    "src/file.ts"
  ],
  "testsAddedOrUpdated": [
    "test/file.test.ts"
  ],
  "commandsRun": [
    {
      "command": "command",
      "result": "passed",
      "summary": "short result"
    }
  ],
  "validationOutput": [
    "validation output or concise summary"
  ],
  "residualRisks": [
    "none"
  ],
  "noStagedFiles": true,
  "diffSummary": "short description of the diff",
  "reviewFindings": [
    "blocker: file.ts:12 - issue found, or no blockers"
  ],
  "manualNotes": "anything else the parent should know"
}
```