---
imports:
  - shared/setup-dev.md
  - shared/dispatch-research-factory.md
name: Kibana API Gap Scanner
description: Finds Kibana API request fields that Terraform resources never use, triages them with an LLM that may only suppress, and files grounded issues for research.
on:
  workflow_dispatch:
  schedule:
    - cron: daily
  push:
    branches:
      - main
    paths:
      - 'generated/kbapi/**'
  # Discovery compiles the whole provider, which does not fit pre_activation (1 vCPU,
  # 15 minute cap). It runs in the dedicated discover job below, ahead of pre_activation.
  needs:
    - discover
concurrency:
  group: kibana-api-gap-scanner
  cancel-in-progress: false
model: "anthropic/claude-sonnet-5"
engine:
  id: claude
  args:
    - "--effort"
    - "high"
  env:
    ANTHROPIC_BASE_URL: "https://openrouter.ai/api"
    ANTHROPIC_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}
# Disable the per-run AI Credits budget guard. The OpenRouter model slug
# "anthropic/claude-sonnet-5" may be absent from the AWF api-proxy's built-in
# pricing table. gh-aw's models.providers frontmatter override does not
# propagate to apiProxy.defaultAiCreditsPricing
# (see https://github.com/github/gh-aw/issues/47365, fix pending in
# https://github.com/github/gh-aw/pull/47571), so with the guard active the
# proxy could reject every request with HTTP 400 (unknown_model_ai_credits).
# Setting -1 omits maxAiCredits from the generated AWF config, letting the
# agent run. The daily guardrail (max-daily-ai-credits, default 5000/day)
# still applies.
max-ai-credits: -1
permissions:
  contents: read
  issues: read
  pull-requests: read
  actions: read
tools:
  cli-proxy: true
  github:
    mode: gh-proxy
    toolsets: [default, issues]
  timeout: 300
  repo-memory:
    - id: kibana-api-gap-scanner
      # NOTE: This branch name must match the ref of the repo-memory checkout step above.
      branch-name: memory/kibana-api-gap-scanner
      file-glob: ["memory/kibana-api-gap-scanner/state.json"]
      create-orphan: true
      max-file-size: 524288
      max-patch-size: 102400
safe-outputs:
  create-issue:
    title-prefix: "[kibana-api-gap] "
    labels: [kibana-api-gap, triaged]
    max: 3
  noop:
    max: 1
    report-as-issue: false
network:
  allowed: [defaults, node, go, openrouter.ai]
if: >-
  needs.discover.outputs.run_agent == 'true'
steps:
  - name: Download Kibana API gap report
    uses: actions/download-artifact@v8.0.1
    with:
      name: kibana-api-gap-report
      path: /tmp/gh-aw/agent
  - name: Verify report artifact path
    run: test -f /tmp/gh-aw/agent/report.json
