---
imports: [shared/setup-dev.md]
name: Dead-code Removal Rotation
description: >-
  Scheduled dead-code cleanup workflow that rotates through highest-confidence
  deadcode candidates (dead with and without tests), performs deterministic
  pre-activation selection, reference classification, and opens at most one
  verified cleanup PR per run.
on:
  workflow_dispatch:
  schedule:
    - cron: daily
  steps:
    - name: Checkout repository
      uses: actions/checkout@v7.0.1
      with:
        persist-credentials: false
        fetch-depth: 1
    - name: Setup Go
      uses: actions/setup-go@v7.0.0
      with:
        go-version-file: go.mod
        cache: false
    # NOTE: This ref must match the repo-memory tool config branch-name below.
    - name: Checkout repo-memory branch
      uses: actions/checkout@v7.0.1
      with:
        ref: memory/ci-deadcode-removal-rotation
        path: gh-aw-repo-memory/ci-deadcode-removal-rotation
        fetch-depth: 1
        persist-credentials: false
      continue-on-error: true
    - name: Export Go paths
      run: |
        echo "GOROOT=$(go env GOROOT)" >> "$GITHUB_ENV"
        echo "GOPATH=$(go env GOPATH)" >> "$GITHUB_ENV"
        echo "GOMODCACHE=$(go env GOMODCACHE)" >> "$GITHUB_ENV"
    - name: Ensure gopls is available
      run: |
        if ! command -v gopls &> /dev/null; then
          go install golang.org/x/tools/gopls@latest
        fi
    - name: Pre-download modules
      run: |
        set -euo pipefail
        echo "go version: $(go version)"
        echo "GOROOT=$(go env GOROOT)"
        echo "GOPATH=$(go env GOPATH)"
        echo "GOMODCACHE=$(go env GOMODCACHE)"
        echo "GOPROXY=$(go env GOPROXY)"
        echo "GOTOOLCHAIN=$(go env GOTOOLCHAIN)"
        time go mod download
    - name: Select dead-code candidate
      id: select_candidate
      run: |
        set -euo pipefail
        result=$(go run ./scripts/ci-deadcode-removal-rotation select \
          --memory gh-aw-repo-memory/ci-deadcode-removal-rotation/memory/ci-deadcode-removal-rotation/memory.json \
          --cooldown-days 14)
        echo "$result"
        found=$(echo "$result" | jq -r '.found')
        echo "found=$found" >> "$GITHUB_OUTPUT"
        if [ "$found" = "true" ]; then
          echo "symbol=$(echo "$result" | jq -r '.symbol')" >> "$GITHUB_OUTPUT"
          echo "symbol_name=$(echo "$result" | jq -r '.symbol_name')" >> "$GITHUB_OUTPUT"
          echo "package=$(echo "$result" | jq -r '.package')" >> "$GITHUB_OUTPUT"
          echo "file=$(echo "$result" | jq -r '.file')" >> "$GITHUB_OUTPUT"
          echo "line=$(echo "$result" | jq -r '.line')" >> "$GITHUB_OUTPUT"
          echo "column=$(echo "$result" | jq -r '.column')" >> "$GITHUB_OUTPUT"
          echo "companion_test_cleanup_eligible=$(echo "$result" | jq -r '.companion_test_cleanup_eligible')" >> "$GITHUB_OUTPUT"
          echo "companion_test_file=$(echo "$result" | jq -r '.companion_test_file')" >> "$GITHUB_OUTPUT"
          echo "impacted_packages=$(echo "$result" | jq -r '.impacted_packages | join(" ")')" >> "$GITHUB_OUTPUT"
          echo "reference_files=$(echo "$result" | jq -r '.reference_files | join(" ")')" >> "$GITHUB_OUTPUT"
          echo "reference_file_count=$(echo "$result" | jq -r '.reference_files | length')" >> "$GITHUB_OUTPUT"
        fi
        filtered=$(echo "$result" | jq -c '.filtered_candidates // []')
        echo "filtered_candidates=$filtered" >> "$GITHUB_OUTPUT"
    - name: Summarize recent outcomes
      id: summarize
      run: |
        set -euo pipefail
        if [ -f gh-aw-repo-memory/ci-deadcode-removal-rotation/memory/ci-deadcode-removal-rotation/memory.json ]; then
          summary=$(go run ./scripts/ci-deadcode-removal-rotation summarize \
            --memory gh-aw-repo-memory/ci-deadcode-removal-rotation/memory/ci-deadcode-removal-rotation/memory.json \
            --days 30)
        else
          summary="No previous attempt memory."
        fi
        echo "summary<<EOF" >> "$GITHUB_OUTPUT"
        echo "$summary" >> "$GITHUB_OUTPUT"
        echo "EOF" >> "$GITHUB_OUTPUT"
