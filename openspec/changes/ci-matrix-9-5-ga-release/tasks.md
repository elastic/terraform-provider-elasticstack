## 1. Promote the 9.5 matrix entry to GA

- [ ] 1.1 In `.github/workflows/provider.yml`, change the `strategy.matrix.version` entry
      `"9.5.0-SNAPSHOT"` to `"9.5.0"`.
- [ ] 1.2 Add `9.5.0` to the `setup-fleet` step's explicit `if:` version list so the promoted entry
      keeps Fleet setup coverage it previously received via the `endsWith(matrix.version,
      '-SNAPSHOT')` match.
- [ ] 1.3 Confirm no other step (`force-install-synthetics`, the pre-pull fleet image step, the
      snapshot PR-warning step, or any script under `Makefile`/`docker-compose.yml`) depends on
      `9.5.0-SNAPSHOT` specifically or otherwise needs updating for this promotion.

## 2. Verify

- [ ] 2.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate ci-matrix-9-5-ga-release --type change`
      and resolve any reported issues.
- [ ] 2.2 After merge, confirm the `9.5.0` matrix job in `Provider CI` runs Fleet setup and is
      blocking (not `continue-on-error`) in the next workflow run.
