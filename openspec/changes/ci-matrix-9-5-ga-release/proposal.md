## Why

The `Provider CI` acceptance matrix in `.github/workflows/provider.yml` currently tracks the
in-development 9.5 line as `9.5.0-SNAPSHOT`. Now that 9.5.0 has reached general availability, the
matrix must test against the released `9.5.0` image instead of the snapshot build. Leaving the
snapshot entry in place means CI keeps exercising an unreleased/mutable build and keeps treating
that version's acceptance failures as non-blocking (`continue-on-error`), even though a GA release
should be held to the same blocking bar as every other released stack version in the matrix.

## What Changes

- Replace the `9.5.0-SNAPSHOT` matrix entry in `.github/workflows/provider.yml` with `9.5.0`, so the
  acceptance matrix builds and tests against the released image rather than the snapshot build.
- Because the promoted entry no longer matches `endsWith(matrix.version, '-SNAPSHOT')`, add `9.5.0`
  explicitly wherever that suffix match was standing in for per-version step gating (notably the
  Fleet setup step's explicit version list), so the promoted version keeps the Fleet coverage it had
  while still labeled as a snapshot.
- As a result, `9.5.0` acceptance-test failures become blocking (`continue-on-error` no longer
  applies) like every other non-snapshot matrix entry, and the snapshot-failure PR warning comment
  no longer applies to `9.5.0`.
- Add an explicit requirement to the `ci-build-lint-test` capability describing how a snapshot-to-GA
  version promotion must be carried out, so this same promotion is done consistently in future stack
  releases and doesn't silently drop per-version step coverage (see Design for the concrete gap this
  closes).

## Capabilities

### Modified Capabilities
- `ci-build-lint-test`: acceptance matrix must promote a snapshot-labeled version to its GA release
  entry without losing per-version step coverage that had been keyed off the snapshot suffix.

## Impact

- `.github/workflows/provider.yml`: `test` job matrix, Fleet setup step condition.
- No provider Go code, generated clients, or documentation outside CI workflow config is affected.
- CI acceptance coverage: `9.5.0` acceptance failures start blocking the `test` job and
  `Test Validation` (previously non-blocking under the snapshot label).