model: "deepseek/deepseek-v4.1-flash"
engine:
  id: claude
  args:
    - "--effort"
    - "high"
  env:
    ANTHROPIC_BASE_URL: "https://openrouter.ai/api"
    ANTHROPIC_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}
    # Run threat detection on Claude Sonnet 5 instead of inheriting the main
    # model. The detection job's AWF config keeps the AI-credits guard active
    # (max-ai-credits: -1 only applies to the agent job), so detection needs a
    # model with pricing configured below.
    GH_AW_MODEL_DETECTION_CLAUDE: "anthropic/claude-sonnet-5"
# Disable the per-run AI Credits budget guard. The OpenRouter model slug
# "deepseek/deepseek-v4.1-flash" may be absent from the AWF api-proxy's built-in
# pricing table and the models.dev catalog. gh-aw (v0.81.6) does not expose
# apiProxy.defaultAiCreditsPricing in frontmatter, so with the guard active the
# proxy could reject every request with HTTP 400 (unknown_model_ai_credits).
# Setting -1 omits maxAiCredits from the generated AWF config (and disables
# token steering), letting the agent run. The daily guardrail
# (max-daily-ai-credits, default 5000/day) still applies.
max-ai-credits: -1
# The daily guardrail is also disabled: it fails closed ("Daily AIC accounting
# record could not be resolved") when prior runs lack resolvable AIC usage.
max-daily-ai-credits: -1
models:
  providers:
    anthropic:
      models:
        "llm-gateway/DeepSeek-V4-Flash":
          "cost":
            "input": "1.4e-7"
            "output": "2.8e-7"
            "cache_read": "8e-08"
            "cache_write": "1e-06"
        # Pricing for the threat-detection model (GH_AW_MODEL_DETECTION_CLAUDE
        # above). The detection job's api-proxy enforces the AI-credits budget
        # and rejects unpriced models with HTTP 400.
        "anthropic/claude-sonnet-5":
          "cost":
            "input": "3e-06"
            "output": "1.5e-05"
            "cache_read": "3e-07"
            "cache_write": "3.75e-06"
timeout-minutes: 30
permissions:
  contents: read
  issues: read
  pull-requests: read
tools:
  cli-proxy: true
  timeout: 900
  repo-memory:
    - id: ci-deadcode-removal-rotation
      # NOTE: This branch name must match the repo-memory checkout step above.
      branch-name: memory/ci-deadcode-removal-rotation
      file-glob: ["memory/ci-deadcode-removal-rotation/memory.json"]
      create-orphan: true
      max-file-size: 524288
      max-patch-size: 102400
safe-outputs:
  create-pull-request:
    labels: [deadcode-cleanup, automated-cleanup, no-changelog]
    max: 1
    patch-format: am
    draft: false
  noop:
    max: 1
    report-as-issue: false
network:
  allowed: [defaults, node, go, openrouter.ai]
if: >-
  needs.pre_activation.outputs.found == 'true'
steps:
  - name: Pre-build custom linter before sandbox execution
    run: make golangci-lint-custom
jobs:
  pre-activation:
    outputs:
      found: ${{ steps.select_candidate.outputs.found }}
      symbol: ${{ steps.select_candidate.outputs.symbol }}
      symbol_name: ${{ steps.select_candidate.outputs.symbol_name }}
      package: ${{ steps.select_candidate.outputs.package }}
      file: ${{ steps.select_candidate.outputs.file }}
      line: ${{ steps.select_candidate.outputs.line }}
      column: ${{ steps.select_candidate.outputs.column }}
      companion_test_cleanup_eligible: ${{ steps.select_candidate.outputs.companion_test_cleanup_eligible }}
      companion_test_file: ${{ steps.select_candidate.outputs.companion_test_file }}
      impacted_packages: ${{ steps.select_candidate.outputs.impacted_packages }}
      reference_files: ${{ steps.select_candidate.outputs.reference_files }}
      reference_file_count: ${{ steps.select_candidate.outputs.reference_file_count }}
      filtered_candidates: ${{ steps.select_candidate.outputs.filtered_candidates }}
      summary: ${{ steps.summarize.outputs.summary }}
