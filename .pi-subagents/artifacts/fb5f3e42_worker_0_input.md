# Task for worker

Makefile review task. In `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests`, the Makefile has a new `targeted-testacc` target that uses `$(eval TARGETED_PKGS := $(shell TARGETED_TESTACC_BASE="$(TARGETED_TESTACC_BASE)" go run ./scripts/targeted-testacc/... --total-shards=$(ACCTEST_TOTAL_SHARDS) --shard-index=$(ACCTEST_SHARD_INDEX) --verbose=$(TARGETED_TESTACC_VERBOSE)))` then checks if `$(TARGETED_PKGS)` is empty. When the tool has verbose enabled, it prints to stderr. There is a risk that if the tool prints anything to stdout in verbose mode or if the shell line contains extra whitespace, the empty check fails or the package list contains noise. Please analyze this exact Makefile snippet and recommend one robust pattern to capture only the tool's stdout package list while preserving stderr output. Also verify whether `--verbose=$(TARGETED_TESTACC_VERBOSE)` with value 0 or 1 is accepted by the Go tool's flag parser (true/false boolean flags accept 0/1/ true/false in Go's flag package). Do not edit files; just return your recommendation.

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