## Context

`elasticstack_elasticsearch_index_template` went through a coherent set of drift defences during its SDK→Plugin-Framework migration. `elasticstack_elasticsearch_component_template` shares almost identical `template.settings` and `template.alias` schema shapes but received none of those defences. The code paths that fix the reported bug already exist in the index_template package; this change lifts them into shared packages and wires them into component_template.

## Goals / Non-Goals

**Goals:**
- Suppress the perpetual plan diff that appears when `template.settings` uses dotted keys in config and nested form arrives from Elasticsearch on read (issue #3897).
- Fix the apply-consistency crash caused by Elasticsearch splitting a `routing`-only alias into `index_routing` + `search_routing`.
- Leave both resources in a structurally symmetric state so future defences only need to be added once.

**Non-Goals:**
- Changing `IndexSettingsValue.StringSemanticEquals` — the existing normalisation is correct.
- Changing how settings are serialised to the Elasticsearch API — ES accepts both dotted and nested; this is purely provider-side state reconciliation.
- Modifying the `elasticstack_elasticsearch_index` resource — it uses a different settings model.
- Adding new schema attributes or user-facing behaviour beyond what is needed to fix the two bugs.

## Decisions

- **Extract into shared packages, not copy.** Copying private helpers into component_template would leave two diverging implementations. Moving + exporting into `aliasutil` and `templateutil` keeps a single source of truth.
- **`AliasObjectType` adoption is wire-safe.** The `AttrTypes` map is identical between the existing plain `types.ObjectType` and `AliasObjectType`, so the tftypes wire format is unchanged. No state migration is required.
- **No separate state migration.** Because the attr types are unchanged the schema version does not need to bump.
- **Single change for both bugs.** Both stem from the same structural asymmetry. Fixing them separately would produce two overlapping refactors with merge-ordering risk.

## Risks / Trade-offs

- [Risk] `AliasObjectType` adoption changes the concrete type of alias set elements stored in state for component_template. If an old provider binary reads state written by the new binary it may not recognise the custom type wrapper. Mitigation: empirical verify-with-old-state test before merge (an upgrade acceptance test or manual binary-swap). If a migration is required, bump `SchemaVersion` and add an upgrader.
- [Risk] `templateilmattachment` resource writes `index.lifecycle.name` into `@custom` component templates. This change is orthogonal to ILM attachment (different code path), but a sanity-check read of `templateilmattachment` code is warranted before closing.

## Open questions

1. **State migration risk** — does switching `component_template`'s alias element from plain `types.ObjectType` to `aliasutil.AliasObjectType` require a state migration? `AttrTypes` are identical so the tftypes wire format should be unchanged, but this needs an empirical binary-swap test before merge.
2. **`templateilmattachment` interaction** — `templateilmattachment` writes programmatically into `@custom` component templates; confirm the new `ModifyPlan` path does not interfere (it only acts when plan and state are both non-null, which should be safe).
3. **Test already written?** `TestAccResourceComponentTemplateDottedSettingsNoDrift` was reportedly added during the research spike. If it already exists in the working tree, adopt it as the regression guard; otherwise write it fresh.
