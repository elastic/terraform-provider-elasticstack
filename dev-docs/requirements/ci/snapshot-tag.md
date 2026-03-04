# `snapshot-tag` — Workflow Requirements

Workflow implementation: `.github/workflows/snapshot-tag.yml`

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

- **[REQ-001] (Trigger)**: The workflow shall run on a daily `schedule` trigger (UTC).
- **[REQ-002] (Trigger)**: The workflow shall support manual runs via `workflow_dispatch`.
- **[REQ-003] (Source)**: The workflow shall create snapshot tags from the current `main` branch HEAD.
- **[REQ-004] (TagName)**: The workflow shall compute the tag name as `v0.0.0-YYYYMMDD` using the current date in UTC.
- **[REQ-005] (Idempotency)**: If the computed tag already exists in the remote repository, the workflow shall succeed without modifying any refs (no-op).
- **[REQ-006] (Concurrency)**: The workflow shall define GitHub Actions `concurrency` controls such that only one snapshot-tag run can execute at a time.
- **[REQ-007] (Safety)**: The workflow shall not force-update or move an existing tag.
- **[REQ-008] (TagType)**: When creating a snapshot tag, the workflow shall create an annotated tag and include a human-readable message.
- **[REQ-009] (Auth/Permissions)**: The workflow shall use the GitHub Actions token and request only the minimal permission needed to push tags (`contents: write`).
- **[REQ-010] (GitIdentity)**: When creating tags, the workflow shall set a deterministic git identity (e.g. `github-actions[bot]`) to make tag provenance clear.
- **[REQ-011] (Observability)**: The workflow shall log whether it created a tag or skipped because the tag already existed.
- **[REQ-012] (Downstream)**: The workflow shall only push the tag; any publishing of artifacts (build/sign/upload) shall be delegated to existing tag-driven release automation.

