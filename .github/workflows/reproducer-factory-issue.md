---
imports:
  - shared/setup-dev.md
  - shared/elastic-stack.md
name: Reproducer Factory Issue Intake
timeout-minutes: 65
description: >-
  Reacts to trusted qualifying `reproducer-factory` issue events or internal workflow dispatch
  requests, suppresses duplicate linked pull requests, and delegates bug reproduction to an agent
  that emits a sticky reproducer comment and optionally opens a single linked PR when the bug is
  reproduced as a passing acceptance test.
on:
  issues:
    types: [opened, labeled]
  workflow_dispatch:
    inputs:
      issue_number:
        description: 'Issue number to reproduce'
        required: true
        type: number
      source_workflow:
        description: 'Source workflow that triggered this dispatch'
        required: false
        type: string
  status-comment: true
  permissions:
    contents: read
    issues: write
    pull-requests: read
  steps:
    - name: Checkout repository
      uses: actions/checkout@v7.0.1
      with:
        persist-credentials: false
        fetch-depth: 1
    - name: Determine intake mode
      id: determine_intake_mode
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const mode = context.eventName === 'workflow_dispatch' ? 'dispatch' : 'issue-event';
          core.setOutput('intake_mode', mode);
          core.info(`Intake mode: ${mode}`);
    - name: Qualify trigger event
      id: qualify_trigger
      if: steps.determine_intake_mode.outputs.intake_mode == 'issue-event'
      uses: actions/github-script@v9.0.0
      env:
        FACTORY_NAME: reproducer-factory
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/qualify-trigger.js');
          await fn({ github, context, core });
    - name: Capture issue context
      id: capture_issue_context
      if: steps.determine_intake_mode.outputs.intake_mode == 'issue-event'
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          core.setOutput('issue_number', context.payload.issue?.number ?? '');
          core.setOutput('issue_title', context.payload.issue?.title ?? '');
          core.setOutput('issue_body', context.payload.issue?.body ?? '');
    - name: Validate dispatch inputs
      id: validate_dispatch_inputs
      if: steps.determine_intake_mode.outputs.intake_mode == 'dispatch'
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/validate-dispatch-inputs.js');
          await fn({ github, context, core });
    - name: Fetch live issue
      id: fetch_live_issue
      if: >-
        steps.determine_intake_mode.outputs.intake_mode == 'dispatch' &&
        steps.validate_dispatch_inputs.outputs.event_eligible == 'true'
      env:
        INPUT_ISSUE_NUMBER: ${{ steps.validate_dispatch_inputs.outputs.issue_number }}
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/fetch-live-issue.js');
          await fn({ github, context, core });
    - name: Fetch issue comments
      id: fetch_issue_comments
      if: >-
        (
          steps.determine_intake_mode.outputs.intake_mode == 'issue-event' &&
          steps.qualify_trigger.outputs.event_eligible == 'true'
        ) || (
          steps.determine_intake_mode.outputs.intake_mode == 'dispatch' &&
          steps.validate_dispatch_inputs.outputs.event_eligible == 'true'
        )
      env:
        INPUT_ISSUE_NUMBER: >-
          ${{ steps.determine_intake_mode.outputs.intake_mode == 'issue-event'
            && steps.capture_issue_context.outputs.issue_number
            || steps.validate_dispatch_inputs.outputs.issue_number }}
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/fetch-issue-comments.js');
          await fn({ github, context, core });
    - name: Fetch prior reproducer comment
      id: fetch_prior_reproducer_comment
      if: >-
        (
          steps.determine_intake_mode.outputs.intake_mode == 'issue-event' &&
          steps.qualify_trigger.outputs.event_eligible == 'true'
        ) || (
          steps.determine_intake_mode.outputs.intake_mode == 'dispatch' &&
          steps.validate_dispatch_inputs.outputs.event_eligible == 'true'
        )
      env:
        INPUT_ISSUE_NUMBER: >-
          ${{ steps.determine_intake_mode.outputs.intake_mode == 'issue-event'
            && steps.capture_issue_context.outputs.issue_number
            || steps.validate_dispatch_inputs.outputs.issue_number }}
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/reproducer-factory/fetch-prior-reproducer-comment.js');
          await fn({ github, context, core });
    - name: Check duplicate PR
      id: check_duplicate_pr
      if: >-
        (
          steps.determine_intake_mode.outputs.intake_mode == 'issue-event' &&
          steps.qualify_trigger.outputs.event_eligible == 'true'
        ) || (
          steps.determine_intake_mode.outputs.intake_mode == 'dispatch' &&
          steps.validate_dispatch_inputs.outputs.event_eligible == 'true'
        )
      uses: actions/github-script@v9.0.0
      env:
        FACTORY_NAME: reproducer-factory
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/check-duplicate-pr.js');
          await fn({ github, context, core });
    - name: Remove trigger label
      id: remove_trigger_label
      if: >-
        steps.determine_intake_mode.outputs.intake_mode == 'issue-event' &&
        steps.qualify_trigger.outputs.event_eligible == 'true' &&
        steps.check_duplicate_pr.outputs.duplicate_pr_found != 'true'
      uses: actions/github-script@v9.0.0
      env:
        FACTORY_NAME: reproducer-factory
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/remove-trigger-label.js');
          await fn({ github, context, core });
    - name: Set phase label
      id: set_phase_label
      if: >-
        (
          steps.determine_intake_mode.outputs.intake_mode == 'issue-event' &&
          steps.qualify_trigger.outputs.event_eligible == 'true'
        ) || (
          steps.determine_intake_mode.outputs.intake_mode == 'dispatch' &&
          steps.validate_dispatch_inputs.outputs.event_eligible == 'true'
        )
      env:
        INPUT_ISSUE_NUMBER: >-
          ${{ steps.determine_intake_mode.outputs.intake_mode == 'issue-event'
            && steps.capture_issue_context.outputs.issue_number
            || steps.validate_dispatch_inputs.outputs.issue_number }}
        PHASE_LABEL_NAME: phase-reproduction
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/phase-label/set.js');
          await fn({ github, context, core });
    - name: Normalize context
      id: normalize_context
      if: always()
      env:
        INTAKE_MODE: ${{ steps.determine_intake_mode.outputs.intake_mode }}
        ISSUE_NUMBER_EVENT: ${{ steps.capture_issue_context.outputs.issue_number }}
        ISSUE_TITLE_EVENT: ${{ steps.capture_issue_context.outputs.issue_title }}
        ISSUE_BODY_EVENT: ${{ steps.capture_issue_context.outputs.issue_body }}
        EVENT_ELIGIBLE_EVENT: ${{ steps.qualify_trigger.outputs.event_eligible }}
        EVENT_ELIGIBLE_REASON_EVENT: ${{ steps.qualify_trigger.outputs.event_eligible_reason }}
        TRIGGER_LABEL_REMOVED_EVENT: ${{ steps.remove_trigger_label.outputs.trigger_label_removed }}
        TRIGGER_LABEL_REMOVED_REASON_EVENT: ${{ steps.remove_trigger_label.outputs.trigger_label_removed_reason }}
        ISSUE_NUMBER_DISPATCH: ${{ steps.fetch_live_issue.outputs.issue_number }}
        ISSUE_TITLE_DISPATCH: ${{ steps.fetch_live_issue.outputs.issue_title }}
        ISSUE_BODY_DISPATCH: ${{ steps.fetch_live_issue.outputs.issue_body }}
        EVENT_ELIGIBLE_DISPATCH: ${{ steps.validate_dispatch_inputs.outputs.event_eligible }}
        EVENT_ELIGIBLE_REASON_DISPATCH: ${{ steps.validate_dispatch_inputs.outputs.event_eligible_reason }}
        ISSUE_COMMENTS: ${{ steps.fetch_issue_comments.outputs.issue_comments }}
        PRIOR_REPRODUCER_COMMENT: ${{ steps.fetch_prior_reproducer_comment.outputs.prior_reproducer_comment }}
        SOURCE_WORKFLOW: ${{ github.event.inputs.source_workflow }}
        DUPLICATE_PR_FOUND: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_found }}
        DUPLICATE_PR_URL: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_url }}
      run: |
        echo "intake_mode=${INTAKE_MODE}" >> "$GITHUB_OUTPUT"
        EOF_DELIM="EOF_$(cat /proc/sys/kernel/random/uuid 2>/dev/null || date +%s%N)"

        if [ "${INTAKE_MODE}" = "issue-event" ]; then
          echo "issue_number=${ISSUE_NUMBER_EVENT}" >> "$GITHUB_OUTPUT"
          echo "issue_title=${ISSUE_TITLE_EVENT}" >> "$GITHUB_OUTPUT"
          {
            echo "issue_body<<${EOF_DELIM}"
            printf '%s\n' "${ISSUE_BODY_EVENT}"
            echo "${EOF_DELIM}"
          } >> "$GITHUB_OUTPUT"
          echo "event_eligible=${EVENT_ELIGIBLE_EVENT}" >> "$GITHUB_OUTPUT"
          echo "event_eligible_reason=${EVENT_ELIGIBLE_REASON_EVENT}" >> "$GITHUB_OUTPUT"
          echo "actor_trusted=true" >> "$GITHUB_OUTPUT"
          echo "actor_trusted_reason=Role-based gate guarantees trust for issue events." >> "$GITHUB_OUTPUT"
          echo "trigger_label_removed=${TRIGGER_LABEL_REMOVED_EVENT}" >> "$GITHUB_OUTPUT"
          echo "trigger_label_removed_reason=${TRIGGER_LABEL_REMOVED_REASON_EVENT}" >> "$GITHUB_OUTPUT"
        else
          echo "issue_number=${ISSUE_NUMBER_DISPATCH}" >> "$GITHUB_OUTPUT"
          echo "issue_title=${ISSUE_TITLE_DISPATCH}" >> "$GITHUB_OUTPUT"
          {
            echo "issue_body<<${EOF_DELIM}"
            printf '%s\n' "${ISSUE_BODY_DISPATCH}"
            echo "${EOF_DELIM}"
          } >> "$GITHUB_OUTPUT"
          echo "event_eligible=${EVENT_ELIGIBLE_DISPATCH}" >> "$GITHUB_OUTPUT"
          echo "event_eligible_reason=${EVENT_ELIGIBLE_REASON_DISPATCH}" >> "$GITHUB_OUTPUT"
          echo "actor_trusted=true" >> "$GITHUB_OUTPUT"
          echo "actor_trusted_reason=Dispatch intake bypasses actor trust check." >> "$GITHUB_OUTPUT"
          echo "trigger_label_removed=false" >> "$GITHUB_OUTPUT"
          echo "trigger_label_removed_reason=Trigger label removal is not required for dispatch intake." >> "$GITHUB_OUTPUT"
          echo "source_workflow=${SOURCE_WORKFLOW}" >> "$GITHUB_OUTPUT"
        fi

        {
          echo "issue_comments<<${EOF_DELIM}"
          printf '%s\n' "${ISSUE_COMMENTS}"
          echo "${EOF_DELIM}"
        } >> "$GITHUB_OUTPUT"

        {
          echo "prior_reproducer_comment<<${EOF_DELIM}"
          printf '%s\n' "${PRIOR_REPRODUCER_COMMENT}"
          echo "${EOF_DELIM}"
        } >> "$GITHUB_OUTPUT"

        echo "duplicate_pr_found=${DUPLICATE_PR_FOUND}" >> "$GITHUB_OUTPUT"
        echo "duplicate_pr_url=${DUPLICATE_PR_URL}" >> "$GITHUB_OUTPUT"
      shell: bash
    - name: Write issue context to files
      id: write_context_files
      if: >-
        steps.normalize_context.outputs.event_eligible == 'true' &&
        steps.normalize_context.outputs.actor_trusted == 'true' &&
        steps.check_duplicate_pr.outputs.duplicate_pr_found != 'true' &&
        steps.normalize_context.outputs.issue_number != ''
      env:
        FACTORY_NAME: reproducer-factory
        ISSUE_BODY: ${{ steps.normalize_context.outputs.issue_body }}
        ISSUE_COMMENTS: ${{ steps.normalize_context.outputs.issue_comments }}
        PRIOR_FACTORY_COMMENT: ${{ steps.normalize_context.outputs.prior_reproducer_comment }}
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/write-context-files.js');
          await fn({ github, context, core });
    - name: Upload issue context artifact
      if: >-
        steps.normalize_context.outputs.event_eligible == 'true' &&
        steps.normalize_context.outputs.actor_trusted == 'true' &&
        steps.check_duplicate_pr.outputs.duplicate_pr_found != 'true' &&
        steps.normalize_context.outputs.issue_number != ''
      uses: actions/upload-artifact@v7.0.1
      with:
        name: reproducer-factory-issue-context
        path: /tmp/reproducer-factory-context/
        if-no-files-found: error
    - name: Finalize gate reason
      id: finalize_gate
      if: always()
      uses: actions/github-script@v9.0.0
      env:
        FACTORY_NAME: reproducer-factory
        EVENT_ELIGIBLE: ${{ steps.normalize_context.outputs.event_eligible }}
        EVENT_ELIGIBLE_REASON: ${{ steps.normalize_context.outputs.event_eligible_reason }}
        ACTOR_TRUSTED: ${{ steps.normalize_context.outputs.actor_trusted }}
        ACTOR_TRUSTED_REASON: ${{ steps.normalize_context.outputs.actor_trusted_reason }}
        DUPLICATE_PR_FOUND: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_found }}
        DUPLICATE_PR_URL: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_url }}
        DUPLICATE_GATE_REASON: ${{ steps.check_duplicate_pr.outputs.gate_reason }}
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/finalize-gate.js');
          await fn({ github, context, core });