jobs:
  discover:
    # A standard runner with a Go build/module cache: a cold build of the provider and its
    # dependency graph needs several CPU minutes. The timeouts fail a hang fast.
    runs-on: ubuntu-latest
    timeout-minutes: 25
    permissions:
      contents: read
      issues: read
    outputs:
      run_agent: ${{ steps.discover_gaps.outputs.run_agent }}
      open_issues: ${{ steps.compute_issue_slots.outputs.open_issues }}
      issue_slots_available: ${{ steps.compute_issue_slots.outputs.issue_slots_available }}
      queue_size: ${{ steps.discover_gaps.outputs.queue_size }}
      gate_reason: ${{ steps.discover_gaps.outputs.gate_reason }}
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
          cache: true
      - name: Compute issue slots
        id: compute_issue_slots
        uses: actions/github-script@v9.0.0
        env:
          ISSUE_SLOTS_LABEL: kibana-api-gap
          ISSUE_SLOTS_CAP: "3"
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          script: |
            const fn = require('${{ github.workspace }}/.github/scripts/workflows/issue-slots/compute.js');
            await fn({ github, context, core });
      # NOTE: This ref must match the repo-memory tool config branch-name below.
      - name: Check for repo-memory branch
        id: memory_branch
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          # Exit code 2 means the branch does not exist yet (first run); any other failure must fail the job.
          set +e
          git ls-remote --exit-code --heads "https://x-access-token:${GH_TOKEN}@github.com/${GITHUB_REPOSITORY}.git" memory/kibana-api-gap-scanner > /dev/null
          status=$?
          set -e
          case "$status" in
            0) echo "exists=true" >> "$GITHUB_OUTPUT" ;;
            2) echo "exists=false" >> "$GITHUB_OUTPUT" ;;
            *) echo "git ls-remote failed with exit code $status" >&2; exit "$status" ;;
          esac
      - name: Checkout repo-memory branch
        if: steps.memory_branch.outputs.exists == 'true'
        uses: actions/checkout@v7.0.1
        with:
          ref: memory/kibana-api-gap-scanner
          path: gh-aw-repo-memory/kibana-api-gap-scanner
          fetch-depth: 1
      - name: Generate filtered Kibana OpenAPI spec
        timeout-minutes: 5
        run: make -C generated/kbapi transform
      # CGO_ENABLED=0 matches the environment discover pins for its package load, so the
      # provider is compiled once and the load reuses the build cache instead of recompiling.
      - name: Build gap scanner
        timeout-minutes: 12
        env:
          CGO_ENABLED: "0"
        run: go build -o "${RUNNER_TEMP}/kibana-api-gaps" ./scripts/kibana-api-gaps
      - name: Discover Kibana API gaps
        id: discover_gaps
        timeout-minutes: 10
        env:
          TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL: "true"
          ISSUE_SLOTS_AVAILABLE: ${{ steps.compute_issue_slots.outputs.issue_slots_available }}
        run: |
          mkdir -p /tmp/gh-aw/agent
          "${RUNNER_TEMP}/kibana-api-gaps" discover \
            --repo . \
            --slots "${ISSUE_SLOTS_AVAILABLE}" \
            --memory gh-aw-repo-memory/kibana-api-gap-scanner/memory/kibana-api-gap-scanner/state.json \
            --report /tmp/gh-aw/agent/report.json
          {
            echo "run_agent=$(jq -r '.run_agent' /tmp/gh-aw/agent/report.json)"
            echo "gate_reason=$(jq -r '.gate_reason' /tmp/gh-aw/agent/report.json)"
            echo "queue_size=$(jq -r '.queue | length' /tmp/gh-aw/agent/report.json)"
          } >> "$GITHUB_OUTPUT"
      - name: Upload Kibana API gap report
        if: success()
        uses: actions/upload-artifact@v7.0.1
        with:
          name: kibana-api-gap-report
          path: /tmp/gh-aw/agent/report.json
          if-no-files-found: error
---

# Kibana API Gap Scanner Worker

Deterministic discovery has already proven that each gap group below is a set of Kibana API request fields that a Terraform resource can reach but never uses. You do **judgement only**: for each queued group, decide whether it is worth exposing in Terraform. You cannot add, remove, or edit fields; you can only return a verdict and a reason.

## Pre-activation context

Do **not** query GitHub issue counts yourself; use only these values.

- **Open `kibana-api-gap` issues**: `${{ needs.discover.outputs.open_issues }}`
- **Issue slots available**: `${{ needs.discover.outputs.issue_slots_available }}`
- **Groups in the triage queue**: `${{ needs.discover.outputs.queue_size }}`
- **Gate reason**: ${{ needs.discover.outputs.gate_reason }}

## Inputs

Read `/tmp/gh-aw/agent/report.json` (downloaded from the discover job artifact). Only its `queue` matters to you. Each queue item has `key`, `resource`, `fingerprint`, `size`, `title`, and `body` (a pre-rendered issue body that ends with a hidden `<!-- kibana-api-gap ... -->` marker). Ignore `groups` and the diagnostics.

## Repository toolchain

