## Why

`elasticstack_elasticsearch_component_template` lacks the plan-time and read-time drift-reconciliation infrastructure that `elasticstack_elasticsearch_index_template` received during its SDK-to-Plugin-Framework migration. This structural asymmetry produces two user-facing bugs:

1. **Settings dotted-vs-nested perpetual diff (the reported bug, issue #3897).** A practitioner who writes `template.settings` with dotted Elasticsearch keys (e.g. `{"index.lifecycle.name":"my-policy"}`) sees a permanent plan diff after apply, because Elasticsearch normalises the stored form to nested JSON and the provider has no `ModifyPlan` hook to rewrite the plan to match the canonical state encoding before display.

2. **Alias routing-split apply error (latent).** A component template with a `routing`-only alias field crashes on apply with `Provider produced inconsistent result after apply`, because Elasticsearch splits `routing` into `index_routing` + `search_routing` on the GET response and the resource has no custom `AliasObjectType` or read-time reconciliation to absorb the split.

`elasticstack_elasticsearch_index_template` already has all three defences — `AliasObjectType` with `ObjectSemanticEquals`, read-time alias canonicalisation, and a `ModifyPlan` hook — housed partly in `internal/elasticsearch/index/template/` and partly in `internal/elasticsearch/index/aliasutil/`. Lifting the remaining private helpers into shared packages and wiring `component_template` into them eliminates the asymmetry once, rather than triggering follow-up issues and a second refactor.

## What Changes

- Extract `alias_type.go`, `alias_canonicalize.go`, and `alias_reconcile.go` from `internal/elasticsearch/index/template/` into `internal/elasticsearch/index/aliasutil/`, exporting previously-private helpers. Update `index_template` callers to use the shared package.
- Extract the settings-half of `reconcilePlanWithPriorStateForSemanticDrift` into a new function in `internal/elasticsearch/index/templateutil/` so both resources share it.
- Adopt `aliasutil.AliasObjectType` as the custom element type for the `alias` set nested block in the `component_template` schema.
- Add read-time alias reconciliation to `component_template` (mirroring `index_template`'s `applyTemplateAliasReconciliationFromReference` + `canonicalizeTemplateAliasSetInModel`).
- Add a `ModifyPlan` implementation to `component_template` that calls both the shared settings reconciler and the shared alias plan-time reconcilers.
- Add and verify acceptance tests: settings dotted-vs-nested no-drift (`TestAccResourceComponentTemplateDottedSettingsNoDrift`), alias routing-split no-crash.
- Add two new requirements (REQ-037 and REQ-038) to the component template spec.

## Capabilities

### New Capabilities

None — no new user-facing resource attributes or data sources.

### Modified Capabilities

- `elasticstack_elasticsearch_component_template`: extend to add plan-time settings drift suppression (REQ-037) and alias routing-split consistency (REQ-038).
- `elasticstack_elasticsearch_index_template`: refactor alias helpers to move from package-private into the shared `aliasutil` package (no behaviour change; internal consumers updated).

## Impact

- `internal/elasticsearch/index/aliasutil/` — add `alias_type.go`, `alias_canonicalize.go`, `alias_reconcile.go` (moved + exported from the template package).
- `internal/elasticsearch/index/templateutil/` — add a shared `ReconcileSettings` (or similar) function.
- `internal/elasticsearch/index/template/` — update callers; existing files may be thinned or left as forwarding shims.
- `internal/elasticsearch/index/componenttemplate/schema.go` — adopt `aliasutil.AliasObjectType` for the alias element type.
- `internal/elasticsearch/index/componenttemplate/read.go` — add read-time alias reconciliation.
- `internal/elasticsearch/index/componenttemplate/modify_plan.go` — new file implementing `resource.ResourceWithModifyPlan`.
- `internal/elasticsearch/index/componenttemplate/resource.go` — register the interface.
- `openspec/specs/elasticsearch-index-component-template/spec.md` — REQ-037 and REQ-038.