env:
  REPRODUCER_FACTORY_ISSUE_NUMBER: ${{ github.event.issue.number || inputs.issue_number }}
concurrency:
  group: reproducer-factory-issue-${{ github.event.issue.number || inputs.issue_number }}
  cancel-in-progress: false
if: >-
  needs.pre_activation.outputs.event_eligible == 'true' &&
  needs.pre_activation.outputs.actor_trusted == 'true' &&
  needs.pre_activation.outputs.duplicate_pr_found != 'true' &&
  needs.pre_activation.outputs.issue_number != ''
steps:
  - name: Download issue context artifact
    uses: actions/download-artifact@v8.0.1
    with:
      name: reproducer-factory-issue-context
      path: /tmp/reproducer-factory-context/
model: "llm-gateway/claude-sonnet-5"
engine:
  id: claude
  args:
    - "--effort"
    - "high"
  env:
    ANTHROPIC_BASE_URL: "https://elastic.litellm-prod.ai/"
    ANTHROPIC_API_KEY: ${{ secrets.CLAUDE_LITELLM_PROXY_API_KEY }}
# Disable the per-run AI Credits budget guard. The model alias
# "llm-gateway/claude-sonnet-5" is a private Elastic LiteLLM alias absent from
# the AWF api-proxy's built-in pricing table. gh-aw's models.providers
# frontmatter override does not propagate to apiProxy.defaultAiCreditsPricing
# (see https://github.com/github/gh-aw/issues/47365, fix pending in
# https://github.com/github/gh-aw/pull/47571), so with the guard active the
# proxy rejects every request with HTTP 400 (unknown_model_ai_credits).
# Setting -1 omits maxAiCredits from the generated AWF config, letting the
# agent run. The daily guardrail (max-daily-ai-credits, default 5000/day)
# still applies.
max-ai-credits: -1
permissions:
  contents: read
  issues: read
  pull-requests: read
