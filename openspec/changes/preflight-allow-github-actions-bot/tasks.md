## 1. Requirements updates

- [ ] 1.1 Sync the delta in `openspec/changes/preflight-allow-github-actions-bot/specs/ci-build-lint-test/spec.md` into `openspec/specs/ci-build-lint-test/spec.md` so `REQ-023–REQ-027` names both allowed bot users.
- [ ] 1.2 Confirm the canonical requirement text and scenarios consistently use the same allowed-email list everywhere the preflight author exception is described.

## 2. Workflow implementation

- [ ] 2.1 Update the preflight logic in `.github/workflows/test.yml` so the "all commits are bot-authored" branch accepts both `198982749+Copilot@users.noreply.github.com` and `41898282+github-actions[bot]@users.noreply.github.com`.
- [ ] 2.2 Preserve the existing no-open-PR behavior and non-`push` event handling while refactoring only as needed to keep the allowed-author list clear.

## 3. Verification

- [ ] 3.1 Run `./node_modules/.bin/openspec validate preflight-allow-github-actions-bot` or `make check-openspec` after syncing the spec.
- [ ] 3.2 Verify that a push with an open PR still runs CI when every commit author email is one of the allowed bot users, and skips CI when any commit author email falls outside that list.
