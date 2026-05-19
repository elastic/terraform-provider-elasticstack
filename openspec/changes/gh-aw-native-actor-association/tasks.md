## 1. Update factory workflow sources

- [ ] 1.1 Add `skip-author-associations: { issues: [none, first_timer, first_time_contributor, contributor] }` to the `on:` block in `research-factory-issue.md`, `code-factory-issue.md`, and `reproducer-factory-issue.md`.
- [ ] 1.2 Add `skip-author-associations` to `change-factory-issue.md` covering both the `issues` and `issue_comment` event types with the same association list.
- [ ] 1.3 Remove the `check_actor_trust` step and all downstream conditions gating on `steps.check_actor_trust.outputs.actor_trusted == 'true'` from all four workflow source files.
- [ ] 1.4 Simplify `normalize_context` in each workflow source to hardcode `actor_trusted=true` for the issue-event path, removing the `ACTOR_TRUSTED_EVENT` / `ACTOR_TRUSTED_REASON_EVENT` env vars and the corresponding output lines.

## 2. Remove the JS trust-check implementation

- [ ] 2.1 Delete `.github/scripts/workflows/lib/factory-runners/check-actor-trust.js`.
- [ ] 2.2 Remove `factoryCheckActorTrust`, `factoryActorTrustWhenSenderMissing`, and their exports from `.github/scripts/workflows/lib/factory-issue-shared.js`.

## 3. Prune actor-trust tests

- [ ] 3.1 Remove `factoryCheckActorTrust` and `factoryActorTrustWhenSenderMissing` test cases from `.github/scripts/workflows/lib/factory-issue-shared.test.mjs`; update `factoryParseFinalizeGateEnv` tests if `actor_trusted` is removed from the env parse contract.
- [ ] 3.2 Remove `checkActorTrust` / `actorTrustWhenSenderMissing` test cases from `.github/scripts/workflows/lib/code-factory-issue.test.mjs`.
- [ ] 3.3 Remove the corresponding actor-trust test cases from `.github/scripts/workflows/lib/change-factory-issue.test.mjs`.

## 4. Regenerate lock files and verify

- [ ] 4.1 Run `gh aw compile` (or equivalent) for all four factory workflow sources to regenerate the `.lock.yml` compiled artifacts.
- [ ] 4.2 Confirm `make check` (or equivalent CI validation) passes with no lint, build, or test failures introduced by the changes.