jobs:
  pre-activation:
    outputs:
      intake_mode: ${{ steps.normalize_context.outputs.intake_mode }}
      issue_number: ${{ steps.normalize_context.outputs.issue_number }}
      issue_title: ${{ steps.normalize_context.outputs.issue_title }}
      issue_body: ${{ steps.normalize_context.outputs.issue_body }}
      event_eligible: ${{ steps.normalize_context.outputs.event_eligible }}
      event_eligible_reason: ${{ steps.normalize_context.outputs.event_eligible_reason }}
      actor_trusted: ${{ steps.normalize_context.outputs.actor_trusted }}
      actor_trusted_reason: ${{ steps.normalize_context.outputs.actor_trusted_reason }}
      duplicate_pr_found: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_found }}
      duplicate_pr_url: ${{ steps.check_duplicate_pr.outputs.duplicate_pr_url }}
      trigger_label_removed: ${{ steps.normalize_context.outputs.trigger_label_removed }}
      trigger_label_removed_reason: ${{ steps.normalize_context.outputs.trigger_label_removed_reason }}
      source_workflow: ${{ steps.normalize_context.outputs.source_workflow }}
      gate_reason: ${{ steps.finalize_gate.outputs.gate_reason }}
tools:
  cli-proxy: true
  github:
    mode: gh-proxy
    toolsets: [issues, pull_requests, repos]
