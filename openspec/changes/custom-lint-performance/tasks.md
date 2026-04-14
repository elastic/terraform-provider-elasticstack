## 1. Define the performance-measurement contract

- [ ] 1.1 Add the `makefile-workflows` delta spec describing the `lint-perf` target, isolated custom-linter measurement scope, profile outputs, and analyzer benchmark capture.

## 2. Add repository-local measurement tooling

- [ ] 2.1 Add a `lint-perf` target to `Makefile` that builds or reuses the repository's custom golangci binary, runs `esclienthelper` and `acctestconfigdirlint` individually against `./...`, fixes concurrency for repeatable comparisons, and writes timing plus CPU, memory, and trace artifacts to a repo-local output directory.
- [ ] 2.2 Add benchmark entry points under `analysis/` for the custom analyzers so `lint-perf` can capture targeted analyzer benchmark output alongside the isolated golangci-lint runs.
- [ ] 2.3 Ensure the target prints or documents where the per-run artifacts are written so contributors can compare before/after optimization runs.

## 3. Optimize `acctestconfigdirlint`

- [ ] 3.1 Refactor `analysis/acctestconfigdirlintplugin/analyzer.go` to iterate `pass.Files`, skip non-`*_test.go` files before AST walking, and inspect only candidate acceptance-test calls rather than traversing every package call expression through `inspect.Analyzer`.
- [ ] 3.2 Preserve the current typed confirmation and diagnostics for `resource.Test` / `resource.ParallelTest` inline `resource.TestCase` handling while removing duplicate local work such as repeated composite-literal element scans and unnecessary filename position materialization.
- [ ] 3.3 Add or update analyzer tests and benchmarks so current compliant and violating cases remain unchanged while the narrowed traversal path is exercised.

## 4. Optimize `esclienthelper`

- [ ] 4.1 Refactor `analysis/esclienthelperplugin/analyzer.go` to precompute the in-scope non-test Elasticsearch files once per pass and reuse that scoped file list across the fact-export and sink-check phases.
- [ ] 4.2 Add run-scoped caches for stable per-function metadata such as Elasticsearch sink parameter indices and imported client-return facts so repeated sink checks stop rescanning the same signatures and reimporting the same facts.
- [ ] 4.3 Refactor sink and provenance helpers to reuse resolved callees where practical while preserving the current conservative provenance model and diagnostics.
- [ ] 4.4 Add or update analyzer tests and benchmarks so current compliant and violating sink behaviors remain unchanged while the cached path is exercised.

## 5. Validate behavior and performance

- [ ] 5.1 Capture before/after isolated measurements for both custom analyzers with `make lint-perf` and inspect the generated timing/profile artifacts to confirm the expected hot paths shrink.
- [ ] 5.2 Run targeted analyzer tests and the relevant repository lint checks to confirm the optimizations do not change the enforced lint behavior.
