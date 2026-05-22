## 1. Syntactic Import-Path Check (replace typed `isAcceptanceTestCall`)

- [ ] 1.1 Add `buildResourceImportAliases(file *ast.File) map[string]bool` helper that iterates `file.Imports` and returns the set of local names that resolve to the `github.com/hashicorp/terraform-plugin-testing/helper/resource` import path (handling explicit alias, no alias, and dot-import)
- [ ] 1.2 Replace `isAcceptanceTestCall(pass, call)` with a syntactic check: look up the selector's `X` identifier in the alias set returned by 1.1, confirm method name is `"Test"` or `"ParallelTest"`
- [ ] 1.3 Update `analysis/acctestconfigdirlintplugin/plugin/plugin.go` `GetLoadMode()` to stop requesting `LoadModeTypesInfo` globally, and ensure the analyzer does not depend on `pass.TypesInfo` in the optimized path
- [ ] 1.4 Remove the `calledFunction` usage from `isAcceptanceTestCall` (and plan follow-up work to remove or rewrite other `TypesInfo`-dependent helpers)
- [ ] 1.5 Update the `run()` function to compute the alias set per file before the inner walk (or pass it as a parameter)
- [ ] 1.6 Verify all existing `analysistest` tests still pass: `go test ./analysis/acctestconfigdirlint/...`

## 2. Targeted Function-Body Traversal (replace `ast.Inspect`)

- [ ] 2.1 Replace the `ast.Inspect(file, ...)` call in `run()` with an explicit loop: `for _, decl := range file.Decls` → type-assert to `*ast.FuncDecl` → check name prefix `"Test"` → iterate `Body.List`
- [ ] 2.2 For each top-level `*ast.ExprStmt`, extract `X` as `*ast.CallExpr`, apply `isCandidateCallExpr` and the new syntactic import check, then call `inspectTestCase`
- [ ] 2.3 Confirm that the non-test-function guard (Decision 4 trade-off) is tested: add a `testdata` case where a helper function (not prefixed `"Test"`) calls `resource.Test` to document the known scope boundary
- [ ] 2.4 Verify all existing `analysistest` tests still pass

## 3. Per-Pass `ReadFile` + Line-Split Cache

- [ ] 3.1 Add a `fileLineCache map[string][]string` variable in `run()` (or a small struct wrapping it)
- [ ] 3.2 Extract a `cachedLines(pass *analysis.Pass, cache map[string][]string, filename string) []string` helper that reads and splits on first access, returns cached slice thereafter
- [ ] 3.3 Update `goEmbedPathsAboveValueSpec` signature to accept the cache (or a getter func) instead of calling `pass.ReadFile` directly
- [ ] 3.4 Thread the cache through the call chain: `run()` → `inspectTestStep` → `isValidEmbeddedCompatConfig` → `goEmbedPathsAboveValueSpec`
- [ ] 3.5 Verify all existing `analysistest` tests still pass

## 4. O(1) `findValueSpecForVar` Replacement

- [ ] 4.1 Add `buildVarSpecIndex(pass *analysis.Pass) map[*types.Var]*ast.ValueSpec` function that iterates `pass.Files` → `GenDecl` (token.VAR) → `ValueSpec` → `Names`, using `pass.TypesInfo.Defs` to resolve each name to `*types.Var`
- [ ] 4.2 Call `buildVarSpecIndex` once at the start of `run()` and store the result
- [ ] 4.3 Replace the body of `findValueSpecForVar` with a single map lookup into the index; update its signature to accept the index as a parameter
- [ ] 4.4 Thread the index through the call chain: `run()` → `inspectTestStep` → `isValidEmbeddedCompatConfig` → `findValueSpecForVar`
- [ ] 4.5 Verify all existing `analysistest` tests still pass

## 5. Benchmark Coverage and Validation

- [ ] 5.1 Extend `benchmark_test.go` with a `BenchmarkAnalyzer_LargePackage` case (or document that the existing benches now reflect post-optimisation numbers) and run `go test ./analysis/acctestconfigdirlint/... -bench=. -benchmem -count=5` to capture stable before/after numbers
- [ ] 5.2 Run `make lint-perf` and verify wall-clock golangci time is meaningfully reduced; commit the new `lint-perf-output/` snapshot as a reference baseline
- [ ] 5.3 Run `make build` to confirm the full provider still compiles
- [ ] 5.4 Run `go vet ./analysis/...` and `make lint` (or `golangci-lint run ./analysis/...`) to confirm no new lint issues