network:
  allowed: [defaults, node, go, elastic.litellm-prod.ai, www.elastic.co]
mcp-servers:
  elastic-docs:
    url: "https://www.elastic.co/docs/_mcp/"
    allowed: ["*"]
checkout:
  fetch-depth: 0
safe-outputs:
  jobs:
    update-reproducer-comment:
      description: Create or update the reproducer-factory sticky comment on the triggering issue
      permissions:
        contents: read
        issues: write
      runs-on: ubuntu-latest
      output: "Reproducer comment created or updated successfully."
      inputs:
        body:
          description: Markdown body of the reproducer comment (without the gha-reproducer-factory marker)
          required: true
          type: string
      steps:
        - name: Checkout repository
          uses: actions/checkout@v7.0.1
          with:
            persist-credentials: false
            fetch-depth: 1
        - name: Create or update reproducer comment
          uses: actions/github-script@v9.0.0
          env:
            REPRODUCER_FACTORY_ISSUE_NUMBER: ${{ github.event.issue.number || inputs.issue_number }}
          with:
            github-token: ${{ secrets.GITHUB_TOKEN }}
            script: |
              const fn = require('${{ github.workspace }}/.github/scripts/workflows/reproducer-factory/update-reproducer-comment.js');
              await fn({ github, context, core });
  create-pull-request:
    draft: false
    labels: [reproducer-factory]
    max: 1
    auto-close-issue: false
  noop:
    max: 1
    report-as-issue: false
