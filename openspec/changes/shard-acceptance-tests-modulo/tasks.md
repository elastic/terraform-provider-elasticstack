## 1. Makefile sharding variables

- [x] 1.1 Add `ACCTEST_TOTAL_SHARDS ?= 1` and `ACCTEST_SHARD_INDEX ?= 0` near the existing `ACCTEST_PARALLELISM` / `ACCTEST_PACKAGE_PARALLELISM` block in `Makefile`.
- [x] 1.2 Replace the `--packages="./..."` argument in the `testacc` recipe with `--packages="$(shell go list ./... | sort | awk '(NR-1) % $(ACCTEST_TOTAL_SHARDS) == $(ACCTEST_SHARD_INDEX)')"`.
- [x] 1.3 Verify that `make testacc` without shard variables is equivalent to today (smoke-test: run `make testacc TESTARGS='-run ^TestAccResourceDashboardEmptyDashboard$$'` and confirm the single test runs).
- [x] 1.4 Verify that `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=0` and `ACCTEST_SHARD_INDEX=1` together cover all packages: `diff <(go list ./... | sort) <(cat <(go list ./... | sort | awk '(NR-1)%2==0') <(go list ./... | sort | awk '(NR-1)%2==1') | sort)` should produce no output.
- [x] 1.5 Update the `makefile-workflows` capability spec (REQ-023–REQ-024 or a new requirement) to require that `testacc` supports `ACCTEST_TOTAL_SHARDS` / `ACCTEST_SHARD_INDEX` with modulo-based package selection and backwards-compatible defaults.

## 2. GitHub Actions workflow change

- [x] 2.1 Locate the workflow source template(s) in `workflows-src/` that define the acceptance-test matrix job.
- [x] 2.2 Add `shard: [0, 1]` as a matrix dimension so each `(version, shard)` pair runs as an independent job with its own runner and Elastic stack.
- [x] 2.3 Thread `ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}` into the `make testacc` step in the template.
- [x] 2.4 Regenerate the compiled workflow files via `make workflow-generate` and confirm `make check-workflows` passes.
- [x] 2.5 Confirm `make check-lint` passes with the updated workflow sources and Makefile.

## 3. Verify and document

- [x] 3.1 Open the PR. Record in the PR description: (a) the estimated wall-clock for each shard derived from the timing data in `speed-up-dashboard-acceptance-tests` PR #2539; (b) the modulo split of the two known-flaky fleet packages (`integration` → shard 0, `integration_policy` → shard 1).
- [x] 3.2 Capture the actual wall-clock for both shards from the first successful CI run and update the PR description with the before/after comparison.
