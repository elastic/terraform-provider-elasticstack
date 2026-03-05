---
name: Schema Coverage Rotation
description: Rotates schema-coverage analysis across stale provider entities and opens actionable test-improvement issues.
on:
  workflow_dispatch:
  schedule:
    - cron: daily
engine: 
  id: copilot
  model: "claude-opus-4.6" 
permissions:
  contents: read
  issues: read
  pull-requests: read
  actions: read
tools:
  cache-memory:
    - id: default
      key: schema-coverage-rotation-memory
      retention-days: 90
  create-orphan: true
safe-outputs:
  create-issue:
    title-prefix: "[schema-coverage] "
    labels: [testing, acceptance-tests, schema-coverage]
    max: 3
  assign-to-agent:
    name: copilot
    model: "claude-sonnet-4.6" 
    custom-agent: acceptance-test-improver
    allowed: [copilot]
    target: "*"
    max: 3
    github-token: ${{ secrets.GH_AW_AGENT_TOKEN }}
---

# Schema Coverage Rotation Worker

You are responsible for running schema-coverage analysis on up to 3 provider entities per run, prioritizing the entities that have not been analyzed for the longest time, while keeping the total number of open `schema-coverage` issues capped at 3.

## Required inputs and references

- Skill instructions: `.agents/skills/schema-coverage/SKILL.md`
- Provider entity sources of truth:
  - `docs/resources/*.md` for resources
  - `docs/data-sources/*.md` for data sources
- Bootstrap memory seed: `.github/aw/memory/schema-coverage-rotation-entities-last-analysed.json`
- Persistent memory path: `/tmp/gh-aw/cache-memory/default/schema-coverage-rotation-entities-last-analysed.json`

## Memory format

Use this JSON object format:

```json
{
  "resources": {
    "elasticstack_example_resource": "2026-03-05T04:15:00Z"
  },
  "data-sources": {
    "elasticstack_example_data_source": null
  }
}
```

Timestamp value rules:
- RFC3339 UTC string when analyzed.
- `null` if known but never analyzed.

## Execution steps

1. Using GitHub search or API, count currently open issues (excluding pull requests) in this repository with label `schema-coverage` using a query such as `is:issue is:open label:"schema-coverage" repo:<this-repo>`, and calculate:
   - `open_schema_coverage_issues` = the count of matching issues
   - `issue_slots_available = max(0, 3 - open_schema_coverage_issues)`
2. If `issue_slots_available` is `0`:
   - Exit immediately.
   - Call `noop` with a short reason indicating the open-issue cap has been reached.
3. Read `.agents/skills/schema-coverage/SKILL.md` and follow it strictly when evaluating coverage.
4. Load `/tmp/gh-aw/cache-memory/default/schema-coverage-rotation-entities-last-analysed.json`.
   - If it does not exist, initialize it from `.github/aw/memory/schema-coverage-rotation-entities-last-analysed.json`.
5. Build the current canonical entity list from docs only:
   - Resources from `docs/resources/*.md` (exclude non-entity pages such as `index.md` if present).
   - Data sources from `docs/data-sources/*.md`.
   - Derive Terraform entity names from doc filenames (without extension), preserving entity type from directory (`resources` vs `data-sources`).
   - Merge and deduplicate names within each type.
   - Ensure discovered resources are present under `resources` with a timestamp or `null`.
   - Ensure discovered data sources are present under `data-sources` with a timestamp or `null`.
6. Select exactly `issue_slots_available` entities by oldest timestamp across both types (`null` first, then oldest datetime), while preserving each selected entity's type (`resource` or `data source`).
7. For each selected entity:
   - Perform schema coverage analysis using the skill rubric.
   - Determine whether there are actionable testing gaps.
   - Update the entity timestamp to the current UTC time after analysis, regardless of whether a gap exists.
8. Persist the updated memory file to `/tmp/gh-aw/cache-memory/default/schema-coverage-rotation-entities-last-analysed.json`.

## Issue creation rules

- Create one issue per analyzed entity only when actionable testing gaps exist.
- Never create more issues than `issue_slots_available` for the current run.

Issue content must include:
- Entity name.
- Entity type (`resource` or `data source`).
- Entity implementation directory path.
- Attributes with no coverage.
- Attributes with poor coverage.
- Prioritized top 5 gaps (or fewer if less exist).
- Concrete acceptance-test additions that would close those gaps.

Do NOT include instructions in the issue body that override the acceptance-test-improver agent's behavior (for example, do not tell it to skip tests, skip builds, or change its workflow).

For each issue created, you MUST call `assign-to-agent` with:
- `name: copilot`
- `custom_agent: acceptance-test-improver`
- the created issue number

If an analyzed entity has no actionable gaps, do not create an issue for it.

If at least one entity was analyzed but none has actionable gaps, you MUST call `noop` with a short reason.
