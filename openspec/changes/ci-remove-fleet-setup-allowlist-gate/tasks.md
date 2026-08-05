## 1. Remove the redundant Fleet setup step

- [x] 1.1 In `.github/workflows/provider.yml`, delete the `setup-fleet` step (`id: setup-fleet`,
      `name: Setup Fleet`) from the Matrix Acceptance Test (`test`) job, including its `if:` allowlist
      condition, `run: make setup-kibana-fleet`, and `env:` block.
- [x] 1.2 Confirm no later step in the `test` job (`force-install-synthetics`, `tf-acceptance`,
      failure-diagnostics, or teardown steps) references the removed step's `id` or step outputs.
- [x] 1.3 Confirm `.github/workflows/copilot-setup-steps.yml` and
      `.github/workflows/shared/elastic-stack.md` are left unchanged — they invoke
      `make setup-kibana-fleet` unconditionally outside this job and are out of scope.

## 2. Update the capability spec

- [x] 2.1 In `openspec/specs/ci-build-lint-test/spec.md`, under
      "Requirement: Acceptance test job structure", update the sentence "Fleet setup and forced
      synthetics installation SHALL run only for configured version subsets" to remove the Fleet-setup
      clause (no per-version Fleet setup step exists anymore) while keeping the forced-synthetics
      clause unchanged.

## 3. Verify

- [x] 3.1 Run
      `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate ci-remove-fleet-setup-allowlist-gate --type change`
      and resolve any reported issues.
- [ ] 3.2 After merge, confirm a `Provider CI` matrix run completes successfully for a representative
      version (e.g. one previously in the allowlist and one previously excluded, such as `9.4.2`)
      without the `setup-fleet` step, and that Fleet-dependent acceptance tests still pass.
