# `snapshot-tag` — Workflow Requirements

Workflow implementation: `.github/workflows/snapshot-tag.yml`

## Purpose

Define scheduled and manual snapshot tagging from `main`, including tag naming, idempotency, concurrency, and provenance.

## Schema

```yaml
on:
  schedule:
    - cron: "<daily cron expression (UTC)>"
  workflow_dispatch: {}

permissions:
  contents: write
```

## Requirements

### Requirement: Schedule and manual triggers (REQ-001–REQ-002)

The workflow SHALL run on a daily `schedule` trigger (UTC). The workflow SHALL support manual runs via `workflow_dispatch`.

#### Scenario: Scheduled run

- GIVEN the cron schedule fires
- WHEN the workflow runs
- THEN it SHALL execute the snapshot-tag logic

### Requirement: Tag source and name (REQ-003–REQ-004)

The workflow SHALL create snapshot tags from the current `main` branch HEAD. The workflow SHALL compute the tag name as `v0.0.0-YYYYMMDD-<git short sha>` using the current date in UTC.

#### Scenario: New snapshot tag

- GIVEN `main` at a known commit
- WHEN the workflow creates a tag
- THEN the tag name SHALL match the required pattern with UTC date and short SHA

### Requirement: Idempotency and concurrency (REQ-005–REQ-006)

If the computed tag already exists in the remote repository, the workflow SHALL succeed without modifying any refs (no-op). The workflow SHALL define GitHub Actions `concurrency` controls such that only one snapshot-tag run can execute at a time.

#### Scenario: Tag already exists

- GIVEN the computed tag exists remotely
- WHEN the workflow runs
- THEN it SHALL succeed without moving or recreating the tag

### Requirement: Safety and tag type (REQ-007–REQ-008)

The workflow SHALL not force-update or move an existing tag. When creating a snapshot tag, the workflow SHALL create an annotated tag and include a human-readable message.

#### Scenario: Create annotated tag

- GIVEN a new tag is created
- WHEN the workflow completes
- THEN the tag SHALL be annotated with a clear message

### Requirement: Auth, identity, and observability (REQ-009–REQ-011)

The workflow SHALL use the GitHub Actions token and request only the minimal permission needed to push tags (`contents: write`). When creating tags, the workflow SHALL set a deterministic git identity (e.g. `github-actions[bot]`) to make tag provenance clear. The workflow SHALL log whether it created a tag or skipped because the tag already existed.

#### Scenario: Minimal permissions

- GIVEN the workflow runs on GitHub Actions
- WHEN permissions are evaluated
- THEN only `contents: write` SHALL be required for tagging

### Requirement: Downstream delegation (REQ-012)

The workflow SHALL only push the tag; any publishing of artifacts (build/sign/upload) SHALL be delegated to existing tag-driven release automation.

#### Scenario: No artifact publish in this workflow

- GIVEN the workflow completes after pushing a tag
- WHEN downstream automation runs
- THEN build/sign/upload SHALL not be inlined in this workflow’s responsibility
