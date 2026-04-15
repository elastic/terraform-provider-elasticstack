## Context

The repository currently ships two custom golangci analyzers through module plugins: `acctestconfigdirlint` and `esclienthelper`. They run in the same repo-wide lint path as other Go linters, so their cost is paid on every local and CI lint invocation.

The current hot spots are different for each analyzer:

- `analysis/acctestconfigdirlintplugin` depends on `inspect.Analyzer`, walks every `*ast.CallExpr` in each package, resolves typed callees, and only then filters to `_test.go` acceptance-test calls. That broad traversal does unnecessary work in mixed packages with large non-test files.
- `analysis/esclienthelperplugin` is already path-scoped, but it repeats filename checks, function/signature scans, and fact imports while walking in-scope files. The analyzer keeps the right conservative behavior, but it recomputes metadata that is stable for the duration of a pass.

The repository also lacks a dedicated measurement loop for these analyzers. `make lint` is too noisy because it includes setup, formatting, docs generation, and the rest of golangci-lint. To make performance work sustainable, the repo needs a first-class command that isolates custom-linter timing and captures benchmark/profile artifacts.

## Goals / Non-Goals

**Goals:**

- Reduce avoidable work in `analysis/acctestconfigdirlintplugin` without changing which code patterns are considered in-scope or which diagnostics are reported.
- Reduce repeated metadata, signature, and fact lookup work in `analysis/esclienthelperplugin` while preserving its current sink detection and conservative provenance rules.
- Add a repository-local `lint-perf` workflow that captures isolated custom-linter timings plus CPU, memory, and trace profiles.
- Add analyzer benchmarks so future changes can compare before/after performance without relying on full aggregate lint wall time.

**Non-Goals:**

- Changing the lint requirements enforced by `acctestconfigdirlint` or `esclienthelper`.
- Replacing the current conservative provenance model in `esclienthelper` with an SSA-based or whole-program dataflow engine.
- Introducing external benchmarking dependencies such as `benchstat` as a requirement for the first version of the measurement target.
- Defining strict runtime budgets that would make lint speed a hard pass/fail contract.

## Decisions

- **Narrow `acctestconfigdirlint` to relevant files before typed call resolution**: Replace the current `inspect.Analyzer`-driven whole-package `CallExpr` preorder with direct iteration over `pass.Files`, early `_test.go` filtering, and a local AST walk for only those files. Keep typed confirmation for `resource.Test` / `resource.ParallelTest`, but move that confirmation behind cheaper file and selector-name guards so non-test files and obviously unrelated calls do not pay for `TypesInfo` lookups.

- **Keep `acctestconfigdirlint` semantics but remove duplicate local work**: Preserve the current inline-literal scope and field-relationship checks, but fold duplicated composite-literal element scans into a single pass and avoid repeated filename position materialization when only the file name is needed.

- **Keep `esclienthelper` in two logical phases but add run-scoped caches**: Preserve the current "export facts first, then inspect sinks" structure so fact availability remains deterministic, but precompute the in-scope file set once per pass and reuse it across both phases. Add caches keyed by `*types.Func` for sink parameter indices, imported return facts, and other stable per-function metadata so repeated sink and provenance checks stop rescanning the same signatures and reimporting the same facts.

- **Refactor `esclienthelper` helpers around resolved callees**: When a sink check or source check already resolves a `*types.Func`, pass that information through helper layers rather than calling `calledFunction` repeatedly for the same expression. Keep the current conservative provenance behavior and existing diagnostic messages intact.

- **Add a single `lint-perf` target for isolated measurement and benchmarks**: The target should build or reuse the repository's custom golangci binary, create a timestamped repo-local output directory, run `esclienthelper` and `acctestconfigdirlint` individually with `--enable-only`, fixed concurrency, and profile flags, then run `go test -bench` for the analyzer packages and capture the outputs in the same report directory. This keeps performance measurement reproducible without conflating it with `make lint`.

- **Benchmark representative analyzer workloads, not only helper functions**: Add benchmarks in the analyzer test packages that execute the analyzers on representative `analysistest` fixtures or equivalent package-scoped workloads. That keeps the benchmarks close to the actual AST/type-analysis cost instead of micro-benchmarking individual helper functions that do not reflect pass-level behavior.

## Risks / Trade-offs

- **[Risk] Early syntactic guards in `acctestconfigdirlint` could accidentally narrow scope too far** -> Mitigation: keep final typed confirmation of the target function and preserve existing analyzer tests for compliant and violating cases.
- **[Risk] Cached metadata in `esclienthelper` could become stale if scoped incorrectly** -> Mitigation: limit caches to immutable per-pass values such as resolved `*types.Func` metadata and imported facts, and do not cache results that depend on mutable `derivedVars` state.
- **[Risk] Benchmark results may not reflect full-repo lint cost on their own** -> Mitigation: pair analyzer benchmarks with isolated `golangci-lint` runs from `lint-perf` so both micro and repo-level measurements are available.
- **[Risk] Performance measurements are noisy across machines and cache states** -> Mitigation: fix concurrency inside `lint-perf`, capture raw profile artifacts, and document that comparisons should be made with warm-cache runs under the same local conditions.

## Migration Plan

1. Add the `makefile-workflows` delta spec for the new `lint-perf` target and the expected measurement artifacts.
2. Implement the `lint-perf` target so contributors can capture a baseline before changing analyzer internals.
3. Add benchmark coverage for the custom analyzers under `analysis/`.
4. Refactor `analysis/acctestconfigdirlintplugin` to narrow traversal to relevant test files and candidate calls while preserving diagnostics.
5. Refactor `analysis/esclienthelperplugin` to precompute in-scope files and cache stable per-function metadata and facts.
6. Run targeted analyzer tests and `lint-perf` before/after comparisons to confirm behavior is unchanged and performance improves.

## Open Questions

- None.
