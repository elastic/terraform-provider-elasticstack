## 1. Syntactic Import-Path Check (replace typed `isAcceptanceTestCall`)

- [x] 1.1 Add `buildResourceImportAliases(file *ast.File) map[string]bool` helper that iterates `file.Imports` and returns the set of local names that resolve to the `github.com/hashicorp/terraform-plugin-testing/helper/resource` import path (handling explicit alias, no alias, and dot-import)
- [x] 1.2 Replace `isAcceptanceTestCall(pass, call)` with a syntactic check: look up the selector's `X` identifier in the alias set returned by 1.1, confirm method name is `"Test"` or `"ParallelTest"`
- [x] 1.3 Deferred to a follow-up change. Flipping `GetLoadMode()` to `LoadModeSyntax` requires removing every remaining `pass.TypesInfo` access (`isTestCaseLit`, `inspectTestStep`'s type check, `isNamedTestCaseDirectoryCall`/`calledFunction`, `isValidEmbeddedCompatConfig`, and `buildVarSpecIndex`), which is out of scope here. This change keeps `LoadModeTypesInfo` and only removes `TypesInfo` from `isAcceptanceTestCall`; the load-mode flip and full `TypesInfo` removal are tracked as a follow-up. Mark this task complete to acknowledge the deferral.
- [x] 1.4 Remove the `calledFunction` usage from `isAcceptanceTestCall` (the remaining `TypesInfo`-dependent helpers stay in place for this change; see task 1.3 for the deferred follow-up)
- [x] 1.5 Update the `run()` function to compute the alias set per file before the inner walk (or pass it as a parameter)
- [x] 1.6 Verify all existing `analysistest` tests still pass: `go test ./analysis/acctestconfigdirlint/...`

## 2. Targeted Function-Body Traversal (replace `ast.Inspect`)

- [x] 2.1 Replace the `ast.Inspect(file, ...)` call in `run()` with an explicit loop: `for _, decl := range file.Decls` → type-assert to `*ast.FuncDecl` → check name prefix `"Test"` → iterate `Body.List`
- [x] 2.2 For each top-level `*ast.ExprStmt`, extract `X` as `*ast.CallExpr`, apply `isCandidateCallExpr` and the new syntactic import check, then call `inspectTestCase`
- [x] 2.3 Confirm that the non-test-function guard (Decision 4 trade-off) is tested: add a `testdata` case where a helper function (not prefixed `"Test"`) calls `resource.Test` to document the known scope boundary
- [x] 2.4 Verify all existing `analysistest` tests still pass

## 3. Per-Pass `ReadFile` + Line-Split Cache

- [x] 3.1 Add a `fileLineCache map[string][]string` variable in `run()` (or a small struct wrapping it)
- [x] 3.2 Extract a `cachedLines(pass *analysis.Pass, cache map[string][]string, filename string) []string` helper that reads and splits on first access, returns cached slice thereafter
- [x] 3.3 Update `goEmbedPathsAboveValueSpec` signature to accept the cache (or a getter func) instead of calling `pass.ReadFile` directly
- [x] 3.4 Thread the cache through the call chain: `run()` → `inspectTestStep` → `isValidEmbeddedCompatConfig` → `goEmbedPathsAboveValueSpec`
- [x] 3.5 Verify all existing `analysistest` tests still pass

## 4. O(1) `findValueSpecForVar` Replacement

- [x] 4.1 Add `buildVarSpecIndex(pass *analysis.Pass) map[*types.Var]*ast.ValueSpec` function that iterates `pass.Files` → `GenDecl` (token.VAR) → `ValueSpec` → `Names`, using `pass.TypesInfo.Defs` to resolve each name to `*types.Var`
- [x] 4.2 Call `buildVarSpecIndex` once at the start of `run()` and store the result
- [x] 4.3 Replace the body of `findValueSpecForVar` with a single map lookup into the index; update its signature to accept the index as a parameter
- [x] 4.4 Thread the index through the call chain: `run()` → `inspectTestStep` → `isValidEmbeddedCompatConfig` → `findValueSpecForVar`
- [x] 4.5 Verify all existing `analysistest` tests still pass

## 5. Benchmark Coverage and Validation

- [x] 5.1 Extend `benchmark_test.go` with a `BenchmarkAnalyzer_LargePackage` case (or document that the existing benches now reflect post-optimisation numbers) and run `go test ./analysis/acctestconfigdirlint/... -bench=. -benchmem -count=5` to capture stable before/after numbers
- [ ] 5.2 Run `make lint-perf` and verify wall-clock golangci time is meaningfully reduced; commit the new `lint-perf-output/` snapshot as a reference baseline
- [ ] 5.3 Run `make build` to confirm the full provider still compiles
- [ ] 5.4 Run `go vet ./analysis/...` and `make lint` (or `golangci-lint run ./analysis/...`) to confirm no new lint issues
