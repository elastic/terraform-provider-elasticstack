# `ci-workflow-shared-imports` — parameterized shared GH AW workflow imports

## Purpose

When multiple GitHub Agentic Workflows share identical or near-identical pre-activation logic, prompt text, or frontmatter blocks, the repository SHALL use a single parameterized shared import component rather than copy-pasting the logic across individual workflow sources. This reduces drift, centralizes maintenance, and preserves identical runtime behavior across all consumers.

## ADDED Requirements

### Requirement: Shared pre-activation workflow components are parameterized

When two or more Agentic Workflows share the same pre-activation steps, pre-activation job outputs, and agent prompt context, the repository SHALL extract that common material into a single shared workflow component under `.github/workflows-src/shared/` with an `import-schema` defining the parameters that vary across consumers. Each consumer SHALL import the shared component via `imports: - uses:` in its frontmatter, passing only the parameters that differ.

#### Scenario: Multiple workflows gate by the same issue-slot mechanism
- **WHEN** three or more workflows each run `actions/github-script@v9` to count open labeled issues, compute `cap - open`, expose `issue_slots_available` as a pre-activation output, skip the agent job when the value is `0`, and inject a `## Pre-activation context` block into the agent prompt
- **THEN** the repository SHALL extract those steps, jobs, and prompt text into one shared workflow component parameterized by `label` and `cap`
- **AND** each workflow SHALL import that component via `uses:` with its own `label` and `cap`, removing all locally duplicated frontmatter and prompt text

### Requirement: Generated shared workflow artifacts are compiled before consumers

The workflow source compiler SHALL produce each generated shared workflow artifact before the GitHub Agentic Workflows compiler resolves import references. The repository manifest SHALL list shared workflow templates so that `make workflow-generate` emits them into `.github/workflows/shared/`.

#### Scenario: Shared component is listed in the manifest
- **WHEN** a maintainer adds a new shared workflow template under `.github/workflows-src/shared/`
- **THEN** the shared component SHALL have an entry in `.github/workflows-src/manifest.json` mapping its template path to an output path under `.github/workflows/shared/`
- **AND** running the workflow source compiler SHALL emit the generated markdown to that output path, creating any necessary parent directories
