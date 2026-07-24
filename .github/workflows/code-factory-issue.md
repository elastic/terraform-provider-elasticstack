---
imports:
  - shared/setup-dev.md
  - shared/elastic-stack.md
name: Code Factory Issue Intake
timeout-minutes: 65
description: >-
  Reacts to trusted qualifying `code-factory` issue events or internal workflow dispatch requests,
  suppresses duplicate linked pull requests, and delegates implementation to an agent that creates
  exactly one linked PR per issue.
on:
  issues:
    types: [opened, labeled]
  bots:
    - github-actions[bot]
  workflow_dispatch:
    inputs:
      issue_number:
        description: 'Issue number to implement'
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
        FACTORY_NAME: code-factory
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
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/code-factory/fetch-issue-comments.js');
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
        FACTORY_NAME: code-factory
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
        FACTORY_NAME: code-factory
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
        PHASE_LABEL_NAME: phase-coding
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

        echo "duplicate_pr_found=${DUPLICATE_PR_FOUND}" >> "$GITHUB_OUTPUT"
        echo "duplicate_pr_url=${DUPLICATE_PR_URL}" >> "$GITHUB_OUTPUT"
      shell: bash
    - name: Sanitize context
      id: sanitize_context
      if: >-
        steps.normalize_context.outputs.event_eligible == 'true' &&
        steps.normalize_context.outputs.actor_trusted == 'true' &&
        steps.check_duplicate_pr.outputs.duplicate_pr_found != 'true'
      env:
        FACTORY_NAME: code-factory
        ISSUE_BODY: ${{ steps.normalize_context.outputs.issue_body }}
        HUMAN_COMMENTS: ${{ steps.fetch_issue_comments.outputs.human_comments }}
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/sanitize-context.js');
          await fn({ github, context, core });
    - name: Upload issue context artifact
      if: >-
        steps.normalize_context.outputs.event_eligible == 'true' &&
        steps.normalize_context.outputs.actor_trusted == 'true' &&
        steps.check_duplicate_pr.outputs.duplicate_pr_found != 'true'
      uses: actions/upload-artifact@v7.0.1
      with:
        name: code-factory-issue-context
        path: /tmp/code-factory-context/
        if-no-files-found: error
    - name: Finalize gate reason
      id: finalize_gate
      if: always()
      uses: actions/github-script@v9.0.0
      env:
        FACTORY_NAME: code-factory
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
if: >-
  needs.pre_activation.outputs.event_eligible == 'true' &&
  needs.pre_activation.outputs.actor_trusted == 'true' &&
  needs.pre_activation.outputs.duplicate_pr_found != 'true' &&
  needs.pre_activation.outputs.issue_number != ''
steps:
  - name: Download issue context artifact
    uses: actions/download-artifact@v8.0.1
    with:
      name: code-factory-issue-context
      path: /tmp/code-factory-context/
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
      sanitized_issue_body: ${{ steps.sanitize_context.outputs.sanitized_issue_body }}
      sanitized_issue_comments: ${{ steps.sanitize_context.outputs.sanitized_issue_comments }}
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
  create-pull-request:
    labels: [code-factory]
    max: 1
    patch-format: am
  noop:
    max: 1
    report-as-issue: false
---

# Code Factory issue intake worker

You implement exactly one GitHub issue labeled `code-factory`. The triggering issue is the sole source of truth for scope and requested behavior. Any acceptance criteria defined by the issue must be met, and your changes must be properly covered by automated testing. Do not broaden the scope beyond what the issue describes unless the repository already requires it to make the implementation viable.

## Pre-activation context

Deterministic pre-activation has already decided that this intake is eligible, the actor is trusted (or dispatch bypasses trust), and there is no open linked `code-factory` pull request for the issue.

- **Gate reason**: ${{ needs.pre_activation.outputs.gate_reason }}
- **Intake mode**: `${{ needs.pre_activation.outputs.intake_mode }}`
- **Issue number**: `${{ needs.pre_activation.outputs.issue_number }}`
- **Issue title**: `${{ needs.pre_activation.outputs.issue_title }}`
- **Issue body** (sanitised): see `/tmp/code-factory-context/issue_body.md`

- **Comment history** (sanitised, human-authored): see `/tmp/code-factory-context/issue_comments.md`

- **Repository**: `${{ github.repository }}`
- **Triggered by**: `@${{ github.actor }}`
- **Required branch**: `code-factory/issue-${{ needs.pre_activation.outputs.issue_number }}`

## Test environment

The Elastic Stack is provisioned in the agent environment. You can run targeted acceptance tests with:

