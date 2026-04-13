## 1. Makefile test aggregation

- [x] 1.1 Add a `hook-test` target that runs `node --test .agents/hooks/*.test.mjs`
- [x] 1.2 Update `test` so it executes Go unit tests plus `workflow-test` and `hook-test`
- [x] 1.3 Remove `workflow-test` from the `check-lint` dependency chain while keeping workflow freshness checks intact

## 2. CI build job alignment

- [x] 2.1 Update `.github/workflows/test.yml` so the `build` job sets up the Node runtime needed for repository JavaScript tests
- [x] 2.2 Add build-job steps that run `make workflow-test` and `make hook-test` before `make build-ci`

## 3. Verification

- [x] 3.1 Run the relevant Makefile targets locally to confirm the new unit-test aggregation works as specified
- [x] 3.2 Run the relevant OpenSpec checks to confirm the updated change artifacts and canonical specs remain valid
