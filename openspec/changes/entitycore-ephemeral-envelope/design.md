## Context

The Plugin Framework's ephemeral resource contract (`ephemeral.EphemeralResource` plus `EphemeralResourceWithConfigure` and `EphemeralResourceWithClose`) is structurally similar to `resource.Resource` but with two crucial differences:

1. **No state.** Values returned by `Open()` live only in `OpenResponse.Result` for the current plan/apply. They are never persisted. `Close()` receives a `CloseRequest` containing only `Private`, not the original config or the prior Result.
2. **Open runs on every plan and apply.** The provider cannot detect or opt out of this; it is inherent to the framework contract.

The existing `entitycore.ResourceBase` / `NewElasticsearchResource[T]` / `NewKibanaResource[T]` envelopes solve a wide range of resource-lifecycle concerns (Configure, Metadata, Schema injection, scoped client resolution, version requirements, composite ID resolution, read-after-write). The two differences above mean those envelopes cannot be directly reused, but the bulk of the prelude logic is identical and worth sharing.

The archived `elasticsearch-security-api-key-ephemeral` change is the only existing ephemeral resource in the provider. Its review-and-CI iteration loop surfaced three classes of bug that this envelope is designed to make structurally impossible.

## Goals

- Provide `entitycore.NewElasticsearchEphemeralResource[T, S]` and `entitycore.NewKibanaEphemeralResource[T, S]` with the same call-site shape as the existing resource and data source envelopes.
- Own all Plugin Framework wiring (factory conversion, Metadata, Configure, Schema with connection-block injection, version-requirement enforcement, scoped client resolution, Result/Private serialization).
- Make Close-state round-tripping safe by construction: the envelope owns the private-state codec and refuses (loudly, at construction) any `S` that contains plugin-framework types.
- Provide a typed Close-time API: the Close callback receives `(ctx, client, CloseRequest[S])` where `client` is already resolved from the snapshotted connection and `S` is unmarshaled — the callback never touches `Private`.
- Migrate `elasticstack_elasticsearch_security_api_key` (ephemeral) to the new envelope in this change to validate the abstraction against a real resource and a real acceptance suite.

## Non-Goals

- `Renew()` support. The framework's `EphemeralResourceWithRenew` is intentionally omitted; no current or near-term resource needs server-side renewal, and adding the seam later is non-breaking.
- Server-side suppression of Open during `terraform plan`. The framework provides no signal to distinguish plan from apply, and no provider-side mechanism can change Terraform's call pattern. Mitigations are documentation-only.
- Refactoring the existing managed `elasticstack_elasticsearch_security_api_key` resource (the non-ephemeral variant). Unchanged.
- Splitting the Elasticsearch and Kibana envelopes into separate spec capabilities. They are parallel, share the same close-state machinery, and are documented together — mirroring the existing `entitycore-datasource-envelope` spec which covers both flavors.
- A custom `golangci-lint` analyzer to enforce plain-Go `S`. The constructor reflect-check is sufficient (see Decisions).

## Decisions

| Topic | Decision |
|---|---|
| Constructor signature | `NewElasticsearchEphemeralResource[T, S](name string, opts ElasticsearchEphemeralOptions[T, S]) ephemeral.EphemeralResource` and the Kibana equivalent. Name + options struct, matching `NewElasticsearchResource[T]`. |
| Connection block injection | Envelope injects an optional `elasticsearch_connection` / `kibana_connection` block, identical pattern to the resource envelope's `Schema` override. Reuses the existing `providerschema.GetEsEphemeralConnectionBlock()` helper; adds a Kibana counterpart if absent. |
| Connection field embedding | Concrete models reuse the existing `entitycore.ElasticsearchConnectionField` and `entitycore.KibanaConnectionField` embeds. No new embed type is introduced. |
| Version requirements | Envelope calls `EnforceVersionRequirements(ctx, client, &model)` in the Open prelude, identical to the resource and data source envelopes. The same `VersionRequirement` opt-in interface applies; no new mechanism is introduced. |
| Open / Close request/response shapes | Structured types: `OpenRequest[T]{Config T}`, `OpenResult[T, S]{Model T, CloseState S}`, `CloseRequest[S]{State S}`, `CloseResponse{}`. Mirrors the resource envelope's `WriteRequest[T]` / `WriteResult[T]`. Future extension (e.g. carrying private-state diagnostics or carrying ID hints) does not require breaking the callback signature. |
| Close callback nullability | **Required**, non-nil. The envelope's plain-Go-S check, connection round-trip, and client resolution are all motivated by Close needing to make API calls; a resource that genuinely needs no Close work can pass a no-op callback in two lines. There is no realistic value in a "no Close callback" branch. |
| Close-state codec | JSON. `S` is JSON-marshaled into a single envelope-owned private slot (`entitycore.ephemeral.user_state`). Sufficient for every realistic ephemeral resource; debuggable; no exotic dependencies. |
| Connection-snapshot codec | Separate envelope-owned private slot (`entitycore.ephemeral.connection`). The snapshot struct uses plain Go types (string, `[]string`, `map[string]string`, `*bool`), avoiding the `tfsdk`-type JSON round-trip bug class. The user never sees the snapshot type. |
| Plain-Go `S` enforcement | **Constructor reflect-check**. At construction time, recursively reflect over the fields of `S` (handling embedded structs, pointers, slices, maps, and arrays) and `panic` with a message of the form `entitycore: ephemeral close state <type> has field <path> of plugin-framework type <type.PkgPath/Name>; Close state must be plain Go types only` if any field's `PkgPath()` is `github.com/hashicorp/terraform-plugin-framework/types`. The check runs on the first construction in any process; the existing `TestNewEphemeralResource_Interfaces`-style test (which every ephemeral resource has) exercises it for free. A custom golangci-lint analyzer was evaluated and rejected: see "Alternatives". |
| Renew support | Not implemented. `EphemeralResourceWithRenew` is not exposed by the envelope. Adding it later is a non-breaking addition (a new `Renew` field on `Options`). |
| Open-on-plan footgun | Document-only. `internal/entitycore/doc.go` describes Open as called on plan and apply; the existing api_key docs template already warns users. No envelope-side mitigation is technically possible. |
| Single spec or two | One spec capability (`entitycore-ephemeral-envelope`) covering both Elasticsearch and Kibana variants. Same pattern as `entitycore-datasource-envelope`. The split between `entitycore-resource-envelope` and `entitycore-kibana-resource-envelope` predates that pattern and is not the precedent to follow for a new capability. |
| api_key migration scope | Same change. Validates the abstraction against a real resource and the existing acceptance suite. The migration is implementation-only — schema shape, validator behavior, version-gating, and public docs are preserved. |
| Existing api_key spec | Not modified. The archived `elasticsearch-security-api-key-ephemeral` change has no synced spec under `openspec/specs/`. The migration is covered by acceptance tests, not by a delta spec. |

