---
name: sdk-to-pf-migration
description: Guides migration of Terraform resources from Plugin SDK to Plugin Framework. Use when migrating SDK resources to PF, planning SDK-to-PF migrations, or when the user asks to migrate a resource to the Plugin Framework.
---

# SDK to Plugin Framework Migration

Migrate Terraform resources from `terraform-plugin-sdk/v2` to `terraform-plugin-framework` while preserving behavior and avoiding breaking changes.

## Prerequisites

- Provider uses mux: `provider/factory.go` combines SDK and PF via `tf6muxserver`. PF resources take precedence when both define the same type.
- Shared clients: `internal/clients/elasticsearch` and `internal/models` are used by both SDK and PF.

## Client Diag Strategy

**Migrate client code to return PF diags** — do not introduce a compatibility layer in the PF resource.

1. **Client layer**: Change client functions (e.g. `PutIlm`, `GetIlm`, `DeleteIlm`) to return `fwdiag.Diagnostics` instead of SDK `diag.Diagnostics`. Use `diagutil.CheckErrorFromFW()` for HTTP errors, `fwdiag.NewErrorDiagnostic()` or `diagutil.FrameworkDiagFromError()` for other errors.
2. **PF resource**: Call the client directly; append returned diags to `resp.Diagnostics` with no conversion.
3. **SDK callers** (if any remain): Update them to use `diagutil.SDKDiagsFromFramework()` when calling the migrated client.

## Migration Workflow

### 1. Migrate Client (if resource-specific)

If the resource has dedicated client functions (e.g. `PutIlm`, `GetIlm`, `DeleteIlm`), migrate them to return `fwdiag.Diagnostics`:

- Change return type from `sdkdiag.Diagnostics` to `fwdiag.Diagnostics`
- Use `diagutil.CheckErrorFromFW()` instead of `diagutil.CheckError()` for HTTP responses
- Use `fwdiag.NewErrorDiagnostic()` / `diagutil.FrameworkDiagFromError()` for errors

Update any SDK callers to use `diagutil.SDKDiagsFromFramework()` when calling these functions.

### 2. Create PF Package

Create `internal/<domain>/<resource>/` (e.g. `internal/elasticsearch/index/ilm/`). Structure:

| File | Purpose |
|------|---------|
| `resource.go` | Resource struct, Metadata, Configure, ImportState |
| `schema.go` | PF schema with `elasticsearch_connection` block |
| `models.go` | Plan/State model types |
| `create.go` | Create |
| `read.go` | Read |
| `update.go` | Update |
| `delete.go` | Delete |
| `resource-description.md` | Embedded description |

**Schema**: Use `providerschema.GetEsFWConnectionBlock(false)` for connection. Replicate SDK schema exactly; preserve attribute names, types, and validation.

**Critical behaviors to preserve**:
- Flatten logic for optional/empty vs absent (e.g. `readonly`/`freeze`/`unfollow` with `enabled: false` when absent)
- Version gating for attributes (reject unsupported per ES version)
- JSON diff suppression: use `jsontypes.NormalizedType` or equivalent for metadata/allocate attributes
- Defaults for missing API values (e.g. `total_shards_per_node: -1` when ES < 7.16)

### 3. Provider Wiring

- **Add**: `provider/plugin_framework.go` — add `NewResource` to `resources()`.
- **Remove**: `provider/provider.go` — remove from `ResourcesMap`.

### 4. Move Acceptance Tests

- Create `internal/<domain>/<resource>/acc_test.go` with package `<resource>_test`.
- Move test functions and `checkResourceDestroy` from old `*_test.go`.
- Move testdata: `testdata/TestAccResourceX*/` into the new package.
- Update imports (version constants, etc.).
- Delete old `*_test.go`.

### 5. SDK Upgrade Test

Add `TestAccResourceXFromSDK` to verify existing SDK-created resources work after upgrade:

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
                        VersionConstraint: "0.14.3", // last SDK version
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

Use the **last provider version where the resource was SDK-based** for `VersionConstraint`.

### 6. Remove Old SDK Resource

- Delete the SDK resource file (e.g. `internal/elasticsearch/index/ilm.go`).
- Move or update shared descriptions if needed.

### 7. Schema Coverage

Run schema-coverage analysis. Add tests for:
- Attributes with no coverage
- Phases/blocks never exercised
- Import (add `TestAccResourceX_importState` if missing)
- Update coverage for optional attributes

## Reference Implementations

- **Security user**: `internal/elasticsearch/security/user/` — PF resource, `TestAccResourceSecurityUserFromSDK`
- **ILM**: `internal/elasticsearch/index/ilm/` — complex schema, version gating
- **Datastream lifecycle**: `internal/elasticsearch/index/datastreamlifecycle/` — simpler PF resource

## Verification

1. `make build`
2. `go test ./internal/<domain>/<resource>/... -v`
3. Acceptance tests: `go test ./internal/<domain>/<resource>/... -v -count=1 -run TestAcc`. Use the testing instructions (dev-docs/high-level/testing.md) for environmental requirements.
4. Dependent resources: run tests for resources that reference the migrated resource (e.g. `templateilmattachment` references `elasticstack_elasticsearch_index_lifecycle`)

## Breaking Changes

Avoid unless unavoidable. Preserve:
- Resource type name (e.g. `elasticstack_elasticsearch_index_lifecycle`)
- Attribute paths and types
- State/ID format (composite ID, import passthrough)
