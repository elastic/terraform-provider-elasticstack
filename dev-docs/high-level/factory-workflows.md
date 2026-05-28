# Factory workflows

This page describes the **label-triggered, issue-scoped** agentic workflows in CI. Each "factory" reacts to a dedicated label on a single GitHub issue, runs a deterministic pre-activation, then delegates work to an agent with bounded outputs (one PR, one sticky comment).

For the broader picture, see [`agentic-development-workflow.md`](./agentic-development-workflow.md). For scheduled scanners that file or fix issues automatically, see [`continuous-quality-workflows.md`](./continuous-quality-workflows.md). For the local OpenSpec loop that picks up after `change-factory`, see [`openspec-workflows.md`](./openspec-workflows.md).

## Front door: issue classifier

Source: [`.github/workflows/issue-classifier.md`](../../.github/workflows/issue-classifier.md)

Every new issue is read by the classifier on `issues.opened`. A daily scheduled run also picks up to **5** oldest untriaged open issues to catch anything that slipped through. The classifier applies exactly one routing label plus `triaged`:

| Label | Means | Typical next step |
|-------|-------|-------------------|
| `needs-research` | Specific feature request (named API or clear product area) | Maintainer applies `research-factory` |
| `needs-reproduction` | Bug with config, error, or repro steps | Maintainer applies `reproducer-factory` |
| `needs-spec` | Problem **and** solution design fully specified (rare) | Maintainer applies `change-factory` |
| `needs-human` | Vague, ambiguous, security, or meta | Maintainer judgment |

The classifier never applies factory trigger labels itself. A maintainer (or future automation) promotes the issue by applying `research-factory`, `reproducer-factory`, `change-factory`, or `code-factory`, or by dispatching the workflow with `workflow_dispatch` and the issue number.

## Trigger labels

| Label | Workflow source | Output |
|-------|-----------------|--------|
| `research-factory` | [`research-factory-issue.md`](../../.github/workflows/research-factory-issue.md) | Sticky implementation-research comment on the issue |
| `reproducer-factory` | [`reproducer-factory-issue.md`](../../.github/workflows/reproducer-factory-issue.md) | Sticky reproducer comment; optional reproduction PR |
| `change-factory` | [`change-factory-issue.md`](../../.github/workflows/change-factory-issue.md) | One OpenSpec proposal PR |
| `code-factory` | [`code-factory-issue.md`](../../.github/workflows/code-factory-issue.md) | One implementation PR that closes the issue |

