## Context

The upstream `gh-aw` semantic function refactor workflow analyzes non-test Go files, clusters functions by purpose, identifies misplaced functions or duplicate implementations, and opens a refactoring issue. Its upstream source is `https://github.com/github/gh-aw/blob/main/.github/workflows/semantic-function-refactor.md`.

This repository already adapted the upstream duplicate-code detector into a local pattern:
- editable workflow source under `.github/workflows-src/`
- generated workflow markdown and compiled lock artifacts under `.github/workflows/`
- a LiteLLM-backed Claude engine configuration using `CLAUDE_LITELLM_PROXY_API_KEY`
- deterministic issue-slot gating before agent activation
- tests that assert the generated workflow and lock file preserve important contracts

The semantic refactor workflow should follow that local pattern rather than copying upstream verbatim. The biggest behavioral difference is issue lifecycle: upstream closes existing `[refactor]` issues and creates at most one new issue; this repository should instead keep open issues as the queue and fill available slots in a dedicated `semantic-refactor` bucket.

## Goals / Non-Goals

**Goals:**

- Add a generated GH AW workflow that performs scheduled and manual semantic refactor analysis for Go source files.
- Keep the workflow traceable to the upstream semantic function refactor source.
- Run the agent through the repository's LiteLLM engine configuration.
- Use the same deterministic open-issue slot model as the duplicate-code detector, keyed by the `semantic-refactor` label and capped at three open issues.
- Produce focused, actionable issues for distinct semantic refactor opportunities.
- Add focused tests for the workflow source, generated workflow, lock metadata, and issue-slot contract.

**Non-Goals:**

- Automatically changing repository code as part of the scheduled analysis workflow.
- Closing or replacing existing refactor issues before each run.
- Defining a comprehensive static analysis engine outside the GH AW prompt.
- Analyzing tests, generated code, workflow files, vendored dependencies, or non-Go files in the initial workflow.

## Decisions

### 1. Use repository workflow-source generation

The workflow will be authored at `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl`, registered in `.github/workflows-src/manifest.json`, generated to `.github/workflows/semantic-function-refactor.md`, and compiled to `.github/workflows/semantic-function-refactor.lock.yml`.

Why:
- This matches the duplicate-code detector and other repository workflow-source conventions.
- It keeps local adaptations reviewable while still committing executable generated artifacts.
- It allows `make workflow-generate`, `make workflow-test`, and `check-workflows` to detect drift.

Alternative considered:
- Copy upstream directly into `.github/workflows/`: rejected because generated artifacts in this repository are not the authoritative edit surface.

### 2. Configure Claude through LiteLLM

The workflow will use the existing local engine shape:

```yaml
engine:
  id: claude
  model: "llm-gateway/gpt-5.5"
  env:
    ANTHROPIC_BASE_URL: "https://elastic.litellm-prod.ai/"
    ANTHROPIC_API_KEY: ${{ secrets.CLAUDE_LITELLM_PROXY_API_KEY }}
```

Why:
- It matches the current duplicate-code detector, schema-coverage rotation, and other recently updated workflows.
- It routes execution through the repository-approved LiteLLM proxy and secret.

Alternative considered:
- Keep upstream `engine: claude`: rejected because it does not encode this repository's LiteLLM routing or model configuration.

### 3. Replace upstream issue closing with deterministic issue slots

Before agent activation, the workflow will count open GitHub issues with the `semantic-refactor` label, subtract that count from a cap of `3`, expose `open_issues`, `issue_slots_available`, and `gate_reason` as pre-activation outputs, and skip the agent job when no slots remain.

Why:
- It uses the same operating model as the duplicate-code detector.
- Open issues become the durable queue of pending semantic refactor work.
- Maintainers can control capacity by closing or relabeling issues.

Alternative considered:
- Preserve upstream's "close existing `[refactor]` issues first" step: rejected because it discards pending work and conflicts with the requested daily cap behavior.

### 4. Use a dedicated semantic-refactor issue bucket

The workflow will use:
- title prefix: `[semantic-refactor] `
- gating label: `semantic-refactor`
- safe-output labels: `semantic-refactor`, `refactoring`, `code-quality`, `automated-analysis`
- safe-output max: `3`

Why:
- A dedicated bucket avoids colliding with general refactor issues.
- The same label drives both capacity and discoverability.
- The title prefix makes automated output easy to filter.

Alternative considered:
- Reuse upstream `[refactor]`: rejected because the bucket would be too broad for deterministic capacity control.

### 5. Create one issue per distinct semantic refactor opportunity

The prompt will instruct the agent to create separate issues for distinct high-value findings, capped by `issue_slots_available`. Examples include misplaced functions, duplicate or near-duplicate functions, scattered helpers, or small clusters of functions that should be extracted or moved together.

Why:
- Separate issues are easier to triage and assign than one broad audit report.
- The issue-slot cap only works well when one issue maps to one actionable unit of work.
- This matches the duplicate-code detector's "one finding per issue" adaptation.

Alternative considered:
- Keep upstream's single comprehensive issue: rejected because it conflicts with the requested "up to 3 issues" workflow and tends to mix unrelated refactors.

## Risks / Trade-offs

- [The workflow may generate low-signal organization suggestions] -> Mitigation: require concrete evidence, meaningful impact, and actionable recommendations before creating an issue.
- [The dedicated label can drift from the title prefix or safe-output labels] -> Mitigation: add workflow-source tests that assert all three remain aligned.
- [Open issues can block new analysis if stale issues remain open] -> Mitigation: this is intentional queue behavior; maintainers can close or relabel stale issues to free capacity.
- [Generated workflow artifacts can drift from source] -> Mitigation: rely on `make workflow-generate`, `check-workflows`, and focused tests.
- [The LiteLLM engine contract can regress during workflow compilation] -> Mitigation: assert generated workflow and lock metadata include the expected model, base URL, and secret wiring.

## Migration Plan

1. Add the OpenSpec capability for the semantic refactor workflow.
2. Add the workflow source template, issue-slot inline script, and manifest registration.
3. Generate `.github/workflows/semantic-function-refactor.md` and `.github/workflows/semantic-function-refactor.lock.yml` with `make workflow-generate`.
4. Add workflow-source tests covering the upstream baseline reference, `semantic-refactor` issue-slot gate, safe-output labels and cap, LiteLLM engine configuration, and prompt issue-creation contract.
5. Run `make workflow-test`, `make workflow-generate`, `make check-workflows`, and OpenSpec validation.

Rollback is straightforward: remove the workflow source, generated artifacts, manifest entry, and tests before the change is archived. If the workflow has already run, existing `semantic-refactor` issues can be closed or relabeled manually.

## Open Questions

- None. The issue bucket is `semantic-refactor`.
