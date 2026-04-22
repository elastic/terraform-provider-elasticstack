## Why

The `openspec-verify-label` workflow currently rejects any pull request that adds files under an active change, which blocks maintainers from using the workflow to review a newly proposed OpenSpec change before implementation starts. That is too restrictive for the common case of a PR that introduces exactly one new change proposal, but allowing a full approval path for a net-new proposal would let the workflow archive or bless a spec change that has not yet gone through the normal modification cycle.

## What Changes

- Update deterministic change selection so the workflow accepts pull requests that touch exactly one active OpenSpec change when every relevant file under that change (non-archive) has status `added` or `modified`—including mixes of added and modified files within the same change.
- Require deterministic pre-activation outputs to distinguish approval-eligible modified-only runs from comment-only runs (when the selected change includes any `added` file), while treating multiple active change ids or unsupported file statuses as ineligible.
- Constrain the review decision so net-new change proposals are comment-only even when verification finds no blocking issues, with that limitation decided by deterministic gating rather than left to agent judgment.
- Update the review guidance so comment-only reviews for net-new change proposals explain that the PR met the usual approval criteria but remained limited because it introduces a new spec change.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-aw-openspec-verification`: allow verification of exactly one added-or-modified active change while deterministically forcing COMMENT-only review outcomes for net-new change proposals

## Impact

- `.github/workflows-src/openspec-verify-label/workflow.md.tmpl`
- `.github/workflows-src/lib/select-change.js`
- `.github/workflows-src/lib/select-change.test.mjs`
- `.github/workflows-src/lib/openspec-verify-label.test.mjs`
- `.github/workflows-src/openspec-verify-label/scripts/select_change.inline.js` and any related workflow prompt text
- `.github/workflows/openspec-verify-label.md`
- `.github/workflows/openspec-verify-label.lock.yml`
- `openspec/specs/ci-aw-openspec-verification/spec.md`
