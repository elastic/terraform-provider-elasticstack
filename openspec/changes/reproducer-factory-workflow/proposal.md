## Why

Bug reports arrive without a systematic way to confirm whether the described failure is reproducible in the current provider, already fixed, or fundamentally ambiguous. Triaging these issues manually is slow and inconsistent — an automated workflow that attempts reproduction and records its findings would compress triage time and produce durable evidence for both open bugs and regressions.

## What Changes

- Add a new `reproducer-factory` GitHub agentic workflow that triggers on the `reproducer-factory` issue label
- The agent reads the issue, locates the relevant provider code, writes a `TestAccReproduceIssue{N}` acceptance test that asserts the failure condition (`ExpectError` / `ExpectNonEmptyPlan`), runs it, and routes to one of three outcomes:
  - **Reproduced**: test passes (bug confirmed) → creates PR on `reproducer-factory/issue-{n}` + sticky comment with root cause analysis
  - **Cannot reproduce**: test cannot be written or run cleanly → sticky comment with 3 specific, codebase-referenced investigation avenues
  - **Appears fixed**: test runs without triggering the expected error → sticky comment with evidence (test output + relevant git history)
- The sticky comment uses marker `<!-- gha-reproducer-factory -->` and is always created or updated (never silent)
- A PR is only created when the reproducing test passes; no PR is emitted for the other two outcomes
- Test file placement: inside the resource's own package when the issue clearly identifies a resource, otherwise in `internal/acctest/reproductions/`

## Capabilities

### New Capabilities

- `ci-reproducer-factory-issue-intake`: Workflow trigger, pre-activation gates (event eligibility, actor trust, duplicate PR suppression), context normalisation, and artifact upload for the reproducer-factory workflow — mirrors the intake pattern of `ci-research-factory-issue-intake` and `ci-code-factory-issue-intake`
- `ci-reproducer-factory-comment-format`: Schema and content rules for the sticky `<!-- gha-reproducer-factory -->` comment, covering all three outcome variants (reproduced / cannot reproduce / appears fixed) and the pipeline metadata JSON block

### Modified Capabilities

## Impact

- New source tree at `.github/workflows-src/reproducer-factory-issue/` (workflow template, intake-constants, inline scripts)
- New compiled lock file at `.github/workflows/reproducer-factory-issue.lock.yml` (generated via `make workflow-generate`)
- New `internal/acctest/reproductions/` directory (created on first use by the agent)
- Requires the `reproducer-factory` GitHub label to exist in the repository
- Requires the same secrets already used by code-factory: `CLAUDE_LITELLM_PROXY_API_KEY`, `GH_AW_GITHUB_MCP_SERVER_TOKEN`, `GH_AW_GITHUB_TOKEN`, `GH_AW_CI_TRIGGER_TOKEN`
- No changes to existing workflows, specs, or provider source code
