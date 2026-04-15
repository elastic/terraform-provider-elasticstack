## Context

After the Kibana/Fleet phase, the provider will already inject a `ProviderClientFactory` and Kibana/Fleet code will resolve typed scoped clients from `kibana_connection`. Elasticsearch will still be the remaining broad-client domain: resources will continue to resolve `*clients.APIClient`, shared sinks in `internal/clients/elasticsearch/` will continue accepting that broad type, and `analysis/esclienthelperplugin` will still enforce correct provenance through expensive static analysis.

This second phase removes that last broad-client path. It extends the factory so Elasticsearch also resolves typed scoped clients from `elasticsearch_connection`, migrates sink boundaries to those typed clients, and then deletes the custom lint rule entirely because the compiler will enforce the same constraint at the API boundary.

## Goals / Non-Goals

**Goals:**
- Extend `ProviderClientFactory` so Elasticsearch entities resolve typed Elasticsearch scoped clients rather than broad `*clients.APIClient` values.
- Change Elasticsearch sink and helper APIs to require typed Elasticsearch scoped clients.
- Preserve current `elasticsearch_connection` behavior while replacing linter-based provenance checks with compile-time type checking.
- Remove `analysis/esclienthelperplugin` and its repository wiring entirely once typed sink boundaries are in place.

**Non-Goals:**
- Revisiting the Kibana/Fleet typed-client design from phase 1 except where the shared factory contract must expand.
- Changing the `elasticsearch_connection` schema surface or coverage rules.
- Keeping a reduced version of the current custom provenance lint as a permanent backstop.
- Redesigning unrelated provider resource behavior beyond client resolution and enforcement.

## Decisions

Extend the existing provider factory rather than introducing a second provider data contract.
The same `ProviderClientFactory` introduced for Kibana/Fleet should gain typed Elasticsearch resolution methods and become the only provider data contract for all entity families.

Alternative considered: leave the phase 1 factory untouched and inject a separate Elasticsearch resolver.
Rejected because that would reintroduce multiple provider data types and prolong the mixed model.

Introduce a concrete `ElasticsearchScopedClient` type and make Elasticsearch sinks accept it directly.
The typed Elasticsearch client should own the Elasticsearch API client plus the existing Elasticsearch-derived helper behavior that resources rely on today, including composite ID generation, cluster identity lookup, version checks, flavor checks, and minimum-version enforcement.

Alternative considered: keep sink signatures broad and rely on resources to resolve the right type before calling them.
Rejected because compile-time enforcement only becomes real when sink signatures stop accepting the broad client type.

Remove transitional legacy Elasticsearch factory methods after migration.
Once Elasticsearch resources and sinks use `ElasticsearchScopedClient`, the temporary broad-client resolution methods kept for phase 1 should be removed from the factory contract so there is no supported path back to the old model.

Alternative considered: keep legacy broad-client methods indefinitely for tests or convenience.
Rejected because that would leave ad-hoc broad-client construction available and weaken the clarity of the new contract.

Delete the custom Elasticsearch provenance analyzer after typed migration completes.
When no in-scope Elasticsearch sink accepts `*clients.APIClient`, the custom lint no longer provides unique protection and should be removed from the repo, test suite, and lint workflow.

Alternative considered: keep the analyzer for belt-and-suspenders verification.
Rejected because it duplicates compiler guarantees while continuing to impose maintenance and performance cost.

## Risks / Trade-offs

- [Risk] Elasticsearch resources use more `APIClient` helper behavior than Kibana/Fleet, so the new scoped type may initially miss methods -> Mitigation: define the typed Elasticsearch contract from actual call-site usage and migrate sink packages first.
- [Risk] Removing the analyzer too early could create an enforcement gap -> Mitigation: delete the analyzer only after all in-scope Elasticsearch sinks and resource call sites no longer accept the broad client type.
- [Risk] Tests and helper utilities may rely on constructing `*clients.APIClient` directly -> Mitigation: add typed test helpers and update acceptance/unit tests alongside the sink migration.
- [Risk] The canonical lint capability could drift from implementation during retirement -> Mitigation: explicitly remove the lint requirements in this change and tie deletion to the typed sink migration tasks.

## Migration Plan

1. Extend `ProviderClientFactory` with typed Elasticsearch resolution methods and add `ElasticsearchScopedClient`.
2. Convert `internal/clients/elasticsearch/` sink and helper packages to accept the typed scoped client.
3. Migrate Framework and SDK Elasticsearch resources/data sources to resolve typed scoped clients from `elasticsearch_connection`.
4. Remove the temporary legacy Elasticsearch factory methods introduced in phase 1.
5. Delete `analysis/esclienthelperplugin`, its tests, `.golangci.yaml` wiring, and associated lint workflow hooks.

## Open Questions

- None. The main dependency is that the Kibana/Fleet-first factory contract has already landed before this change begins.
