## Why

Some GitHub issues land too underspecified for `change-factory` to safely author an OpenSpec proposal: the capability area, target API surface, or "done" criteria aren't yet clear. Today the only path is for a maintainer to investigate by hand and rewrite the issue, or for `change-factory` to bail out with a `noop` and a clarifying comment. Neither makes durable progress, and the investigative work isn't captured in a form the next stage can reuse.

A new `research-factory` workflow inserts a deep-research pass *before* `change-factory`. It enriches the issue in place — appending a stable, human-readable research block to the issue body — comparing at least two approaches, surfacing open questions, and grounding decisions in the Elastic docs and a full repository checkout. The block becomes the authoritative scope when `change-factory` runs. Re-applying the trigger label re-runs the research with full awareness of any new comments or edits, so the issue makes progress over multiple loops instead of restarting cold.

## What Changes

- Add a new agentic workflow `research-factory` (label `research-factory`, branch-naming convention not applicable — this workflow does not open PRs).
- Triggers: `issues.opened` / `issues.labeled` for the `research-factory` label, plus `workflow_dispatch` accepting an `issue_number` input so a future issue classifier can dispatch research without applying the label.
- Per-issue concurrency: at most one running research session per issue; new triggers queue rather than cancel the in-flight run.
- The agent has a self-imposed 25-minute research budget, with a 35-minute job-level hard timeout to absorb setup, MCP latency, and post-processing.
- The agent has a full `main` checkout, the `elastic-docs` MCP server, and read-only access to issue comments. It must not write code or open PRs.
- The agent's sole durable output is the issue body: it appends or rewrites a single gated section delimited by `<!-- implementation-research:start -->` and `<!-- implementation-research:end -->` containing problem framing, ≥2 candidate approaches with pros/cons, a recommendation, and open questions.
- Re-runs synthesize the next block from the original issue + all comments + the current block contents. The block contract states the section is regenerated each run; users influence the next run by commenting or editing outside the block (and may edit inside, but those edits are read as input rather than preserved verbatim).
- Pre-activation extends the existing factory-intake plumbing to fetch the issue's full comment history (chronological, human-authored) so the agent has the full conversation as context.
- The workflow's framework status comment carries any "what changed this run" narrative; no separate `add-comment` is used for run summaries.
- **BREAKING for `change-factory` semantics**: when the research block is present in the issue body, `change-factory` SHALL treat it as the exclusive source of scope — adopting its `Recommendation` as the proposal spine and carrying its `Open questions` into `design.md`. When the block is absent, `change-factory` retains its current "title and body are authoritative" behavior unchanged. This is non-breaking for issues that bypass research-factory entirely.
- Out of scope for v1: the issue classifier itself, the `repro-factory` workflow, schema-level lint validation of the block, persistent state outside the issue body, and any modification to `code-factory`.

## Capabilities

### New Capabilities

- `ci-research-factory-issue-intake`: end-to-end requirements for the new agentic workflow — triggers, dispatch, intake gating, concurrency, timeout budget, prompt obligations (compare ≥2 approaches, time-box awareness, partial-output preference), Elastic docs MCP wiring, comment-history capture, label removal, status-comment behavior, and the `update-issue` safe output that produces the block.
- `ci-implementation-research-block-format`: the wire format of the in-body research block — exact marker comments, mandatory subsections (`Problem framing`, `Approaches considered` with ≥2 H4 children, `Recommendation`, `Open questions`), provenance header, and the regenerated-each-run social contract. Both the research-factory producer and the change-factory consumer are bound to this contract.

### Modified Capabilities

- `ci-change-factory-issue-intake`: change-factory must detect the research block when present and treat it as the exclusive authoritative source for scope (recommendation drives the proposal, open questions land in `design.md`), while preserving today's title-and-body-only behavior when the block is absent.

## Impact

- **New workflow files**: `.github/workflows-src/research-factory-issue/workflow.md.tmpl` plus its inline scripts directory, generating `.github/workflows/research-factory-issue.md` and `.github/workflows/research-factory-issue.lock.yml` via `make workflow-generate`.
- **Shared library extension**: `.github/workflows-src/lib/factory-issue-shared.js` is the natural home for any new helpers (issue-comment fetching, dispatch-issue resolution) so research-factory, code-factory, and a future repro-factory share intake logic. New helpers ship with unit tests in `.github/workflows-src/lib/`.
- **Modified workflow source**: `.github/workflows-src/change-factory-issue/workflow.md.tmpl` adds research-block-aware prompt instructions; the regenerated `.github/workflows/change-factory-issue.{md,lock.yml}` change accordingly.
- **Modified spec**: `openspec/specs/ci-change-factory-issue-intake/spec.md` gains a research-aware requirement.
- **GitHub repository configuration**: a new `research-factory` label needs to exist in the repo; coordination with maintainers, no code change required beyond documenting it.
- **No runtime impact** on the Terraform provider, generated clients, acceptance tests, or release flow. No new secrets. Reuses the existing `CLAUDE_LITELLM_PROXY_API_KEY` and `GITHUB_TOKEN`.
- **Network policy**: same allowlist as change-factory (`defaults`, `node`, `elastic.litellm-prod.ai`, `www.elastic.co`).
