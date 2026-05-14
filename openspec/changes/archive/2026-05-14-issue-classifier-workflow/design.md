## Context

The repository has a growing suite of GitHub Agentic Workflows (gh-aw) that automate issue lifecycle: `research-factory` researches feature requests, `reproducer-factory` reproduces bugs, `change-factory` generates implementation specs. All three consume issues via label-triggered intake. Without automatic triage, issues accumulate unlabeled and the factory pipelines never fire.

The existing factory workflows establish patterns this workflow must follow: YAML frontmatter + markdown agent instructions, pre-activation deterministic JS gates, safe-outputs for all GitHub mutations, and the `sanitizeUserContent` library for adversarial input handling.

## Goals / Non-Goals

**Goals:**
- Classify every new issue within minutes of filing (issue-opened trigger)
- Drain a backlog of at most 5 untriaged issues per day (scheduled trigger)
- Apply `triaged` + one `needs-*` label and post a classification comment per issue
- Remain idempotent: never re-classify an issue that already has `triaged`

**Non-Goals:**
- Closing or re-opening issues
- Classifying pull requests
- Pre-activation sanitization pipeline (Path A) — deferred; agent fetches via GitHub MCP (Path B)
- Additional categories beyond the four (`needs-research`, `needs-reproduction`, `needs-spec`, `needs-human`)

## Decisions

### D1: Dual trigger (issue-opened + schedule) over cron-only

**Decision**: Trigger on `issues: [opened]` for immediate classification and `schedule: daily` as a backlog sweep.

**Rationale**: A cron-only approach introduces up to 24-hour lag on new issues. The event trigger eliminates this without adding complexity — each trigger mode has its own pre-activation path. The daily sweep catches issues that slipped through (e.g., during a failing run or a burst that overwhelmed the event queue).

**Alternative considered**: Cron-only with a 48-hour window for "new" issues. Rejected: unnecessary complexity and lag.

### D2: Single workflow file with mode detection

**Decision**: One `workflow.md.tmpl` that detects trigger mode in pre-activation (`event`, `scheduled`, `dispatch`) rather than two separate workflow files.

**Rationale**: The classification logic and safe-outputs config are identical across all modes. Splitting would duplicate the agent prompt and safe-outputs definition. The pattern is established by existing factory workflows that handle both `issues:` events and `workflow_dispatch` in a single file.

### D3: Path B sanitization (agent fetches via GitHub MCP)

**Decision**: Agent fetches issue body via GitHub MCP tools. No pre-activation sanitization pipeline.

**Rationale**: The classifier makes a labeling decision — it does not execute code, open PRs, or produce artifacts derived from issue content. The injection surface is meaningfully lower than the factory workflows. Path A (pre-activation fetch + `sanitizeUserContent` + `/tmp/` files) can be added later if prompt injection becomes a concern.

**Alternative considered**: Full Path A pipeline matching factory workflow pattern. Deferred to a follow-up change.

### D4: No custom concurrency configuration

**Decision**: Rely on gh-aw auto-generated concurrency groups.

**Rationale**: gh-aw automatically assigns `gh-aw-{workflow}-{issue.number}` for issue-event triggers (per-issue serialisation) and `gh-aw-{workflow}` for scheduled triggers (single backlog run). These are exactly the semantics needed with no explicit configuration.

### D5: `hide-older-comments: true` on `add-comment`

**Decision**: Use `hide-older-comments: true` with `allowed-reasons: [outdated]` and `footer: false`.

**Rationale**: If a classification is corrected and the workflow re-runs, the old comment should be hidden rather than leaving a confusing thread. The `footer: false` avoids the generic "AI-generated" attribution on a public-facing comment that already explains its own origin via the `<!-- gha-issue-classifier -->` marker.

### D6: Fixed cap of 5 issues for scheduled path

**Decision**: Scheduled path processes up to 5 untriaged issues per run, selected newest-first. No age filter.

**Rationale**: With the event trigger handling new issues, the scheduled path is purely a backlog drain. Age filtering is unnecessary — the `triaged` label is the idempotency guard. A fixed cap of 5 limits cost per run without letting backlogs grow unbounded (5/day will clear a backlog of 35 in a week).

## Risks / Trade-offs

**[Risk] Overlap between event trigger and scheduled trigger on the same issue** → `add-labels` is idempotent; `hide-older-comments: true` means a second comment replaces the first. Acceptable.

**[Risk] Misclassification by the agent** → Classification comment explicitly invites correction. Maintainers can remove the `needs-*` label and re-triage manually. Low blast radius.

**[Risk] Prompt injection via issue content** → Path B defers sanitization. Mitigated by: (1) agent is read-only and has no direct write access, (2) safe-outputs allowlist restricts labels to the five valid values, (3) comment content comes from the agent's own reasoning, not echoed from the issue. Acceptable for v1.

**[Risk] Backlog growth during periods of high issue volume** → Daily cap of 5 means a large backlog drains slowly. Could increase cap or add a second scheduled run per day if this becomes a problem.
