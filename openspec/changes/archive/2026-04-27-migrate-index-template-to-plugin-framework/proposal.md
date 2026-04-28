## Why

`elasticstack_elasticsearch_index_template` (resource and data source) is one of the last large schemas still implemented on `terraform-plugin-sdk/v2`. It also carries some of the trickiest bespoke SDK behavior in this provider — alias routing diff suppression, `DiffIndexSettingSuppress`, an 8.x `allow_custom_routing` workaround, and version gating — which are difficult to evolve and review while the resource straddles two frameworks. Migrating to the Plugin Framework brings it in line with the modernized Elasticsearch index family (`ilm`, `index`, `datastreamlifecycle`, `index_template_ilm_attachment`), enables the `resourcecore.Core` testing contract, and replaces SDK-specific workarounds with first-class custom types that use semantic equality.

The migration is additive in user-visible terms: the resource type name, attribute paths, block syntax, identity format (`<cluster_uuid>/<template_name>`), and import behavior are all preserved. Existing state authored by the SDK implementation must continue to work after upgrading.

## What Changes

- Add a new Plugin Framework resource implementation under `internal/elasticsearch/index/template/` that mirrors today's `internal/elasticsearch/index/template.go` schema and behavior, embedding `resourcecore.Core`.
- Add a Plugin Framework data source under the same package, sharing the read path with the resource.
- Convert all `MaxItems: 1` blocks in the resource schema (`data_stream`, `template`, `template.lifecycle`, `template.data_stream_options`, `template.data_stream_options.failure_store`, `template.data_stream_options.failure_store.lifecycle`) from the SDK list/set-with-`MaxItems:1` idiom to Plugin Framework `SingleNestedBlock` declarations. Mirror the same shape on the data source.
- Bump the resource state schema version from `0` to `1` and provide an `UpgradeState` implementation that rewrites list-/set-shaped values (`[{…}]` or `[]`) to plain object-shaped values (`{…}` or `null`) for each of the six converted blocks.
- Replace the SDK alias-routing `DiffSuppressFunc` (`suppressAliasRoutingDerivedDiff`) and the post-read "preserve user routing" workaround with an `ObjectValuableWithSemanticEquals` custom type for the alias element.
- Replace the SDK `DiffIndexSettingSuppress` with a new `IndexSettingsValue` custom type in `internal/utils/customtypes/` that implements `StringValuableWithSemanticEquals` and validates JSON-object input.
- Use `jsontypes.Normalized` for `metadata`, `template.mappings`, and `template.alias.filter`.
- Migrate `clients/elasticsearch.PutIndexTemplate` / `GetIndexTemplate` / `DeleteIndexTemplate` to return `fwdiag.Diagnostics`. (No SDK callers remain after migration; component template does not use these helpers.)
- Register the new resource and data source in `provider/plugin_framework.go`; remove the SDK entries from `provider/provider.go`.
- Move and port the existing `template_test.go` and `template_data_source_test.go` into the new package; add `TestAccResourceIndexTemplateFromSDK` covering upgrade from the last SDK-based provider release.
- Delete the SDK files (`template.go`, `template_data_source.go`, and SDK-only helpers no longer used elsewhere).

Component template is intentionally **out of scope**. The shared SDK helpers (`expandTemplate`, `flattenTemplateData`, `extractAliasRoutingFromTemplateState`, `preserveAliasRoutingInFlattenedAliases`, `hashAliasByName`, `stringIsJSONObject`) are duplicated in the new PF package; the SDK copies remain in place for `component_template.go` until it is migrated separately.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-index-template`: clarifies that alias routing diff parity is achieved through semantic equality on the alias element rather than per-field plan-time diff suppression, and that index settings comparison is implemented via a custom type with semantic equality. The user-observable contract is preserved.

## Impact

- Affected code is centered in `internal/elasticsearch/index/template/` (new), `internal/utils/customtypes/index_settings_value.go` (new), `internal/clients/elasticsearch/index.go` (CRUD diags type change), `provider/plugin_framework.go`, and `provider/provider.go`. The shared `internal/elasticsearch/index/commons.go` and `internal/elasticsearch/index/component_template.go` continue to compile unchanged.
- Affected interfaces are internal: the public Terraform schema, attribute paths, identity format, import behavior, and resource type name are unchanged. Plan diffs for routing-only alias configurations and dotted/nested `settings` JSON remain suppressed.
- The change is intended to be backward compatible for existing state. The state schema version bump from `0` to `1` carries an upgrader that reshapes the six single-item collections to plain objects so existing tfstate continues to work without manual intervention or breaking HCL changes (block syntax `data_stream { … }` is preserved).
- A dedicated SDK → Plugin Framework upgrade acceptance test (`TestAccResourceIndexTemplateFromSDK`) is required by REQ-042. It pins the last SDK-based provider release via `ExternalProviders`, applies a configuration that exercises every collapsed block plus the routing-only alias case, then re-applies the equivalent configuration with the in-tree Plugin Framework provider and asserts a no-op upgrade. This is the canonical regression guard for both the schema-shape change and the alias semantic equality contract.
- Downstream resources that reference `elasticstack_elasticsearch_index_template` in acceptance tests (notably `elasticsearch_index_template_ilm_attachment`) are exercised as part of verification.
