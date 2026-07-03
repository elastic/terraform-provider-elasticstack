# Task for worker

Fix the following issues in the existing `scripts/targeted-testacc/` scaffold. Read all files in that directory first, plus the spec at `openspec/changes/selective-acceptance-tests/specs/selective-acceptance-tests/spec.md`.

**Required fixes:**
1. **Empty/unresolvable diff → full suite.** In `main.go`, when `GitDiff(baseline)` returns an error or returns zero changed files, the tool must emit *all* acceptance test packages (subject to sharding). Currently it returns early with no output. Distinguish this from the "docs-only" case where `len(changedFiles) > 0` but `classified.HasCode` is false — that case should still emit zero packages.
2. **Filter phase-2 consumers to acceptance-test packages.** In `main.go`, intersect the results of `FindTestConsumers` with the `accSet` before adding them to `phase2Packages` / `phaseReasons`.
3. **Fix acceptance-test regex false positives.** In `acctestpackages.go`, change `testAccFuncRE` to `regexp.MustCompile("func\\s+TestAcc[A-Z]")` so it does not match functions like `TestAcceptanceServerInfo_*`.
4. **Skip unparseable Go files gracefully.** In `entityname.go`, change `extractFromFile` so that if `parser.ParseFile` returns an error it skips the file and returns an empty slice with a nil error (do not abort the whole run). Keep the `parser.AllErrors` flag.
5. **Avoid mutating caller slices in `uniqStrings` / `stringsSorted`.** In `depgraph.go`, make `stringsSorted` copy the slice before sorting; make `uniqStrings` copy before slicing in. Ensure callers still receive deduplicated sorted results unchanged.

**Optional low-priority:**
- Add a verbose-mode warning in `componentName` when an unknown component suffix is encountered, but do not change current behavior for known ones.

**After fixes, run:**
- `go vet ./scripts/targeted-testacc/...`
- `go build ./scripts/targeted-testacc/...`
- `go run ./scripts/targeted-testacc/... --base=HEAD --dry-run` (should now print all acc-test packages because the diff is empty)
- `go run ./scripts/targeted-testacc/... --base=HEAD~1 --dry-run` (report selected count; should be 0 or a small number and phase-2 consumers should only be acceptance-test packages)

Create one or more small focused git commits for these fixes. Do NOT push. Report back the files changed, commits created, and command results.

## Acceptance Contract
Acceptance level: checked
Completion is not accepted from prose alone. End with a structured acceptance report.

Criteria:
- criterion-1: Implement the requested change without widening scope

Required evidence: changed-files, tests-added, commands-run, residual-risks, no-staged-files

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