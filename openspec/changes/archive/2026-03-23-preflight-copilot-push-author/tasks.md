## 1. Requirements in main tree

- [x] 1.1 Sync the delta in `openspec/changes/preflight-copilot-push-author/specs/ci-build-lint-test/spec.md` into `openspec/specs/ci-build-lint-test/spec.md` (or archive the change per project workflow so canonical spec matches REQ-023–REQ-027).

## 2. Workflow implementation

- [x] 2.1 In `.github/workflows/test.yml` preflight `github-script` step, for `push` events: set `should_run=true` if there is no open PR **or** if every `context.payload.commits` entry has `author.email === '198982749+Copilot@users.noreply.github.com'`; otherwise `should_run=false`.
- [x] 2.2 Preserve existing `pull_request` / `workflow_dispatch` / `ready_for_review` branches unchanged aside from any refactor needed to share the push branch logic clearly.

## 3. Verification

- [x] 3.1 Run `openspec validate --all` (or project `make check-openspec` after sync) so the updated canonical spec validates.
- [ ] 3.2 Manually confirm a push with mixed author emails skips `build`/`lint`/`test`, and an all-Copilot push without an open PR runs them.
