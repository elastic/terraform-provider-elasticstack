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

You are responsible for running schema-coverage analysis on 3 provider entities per run, prioritizing the entities that have not been analyzed for the longest time.

## Required inputs and references

- Skill instructions: `.agents/skills/schema-coverage/SKILL.md`
- Provider entity sources of truth:
  - `provider/plugin_framework.go` (`resources()` and `dataSources()` lists)
  - `provider/provider.go` (`ResourcesMap` and `DataSourcesMap` lists)
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

1. Read `.agents/skills/schema-coverage/SKILL.md` and follow it strictly when evaluating coverage.
2. Load `/tmp/gh-aw/cache-memory/default/schema-coverage-rotation-entities-last-analysed.json`.
   - If it does not exist, initialize it from `.github/aw/memory/schema-coverage-rotation-entities-last-analysed.json`.
3. Build the current canonical entity list from both provider implementations:
   - Resources:
     - `provider/plugin_framework.go` `resources()` registrations.
     - `provider/provider.go` `ResourcesMap` registrations.
   - Data sources:
     - `provider/plugin_framework.go` `dataSources()` registrations.
     - `provider/provider.go` `DataSourcesMap` registrations.
   - Merge and deduplicate names within each type.
   - Ensure discovered resources are present under `resources` with a timestamp or `null`.
   - Ensure discovered data sources are present under `data-sources` with a timestamp or `null`.
4. Select exactly 3 entities by oldest timestamp across both types (`null` first, then oldest datetime), while preserving each selected entity's type (`resource` or `data source`).
5. For each selected entity:
   - Perform schema coverage analysis using the skill rubric.
   - Determine whether there are actionable testing gaps.
   - Update the entity timestamp to the current UTC time after analysis, regardless of whether a gap exists.
6. Persist the updated memory file to `/tmp/gh-aw/cache-memory/default/schema-coverage-rotation-entities-last-analysed.json`.

## Issue creation rules

Create one issue per analyzed entity only when actionable testing gaps exist.

Issue content must include:
- Entity name.
- Entity type (`resource` or `data source`).
- Entity implementation directory path.
- Attributes with no coverage.
- Attributes with poor coverage.
- Prioritized top 5 gaps (or fewer if less exist).
- Concrete acceptance-test additions that would close those gaps.

For each issue created, you MUST call `assign-to-agent` with:
- `name: copilot`
- `custom_agent: acceptance-test-improver`
- the created issue number

If an analyzed entity has no actionable gaps, do not create an issue for it.

If none of the 3 analyzed entities has actionable gaps, you MUST call `noop` with a short reason.