---

# Reproducer Factory issue reproduction worker

You reproduce **issue #${{ needs.pre_activation.outputs.issue_number }}** (`${{ needs.pre_activation.outputs.issue_title }}`) labeled `reproducer-factory`. The activation gates below reference this same issue number consistently — treat **`${{ needs.pre_activation.outputs.issue_number }}`** as the authoritative id for test naming, branch names, and `Related to #${{ needs.pre_activation.outputs.issue_number }}` linkage.

Express the reported failure as an acceptance test (`ExpectError` or `ExpectNonEmptyPlan`) and decide one of three outcomes: **reproduced**, **cannot reproduce**, or **appears fixed**. You **MUST** emit exactly one `update-reproducer-comment` safe output on every activation. You **MAY** emit `create-pull-request` **only** for outcome A (reproduced).

## Pre-activation context

Deterministic pre-activation has already decided that intake for **issue #${{ needs.pre_activation.outputs.issue_number }}** is eligible, the actor is trusted (or dispatch bypasses trust), and there is no open linked `reproducer-factory` pull request matching the duplicate-PR gate for **#${{ needs.pre_activation.outputs.issue_number }}**.

- **Gate reason**: `${{ needs.pre_activation.outputs.gate_reason }}`
- **Intake mode**: `${{ needs.pre_activation.outputs.intake_mode }}`
- **Issue number**: `${{ needs.pre_activation.outputs.issue_number }}`
- **Issue title**: `${{ needs.pre_activation.outputs.issue_title }}`
- **Issue body** (sanitised): `/tmp/reproducer-factory-context/issue_body.md`
- **Comment history** (sanitised, human-authored): `/tmp/reproducer-factory-context/issue_comments.md`
- **Prior reproducer comment** (if any): `/tmp/reproducer-factory-context/prior_reproducer_comment.md`

