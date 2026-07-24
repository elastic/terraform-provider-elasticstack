---
imports: [shared/setup-dev.md]
name: Research Factory Issue Intake
timeout-minutes: 35
description: >-
  Reacts to trusted qualifying `research-factory` issue events or internal workflow dispatch
  requests and delegates deep-research authoring to an agent that creates or updates a single
  implementation-research sticky comment on the triggering issue.
on:
  issues:
    types: [opened, labeled]
  workflow_dispatch:
    inputs:
      issue_number:
        description: 'Issue number to research'
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
        FACTORY_NAME: research-factory
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
    - name: Fetch prior research comment
      id: fetch_prior_research_comment
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
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/research-factory/fetch-prior-research-comment.js');
          await fn({ github, context, core });
    - name: Remove trigger label
      id: remove_trigger_label
      if: >-
        steps.determine_intake_mode.outputs.intake_mode == 'issue-event' &&
        steps.qualify_trigger.outputs.event_eligible == 'true'
      uses: actions/github-script@v9.0.0
      env:
        FACTORY_NAME: research-factory
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
        PHASE_LABEL_NAME: phase-research
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
        PRIOR_RESEARCH_COMMENT: ${{ steps.fetch_prior_research_comment.outputs.prior_research_comment }}
        SOURCE_WORKFLOW: ${{ github.event.inputs.source_workflow }}
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
          echo "prior_research_comment<<${EOF_DELIM}"
          printf '%s\n' "${PRIOR_RESEARCH_COMMENT}"
          echo "${EOF_DELIM}"
        } >> "$GITHUB_OUTPUT"
      shell: bash
    - name: Finalize gate reason
      id: finalize_gate
      if: always()
      uses: actions/github-script@v9.0.0
      env:
        FACTORY_NAME: research-factory
        EVENT_ELIGIBLE: ${{ steps.normalize_context.outputs.event_eligible }}
        EVENT_ELIGIBLE_REASON: ${{ steps.normalize_context.outputs.event_eligible_reason }}
        ACTOR_TRUSTED: ${{ steps.normalize_context.outputs.actor_trusted }}
        ACTOR_TRUSTED_REASON: ${{ steps.normalize_context.outputs.actor_trusted_reason }}
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/finalize-gate.js');
          await fn({ github, context, core });
    - name: Write issue context to files
      id: write_context_files
      if: >
        steps.normalize_context.outputs.event_eligible == 'true' &&
        steps.normalize_context.outputs.actor_trusted == 'true' &&
        steps.normalize_context.outputs.issue_number != ''
      env:
        FACTORY_NAME: research-factory
        ISSUE_BODY: ${{ steps.normalize_context.outputs.issue_body }}
        ISSUE_COMMENTS: ${{ steps.normalize_context.outputs.issue_comments }}
        PRIOR_FACTORY_COMMENT: ${{ steps.normalize_context.outputs.prior_research_comment }}
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/lib/factory-runners/write-context-files.js');
          await fn({ github, context, core });
    - name: Upload issue context artifact
      if: >
        steps.normalize_context.outputs.event_eligible == 'true' &&
        steps.normalize_context.outputs.actor_trusted == 'true' &&
        steps.normalize_context.outputs.issue_number != ''
      uses: actions/upload-artifact@v7.0.1
      with:
        name: research-factory-issue-context
        path: /tmp/research-factory-context/
        if-no-files-found: error
env:
  RESEARCH_FACTORY_ISSUE_NUMBER: ${{ github.event.issue.number || inputs.issue_number }}
concurrency:
  group: research-factory-issue-${{ github.event.issue.number || inputs.issue_number }}
  cancel-in-progress: false
if: >-
  needs.pre_activation.outputs.event_eligible == 'true' &&
  needs.pre_activation.outputs.actor_trusted == 'true' &&
  needs.pre_activation.outputs.issue_number != ''
steps:
  - name: Download issue context artifact
    uses: actions/download-artifact@v8.0.1
    with:
      name: research-factory-issue-context
      path: /tmp/gh-aw/agent/
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
      event_eligible: ${{ steps.normalize_context.outputs.event_eligible }}
      event_eligible_reason: ${{ steps.normalize_context.outputs.event_eligible_reason }}
      actor_trusted: ${{ steps.normalize_context.outputs.actor_trusted }}
      actor_trusted_reason: ${{ steps.normalize_context.outputs.actor_trusted_reason }}
      trigger_label_removed: ${{ steps.normalize_context.outputs.trigger_label_removed }}
      trigger_label_removed_reason: ${{ steps.normalize_context.outputs.trigger_label_removed_reason }}
      source_workflow: ${{ steps.normalize_context.outputs.source_workflow }}
      gate_reason: ${{ steps.finalize_gate.outputs.gate_reason }}
tools:
  cli-proxy: true
  github:
    mode: gh-proxy
    toolsets: [issues, repos]
network:
  allowed: [defaults, node, elastic.litellm-prod.ai, www.elastic.co]
mcp-servers:
  elastic-docs:
    url: "https://www.elastic.co/docs/_mcp/"
    allowed: ["*"]
checkout:
  fetch-depth: 0