## Risks / Trade-offs

- **Constructor panic at runtime if S has tfsdk types.** A misuse surfaces as `panic` at provider start (or on first call to `EphemeralResources(ctx)` if registration is lazy). This is loud and unambiguous, but it means a misuse caught only in production would crash the provider. The mitigation is that every new ephemeral resource ships with at least one unit test that constructs the resource (the interface-implements test pattern), and that test runs in the standard `go test ./...` pass. If a future resource lacks even that test, we accept a small risk of a late-binding panic in exchange for the cost-free guard.
- **Generic-instantiation explosion.** Two type parameters per constructor (`T`, `S`) plus the structured request/response types create more generic surface than the resource envelope. The downstream effect on compile time is negligible — the resource envelope already uses generics at this scale — but the API reads as denser. Mitigated by clear `doc.go` examples and by the api_key migration as a reference implementation.
- **JSON serialization of `S`** silently elides unexported fields and types that don't round-trip cleanly (e.g. `time.Time` with custom locations). The plain-Go-only rule plus standard JSON conventions are sufficient for the resource roster we expect; we accept the constraint that `S` must be JSON-round-trippable in addition to being free of `tfsdk` types.
- **Coupling api_key to a new envelope in the same change.** A bug in the envelope blocks the api_key migration. Mitigation: keep the migration in a separate commit at the end of the implementation task list so it can be reverted independently if the envelope work goes sideways. The migration commit is small (the schema/validator/test code is preserved verbatim).

## Alternatives

### Custom `golangci-lint` analyzer for plain-Go `S`

Considered and rejected for this change. The repo already has the plugin infrastructure (`analysis/acctestconfigdirlintplugin`), so it is technically tractable. However:

- The analyzer would need to resolve `S` from a generic-instantiation call site (`entitycore.NewElasticsearchEphemeralResource[apikey.tfModel, apikey.closeState](...)`), which is meaningfully more complex than the existing analyzer's "find a call and inspect a literal" pattern. `go/types` exposes `*types.Named.TypeArgs()` but the surface area to cover (type aliases, generic `S`, transitive embedded fields) roughly quintuples the analyzer LOC.
- The constructor reflect-check costs ~30 LOC, lives next to the type it protects, and is exercised by any test that touches the constructor — which every new ephemeral resource already ships with.
- If a second structural rule about ephemerals appears later (e.g. "T must embed `ElasticsearchConnectionField`", "S must have JSON tags on exported fields"), the analyzer's marginal cost drops and an analyzer becomes the better factoring. For a single rule, the runtime check wins.

The decision is revisitable when the second rule arrives.

### Bring-your-own codec for `S`

Considered: the envelope owns the private-state key and the bytes, the user supplies `Encode(S) ([]byte, diag.Diagnostics)` and `Decode([]byte) (S, diag.Diagnostics)`. Rejected: every realistic ephemeral resource will wrap `encoding/json` with a small struct. Mandating a callback per resource buys flexibility nobody needs and costs boilerplate everybody pays.

### Single envelope covering both flavors via a runtime client-kind switch

Considered: one `NewEphemeralResource[T, S]` with a `Component` enum. Rejected for the same reason the resource envelopes are split: the scoped client type differs (`*clients.ElasticsearchScopedClient` vs `*clients.KibanaScopedClient`), which would force the Open/Close callbacks to accept `any` and downcast. The existing split-by-flavor pattern is the right precedent.

## Open Questions

1. **Kibana ephemeral connection block helper** — does `providerschema` already export an ephemeral-namespaced `GetKbEphemeralConnectionBlock()`? If not, the change adds one (parallel to `GetEsEphemeralConnectionBlock`). Implementation will confirm and surface a tiny addition if needed.
2. **Reserved private-state keys** — `entitycore.ephemeral.user_state` and `entitycore.ephemeral.connection` are proposed names. Confirm there is no existing convention for entitycore-owned private-state keys elsewhere in the codebase; if there is, follow it.
3. **Should the envelope expose a hook to merge user-owned private-state entries?** Out of scope for v1. A resource that needs additional private-state slots beyond `S` can use a richer `S` struct. Revisit only if a real consumer needs separate slots (e.g. for incremental writes).

## Migration / State

Ephemeral resources have no persistent state. The api_key migration does not require a state upgrade and produces no diff against existing acceptance fixtures.
