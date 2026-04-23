# Kibana Plugin Framework Generic Resource - Adoption Considerations

This document outlines follow-on adoption considerations for other Kibana Plugin Framework resources that may wish to migrate to the shared generic resource framework.

## When to Adopt

The framework is most beneficial for resources that:

- Implement standard CRUD operations (Create, Read, Update, Delete)
- Use Kibana scoped-client resolution with version enforcement
- Have space-aware composite IDs (`<space_id>/<resource_id>`)
- Follow the read-after-write pattern for authoritative state
- Need remote-not-found handling during refresh

## Model Contract Implementation

Resources must implement the `ModelContract` interface:

```go
type ModelContract[CreateRequest any, UpdateRequest any, Remote any] interface {
    KibanaConnectionModel
    IDModel
    SpaceIDModel
    VersionRequirement() VersionRequirement
    ToCreateRequest(ctx context.Context) (CreateRequest, diag.Diagnostics)
    ToUpdateRequest(ctx context.Context) (UpdateRequest, diag.Diagnostics)
    PopulateFromRemote(ctx context.Context, spaceID string, remote Remote) diag.Diagnostics
}
```

### Key Implementation Notes

1. **KibanaConnectionModel**: Return the `kibana_connection` list attribute
2. **IDModel**: Implement `GetID()` and `SetID()` for the computed `id` attribute
3. **SpaceIDModel**: Implement `GetSpaceID()` and `SetID()` for space-aware resources
4. **VersionRequirement**: Return minimum version and error messages
5. **Request Builders**: Convert Terraform model to API request types
6. **PopulateFromRemote**: Map API response back to Terraform model state

## Assembly Pattern

Each resource defines an `Assembly` that binds the model, API, and schema:

```go
type myResourceAssembly struct{}

func (a myResourceAssembly) TypeNameSuffix() string { return "kibana_my_resource" }
func (a myResourceAssembly) API() pfresource.ResourceAPI[CreateReq, UpdateReq, *RemoteModel] {
    return &myResourceAPI{}
}
func (a myResourceAssembly) NewModel() *myModel { return &myModel{} }
func (a myResourceAssembly) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    pfresource.ImportStateCompositeID(ctx, req, resp, "id", "space_id")
}
```

## Resource CRUD

The resource struct holds an `Orchestrator` and delegates operations:

```go
type MyResource struct {
    orchestrator pfresource.Orchestrator[CreateReq, UpdateReq, *RemoteModel, *myModel]
}

func (r *MyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan myModel
    // ... get plan from request
    updated, diags := r.orchestrator.Create(ctx, &plan, spaceID)
    // ... set state
}
```

## Transport API

Implement the `ResourceAPI` interface for your resource family:

```go
type ResourceAPI[CreateRequest any, UpdateRequest any, Remote any] interface {
    Create(ctx context.Context, client *kibanaoapi.Client, spaceID string, request CreateRequest) (string, diag.Diagnostics)
    Get(ctx context.Context, client *kibanaoapi.Client, spaceID string, resourceID string) (Remote, bool, diag.Diagnostics)
    Update(ctx context.Context, client *kibanaoapi.Client, spaceID string, resourceID string, request UpdateRequest) diag.Diagnostics
    Delete(ctx context.Context, client *kibanaoapi.Client, spaceID string, resourceID string) diag.Diagnostics
}
```

## Migration Checklist

1. Create `implementation.go` with model interface implementations
2. Refactor `resource.go` to use `Orchestrator`
3. Create or adapt transport API in `kibanaoapi/agentbuilderapi/` pattern
4. Remove bespoke `create.go`, `read.go`, `update.go`, `delete.go` files
5. Preserve existing schema, import behavior, and version gates
6. Run unit and acceptance tests to verify behavior

## Current Adopters

- `internal/kibana/agentbuilderagent/` - Agent Builder agents
- `internal/kibana/agentbuildertool/` - Agent Builder tools
- `internal/kibana/agentbuilderworkflow/` - Agent Builder workflows

## Future Enhancements (Out of Scope)

- Data source support in the generic framework
- Passthrough ID import helper (for non-space-aware resources)
- Post-read/post-write hooks for specialized diagnostics
