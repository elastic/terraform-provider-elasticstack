## Context

The `acctestconfigdirlint` analyzer runs as a `go/analysis` pass inside golangci-lint. Profiling (`make lint-perf`) shows **~29.5 s wall-clock** in the golangci path and **~1.6 s/op** in the isolated benchmark. The dominant costs are:

1. **Full type-checking** (`go/types.checkFiles`, 33% golangci CPU / 17% bench CPU) — driven by `go/packages.Load` requesting `NeedTypesInfo` for all 140 packages, most of which contain no acceptance tests.
2. **Disk I/O** (`os.ReadFile` / `syscall.Open`, 43.8% golangci CPU / 15.5% bench CPU) — every `.go` source file read once by the parser and again by `goEmbedPathsAboveValueSpec`.
3. **GC pressure** (`runtime.madvise` / `mallocgc`, 27% flat golangci CPU) — 1.52 GB allocated per benchmark run, mostly from `go/types.recordTypeAndValue` (0.47 GB, 31% of all allocations).
4. **`findValueSpecForVar` linear scan** — O(files × decls × specs × names) per `ExternalProviders`+`Config` match.
5. **`ast.Inspect` full-tree walk** — visits every AST node in every test file, including deep expression subtrees that will all be rejected by the candidate guard.

The analyzer's own logic contributes zero measurable flat CPU samples; all savings must come from reducing what the infrastructure does on our behalf.

## Goals / Non-Goals

**Goals:**
- Eliminate full type-checking for packages that contain no acceptance test candidates.
- Read each source file from disk at most once per analyzer pass.
- Replace the O(N) `findValueSpecForVar` scan with an O(1) lookup built once at pass start.
- Replace the `ast.Inspect` full-tree walk with a targeted descent limited to test-function bodies.
- Keep all existing diagnostic messages, rules, and observable analyzer behaviour identical.
- Maintain or improve benchmark coverage so regressions are caught.