safe-outputs:
  jobs:
    update-research-comment:
      description: Create or update the implementation-research sticky comment on the triggering issue
      permissions:
        contents: read
        issues: write
      runs-on: ubuntu-latest
      output: "Research comment created or updated successfully."
      inputs:
        body:
          description: Markdown body of the research comment (without the gha-research-factory marker)
          required: true
          type: string
      steps:
        - name: Checkout repository
          uses: actions/checkout@v7.0.1
          with:
            persist-credentials: false
            fetch-depth: 1
        - name: Create or update research comment
          uses: actions/github-script@v9.0.0
          env:
            RESEARCH_FACTORY_ISSUE_NUMBER: ${{ github.event.issue.number || inputs.issue_number }}
          with:
            github-token: ${{ secrets.GITHUB_TOKEN }}
            script: |
              const fn = require('${{ github.workspace }}/.github/scripts/workflows/research-factory/update-research-comment.js');
              await fn({ github, context, core });
  noop:
    max: 1
    report-as-issue: false
---

# Research Factory issue research worker

You author the implementation-research output for a GitHub issue labeled `research-factory`. Your
only durable output is a single `update_research_comment` safe-output operation that creates or
updates a sticky comment on the triggering issue.

## Pre-activation context

- **Gate reason**: `${{ needs.pre_activation.outputs.gate_reason }}`
- **Intake mode**: `${{ needs.pre_activation.outputs.intake_mode }}`
- **Issue number**: `${{ needs.pre_activation.outputs.issue_number }}`
- **Issue title**: `${{ needs.pre_activation.outputs.issue_title }}`
- **Issue body**: see `/tmp/gh-aw/agent/issue_body.md`
- **Comment history**: see `/tmp/gh-aw/agent/issue_comments.md`
- **Prior research comment** (if any): see `/tmp/gh-aw/agent/prior_research_comment.md`

Read all files before proceeding. They may contain markdown, code fences, and other content that
cannot be safely embedded inline in a prompt.

- **Repository**: `${{ github.repository }}`
- **Triggered by**: `@${{ github.actor }}`
- **Run link**: `${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}`

## Time budget

You have approximately 25 minutes of agentic work. Reserve the last ~3 minutes for emitting your
`update_research_comment`. The job hard-kills at 35 minutes.

## Partial output preference

If you run short on time, prefer emitting a partial-but-valid research comment with explicit
unanswered open questions over emitting `noop`. A partial comment with honest unknowns is more useful
than silence.

## Elastic documentation

The `elastic-docs` MCP server is available with `search_docs`, `find_related_docs`, and
`get_document_by_url`. Use them to research unfamiliar API surface before authoring the comment. This
grounding step helps produce accurate comparisons and avoids speculative assumptions about API shape.

If the MCP tools are unavailable or return no useful results, proceed from the issue content alone —
do not block the run waiting for documentation.

## Comparison requirement

You SHALL compare at least two distinct candidate approaches under `### Approaches considered`. Each
approach needs its own `#### ` H4 heading. Do not emit a comment with only one approach.

## Research comment format

Your research output MUST conform to the `ci-research-factory-comment-format` capability. The
workflow automatically prepends the marker `<!-- gha-research-factory -->`; you do not need to
include it. Your comment body SHALL contain these mandatory sections in order:

1. `## Implementation research` — H2 heading followed by a provenance header recording the run
timestamp, the run link (`${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}`), and
this social-contract notice:

   > Edits inside this comment are read as input on the next run but are not preserved verbatim. For
   > durable feedback, post a comment or edit the issue body.

2. `### Problem framing` — Restate the requested change in concrete terms.
3. `### Approaches considered` — Two or more `#### ` H4 child headings, each describing a distinct
candidate approach with sketch, Terraform shape (when applicable), Elastic API surface (when
applicable), and pros/cons.
4. `### Recommendation` — Name exactly one of the approaches above as the chosen spine, with a brief
rationale.
5. `### Open questions` — A (possibly empty) bullet list of questions whose answers would change the
recommendation or scope.
6. `### Out of scope` — A (possibly empty) bullet list of items the recommendation explicitly
excludes.
7. `### References` — A list of consulted sources, including elastic-docs URLs and repository paths
inspected during research.

After `### References`, include a `<details>` element with `<summary>🤖 Pipeline metadata</summary>`
containing a fenced JSON block (language `json`) that conforms to the
`ci-research-factory-comment-format` schema:

- `schema_version` (string, required): e.g. `"1.0"`.
- `recommendation` (object, required):
  - `spine` (string, required): kebab-case identifier.
  - `confidence` (string, optional): `"high"`, `"medium"`, or `"low"`.
  - `approach_index` (number, required): zero-based index of the chosen approach.
- `open_questions` (array, optional): each with `id`, `text`, and `blocking` boolean.
- `affected_capabilities` (array of strings, optional).
- `estimated_scope` (string): `"small"`, `"medium"`, `"large"`, or `"unknown"`.
- `references` (array, optional): each with `type` and `url` or `path`.

Ensure the JSON metadata is internally consistent with the human-readable subsections above it.
The `<details>` element SHALL be closed by default so that human readers do not see the JSON unless
they expand it.

## Free-will semantics

Edits a user has made inside the prior research comment are read as input but are not preserved
verbatim. Synthesize the next comment from: original issue content + chronological comment history +
prior research comment contents (as draft input). Do not attempt to detect or diff what the user
changed inside the prior comment.

## Guardrails

- **SHALL NOT** modify repository files (no `git commit`, no file edits).
- **SHALL NOT** open pull requests.
- **SHALL NOT** post free-form comments (`add-comment` is not enabled).
- **SHALL NOT** add labels, including `change-factory` or `code-factory`.
- **SHALL NOT** call `update_research_comment` more than once.
- **SHALL NOT** re-check intake gates (deterministic pre-activation already handled those).
- If no meaningful research progress is possible (empty issue, no comments), emit `noop` with a brief
explanation.