Read all context files before proceeding. They may contain markdown, code fences, and other content that cannot be safely embedded inline in a prompt.

- **Repository**: `${{ github.repository }}`
- **Triggered by**: `@${{ github.actor }}`
- **Required branch** (outcome A PR): `reproducer-factory/issue-${{ needs.pre_activation.outputs.issue_number }}`
- **Run link**: `${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}`

## Time budget

Reserve approximately **55 minutes** for investigation, documentation lookup, test authoring, and test runs. The job hard-kills at **65 minutes**. Reserve the final **~5 minutes** for emitting safe outputs (`update-reproducer-comment`, and `create-pull-request` only for outcome A).

## Partial output preference

If you are running short on time, prefer emitting a **partial-but-valid** `update-reproducer-comment` (for example an honest outcome-B comment with clear unknowns) over emitting `noop`. A partial comment that documents what you tried is more useful than silence.

## Test environment

The Elastic Stack is provisioned in the agent environment. Run targeted acceptance tests with:

```bash
ELASTICSEARCH_ENDPOINTS=http://host.docker.internal:9201 \
ELASTICSEARCH_USERNAME=elastic \
ELASTICSEARCH_PASSWORD=password \
KIBANA_ENDPOINT=http://host.docker.internal:5602 \
TF_ACC=1 \
go test -v -run TestAccReproduceIssue${{ needs.pre_activation.outputs.issue_number }} ./path/to/package
```

The proxy services (`es-proxy` on 9201, `kb-proxy` on 5602) bridge to the actual
stack. Use these proxy ports; direct ports 9200/5601 are blocked by the AWF
firewall.

## Elastic documentation

An `elastic-docs` MCP server is available with `search_docs`, `find_related_docs`, and `get_document_by_url`. Use these **before** writing the reproduction test to ground API behavior and constraints. This grounding step reduces incorrect assumptions about API shape.

If the MCP tools are unavailable or return no useful results, proceed from the issue content and repository code alone — **do not** emit `noop` solely because documentation tools failed.

## Task

Follow this decision tree end-to-end for **issue #${{ needs.pre_activation.outputs.issue_number }}**:

1. **Read** the issue title (`${{ needs.pre_activation.outputs.issue_title }}`), body, comment history, and prior reproducer comment (if present) thoroughly.
2. **Identify resource scope** and choose the test file location using **Test file placement**.
3. **Write** `TestAccReproduceIssue${{ needs.pre_activation.outputs.issue_number }}` using `resource.TestStep` with `ExpectError` or `ExpectNonEmptyPlan` (as appropriate) to assert the failure described in the issue.
4. **Run** the acceptance test against the live Elastic Stack (see **Test environment**). If the test passes with `ExpectError` or `ExpectNonEmptyPlan`, the failure is reproduced. If it fails for a different reason, iterate on the config. If it passes unexpectedly, consider outcome C.
5. **Route by result:**
   - **Test reproduces the failure** (`ExpectError` or `ExpectNonEmptyPlan` passes against the live stack) → **Outcome A (reproduced)**
     Emit `update-reproducer-comment` with the outcome-A body, then emit `create-pull-request` on branch `reproducer-factory/issue-${{ needs.pre_activation.outputs.issue_number }}`. The PR body **MUST** include `Related to #${{ needs.pre_activation.outputs.issue_number }}` (do **not** use `Closes`).
   - **Cannot build a credible test config or cannot determine from static analysis** → **Outcome B (cannot reproduce)**
     Emit `update-reproducer-comment` with the outcome-B body. **Do not** emit `create-pull-request`.
   - **Static analysis suggests the issue no longer applies** (e.g. recently merged fix, code path changed) → **Outcome C (appears fixed)**
     Emit `update-reproducer-comment` with the outcome-C body, including the test configuration (inline or fenced), evidence from code/git history that the failure condition no longer applies. **Do not** emit `create-pull-request`.

## Test file placement

