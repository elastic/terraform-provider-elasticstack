## Why

The custom analyzers under `analysis/` now run as part of normal repository linting, so avoidable traversal and repeated typed lookups add noticeable cost to every contributor and CI `make lint` / `make check-lint` run. The repository also lacks a repeatable way to isolate those analyzers, capture profiles, and compare benchmark results, which makes performance work harder to validate and maintain.

## What Changes

- Optimize `analysis/acctestconfigdirlintplugin` so it limits work to relevant test files and candidate acceptance-test calls instead of traversing every package call expression before filtering.
- Optimize `analysis/esclienthelperplugin` so it precomputes in-scope files, reuses resolved function metadata, and caches repeated sink/fact lookups while preserving the current lint contract and diagnostics.
- Add a dedicated `make lint-perf` target that runs the repository's custom golangci binary in isolated custom-linter mode, captures timing/profile artifacts, and runs analyzer benchmarks for comparison work.
- Add benchmark and regression coverage for the custom analyzers so contributors can measure performance improvements without relying on full `make lint` wall time.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `makefile-workflows`: add a dedicated make target for isolated custom-linter performance measurement and benchmark capture

## Impact

- **Specs**: delta spec under `openspec/changes/custom-lint-performance/specs/makefile-workflows/spec.md`
- **Makefile**: new `lint-perf` target and any supporting recipe wiring
- **Analyzer implementation**: `analysis/acctestconfigdirlintplugin` and `analysis/esclienthelperplugin`
- **Tests and benchmarks**: analyzer regression coverage plus new benchmark entry points under `analysis/`
- **Developer workflow**: contributors gain a repository-local way to capture before/after timing and profile artifacts for custom lint rules
