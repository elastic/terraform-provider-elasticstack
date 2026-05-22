# acctestconfigdirlint-perf Specification

## Purpose
TBD - created by archiving change acctestconfigdirlint-perf. Update Purpose after archive.
## Requirements
### Requirement: Analyzer uses syntactic import-path check for acceptance-test call detection
The analyzer SHALL identify `resource.Test` and `resource.ParallelTest` calls using a syntactic check against the imported package path (`github.com/hashicorp/terraform-plugin-testing/helper/resource`) rather than a full `go/types` type-info lookup. The check SHALL handle standard imports and explicit aliases correctly.
#### Scenario: Standard import alias resolved syntactically
- **WHEN** a test file imports `"github.com/hashicorp/terraform-plugin-testing/helper/resource"` without an explicit alias
- **THEN** the analyzer SHALL recognize calls of the form `resource.Test(...)` and `resource.ParallelTest(...)` as acceptance-test call candidates without requiring full package type-checking

#### Scenario: Explicit import alias resolved syntactically
- **WHEN** a test file imports the resource package with an explicit alias (e.g., `tftest "github.com/hashicorp/..."`)
- **THEN** the analyzer SHALL recognize calls using that alias (e.g., `tftest.Test(...)`) as acceptance-test call candidates

#### Scenario: Unrelated package with same method name is not matched
- **WHEN** a test file calls `someOtherPkg.Test(...)` where `someOtherPkg` does not import the resource package path
- **THEN** the analyzer SHALL NOT treat that call as an acceptance-test candidate

### Requirement: Analyzer traversal is limited to test function bodies
The analyzer SHALL traverse only `*ast.FuncDecl` nodes whose names begin with `"Test"` inside `_test.go` files that import the resource package. Within each such function body, the analyzer SHALL descend into nested blocks (including `t.Run` closures, `if` blocks, and `for` loops) to find `resource.Test` / `resource.ParallelTest` calls.

#### Scenario: Call at top level of test function is evaluated
- **WHEN** `resource.Test(t, ...)` appears as a direct statement in a function named `TestFoo`
- **THEN** the analyzer SHALL evaluate the call and inspect the `resource.TestCase` argument

#### Scenario: Non-test function is not traversed
- **WHEN** a `_test.go` file contains a helper function not prefixed with `"Test"` that also contains a `resource.Test` call
- **THEN** the analyzer SHALL NOT evaluate that call (helper functions are not direct acceptance test entry points)

#### Scenario: Call nested inside a test function body is evaluated
- **WHEN** `resource.Test(t, ...)` appears inside a nested block within a `Test*`-prefixed function (e.g. inside `t.Run(name, func(t *testing.T) { ... })`, an `if` block, or a `for` loop)
- **THEN** the analyzer SHALL still evaluate the call and inspect the `resource.TestCase` argument

### Requirement: Source file content is cached within a single analyzer pass
Within a single invocation of the analyzer's `run` function, each source file SHALL be read from disk and split into lines at most once. Subsequent accesses to the same file's line content SHALL use the cached result.

#### Scenario: Multiple compat-step matches in the same file use cached lines
- **WHEN** a single `_test.go` file contains more than one `ExternalProviders`+`Config` step that triggers the `//go:embed` path scan
- **THEN** the analyzer SHALL read and split that file from disk exactly once, reusing the line slice for all subsequent accesses within the same pass

#### Scenario: Cache is not shared across packages or passes
- **WHEN** the analyzer runs on two different packages in the same golangci invocation
- **THEN** each package's pass SHALL use its own independent file-content cache with no cross-package sharing

### Requirement: Package-level var-to-spec lookup is O(1) per query
The analyzer SHALL build a complete `*types.Var` → `*ast.ValueSpec` index at pass entry (one traversal of all package-level `GenDecl`s) and use that index for all subsequent lookups within the pass. Linear scans of all files and declarations to locate a single `ValueSpec` are not permitted within the hot path.

#### Scenario: Index is built once per pass
- **WHEN** the analyzer pass starts for a package
- **THEN** the analyzer SHALL construct the var-to-spec index by traversing all package files exactly once before any `ExternalProviders`+`Config` checks are performed

#### Scenario: Lookup resolves immediately
- **WHEN** the analyzer needs the `ValueSpec` for a given `*types.Var`
- **THEN** it SHALL retrieve it via a single map lookup without iterating over any file, declaration, or spec list

