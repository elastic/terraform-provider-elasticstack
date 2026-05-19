## Why

The provider is already fully migrated to the Terraform Plugin Framework, yet `github.com/hashicorp/terraform-plugin-sdk/v2` remains a direct dependency in `go.mod`. The only remaining imports are a handful of utility functions (`helper/logging.IsDebugOrHigher`, `helper/schema.Set`) and one stray test import of the SDK v2 `acctest` package. Removing this direct dependency cleans up the module declaration, signals that the provider is Plugin Framework-only, and eliminates the risk of future regressions where SDK v2 types/functions are reintroduced.

## What Changes

- **Remove** `internal/utils/typeutils/schema.go` and `internal/utils/typeutils/schema_test.go` (dead code; `ExpandStringSet` is unused in production).
- **Replace** all uses of `github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging.IsDebugOrHigher()` with a new internal helper in `internal/debugutils/logging.go`.
- **Deduplicate** the `varsAreSensitive` pattern used in fleet integration policy schema definitions into a shared helper (co-located with the `IsDebugOrHigher` replacement).
- **Fix** the single stray test import in `internal/elasticsearch/index/templateilmattachment/acc_test.go` to use `terraform-plugin-testing/helper/acctest` instead of the SDK v2 version.
- **Drop** the direct `require` line for `github.com/hashicorp/terraform-plugin-sdk/v2` from `go.mod` and run `go mod tidy`; SDK v2 will remain as an indirect dependency via `terraform-plugin-testing`.

## Capabilities

### New Capabilities
<!-- No new user-facing capabilities. -->

### Modified Capabilities
<!-- No spec-level behavior changes. This is an internal dependency cleanup. -->

## Impact

- **go.mod / go.sum**: SDK v2 removed from `require` block; downgraded to `// indirect`.
- **`internal/clients/config/elasticsearch.go`, `internal/clients/fleet/client.go`, `internal/clients/kibanaoapi/client.go`**: Import path change for `IsDebugOrHigher`.
- **`internal/fleet/integration_policy/schema.go`, `internal/fleet/integration_policy/schema_v2.go`**: Import path change; use shared `IsSensitiveInSchema()` helper.
- **`internal/utils/typeutils/schema.go`, `internal/utils/typeutils/schema_test.go`**: Deleted.
- **`internal/elasticsearch/index/templateilmattachment/acc_test.go`**: Import path change for `acctest`.
