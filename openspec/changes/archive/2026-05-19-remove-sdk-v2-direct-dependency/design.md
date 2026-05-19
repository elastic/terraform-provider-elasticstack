## Context

The `terraform-provider-elasticstack` provider completed its migration from the legacy Terraform Plugin SDK v2 to the Terraform Plugin Framework some time ago. All resources and data sources are now implemented using Plugin Framework types (`terraform-plugin-framework`, `terraform-plugin-framework-jsontypes`, etc.).

Despite this, `github.com/hashicorp/terraform-plugin-sdk/v2` remains as a **direct dependency** in `go.mod` (v2.40.1). An audit of all imports reveals that the provider code only uses three things from the SDK v2 module:

1. `helper/logging.IsDebugOrHigher()` — called in 5 places to decide whether sensitive integration-policy variables should be logged.
2. `helper/schema.Set` — used only in `internal/utils/typeutils/schema.go`, which defines `ExpandStringSet`. This function is **unused** in production code.
3. `helper/acctest` — imported in exactly one test file (`templateilmattachment/acc_test.go`). Every other test in the repo already uses the `terraform-plugin-testing` equivalent.

`terraform-plugin-testing` (already a direct dependency at v1.16.0) transitively depends on SDK v2, so the module will remain in the dependency graph. The goal is solely to eliminate **direct** usage.

## Goals / Non-Goals

**Goals:**
- Eliminate every direct import path of `github.com/hashicorp/terraform-plugin-sdk/v2` from provider code.
- Remove the direct `require` entry from `go.mod`.
- Preserve all existing behavior (sensitive field masking, debug transport toggling, random string generation in tests).

**Non-Goals:**
- Removing SDK v2 from the transitive dependency graph (blocked by `terraform-plugin-testing`).
- Changing any resource schema, behavior, or acceptance test logic.
- Introducing new external dependencies.

## Decisions

### 1. Re-implement `IsDebugOrHigher()` internally instead of using an external helper

**Rationale:** The SDK v2 function is trivial (~10 lines). Re-implementing avoids pulling in a large module for a single one-liner. The replacement lives in `internal/debugutils` because it is already the home of debug-logging infrastructure (`debugRoundTripper`, `PrettyPrintJSONLines`).

**Alternative considered:** Use `tflog` context-scoped checks. Rejected because `terraform-plugin-log` does not expose a "current log level" API by design.

### 2. Co-locate the `varsAreSensitive` shared helper with the new logging helper

**Rationale:** Both fleet integration policy schemas (`schema.go` and `schema_v2.go`) contain the identical expression:
```go
varsAreSensitive := !logging.IsDebugOrHigher() && os.Getenv("TF_ACC") != "1"
```
Creating a small exported function `IsSensitiveInSchema() bool` in `internal/debugutils` removes duplication and makes the intent explicit.

### 3. Delete `internal/utils/typeutils/schema.go` rather than rewrite it for Plugin Framework types

**Rationale:** The package contains only `ExpandStringSet`, which operates on SDK v2 `*schema.Set`. `grep` confirms zero call sites outside its own test file. Dead-code deletion is simpler and less error-prone than a pointless rewrite.

### 4. Fix the single stray `acctest` import rather than ignore it

**Rationale:** The SDK v2 `helper/acctest` and the `terraform-plugin-testing` `helper/acctest` packages are API-compatible for the functions we use (`RandStringFromCharSet`, `CharSetAlphaNum`). A one-line import path change resolves the last SDK v2 test reference.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Hand-rolled `IsDebugOrHigher()` might diverge from SDK v2 behavior over time | Implementation is a direct mirror of the SDK v2 logic (`TF_LOG` env var, case-insensitive match against "DEBUG"/"TRACE"). The surface area is tiny enough that divergence is unlikely. |
| `go mod tidy` might remove `sdk/v2` from `go.sum` if no direct import remains, then re-add it later via testing | Acceptable. The desired outcome is removing the **direct** require, not eradicating it from `go.sum`. |
| Deleting `typeutils/schema.go` might break something in an unreleased branch | `grep` across `main` confirms zero callers. CI build (`make build`) will confirm. |
