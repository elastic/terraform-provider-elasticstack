## 1. Update deterministic change selection

- [x] 1.1 Extend `.github/workflows-src/lib/select-change.js` so a PR touching exactly one active non-archive change remains eligible when its files are limited to `added` and `modified`, while continuing to reject multiple change ids and unsupported statuses.
- [x] 1.2 Publish deterministic review-disposition data for the selected change, including a human-readable disposition reason, so downstream workflow steps can distinguish approval-eligible modified-only runs from comment-only net-new change proposals.
- [x] 1.3 Update `.github/workflows-src/lib/select-change.test.mjs` to cover modified-only approval-eligible selection, net-new comment-only selection, mixed added/modified files within one change, and unsupported-status rejection.

## 2. Update workflow instructions and generated artifacts

- [x] 2.1 Update `.github/workflows-src/openspec-verify-label/workflow.md.tmpl` and any included script wiring so the agent prompt consumes the deterministic review disposition and disposition reason instead of inferring approval eligibility from PR files.
- [x] 2.2 Require the review body instructions to explain when a net-new spec change is limited to `COMMENT` despite otherwise meeting the normal approval criteria, and ensure archive/push steps remain unreachable for that path.
- [x] 2.3 Regenerate `.github/workflows/openspec-verify-label.md` and `.github/workflows/openspec-verify-label.lock.yml` from the workflow source.

## 3. Verify the change

- [x] 3.1 Run the relevant workflow-source tests and OpenSpec validation checks for `allow-single-new-change-proposal-review`.
- [x] 3.2 Confirm the rendered workflow and prompt reflect the deterministic comment-only limitation and clear reviewer-facing explanation for net-new change proposals.