- **Default**: `internal/acctest/reproductions/issue_${{ needs.pre_activation.outputs.issue_number }}_acc_test.go` defining `TestAccReproduceIssue${{ needs.pre_activation.outputs.issue_number }}`.
- **Resource package** when the issue clearly identifies one Terraform resource: the same `issue_${{ needs.pre_activation.outputs.issue_number }}_acc_test.go` / `TestAccReproduceIssue${{ needs.pre_activation.outputs.issue_number }}` pattern under that resource’s package (example: `internal/kibana/alertingrule/issue_${{ needs.pre_activation.outputs.issue_number }}_acc_test.go`) when the issue names an `elasticstack_*` type or an unambiguous human description (“alerting rule resource”, “Kibana dashboard”, etc.).
- **Use the fallback path** when the issue is ambiguous, spans multiple resources, or describes a provider-level concern that is not attributable to one resource package.

## Comment format

The reproducer comment **MUST** conform to the `ci-reproducer-factory-comment-format` capability. The workflow automatically prepends `<!-- gha-reproducer-factory -->`; **do not** include that marker in the body you pass to `update-reproducer-comment`.

1. **`## Bug reproduction`** — Open the body with this H2 heading. Immediately below it, add a provenance header that records the run timestamp, the run link (`${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}`), and a one-line notice that the comment is replaced on each run, e.g.:

   > This comment is replaced on each workflow run.

2. **Outcome-specific sections** (after the provenance header):

- **Outcome A**: `### Summary`, `### Root cause`, `### Reproduction test`, then `### References`, then a closed `<details>` block for pipeline metadata JSON.
- **Outcome B**: `### Summary`, `### Investigation avenues` with **exactly three** numbered items — each item **must** reference a concrete file path, Go symbol, git commit, or API endpoint (not vague advice).
- **Outcome C**: `### Summary`, `### Evidence`, then `### References`, then the `<details>` JSON block.

Place `### References` **before** the `<details>` metadata block.

3. **Pipeline metadata block** — After `### References`, include exactly one HTML `<details>` element (**no** `open` attribute) with `<summary>🤖 Pipeline metadata</summary>` containing exactly one fenced ```json``` code block. The JSON object MUST match `ci-reproducer-factory-comment-format` (`openspec/specs/ci-reproducer-factory-comment-format/spec.md`):

- `schema_version` (string, required): e.g. `"1.0"`.
- `outcome` (string, required): `"reproduced"`, `"cannot-reproduce"`, or `"appears-fixed"`.
- `test_name` (string, optional): `TestAccReproduceIssue{N}` — present for outcomes A and C.
- `test_file` (string, optional): relative path to the test file — present for outcome A.
- `pr_url` (string, optional): reproduction PR URL — present for outcome A.
- `references` (array, optional): each entry has `type` (`"elastic-docs"`, `"repo-path"`, `"issue"`, `"pr"`, `"external"`) and `url` or `path`.

The JSON `outcome` and optional fields MUST be internally consistent with the human-readable outcome sections above.

## Pull request contract

**Outcome A only:**

- Branch: `reproducer-factory/issue-${{ needs.pre_activation.outputs.issue_number }}`
- Label: `reproducer-factory` (via `create-pull-request` safe output)
- PR body **MUST** contain `Related to #${{ needs.pre_activation.outputs.issue_number }}`. Do **not** use `Closes #${{ needs.pre_activation.outputs.issue_number }}` — the reproduction confirms the bug for **#${{ needs.pre_activation.outputs.issue_number }}**; it does not resolve the issue.
- Include **only** the reproduction test file — no unrelated changes.

## Guardrails

- Do **not** re-check pre-activation gates (event eligibility, actor trust, duplicate PR state); pre-activation already handled them.
- Emit **only** configured safe outputs — **no** free-form issue comments, **no** label changes, **no** extra pull requests.
- Emit **at most one** `update-reproducer-comment` and **at most one** `create-pull-request`.
- If you cannot complete a full investigation, prefer a **partial** outcome-B `update-reproducer-comment` that states unknowns honestly over `noop`.
- Do **not** emit `noop` solely because the `elastic-docs` MCP server is unavailable.
- Do **not** include `<!-- gha-reproducer-factory -->` in the markdown you pass to `update-reproducer-comment`; the workflow adds it.