## Context

`internal/resourcecore` already defines the provider-wide Plugin Framework resource wiring for Terraform type-name construction, provider-data conversion, and client-factory storage. The first rollout intentionally stopped at four pilot resources so the team could validate promoted-method behavior, import boundaries, and readability before widening adoption.

Issue [#2454](https://github.com/elastic/terraform-provider-elasticstack/issues/2454) shows that many remaining Plugin Framework resources still duplicate the same `Configure` and `Metadata` bodies that `resourcecore` was created to own. Most of that duplication is in Fleet and Kibana resources whose `Configure` behavior already matches the canonical early-return semantics in `resourcecore`. For this change, **one** additional bootstrap mismatch is explicitly in scope: assigning the converted factory **before** returning when conversion diagnostics are present (instead of `resourcecore`’s early return without replacing the stored factory). After re-auditing `elasticstack_elasticsearch_security_api_key`, the previously cited package-level `configuredResources` slice turned out to be unused, so the resource is now included in the rollout as well.

## Goals / Non-Goals

**Goals:**
- Migrate the remaining compatible Plugin Framework resources that manually duplicate `client`, `Configure`, and `Metadata` wiring to embed `*resourcecore.Core`.
- Preserve each migrated resource's Terraform type name, import semantics, CRUD logic, schema, and state-upgrade behavior exactly.
- Extend verification so the broader rollout remains safe after moving beyond the original four-resource pilot.
- Capture the rollout requirement in the existing `provider-framework-resource-core` capability rather than leaving the spec scoped only to pilots.

**Non-Goals:**
- Changing Terraform resource names, schema shape, import identifiers, or external API behavior.
- Converting Plugin Framework data sources or SDK resources.
- Migrating Plugin Framework resources whose `Configure` does more than bootstrap wiring and the accepted assign-before-return pattern.
- Moving CRUD, validation, or state-upgrade logic into `resourcecore`.

## Decisions

### Scope the rollout to resources compatible with `resourcecore`, plus an approved `Configure` edge case

This change migrates resources whose bootstrap is equivalent to the shared core **or** whose **only** remaining mismatch is assigning the converted factory before returning when conversion diagnostics are present (see [`resource-inventory.md`](./resource-inventory.md)). After migration, those resources use [`Core.Configure`](../../../internal/resourcecore/core.go) (early return without assigning a new factory when errors are present). `elasticstack_elasticsearch_security_api_key` is now included because its previously suspected package-level side effect (`configuredResources`) was unused and has been removed.

**Alternative considered:** treat assign-before-return Elasticsearch resources as permanently out of scope. Rejected after scope revision: that single mismatch is accepted so the rollout can remove duplicated bootstrap without tackling unrelated `Configure` side effects.

### Preserve resource-local import ownership and all non-bootstrap behavior

Each migrated resource will embed `*resourcecore.Core`, initialize it in the constructor with the correct component and literal resource-name suffix, and replace direct client-field reads with `r.Client()`. The concrete resource will keep its explicit `ImportState`, schema, CRUD, and state-upgrade methods unchanged.

This preserves the resourcecore contract that import behavior remains local to the resource, while still removing the duplicated `client` field plus `Configure`/`Metadata` bodies.

**Alternative considered:** add more behavior to `resourcecore`, such as default import handling or shared CRUD helpers. Rejected because issue `#2454` is specifically about bootstrap duplication, and widening the abstraction would create a much larger design surface.

### Use literal component and resource-name pairs to preserve exact type names

Every migrated resource will configure the core with the exact component and literal suffix that reproduces its existing Terraform type name, including legacy spellings. This keeps the rollout safe for resources such as `agentbuilder_workflow`-style packages whose Go package names, type names, and Terraform suffixes are not always trivially derivable from one another.

**Alternative considered:** derive the core suffix from package names automatically. Rejected because the current provider still contains legacy naming variants and inconsistent package naming patterns.

### Verify the widened rollout with a provider registry test plus targeted package coverage

The existing `internal/resourcecore` unit and conformance tests already protect core semantics. This change will add a provider-package unit test that iterates the resources registered in `provider/plugin_framework.go`, instantiates each registered resource, and asserts that the concrete resource embeds `*resourcecore.Core`. That provider-level inventory test will be complemented by targeted tests for the shared core and for representative migrated packages across the widened rollout, especially resources with passthrough import, custom import, and no import support.

**Alternative considered:** rely only on compilation and the existing shared-core tests. Rejected because the widened rollout spans many packages and import variants, and the provider registry is the actual source of truth for what the provider exposes.

## Risks / Trade-offs

- [A resource may look mechanically compatible while still relying on slightly different `Configure` behavior] → Audit each candidate resource before conversion and leave outliers for a separate follow-up.
- [Embedding can obscure where the client now comes from during review] → Keep constructor initialization explicit and use `r.Client()` consistently so client access remains easy to trace.
- [Literal suffix mistakes could silently change Terraform type names] → Preserve type-name strings by mapping each resource to an explicit component plus literal suffix and cover representative names in tests.
- [The rollout may touch many packages at once] → Convert resources in small groups, grouped by component or import shape, so regressions are easier to isolate.

## Migration Plan

1. Inventory the remaining Plugin Framework resources that still duplicate canonical `client` / `Configure` / `Metadata` wiring and separate compatible resources from outliers. **Task 1 output:** see [`resource-inventory.md`](./resource-inventory.md). **Revised scope:** audited Elasticsearch resources whose only bootstrap mismatch is assign-before-return are migrated in this change; after removing the unused `configuredResources` slice, `elasticstack_elasticsearch_security_api_key` is migrated too.
2. Convert compatible resources to embed `*resourcecore.Core`, initialize the core in constructors, and replace direct client-field usage with `Client()`.
3. Keep explicit `ImportState`, schema, CRUD, and state-upgrade methods unchanged for every converted resource.
4. Add the provider-package registry test for `provider/plugin_framework.go`, then run targeted tests for `./internal/resourcecore/...` and representative migrated packages that cover passthrough import, custom import, and no-import resource shapes.
5. Leave any audited incompatibilities out of scope for this change and capture them in follow-up work if needed.

Rollback is straightforward because the rollout is internal-only: each converted resource can revert to its prior explicit `client`, `Configure`, and `Metadata` implementation without state migration.

## Open Questions

- Whether the provider registry test should cover only the standard resource set from `Provider.resources(...)` or also the experimental resource set returned by `Provider.experimentalResources(...)`.
