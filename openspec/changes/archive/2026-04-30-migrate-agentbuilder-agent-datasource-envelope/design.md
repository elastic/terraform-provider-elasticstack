## Context

The `entitycore` package currently has two data source patterns:

- `DataSourceBase`, which provides embedded Configure, Metadata, and client factory access for data sources that keep a concrete `Read` method.
- `NewKibanaDataSource`, which wraps a schema factory and read callback with generic config decode, Kibana scoped client resolution, callback invocation, and state persistence.

The Agent Builder agent data source currently uses neither pattern. It stores the provider client factory directly, implements `Metadata` and `Configure`, injects its own `kibana_connection` schema block, and performs the whole read flow locally.

The read path has two version concerns. The Agent Builder agent API requires `minKibanaAgentBuilderAPIVersion` before the data source can operate at all. Workflow-type tool dependency export requires `minVersionAdvancedAgentConfig`, but only after reading the agent and discovering workflow-type tools.

## Goals / Non-Goals

**Goals:**

- Migrate the Agent Builder agent data source to the generic Kibana data source envelope.
- Add optional model-specific pre-read version requirements to the Kibana envelope.
- Preserve the Terraform type name `elasticstack_kibana_agentbuilder_agent`.
- Preserve existing data source behavior for ID resolution, spaces, `kibana_connection`, dependency export, tool ordering, and workflow YAML export.
- Keep conditional version checks that depend on fetched API data in entity-specific read logic.

**Non-Goals:**

- Do not migrate Agent Builder resources to a resource envelope.
- Do not change the user-visible data source schema.
- Do not make version requirements mandatory for every `KibanaDataSourceModel`.
- Do not attempt to express post-fetch conditional requirements through the generic envelope in this change.

## Decisions

### 1. Use an optional interface for version requirements

**Chosen:** Add an optional interface implemented by models that need static pre-read server version checks, for example:

```go
type DataSourceVersionRequirement struct {
	MinVersion   *version.Version
	ErrorMessage string
}

type KibanaDataSourceWithVersionRequirements interface {
	GetVersionRequirements() ([]DataSourceVersionRequirement, diag.Diagnostics)
}
```

The generic Kibana data source detects this with a type assertion on the decoded model.

**Alternative:** Add `GetVersionRequirements` directly to `KibanaDataSourceModel`.

**Rationale:** Most Kibana envelope data sources should not need to implement no-op version methods. Optional capabilities keep the base model contract focused on connection access while allowing model-specific behavior where needed.

---

### 2. Enforce static requirements after scoped client resolution

**Chosen:** In `genericKibanaDataSource.Read`, enforce optional version requirements after `GetKibanaClient` succeeds and before `readFunc` is called.

**Rationale:** Version enforcement needs a `*clients.KibanaScopedClient`, and failing before the entity callback prevents API calls against unsupported servers. This keeps the envelope responsible for the full pre-read gate.

---

### 3. Keep workflow-tool version gating local

**Chosen:** Leave the `minVersionAdvancedAgentConfig` workflow-tool check inside the Agent Builder read callback.

**Rationale:** The workflow API requirement is conditional on the returned agent's tool references and only matters when `include_dependencies` causes tool expansion. The generic envelope should not grow a post-fetch lifecycle hook solely for this data-dependent case.

---

### 4. Schema factory omits connection blocks

**Chosen:** Convert the Agent Builder data source schema method into a factory that returns the entity attributes only, with no `kibana_connection` block. The envelope injects `kibana_connection`.

**Rationale:** This matches the existing envelope contract and avoids duplicate connection block definitions.

---

### 5. Preserve model field tags unless embedding is clearly beneficial

**Chosen:** Prefer adding `GetKibanaConnection() types.List` to `agentDataSourceModel` over embedding `entitycore.KibanaConnectionField`, unless implementation shows embedding produces a cleaner diff.

**Rationale:** The data source model already has a `KibanaConnection types.List` field used by existing schema/tests. A simple accessor satisfies the envelope constraint with minimal state-model churn.

## Risks / Trade-offs

- **Envelope tests may be hard to make fully isolated from real client behavior.** The existing envelope tests already exercise unconfigured-client diagnostics. Version-hook tests should use the available test helpers where possible and focus on no-op behavior, interface detection, diagnostic propagation, and callback short-circuiting.
- **Type assertion on a value model may miss pointer receiver implementations.** Implement `GetVersionRequirements` on the value type unless there is a strong reason not to. This matches the decoded model value passed through the envelope.
- **Schema injection changes where `kibana_connection` is declared.** The final schema should remain equivalent because the envelope injects the same provider schema block. Tests should assert the block is still present.
- **Static and conditional version checks split across layers.** This is intentional: static API requirements belong to the envelope, while data-dependent feature requirements remain in the callback.

## Migration Plan

1. Add the optional version-requirements types and enforcement path to `internal/entitycore/data_source_envelope.go`.
2. Add or update entitycore tests for the optional hook and existing schema/metadata behavior.
3. Convert Agent Builder data source construction to `entitycore.NewKibanaDataSource`.
4. Convert the Agent Builder schema method to a schema factory without `kibana_connection`.
5. Add `GetKibanaConnection` and `GetVersionRequirements` to `agentDataSourceModel`.
6. Refactor `DataSource.Read` into `readAgentDataSource`, removing envelope-owned orchestration.
7. Run focused unit tests and, where the environment supports it, relevant acceptance tests.
