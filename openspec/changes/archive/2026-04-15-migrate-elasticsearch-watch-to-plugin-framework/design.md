## Context

`elasticstack_elasticsearch_watch` is still implemented in `internal/elasticsearch/watcher/watch.go` with the Terraform Plugin SDK, while the provider already serves Plugin Framework resources through the muxed proto6 provider. The watch migration therefore has two moving parts: the watch-specific Elasticsearch client helpers must return framework diagnostics, and the new Framework resource must preserve the existing resource contract, including schema shape, import format, composite IDs, and upgrade compatibility with SDK-managed state.

This change is intentionally separate from the watch hardening change. The migration should build on the corrected watch behavior and stronger acceptance suite rather than mixing framework porting with behavior fixes.

## Goals / Non-Goals

**Goals:**
- Reimplement the watch resource with the Terraform Plugin Framework while preserving the existing external resource contract.
- Convert watch-specific Elasticsearch helper functions to framework diagnostics so the Framework resource can call them directly.
- Preserve the composite `id`, import ID format, connection override behavior, and watch CRUD semantics.
- Add upgrade coverage proving a watch created by the last SDK-backed provider release can be managed by the new Framework resource without recreation.

**Non-Goals:**
- Changing the resource type name, attribute names, defaults, or import format.
- Redesigning watch behavior beyond the separately tracked hardening fix.
- Introducing a broad compatibility layer that leaves the watch helper APIs SDK-first after the migration.

## Decisions

Model the migrated resource after existing Elasticsearch Plugin Framework resources.
The new watch package should follow the repo's standard Framework layout with `resource.go`, `schema.go`, `models.go`, and CRUD files. It should use `providerschema.GetEsFWConnectionBlock()` plus `ProviderClientFactory.GetElasticsearchClient()` so connection override handling matches other Elasticsearch resources.

Alternative considered: keep most logic inside the existing SDK file and adapt it minimally.
Rejected because it would preserve SDK-specific patterns and make future maintenance inconsistent with the rest of the Framework resources.

Use framework-native diagnostics in `internal/clients/elasticsearch/watch.go`.
The watch helpers are only used by the watch resource, so converting them to framework diagnostics directly keeps the Framework resource simple and avoids adding conversion glue in each CRUD method.

Alternative considered: leave the helpers SDK-based and convert diagnostics inside the new resource.
Rejected because it preserves the wrong abstraction boundary and makes the migrated resource responsible for compatibility concerns that belong in the helper layer.

Represent JSON attributes with normalized JSON handling in the Framework schema.
The SDK resource uses JSON strings plus `DiffSuppressFunc` to avoid semantically meaningless diffs. The Framework resource should preserve that behavior with the repo's normalized JSON string type so existing watch JSON continues to round-trip without perpetual diffs.

Alternative considered: use plain `types.String` attributes for all JSON fields.
Rejected because it would reintroduce formatting-sensitive diffs and weaken parity with the SDK behavior.

Add an SDK-to-Framework acceptance test pinned to the last SDK-backed release.
The migration needs an explicit upgrade test that creates the resource with the last published SDK implementation and then refreshes it with the new Framework implementation.

Alternative considered: rely only on regular acceptance tests after the port.
Rejected because those tests do not prove existing released state upgrades cleanly.

Only add a Framework state upgrader if the upgrade test proves one is necessary.
The desired outcome is a straight migration with no explicit upgrader. A state upgrader should be introduced only if the new schema or custom JSON types cannot consume SDK-authored state directly.

## Risks / Trade-offs

- [Risk] The Framework JSON type or state model may not deserialize existing SDK state exactly as expected. -> Mitigation: add a `FromSDK` acceptance test first and introduce a minimal state upgrader only if the test demonstrates a compatibility gap.
- [Risk] Moving the resource from SDK to Framework could accidentally change ID, import, or default behavior. -> Mitigation: preserve the existing acceptance coverage and extend it with explicit upgrade coverage.
- [Risk] Registering the resource on both sides of the mux would create undefined behavior. -> Mitigation: add the new Framework resource to `provider/plugin_framework.go` and remove the SDK registration from `provider/provider.go` in the same change.
- [Risk] The helper migration could leave lingering SDK callers or conversions. -> Mitigation: keep the watch helper changes scoped to watch-only call sites and verify no other package still imports the old signatures.

## Migration Plan

1. Convert `internal/clients/elasticsearch/watch.go` to framework diagnostics.
2. Replace the SDK watch resource implementation with a Framework package that preserves schema, ID, import, and connection behavior.
3. Register the Framework watch resource and remove the SDK registration from the muxed SDK provider.
4. Move or adapt watch acceptance coverage into the Framework package layout and add a `FromSDK` upgrade test pinned to `0.14.3`.
5. Run build and focused watch tests; add a state upgrader only if the upgrade path fails without one.

## Open Questions

- Whether a dedicated Framework state upgrader is needed should be decided by the upgrade test, not assumed up front.
