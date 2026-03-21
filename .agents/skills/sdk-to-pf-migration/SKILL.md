---
name: sdk-to-pf-migration
description: Guides migration of Terraform resources from Plugin SDK to Plugin Framework. Use when migrating SDK resources to PF, planning SDK-to-PF migrations, or when the user asks to migrate a resource to the Plugin Framework.
---

# SDK to Plugin Framework Migration

Migrate Terraform resources from `terraform-plugin-sdk/v2` to `terraform-plugin-framework` while preserving behavior and avoiding breaking changes.

## Prerequisites

- Provider uses mux: `provider/factory.go` combines SDK and PF via `tf6muxserver`. PF resources take precedence when both define the same type.
- Shared clients: `internal/clients/elasticsearch` (and analogous client packages) plus `internal/models` are used by both SDK and PF.

## Client Diag Strategy

**Migrate client code to return PF diags** — do not introduce a compatibility layer in the PF resource.

1. **Client layer**: Change resource-specific client functions (CRUD helpers the resource calls) to return `fwdiag.Diagnostics` instead of SDK `diag.Diagnostics`. Use `diagutil.CheckErrorFromFW()` for HTTP errors, `fwdiag.NewErrorDiagnostic()` or `diagutil.FrameworkDiagFromError()` for other errors.
2. **PF resource**: Call the client directly; append returned diags to `resp.Diagnostics` with no conversion.
3. **SDK callers** (if any remain): Update them to use `diagutil.SDKDiagsFromFramework()` when calling the migrated client.

## Migration Workflow

### 1. Migrate Client (if resource-specific)

If the resource has dedicated client functions, migrate them to return `fwdiag.Diagnostics`:

- Change return type from `sdkdiag.Diagnostics` to `fwdiag.Diagnostics`
- Use `diagutil.CheckErrorFromFW()` instead of `diagutil.CheckError()` for HTTP responses
- Use `fwdiag.NewErrorDiagnostic()` / `diagutil.FrameworkDiagFromError()` for errors

Update any SDK callers to use `diagutil.SDKDiagsFromFramework()` when calling these functions.

*(Example: ILM used `PutIlm` / `GetIlm` / `DeleteIlm` — same pattern applies to any named CRUD helpers.)*

### 2. Create PF Package

Create `internal/<domain>/<resource>/` (path mirrors where the SDK resource lived). Typical layout:

| File | Purpose |
|------|---------|
| `resource.go` | Resource struct, Metadata, Configure, ImportState |
| `schema.go` | PF schema including the appropriate connection block for the backend |
| `models.go` | Plan/State model types |
| `create.go` | Create |
| `read.go` | Read |
| `update.go` | Update |
| `delete.go` | Delete |
| `resource-description.md` | Embedded description |

**Schema**: Use the same connection block helpers as other PF resources in this provider (e.g. `providerschema.GetEsFWConnectionBlock(false)` for Elasticsearch-backed resources). Replicate the SDK schema exactly; preserve attribute names, types, and validation.

**Critical behaviors to preserve** (audit the SDK implementation for each):

- **Optional vs absent**: Flatten/expand rules where the API distinguishes “not set” from “set to empty or disabled” — e.g. ILM phase actions where `readonly` / `freeze` / `unfollow` use `enabled: false` when absent in config.
- **Version gating**: Attributes that are only valid for certain stack versions must still be rejected or handled the same way in PF.
- **JSON / structured attributes**: If the SDK used normalized JSON types to avoid perpetual diffs, keep that approach (e.g. `jsontypes.NormalizedType` or the project’s equivalent for metadata-like blobs).
- **API defaults**: When the API omits fields that Terraform previously surfaced with sentinel defaults, preserve that mapping (e.g. ILM’s `total_shards_per_node: -1` when `ES < 7.16`).

### 3. Provider Wiring

- **Add**: `provider/plugin_framework.go` — register `NewResource` in `resources()`.
- **Remove**: `provider/provider.go` — remove the type from `ResourcesMap`.

### 4. Move Acceptance Tests

- Create `internal/<domain>/<resource>/acc_test.go` with package `<resource>_test`.
- Move test functions and helpers such as `checkResourceDestroy` from the old `*_test.go`.
- Move testdata: `testdata/TestAccResourceX*/` into the new package.
- Update imports (version constants, etc.).
- Delete the old `*_test.go`.

### 5. SDK Upgrade Test

Add `TestAccResourceXFromSDK` to verify existing SDK-created state works after upgrade:

```go
//go:embed testdata/TestAccResourceXFromSDK/create/resource.tf
var sdkCreateConfig string

func TestAccResourceXFromSDK(t *testing.T) {
    name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
    resource.Test(t, resource.TestCase{
        PreCheck: func() { acctest.PreCheck(t) },
        Steps: []resource.TestStep{
            {
                ExternalProviders: map[string]resource.ExternalProvider{
                    "elasticstack": {
                        Source:            "elastic/elasticstack",
                        VersionConstraint: "0.14.3", // last SDK version for this resource
                    },
                },
                Config: sdkCreateConfig,
                ConfigVariables: config.Variables{"name": config.StringVariable(name)},
                Check: resource.ComposeTestCheckFunc(...),
            },
            {
                ProtoV6ProviderFactories: acctest.Providers,
                ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
                ConfigVariables: config.Variables{"name": config.StringVariable(name)},
                Check: resource.ComposeTestCheckFunc(...),
            },
        },
    })
}
```

Use the **last provider version where the resource was still SDK-based** for `VersionConstraint`.

### 6. Remove Old SDK Resource

- Delete the SDK resource implementation file(s).
- Move or update shared descriptions if they were only referenced from the old location.

### 7. Schema Coverage

Run schema-coverage analysis. Add tests for:

- Attributes with no coverage
- Nested blocks or phases never exercised
- Import (add `TestAccResourceX_importState` if missing)
- Update coverage for optional attributes

## Reference Implementations

Use these PF migrations as patterns (complexity varies):

- **Security user**: `internal/elasticsearch/security/user/` — includes `TestAccResourceSecurityUserFromSDK`
- **ILM**: `internal/elasticsearch/index/ilm/` — large schema, version gating, JSON normalization
- **Data stream lifecycle**: `internal/elasticsearch/index/datastreamlifecycle/` — smaller surface area

## Verification

1. `make build`
2. `go test ./internal/<domain>/<resource>/... -v`
3. Acceptance tests: `go test ./internal/<domain>/<resource>/... -v -count=1 -run TestAcc`. Follow [testing](dev-docs/high-level/testing.md) for environment requirements.
4. **Downstream resources**: Run tests for any resource that references the migrated type (e.g. resources that attach to ILM policies reference `elasticstack_elasticsearch_index_lifecycle`).

## Breaking Changes

Avoid unless unavoidable. Preserve:

- Resource type name (e.g. `elasticstack_elasticsearch_index_lifecycle` for ILM)
- Attribute paths and types
- State/ID format (composite ID, import passthrough)