> Maintainer setup: these labels must exist in repo settings before the workflows can be triggered by label events. See [`contributing.md`](./contributing.md#factory-workflow-labels-maintainers).

All factories also support `workflow_dispatch` with an `issue_number` input, which is how the continuous-quality scanners fan out to `code-factory` (see [`continuous-quality-workflows.md`](./continuous-quality-workflows.md)).

## Phase labels

After a factory runs, the issue carries exactly one `phase-*` label so the pipeline stage is visible from the issue list:

| Phase label | Set by |
|-------------|--------|
| `phase-research` | `research-factory` |
| `phase-reproduction` | `reproducer-factory` |
| `phase-specification` | `change-factory` |
| `phase-coding` | `code-factory` |

Behavior is pinned by the [`ci-factory-pipeline-phase-labels`](../../openspec/specs/ci-factory-pipeline-phase-labels/spec.md) spec.

## Shared mechanics

Every factory uses the same deterministic pre-activation pattern, implemented under [`.github/scripts/workflows/lib/factory-runners/`](../../.github/scripts/workflows/lib/factory-runners/):

1. **Qualify trigger** — confirms the label or dispatch input is valid.
2. **Capture context** — reads issue title, body, comments, and any prior sticky comment.
3. **Suppress duplicate PRs** — refuses to start when an open branch like `<factory>/issue-<N>` already has a linked PR.
4. **Sanitize context** — strips embedded instructions / prompt-injection patterns from issue content before handing it to the agent.
5. **Remove trigger label** and **set phase label**.
6. **Upload context artifact** that the agent job downloads.

The agent then runs against an LLM gateway model with the relevant tools (`elastic-docs` MCP, `github` gh-proxy toolset, sometimes a live Elastic Stack) and emits **bounded safe outputs** — usually at most one PR and one comment per run.

## `research-factory` — feature research

Adds a deep-research pass **before** `change-factory`. It compares at least two candidate approaches, surfaces open questions, and grounds decisions in Elastic documentation.

### Trigger

- Apply the `research-factory` label to an issue.
- Or dispatch the workflow with `issue_number`.

### Output: sticky comment

A single comment delimited by `<!-- gha-research-factory -->` containing problem framing, two or more candidate approaches, a recommendation, open questions, and references. The exact section list and metadata JSON shape are pinned by [`ci-research-factory-comment-format`](../../openspec/specs/ci-research-factory-comment-format/spec.md); the agent prompt is in [`research-factory-issue.md`](../../.github/workflows/research-factory-issue.md).

### Social contract

- The block is **regenerated on every run**. Edits you make inside it are read as input on the next re-run but are **not preserved verbatim**.
- For durable feedback, post a comment or edit content **outside** the block.
- To trigger a fresh research pass, re-apply the `research-factory` label.

### How `change-factory` consumes it

When the research comment is present, `change-factory` treats it as the **exclusive authoritative scope**:

- `### Recommendation` becomes the proposal spine.
- `### Open questions` is copied into `design.md` as `## Open questions`.
- `### Approaches considered` is treated as already-evaluated context.

If the comment is absent, `change-factory` falls back to issue title + body. `change-factory` never modifies the research comment.

### What it does not do

- Does not open pull requests.
- Does not write code.
- Does not apply `change-factory` — promotion is a human action.

## `reproducer-factory` — bug reproduction

Confirms a reported bug reproduces against a live Elastic Stack and (when it does) lands a failing acceptance test on a dedicated branch.

### Trigger

- Apply the `reproducer-factory` label to an issue.
- Or dispatch with `issue_number`.

### Outputs

Every run emits exactly one **sticky reproducer comment** (`<!-- gha-reproducer-factory -->`) with one of three outcomes:

| Outcome | Comment sections | PR? |
|---------|------------------|-----|
| **A — reproduced** | Summary, Root cause, Reproduction test, References, metadata | Yes, on `reproducer-factory/issue-<N>` containing only the reproduction test, body `Related to #<N>` |
| **B — cannot reproduce** | Summary, three concrete Investigation avenues | No |
| **C — appears fixed** | Summary, Evidence, References, metadata | No |

The comment format is pinned by [`ci-reproducer-factory-comment-format`](../../openspec/specs/ci-reproducer-factory-comment-format/spec.md).

### Test file placement

- **Default**: `internal/acctest/reproductions/issue_<N>_acc_test.go` defining `TestAccReproduceIssue<N>`.
- **Resource package** when the issue clearly identifies one Terraform resource (e.g. `internal/kibana/alertingrule/issue_<N>_acc_test.go`).

### Environment

Runs against a provisioned Elastic Stack inside the agent environment, talking to it via proxy ports `9201` (Elasticsearch) and `5602` (Kibana). Direct ports `9200`/`5601` are blocked by the AWF firewall.

### Time budget

Reserves ~55 minutes of agent work with a 65-minute hard kill. If time runs short, the agent prefers a partial-but-valid outcome-B comment over emitting `noop`.

## `change-factory` — OpenSpec proposal authoring

Authors exactly one OpenSpec change directory and opens it as a proposal PR.

### Trigger

- Apply the `change-factory` label.
- Or post a `/change-factory` slash command on the issue (captured as `human_direction`).
- Or dispatch with `issue_number`.

### Scope authority (priority order)

1. **Human direction** (slash command text) — final say, overrides everything below.
2. **Implementation-research comment** (`<!-- gha-research-factory -->`) — exclusive baseline when present.
3. **Issue title and body** — default.

`change-factory` never modifies the research comment.

### Output: proposal PR

- Branch: `change-factory/issue-<N>`
- Labels: `change-factory`, `no-changelog`
- Body: includes the literal phrase `Related to #<N>` (not `Closes`) so merging does not auto-close the issue
- Contents: **only** files under `openspec/changes/<change-id>/`:
  - `proposal.md`, `design.md`, `tasks.md`
  - `.openspec.yaml`
  - Delta specs under `specs/<capability>/spec.md`

The change is **apply-ready**: an implementer can start from `tasks.md` and delta specs without further research.

### Out of scope for this run

`change-factory` is proposal-only. It deliberately does **not** run `make build`, `go test`, acceptance tests, modify provider source, or touch generated clients.

## `code-factory` — direct implementation

Implements a single GitHub issue end-to-end on a dedicated branch and opens an implementation PR.

### Trigger

- Apply the `code-factory` label.
- Or dispatch with `issue_number` and `source_workflow` (used by the continuous-quality scanners).

### Intent

This path **bypasses OpenSpec** by design. It is intended for:

- Quality / refactor / test-gap issues filed by the [continuous-quality workflows](./continuous-quality-workflows.md).
- Mechanical bug fixes where no spec is impacted.

For new behavior or contract changes, use `change-factory` → local implementation loop instead (see [`openspec-workflows.md`](./openspec-workflows.md)).

### Output: implementation PR

- Branch: `code-factory/issue-<N>`
- Label: `code-factory`
- Body: includes `Closes #<N>` (deterministic linkage for duplicate-PR suppression on re-runs)

### Verification the agent runs before opening the PR

In order:

1. `make fmt` (must produce no diff)
2. `make check-lint`
3. `make build`
4. `go test ./...`
5. Targeted acceptance tests per [`testing.md`](./testing.md)

If `make fmt` produces a diff or any step fails, the agent fixes and re-runs rather than opening a broken PR.

## Pre-activation patterns worth knowing

A few recurring gates that show up across factories:

- **Trust gate** — issue events are trusted by virtue of the GitHub role check that delivered the event; `workflow_dispatch` paths bypass trust by design (only repo collaborators with `actions: write` can dispatch).
- **Duplicate PR suppression** — every factory checks for an existing open PR on its branch convention (`<factory>/issue-<N>`) and refuses to start a second run while one is open. This is intentional: the human-led implementation loop or PR review supersedes a fresh agent run.
- **Sanitization** — sanitized issue body and comment history are written to `/tmp/<factory>-context/` and the agent reads them from disk rather than from inline prompt expansion. Treat any prompt-injection in the issue body as data, never as instructions.
- **Safe outputs** — the workflow declares the maximum number of comments / PRs / labels the agent may emit; the runner enforces those caps regardless of what the agent tries.

## See also

- Local implementation loop after `change-factory`: [`openspec-workflows.md`](./openspec-workflows.md)
- Maintainer setup of factory and phase labels: [`contributing.md`](./contributing.md#factory-workflow-labels-maintainers)