---

# Dead-code Removal Rotation Worker

You are responsible for safely removing one dead-code candidate per run and opening a verified cleanup PR.

## Pre-activation context

A deterministic pre-activation step has already selected the candidate and analyzed its references. Use only the values below.

- **Candidate found**: `${{ needs.pre_activation.outputs.found }}`
- **Symbol**: `${{ needs.pre_activation.outputs.symbol }}`
- **Package**: `${{ needs.pre_activation.outputs.package }}`
- **File**: `${{ needs.pre_activation.outputs.file }}`
- **Line**: `${{ needs.pre_activation.outputs.line }}`
- **Column**: `${{ needs.pre_activation.outputs.column }}`
- **Companion test cleanup eligible**: `${{ needs.pre_activation.outputs.companion_test_cleanup_eligible }}`
- **Companion test file**: `${{ needs.pre_activation.outputs.companion_test_file }}`
- **Impacted packages**: `${{ needs.pre_activation.outputs.impacted_packages }}`
- **Reference files**: `${{ needs.pre_activation.outputs.reference_files }}`
- **Reference file count**: `${{ needs.pre_activation.outputs.reference_file_count }}`

## Recent outcome summary

```
${{ needs.pre_activation.outputs.summary }}
```

## Pre-activation filtered candidates

Before proceeding with the main task, record the following candidates that pre-activation filtered due to invalid or excluded references. This avoids repeating costly gopls analysis on future runs.

```json
${{ needs.pre_activation.outputs.filtered_candidates }}
```

Run:

```bash
echo '${{ needs.pre_activation.outputs.filtered_candidates }}' | jq '[.[] | {symbol, package}]' | \
go run ./scripts/ci-deadcode-removal-rotation record-batch \
  --memory /tmp/gh-aw/repo-memory/ci-deadcode-removal-rotation/memory/ci-deadcode-removal-rotation/memory.json \
  --reason invalid_candidate_references
```

After recording the filtered candidates, proceed with the main task below.

## Task

1. **Read the candidate file** (`${{ needs.pre_activation.outputs.file }}`) and locate the dead function at line `${{ needs.pre_activation.outputs.line }}`.
2. **Remove the dead function** including its doc comment block. Ensure the file remains syntactically valid.
3. **Companion test cleanup (conditional)**:
   - Only if `${{ needs.pre_activation.outputs.companion_test_cleanup_eligible }}` is `true`:
     - Read the companion test file `${{ needs.pre_activation.outputs.companion_test_file }}`.
     - **Before deleting any tests**, search the file for the strings `resource.Test` or `resource.ParallelTest`.
     - If either string is present, **abort immediately** without making any changes.
       - Record the attempt as `invalid_candidate_acceptance_test` using:

         ```
         go run ./scripts/ci-deadcode-removal-rotation record \
           --memory /tmp/gh-aw/repo-memory/ci-deadcode-removal-rotation/memory/ci-deadcode-removal-rotation/memory.json \
           --symbol "${{ needs.pre_activation.outputs.symbol }}" \
           --package "${{ needs.pre_activation.outputs.package }}" \
           --reason invalid_candidate_acceptance_test \
           --context '{"referenceFileCount":${{ needs.pre_activation.outputs.reference_file_count }},"testCleanupEligible":true}'
         ```

       - Then call `noop` with a concise reason.
     - If the backstop passes, remove only the test functions that reference `${{ needs.pre_activation.outputs.symbol_name }}` (and their doc comments).
   - If `${{ needs.pre_activation.outputs.companion_test_cleanup_eligible }}` is `false`:
     - Do **not** delete any tests. Only remove the dead symbol.
