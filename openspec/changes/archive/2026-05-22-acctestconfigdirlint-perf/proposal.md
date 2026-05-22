## Why

The `acctestconfigdirlint` custom linter takes **~1.6 s per benchmark iteration** and **~29.5 s wall-clock** in golangci (`--concurrency=1`) over `./internal/...`. Profiling reveals that almost all of this time is spent in infrastructure the analyzer drives but does not control — full type-checking of every package, repeated disk I/O, and GC pressure from 1.52 GB allocated per run — rather than in the analyzer logic itself. Because this linter runs in every `make lint` invocation on the hot path, the overhead compounds across local developer loops and CI minutes.

## What Changes

- **Narrow the `go/packages` load mode**: replace the full `NeedTypesInfo` load (which triggers complete type-checking of every package) with a syntactic import-path check so non-acceptance-test packages skip type-checking entirely.
- **Add a syntactic pre-filter for candidate files**: skip packages that contain no `_test.go` files before any AST walk.
- **Cache `pass.ReadFile` + line splitting** in `goEmbedPathsAboveValueSpec` so each source file is read from disk and split into lines at most once per analyzer pass.
- **Replace `findValueSpecForVar` linear scan with an O(1) pre-built index** mapping `*types.Var` → `*ast.ValueSpec`, computed once at pass start.
- **Replace `ast.Inspect` full-tree walk with a targeted function-body traversal** that only visits top-level statement lists in `_test.go` functions.

## Capabilities

### New Capabilities

- `acctestconfigdirlint-perf`: Performance characteristics of the `acctestconfigdirlint` analyzer — the load mode used, the traversal strategy, and the caching contract for file reads and var-to-spec lookups within a single analyzer pass.

### Modified Capabilities

- `acceptance-test-config-directory-lint`: The analyzer's observable behaviour (diagnostics, in-scope rules, error messages) is unchanged. Only internal execution strategy changes, so no requirement-level behaviour change is expected and no delta spec is needed.

## Impact

- `analysis/acctestconfigdirlintplugin/analyzer.go` — load-mode narrowing, targeted walk replacing `ast.Inspect`
- `analysis/acctestconfigdirlintplugin/embed_compat.go` — `ReadFile`/line-split cache; `findValueSpecForVar` replaced by pre-built index
- `analysis/acctestconfigdirlint/benchmark_test.go` — benchmark coverage extended to verify the speedup
- No public API changes; the `Analyzer` export and all diagnostic messages remain identical
- No changes to golangci configuration or `.golangci.yml`
