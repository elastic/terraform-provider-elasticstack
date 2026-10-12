## Why

`research-factory` produces its implementation-research comment in a single pass. Nothing objectively assesses whether that research is good enough to become an OpenSpec proposal, so a maintainer has to read and judge every research comment before deciding whether to apply `change-factory`. A model grading its own output inflates its assessment, so self-review inside the same pass is not a substitute.

This matters now because research is about to be requested automatically at volume (see the `kibana-api-gap-scanner` change). Without an objective quality bar, automated research either reaches maintainers half-finished or forces them to review every comment by hand, which defeats the purpose of the automation.

## What Changes

- Every `research-factory` run (human-requested or dispatched by another workflow) iterates its research through a propose -> independent critique -> revise loop instead of a single pass.
- The critique is performed by a critic that is independent of the research author: a separate critic role running on a different model.
- Research is judged against a two-part **done gate**:
  - **Hard checklist** (all must pass): *Grounded* (every claimed capability cites a verifiable source: an Elastic Stack API specification node, Elastic documentation, or existing client or provider code), *Mapped* (each new capability has a proposed Terraform schema mapping), *Compatible* (additive vs breaking assessed, with a migration note when breaking), *Versioned* (minimum Stack version and gating strategy named), *Testable* (acceptance-test outline including unset/empty and update cases), *Idiomatic* (design follows existing repository patterns rather than inventing new ones).
  - **Convergence signal**: the critic's quality score is at or above a threshold for consecutive rounds, or the critic has no further actionable feedback.
- The loop is bounded by a maximum number of rounds. Reaching the bound without satisfying the gate fails safe.
- Each run ends in exactly one gate outcome, made visible on the issue:
  - converged, checklist green, critic available, and no blocking open questions -> the issue is labelled `ready-for-change-factory`;
  - otherwise, including runs that hit the round bound, the session time box, or only partially complete -> the issue is labelled `research-needs-human`.
- The outcome label is derived from the published research metadata by a fixed rule, not taken from the agent's own claim. Contradictory, missing, or malformed metadata fails safe to `research-needs-human`, and the comment is corrected so it never disagrees with the label.
- When a new research run starts, any outcome label from a previous run is removed, so a stale outcome is never visible.
- The research comment gains a visible quality-gate section (outcome, checklist, score, outstanding critic feedback), and its machine-readable metadata records the gate outcome, checklist results, final critic score, and number of rounds, so downstream consumers and maintainers can see why research did or did not pass.
- The research session's time box grows to fit the bounded critique rounds.
- `research-factory` still never promotes an issue to `change-factory`; the `ready-for-change-factory` label is a signal to maintainers, not a trigger.
- **Out of scope** (explicit non-goals):
  - Automatically applying `change-factory` (or dispatching it) when the gate passes. Promotion stays a human decision until the critic's pass rate has been observed against maintainer judgement.
  - A critique loop that spans multiple workflow runs (for example, re-triggering research via labels between rounds).
  - Changes to `change-factory`, `code-factory`, or `reproducer-factory` behaviour.
  - Changes to how research is requested (labels, dispatch inputs, trust checks).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ci-research-factory-issue-intake`: pre-activation also clears stale outcome labels; the research-only requirement allows outcome-label changes as a durable output; the research session gains a bounded, independent critique loop and a done gate; the time box grows; the workflow applies exactly one gate-outcome label (`ready-for-change-factory` or `research-needs-human`); the "does not promote the issue to a downstream factory" requirement changes. Today it forbids the workflow from applying labels at all; it will allow exactly the two gate-outcome labels while still forbidding any factory trigger label.
- `ci-research-factory-comment-format`: a new quality-gate subsection; the JSON metadata block gains gate outcome, checklist results, critic scores, round count, and model identifiers; the done-gate rules and the rule deriving the outcome label from metadata are defined here.


## Impact

- `.github/workflows/research-factory-issue.md` and its compiled lock file: pre-activation (stale outcome-label removal), agent prompt, critic configuration, time box, and the custom comment script (outcome derivation and labelling).
- Research comment format consumers: `change-factory` reads the research comment; new metadata fields must be additive so it keeps working.
- Repository labels: `ready-for-change-factory` and `research-needs-human` must be provisioned. Both are new and owned by research-factory. They are deliberately separate from the issue classifier's `needs-human`, which keeps its meaning and is never touched by research-factory. Maintainer docs list the required labels.
- LLM gateway usage: an additional model is used for critique; per-run cost and duration increase.
- Docs: `dev-docs/high-level/factory-workflows.md` (research-factory section).