4. **Clean up newly orphaned private helpers (bounded)**:
   - The selected function remains the only root candidate. You may additionally remove at most **three private helper functions**, only in the same package and production file as the selected function.
   - A helper must have been reachable from the selected function before this cleanup, and become unreferenced solely because of removals in this run. Check references across the repository before removing each helper; do not infer safety from an `unused` diagnostic alone.
   - Never remove exported functions, methods, package initialization, variables, types, or unrelated pre-existing dead code under this exception.
   - Do not delete additional tests under this exception. A helper with remaining references (including test references) is not eligible.
   - Repeat reference checking after each helper removal. If more than three helpers or changes outside this boundary are required, revert the cleanup, record `build_failed` with explanatory context, call `noop`, and stop.
   - Include every removed helper and its reference evidence in the PR body.
5. **Verify** the cleanup before opening a PR:
   - Run the full build **exactly once**, using the sandbox's memory-aware settings from the first attempt:

     ```bash
     GOMEMLIMIT=4GiB GOGC=20 make build
     ```

     `GOMEMLIMIT` is a soft Go runtime limit, leaving headroom under the sandbox's 6 GiB cgroup limit; reduced GC target further reduces peak memory. This remains the full `make build` gate, not a compilation-only substitute.
   - Preserve its exit status: run it directly, or capture the status before printing logs; never hide a failing build behind a successful pipe or log command.
   - **Do not retry the build**, including after a kill, timeout, or failure, and do not try alternative lint flags, GC settings, or compilation-only commands as a substitute. Reserve the remaining execution time for recording the outcome and calling `noop`.
   - If the build fails or times out:
     - Record the attempt as `build_failed` (or `verification_timeout` on timeout):

       ```
       go run ./scripts/ci-deadcode-removal-rotation record \
         --memory /tmp/gh-aw/repo-memory/ci-deadcode-removal-rotation/memory/ci-deadcode-removal-rotation/memory.json \
         --symbol "${{ needs.pre_activation.outputs.symbol }}" \
         --package "${{ needs.pre_activation.outputs.package }}" \
         --reason <reason>
       ```

     - Revert the candidate and helper cleanup edits, without modifying repo-memory.
     - Then call `noop` with a concise reason, including the build exit status and whether it was killed or timed out.
     - Stop without creating a PR.
   - Run unit tests for the impacted packages:

     ```
     go test -v ${{ needs.pre_activation.outputs.impacted_packages }}
     ```

     If any test fails:
     - Record the attempt as `tests_failed`.
     - Then call `noop` with a concise reason.
     - Stop without creating a PR.
6. **Format** the cleaned files:
   - Run `make fmt`.
7. **Open a cleanup PR** using the `create-pull-request` safe output **only** if verification succeeds.
   - Title format: `[deadcode] Remove ${{ needs.pre_activation.outputs.symbol }}`
   - Body must include:
     - A short description of the removed symbol.
     - The recent outcome summary.
     - A note that the PR was generated by the dead-code removal rotation workflow and that maintainers should review, merge, or close it manually.
8. **Record success** after opening the PR:

   ```
   go run ./scripts/ci-deadcode-removal-rotation record \
     --memory /tmp/gh-aw/repo-memory/ci-deadcode-removal-rotation/memory/ci-deadcode-removal-rotation/memory.json \
     --symbol "${{ needs.pre_activation.outputs.symbol }}" \
     --package "${{ needs.pre_activation.outputs.package }}" \
     --reason pr_created \
     --context '{"referenceFileCount":${{ needs.pre_activation.outputs.reference_file_count }},"testCleanupEligible":${{ needs.pre_activation.outputs.companion_test_cleanup_eligible }}}'
   ```

9. If you cannot safely proceed at any point, record the appropriate reason code and call `noop`.

## Guardrails

- Remove only one root dead-code candidate per run. The sole exception is up to three newly orphaned private helper functions satisfying all boundaries in task step 4; never expand cleanup to unrelated dead code.
- Never delete tests unless pre-activation has explicitly marked the candidate as eligible for companion test cleanup.
- Never bypass the `resource.Test` / `resource.ParallelTest` backstop.
- Do not open a PR if verification fails.
- Every run MUST end with exactly one safe-output call — either `create-pull-request` or `noop`. Never stop mid-task without calling one of them. Recording the attempt after `create-pull-request` is expected and is not a second safe output.
- Keep changes minimal and focused.