```bash
ELASTICSEARCH_ENDPOINTS=http://host.docker.internal:9201 \
ELASTICSEARCH_USERNAME=elastic \
ELASTICSEARCH_PASSWORD=password \
KIBANA_ENDPOINT=http://host.docker.internal:5602 \
TF_ACC=1 \
go test -v -run TestAccResourceName ./path/to/package
```

## Elastic documentation

An `elastic-docs` MCP server is available with three tools: `search_docs`, `find_related_docs`, and
`get_document_by_url`. Before writing implementation code, use `search_docs` to look up API
behavior, parameters, and constraints for the feature described in the issue. This grounding step
helps produce accurate implementations and avoids speculative assumptions about API shape.

If the MCP tools are unavailable or return no useful results, proceed with implementation from the
issue content alone - do not block the run waiting for documentation.

## Task

Implement the triggering issue on branch `code-factory/issue-${{ needs.pre_activation.outputs.issue_number }}` and create exactly one linked pull request labeled `code-factory`.

1. Read the issue title and body carefully and treat them as authoritative.
2. Create or update the implementation on branch `code-factory/issue-${{ needs.pre_activation.outputs.issue_number }}`.
3. Verify your changes according to the "## Verification tasks" section.
4. Open exactly one pull request for that branch using the `create-pull-request` safe output.
5. Preserve canonical issue linkage metadata for deterministic reruns by including `Closes #${{ needs.pre_activation.outputs.issue_number }}` in the PR body - this is the stable identifier that prevents duplicate PR creation on future workflow runs.
6. Keep the pull request labeled `code-factory`.

## Verification tasks

Run these steps **in order** before committing. Wait for each to complete fully.

1. **Format code first — before committing**: `make fmt` must succeed and produce no diff.
   This runs both `go fmt ./...` and `terraform fmt --recursive`. Both must pass cleanly.
2. **Lint**: `make check-lint` must succeed.
3. **Build**: `make build` must succeed. Wait for it to finish completely.
4. **Unit tests**: `go test ./...` must pass.
5. **Acceptance tests**: Run targeted acceptance tests against the live Elastic Stack. Use the connection variables shown in **Test environment**. If tests fail, check whether the failure is related to your changes before proceeding.

## Pull request contract

The linked pull request must:

- use branch `code-factory/issue-${{ needs.pre_activation.outputs.issue_number }}`
- be the only open `code-factory` pull request for this issue
- include explicit issue linkage via `Closes #${{ needs.pre_activation.outputs.issue_number }}` in the PR body
- include a valid `## Changelog` section (see **PR body** below) — the CI changelog check will block the PR if this is missing or malformed
- stay focused on implementing the triggering issue only

## PR body

The pull request body must include all three of the following blocks, in order:

### Changelog (required — CI validates this)

Select `Customer impact` based on the nature of the change:
- `fix` — user-visible bug fix
- `enhancement` — new capability or attribute exposed to users
- `breaking` — removes or incompatibly changes existing behaviour
- `none` — internal only (refactoring, test coverage, CI, docs)

Most code-factory issues fall into one of these categories:
- Refactoring / deduplication / test-coverage gaps → `none`
- User-visible bug fixes → `fix`
- New resource attributes or resources → `enhancement`

When in doubt, prefer `none` over inventing customer impact that isn't described in the issue.

```
## Changelog
Customer impact: <none|fix|enhancement|breaking>
Summary: <one-line description of user impact; omit if Customer impact is none>
```

For `breaking`, also add a `### Breaking changes` block with a prose description, terminated by `<!-- /breaking-changes -->`.

### Issue linkage (required — prevents duplicate PRs on rerun)

```
Closes #${{ needs.pre_activation.outputs.issue_number }}
```

### Detailed changes (required)

Describe intent, approach, notable design decisions, and follow-up work.

## Guardrails

- Do not re-check trigger eligibility, actor trust, or duplicate PR state; deterministic pre-activation already handled those checks.
- Run `make fmt` **before** committing — unformatted code will fail CI. `make check-lint` and `make build` must also succeed.
- Run targeted acceptance tests (`TF_ACC=1`) against the live Elastic Stack using the endpoints in **Test environment**.
- Do not open a second pull request for the same issue.
- Do not change the branch naming convention.
- Do not open issues in this workflow.
- Do not use another issue, pull request, or external request as the source of truth over the triggering issue.
- Do not omit or leave placeholder text in the `## Changelog` section — `Customer impact:` must be one of `none`, `fix`, `enhancement`, or `breaking`, and `Summary:` must be a real sentence unless `Customer impact: none`.
- If you cannot make progress safely, use `noop` with a concise explanation instead of opening an extra pull request.