## 1. Makefile test aggregation

- [ ] 1.1 Add a `hook-test` target that runs `node --test .agents/hooks/*.test.mjs`
- [ ] 1.2 Update `test` so it executes Go unit tests plus `workflow-test` and `hook-test`
- [ ] 1.3 Remove `workflow-test` from the `check-lint` dependency chain while keeping workflow freshness checks intact

## 2. CI build job alignment

- [ ] 2.1 Update `.github/workflows/test.yml` so the `build` job sets up the Node runtime needed for repository JavaScript tests
- [ ] 2.2 Add build-job steps that run `make workflow-test` and `make hook-test` before `make build-ci`

## 3. Verification

- [ ] 3.1 Run the relevant Makefile targets locally to confirm the new unit-test aggregation works as specified
- [ ] 3.2 Run the relevant OpenSpec checks to confirm the updated change artifacts and canonical specs remain valid