**Non-Goals:**
- Changing golangci configuration, `.golangci.yml`, or the `make lint` invocation.
- Modifying the `acceptance-test-config-directory-lint` spec requirements (behaviour is unchanged).
- Parallelising the analyzer internally (golangci controls concurrency).
- Caching results across golangci invocations (that is golangci's cache responsibility).

## Decisions

### Decision 1 — Syntactic import-path check instead of full `NeedTypesInfo` for the acceptance-test call guard

**Problem**: `isAcceptanceTestCall` currently uses `pass.TypesInfo.Uses[sel]` to resolve the called function, which requires `NeedTypesInfo`. This forces full type-checking of every package.

**Chosen approach**: Replace the typed lookup with an import-alias-aware syntactic check. At the top of `run()`, build a `map[string]bool` of local names that import `github.com/hashicorp/terraform-plugin-testing/helper/resource` (iterating `file.Imports`). Then `isAcceptanceTestCall` becomes: "is the selector's X an identifier in that map, and is the method name `Test` or `ParallelTest`?". This is deterministic from syntax alone.

**Why not keep typed lookup for just these two calls?** The `go/analysis` framework does not support requesting `NeedTypesInfo` for a subset of packages; it is all-or-nothing per package. So any use of `TypesInfo` in `run()` keeps full type-checking for every package. The syntactic approach is unambiguous here because `resource.Test` / `resource.ParallelTest` are the only two public functions in that package with those names, and the import path is canonical.

**Alternative considered**: Keep `NeedTypesInfo` but add a pre-filter that exits `run()` early if the package has no `_test.go` files. This is simpler but still pays the full type-check cost for packages that do have test files (including non-acceptance-test packages with unit tests).

**Trade-off**: The syntactic check will produce a false positive if user code shadows the import alias with a local variable also named `resource` — an extremely unlikely pattern in acceptance test files, and even then the impact is a missed lint issue (false negative), not a false positive.

### Decision 2 — Per-pass `ReadFile` + line-split cache in `goEmbedPathsAboveValueSpec`

**Problem**: `goEmbedPathsAboveValueSpec` calls `pass.ReadFile(filename)` and `strings.Split(..., "\n")` on every invocation, which is once per `ExternalProviders`+`Config` match. In a package with multiple compat steps in the same file this causes repeated syscalls and allocations.

**Chosen approach**: Add a `map[string][]string` cache (filename → lines) as a closure variable in `run()`, passed into `goEmbedPathsAboveValueSpec`. On first access for a filename the file is read and split; subsequent calls return the cached slice.

**Alternative considered**: Cache at the `isValidEmbeddedCompatConfig` call site by pre-computing embed paths for all var decls up-front. Rejected because it would unconditionally pay the cost for packages with no compat steps.

### Decision 3 — Pre-built `*types.Var` → `*ast.ValueSpec` index

**Problem**: `findValueSpecForVar` iterates over all files, all `GenDecl`s, all `ValueSpec`s, and all names to find one `*ast.ValueSpec`. Called once per compat-config expression, it is O(N) in package size.

**Chosen approach**: At `run()` entry, build a `map[*types.Var]*ast.ValueSpec` by walking `pass.Files` once. `findValueSpecForVar` is replaced by an O(1) map lookup. Because the `TypesInfo.Defs` map is keyed by `*ast.Ident` already, the scan cost is paid only once regardless of how many compat matches are found.

**Note**: This decision still requires `pass.TypesInfo` for the `Defs` lookup — but only in packages where the analyzer actually runs its full logic (i.e. packages with `_test.go` files containing acceptance test calls). Combined with Decision 1's syntactic guard, packages without candidates skip this entirely.

### Decision 4 — Targeted function-body traversal replacing `ast.Inspect`

**Problem**: `ast.Inspect(file, ...)` visits every node in the file. Most are rejected immediately by `isCandidateCallExpr`. For large test files with thousands of expressions this wastes visitor allocations and time.

**Chosen approach**: Replace the `ast.Inspect` loop with an explicit walk:
1. For each file: iterate `file.Decls`.
2. For each `*ast.FuncDecl` whose name begins with `"Test"` (fast string prefix check): iterate `Body.List` (top-level statements).
3. For each `*ast.ExprStmt`: check if its `X` is a `*ast.CallExpr` — these are the only positions where `resource.Test(t, ...)` can appear.
4. Apply `isCandidateCallExpr` and the syntactic import check, then `inspectTestCase`.

This avoids allocating closure state for nodes that will never be acceptance-test calls.

**Alternative considered**: Keep `ast.Inspect` but add an early-exit return `false` for non-call nodes at depth > 2. This is fragile and would miss calls inside `if`/`for` blocks (though those are rare in practice). The explicit walk is cleaner.

**Trade-off**: The targeted walk will miss `resource.Test` calls that appear inside nested blocks (e.g., inside an `if` or helper function defined in the test file). In practice, acceptance tests always call `resource.Test` / `resource.ParallelTest` at the top level of the test function body. If this assumption is violated the analyzer produces false negatives (missed issues), not false positives. The existing `isCandidateCallExpr` + typed-lookup path already had the same implicit scope assumption.

### Decision 5 — Early exit for files with no `_test.go` suffix (unchanged, but documented)

The existing `strings.HasSuffix(filename, "_test.go")` guard in `run()` is kept. With Decision 1 in place the guard is now the only cost for non-test files (no more type-check overhead for the package). No change needed here.

## Risks / Trade-offs

- **Syntactic import check mis-fires on aliased imports** → Mitigation: the check is conservative — it collects all import specs whose path is the resource package, including aliased ones (using `imp.Name` if set, else the last path segment). Blank imports (`. "..."`) resolve the selector-X check differently but are disallowed by goimports conventions for this package.
- **Targeted walk misses deeply nested `resource.Test` calls** → Mitigation: by convention acceptance tests always call `resource.Test` at function-body top level; this is enforced by the existing rule structure and would be caught by the `analysistest` test suite if violated.
- **Index-building cost in large packages** → Building the `*types.Var` → `*ast.ValueSpec` map costs one pass over all `GenDecl`s, which is O(package declarations). This replaces the existing repeated O(N) scans and is strictly better.
- **Cache invalidation** → The per-pass `ReadFile` cache is scoped to a single `run()` invocation; it is not shared across packages or runs. There is no stale-data risk.

## Migration Plan

All changes are internal to `analysis/acctestconfigdirlintplugin/`. The `Analyzer` variable, package path, and diagnostic messages are unchanged. No consumer changes are required. The existing `analysistest`-based test suite validates behaviour parity. After implementation, `make lint-perf` should be re-run and the new benchmark numbers committed alongside the code changes for reference.

## Open Questions

- None. All decisions above are unambiguous given the profile data and the analyzer's narrow scope.