Deterministic workflow steps have already installed Go from `go.mod`, exported `GOROOT`, `GOPATH`, and `GOMODCACHE` for AWF, installed Node from `package.json`, and run `make setup`. Use the resulting `go` tooling as-is. Do **not** install alternate toolchains.

## Triage loop

Work through the queue **strictly in order**, one group at a time. For each group:

1. Read the pre-rendered issue body for the group (its fields, OAS pointers, and any `x-state` version) and the resource implementation under `internal/`.
2. Decide exactly one verdict, `file` or `suppress`, and write **one** short reason for it.
3. If the verdict is `file`, call `create-issue` for that group (see below).

Stop as soon as **either** of these is true:

- the number of `file` verdicts equals `${{ needs.discover.outputs.issue_slots_available }}`; or
- 15 groups have been triaged.

`suppress` verdicts never consume a slot. Never skip a group to reach a later one: your verdicts must be a prefix of the queue in queue order. `record` also rejects a prefix that stops early, so keep triaging until the slots are filled with `file` verdicts or the whole queue is covered.

### When to suppress

Suppress **only** when you are highly confident the group should not be exposed in Terraform, and only for one of these reasons:

- **Server-managed or computed state**: values the server sets or that users cannot meaningfully control (revision counters, timestamps, generated identifiers, internal status).
- **Internal or telemetry fields**: fields for Kibana internals, analytics, or UI bookkeeping.
- **Runtime or imperative actions** that make no sense as infrastructure as code (for example "execute now", "snooze", "refresh").
- **Already covered by an opaque attribute** the resource exposes (for example a JSON-string attribute that carries the whole structure) which the deterministic analysis missed.

**If in doubt, file.** A group you are not highly confident about is filed.

## Issue creation

For each `file` verdict call `create-issue` once:

- **title**: the queue item's `title`, unchanged (the `[kibana-api-gap] ` prefix is added automatically, so do not repeat it).
- **body**: the exact stdout of `go run ./scripts/kibana-api-gaps render-issue --report /tmp/gh-aw/agent/report.json --key '<queue key>' --reason '<your reason>'`, pasted byte for byte. The command inserts the `Triage reason:` paragraph before the hidden marker and strips newlines and HTML comment delimiters from the reason. Never edit, reflow, or add to the body yourself.

Do not apply any labels yourself and do not add any factory trigger label; the safe output applies `kibana-api-gap` and `triaged`.

### Issue title length guardrail

GitHub issue titles are limited to **256 characters total**, including the `title-prefix` that `create-issue` prepends automatically. This workflow's prefix is `"[kibana-api-gap] "` (17 characters), leaving **239** characters for the title. The pre-rendered titles are short; if one would exceed the limit, truncate the field part only, never the resource name.

## Recording verdicts

After the last verdict (and after all `create-issue` calls), write `/tmp/gh-aw/agent/verdicts.json`, a JSON array in triage order with one object per triaged group and no other fields:

```json
[
  {"key": "<queue key>", "verdict": "file", "reason": "<one short reason>"},
  {"key": "<queue key>", "verdict": "suppress", "reason": "<one short reason>"}
]
```

Then run `record` exactly **once**:

```
go run ./scripts/kibana-api-gaps record \
  --report /tmp/gh-aw/agent/report.json \
  --verdicts /tmp/gh-aw/agent/verdicts.json \
  --memory /tmp/gh-aw/repo-memory/kibana-api-gap-scanner/memory/kibana-api-gap-scanner/state.json
```

`record` validates the verdicts against the report (known keys, queue order, no more `file` verdicts than slots, no early stop before the slots are filled or the queue is covered, non-empty reasons). If it rejects the file, fix `verdicts.json` and run it again; it must succeed before you finish. Never write `state.json` by hand.

If you filed nothing and the queue was non-empty, still call `record`, then call `noop` with a short reason.

## Dispatch

After creating all issues for this run (or if no issues were created), call the `dispatch_research_factory` safe output tool once with `dispatch: true` to dispatch the `research-factory` workflow for each created issue.
