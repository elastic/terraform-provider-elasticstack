## Context

`ci-aw-openspec-verification` currently treats any added file under `openspec/changes/<id>/` as an ineligible case. That makes the deterministic selector simple, but it prevents maintainers from using the workflow to review a newly introduced change proposal before it has any follow-up modification cycle. The requested behavior changes two parts of that contract at once: the selector must accept a single active change even when that change is new in the PR, and the downstream review path must know whether approval is allowed without re-deriving PR file state inside the agent prompt.

The existing workflow already centralizes change selection in `.github/workflows-src/lib/select-change.js` and exposes scalar pre-activation outputs to the agent job. That makes this change a good fit for extending the deterministic result object rather than asking the agent to inspect file statuses again.

## Goals / Non-Goals

**Goals:**
- Allow the workflow to verify pull requests that touch exactly one non-archive OpenSpec change even when that change includes added files.
- Determine before agent reasoning whether the selected change is approval-eligible or restricted to comment-only.
- Require the review body to clearly explain when a net-new change proposal satisfied normal quality gates but still cannot be approved.
- Preserve the existing ineligible behavior for multi-change PRs and unsupported file statuses such as `removed` or `renamed`.

**Non-Goals:**
- Changing how the agent performs OpenSpec verification once a change has been selected.
- Allowing approval for PRs that introduce any net-new files under the selected active change.
- Expanding workflow scope beyond the `ci-aw-openspec-verification` contract.

## Decisions

Use deterministic review eligibility outputs in addition to the selected change id.
The selector should continue to identify a single active change from PR files, but it should also publish a second classification that tells the agent whether the run is `approval-eligible` or `comment-only`. A comment-only result should be emitted whenever the selected change includes one or more `added` files; a fully approval-eligible result should be emitted only when every relevant file for the selected change is `modified`.

Alternative considered: let the agent infer comment-only status from the diff.
Rejected because the user explicitly wants this limitation decided by deterministic preconditions, and duplicating file-status logic in the prompt would reintroduce ambiguity.

Allow exactly one active change id with statuses limited to `added` and `modified`.
The selection logic should still reject zero active changes, multiple active change ids, and unsupported statuses. Within a single selected change, `added` and `modified` should both be allowed so that a new proposal can still be reviewed if the PR also updates related files under the same change directory.

Alternative considered: allow only all-added net-new changes and reject mixed added/modified status sets.
Rejected because the PR diff represents the whole branch relative to base, and a legitimate net-new proposal may contain both added and modified files by the time maintainers request verification.

Drive the review outcome contract from deterministic gating.
The prompt should consume the new pre-activation outputs and state that only `approval-eligible` runs may submit `APPROVE`. When the run is marked comment-only, the agent should still perform normal verification and relevance review, but it must submit `COMMENT` even if there are zero CRITICAL issues and zero `unassociated` files.

Alternative considered: keep the current APPROVE rule and add a prose warning for net-new proposals.
Rejected because it leaves room for accidental approval and does not satisfy the requirement that the limitation be computed before agent reasoning.

Require an explicit comment-only explanation in the review body.
When deterministic gating marks the run comment-only because the selected change contains added files, the prompt should require a short explanation such as "This PR meets the approval criteria but is limited to comment only due to implementing a net-new spec change." This keeps the outcome understandable to maintainers and contributors.

Alternative considered: rely on the generic selection reason output alone.
Rejected because the user wants the limitation explained in the review comment itself, not only in workflow logs or hidden step outputs.

## Risks / Trade-offs

- Broader eligibility could let more PRs reach the agent job -> Keep the one-change restriction and continue rejecting multiple ids and unsupported statuses.
- Mixed added/modified files may blur whether a change is "new" -> Treat any added file under the selected active change as comment-only to keep the rule deterministic and conservative.
- Review guidance could drift from deterministic selector behavior -> Use explicit pre-activation outputs for review mode and reason, then reference those outputs directly in the prompt.
- Maintainers may expect archive-on-comment for clean net-new proposals -> Preserve the existing archive-after-APPROVE-only rule so comment-only runs never archive.

## Migration Plan

- Extend `.github/workflows-src/lib/select-change.js` and its tests to classify single-change PRs as approval-eligible, comment-only, or ineligible.
- Surface the new review-mode outputs through the workflow source and compiled artifacts.
- Update the prompt so the agent explains deterministic comment-only limitations and never approves or archives a net-new change proposal.
- Recompile the workflow lock file and run the relevant OpenSpec and workflow validation checks.

## Open Questions

- None.
