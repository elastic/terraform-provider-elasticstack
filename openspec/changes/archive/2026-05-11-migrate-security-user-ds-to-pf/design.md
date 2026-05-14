## Context

The `elasticstack_elasticsearch_security_user` data source currently lives in `internal/elasticsearch/security/user_data_source.go` as a Terraform Plugin SDK v2 `*schema.Resource`. It is relatively simple: a required `username` input, computed scalar outputs (`full_name`, `email`, `enabled`), a computed set of strings (`roles`), and a computed JSON string (`metadata`). The resource portion is already PF-based.

## Goals / Non-Goals

**Goals:**
- Migrate the security user data source to Plugin Framework.
- Wrap with `entitycore.NewElasticsearchDataSource`.
- Maintain 1:1 schema parity.
- Preserve not-found semantics (remove from state without error).

**Non-Goals:**
- No changes to the security user resource.
- No new attributes.

## Decisions

### Decision: `metadata` uses `jsontypes.NormalizedType{}`

**Chosen:** The `metadata` attribute uses `CustomType: jsontypes.NormalizedType{}`.

**Rationale:** Consistent with PF JSON string convention in this provider.

### Decision: Nil fields default to empty strings

**Chosen:** In the read callback, when `user.Email` or `user.FullName` are nil, the model sets `types.StringValue("")` rather than `types.StringNull()`.

**Rationale:** Matches the current SDK behavior which explicitly sets `""` via `d.Set("email", "")`. This avoids a semantic drift where PF would naturally use `Null` for absent values.

### Decision: Not-found returns an empty id without diagnostics

**Chosen:** When `elasticsearch.GetUser` returns nil with no error, the callback returns a model with `ID` set to `types.StringValue("")` and returns no warning or error diagnostic.

**Rationale:** This preserves the current SDK behavior as closely as the PF data source envelope allows: Terraform persists the returned model, but the empty `id` signals that the user was not found without introducing a new diagnostic.

## Risks / Trade-offs

- **Not-found behavior approximation.** SDK data sources clear their ID directly with `d.SetId("")`; PF data sources persist the callback result. The callback must therefore return an empty `id` and avoid warning diagnostics so callers see the same no-error missing-user behavior.

## Migration Plan

No data migration required. Schema shape preserved.

## Open Questions

None.
