# Delta Spec: `elasticstack_elasticsearch_index_template` Plugin Framework migration

Base spec: `openspec/specs/elasticsearch-index-template/spec.md`
Last requirement in base spec: REQ-037
This delta MODIFIES requirements REQ-024 (alias state mapping invariants), REQ-029 (read state mapping for routing), and REQ-031 (alias routing plan diff suppression), and ADDS requirements REQ-038 to REQ-042.

---

This delta defines the target behavior introduced by change `migrate-index-template-to-plugin-framework`. It keeps the user-observable contract for `elasticstack_elasticsearch_index_template` unchanged and clarifies the mechanism by which alias routing parity and index settings comparison are achieved under the Plugin Framework.

## MODIFIED Requirements

### Requirement: JSON and object mapping (REQ-018–REQ-025)

`metadata` SHALL be validated as JSON by schema and parsed as JSON during create/update; if parsing fails, the resource SHALL return an error diagnostic and SHALL not call the Put API. `template.mappings` and `template.settings` SHALL be validated as JSON objects by schema and parsed into objects during create/update. `template.alias.filter` SHALL be validated as JSON by schema and parsed into an object when non-empty during create/update. `template.alias` SHALL be mapped as a set keyed by alias name in API payload/state conversion. Set membership SHALL be determined by the alias element's semantic equality (see REQ-031); two alias values that differ only in API-derived `index_routing` or `search_routing` SHALL be treated as the same set member. Alias routing and flag fields SHALL be copied directly between Terraform values and API model fields. `template.lifecycle` SHALL be mapped as at most one lifecycle object with `data_retention`. `data_stream.hidden` SHALL be sent when present. `data_stream.allow_custom_routing` SHALL be sent only when `true`, except that on updates it SHALL also be sent when prior state had `allow_custom_routing=true` (8.x workaround behavior).

#### Scenario: Invalid metadata JSON

- GIVEN invalid `metadata` JSON
- WHEN create/update runs
- THEN the provider SHALL error before calling Put

#### Scenario: Routing-only alias remains a single set member

- GIVEN an alias configured with only `routing = "x"` (no `index_routing` or `search_routing` in config)
- WHEN refresh populates `index_routing = "x"` and `search_routing = "x"` from the API
- THEN the alias set in state SHALL contain exactly one element for that `name`

### Requirement: Read state mapping (REQ-026–REQ-030)

On read, the resource SHALL set `name`, `composed_of`, `ignore_missing_component_templates`, `index_patterns`, `priority`, and `version` from the API response. On read, when API `metadata` is present, it SHALL be serialized into a JSON string and stored in state. On read, when API `template` is present, it SHALL be flattened into `template` state, including aliases, lifecycle, mappings, and settings. On read, when API `data_stream` is present, it SHALL be flattened into a single `data_stream` block and include only fields present in API response. The provider SHALL NOT post-process the flattened alias state to re-inject the user-configured `routing` value; equivalence between user-configured `routing` and API-derived `index_routing`/`search_routing` SHALL instead be preserved through the alias element's semantic equality (see REQ-031), so that a refreshed alias compares equal to the prior state value when only API-derived routing fields differ.

#### Scenario: User routing preserved without state rewriting

- GIVEN the user set alias `routing = "x"` and the API echoes `index_routing = "x"` / `search_routing = "x"`
- WHEN read refreshes state
- THEN the framework SHALL retain the prior state value for the alias element via semantic equality
- AND no diff SHALL be produced for the routing-only alias

### Requirement: Alias routing plan diff suppression (REQ-031)

The provider SHALL treat a refreshed alias element as semantically equal to its prior state value when both elements have the same `name`, `is_hidden`, `is_write_index`, and `routing`, equivalent JSON `filter`, and:

- For each of `index_routing` and `search_routing`: either the values are equal, or the prior state value is null/empty AND the new value equals the new value's `routing` field AND `routing` is non-empty.

This semantic equality SHALL be implemented as an `ObjectValuableWithSemanticEquals` on the alias element type. As a consequence, plans SHALL NOT show diffs for `index_routing` or `search_routing` when those attributes are unset in configuration and their refreshed values equal the configured `routing`. The framework SHALL apply the same equivalence during refresh, plan, and post-apply state comparison, so the resource SHALL NOT produce "Provider produced inconsistent result after apply" errors for routing-only alias configurations.

#### Scenario: Routing-only alias config

- GIVEN a configuration which configures only the `routing` attribute on an alias
- WHEN apply completes and state is refreshed
- THEN `search_routing` and `index_routing` in state SHALL match the `routing` attribute as documented
- AND no diff SHALL be reported on subsequent plans

#### Scenario: Apply-time consistency

- GIVEN a routing-only alias configuration that previously triggered SDK diff suppression
- WHEN apply runs and Terraform compares the post-apply state to the planned state
- THEN no inconsistency error SHALL be raised

#### Scenario: Explicit override is honored

- GIVEN an alias with `routing = "x"` and an explicit `index_routing = "y"` in configuration
- WHEN refresh returns `index_routing = "y"`
- THEN the prior state value `"y"` SHALL be retained AND no diff SHALL be reported

## ADDED Requirements

### Requirement: Plugin Framework implementation (REQ-038)

The resource and data source SHALL be implemented on `terraform-plugin-framework`. The resource SHALL embed `*resourcecore.Core` constructed with `resourcecore.New(resourcecore.ComponentElasticsearch, "index_template")`. The data source SHALL implement `datasource.DataSourceWithConfigure` directly. Both SHALL be registered through `provider/plugin_framework.go`. The Terraform resource type name (`elasticstack_elasticsearch_index_template`), attribute paths, block HCL syntax (`data_stream { … }`, `template { … }`, etc.), computed-attribute set, identity format (`<cluster_uuid>/<template_name>`), and import behavior SHALL be unchanged from the previous Plugin SDK v2 implementation.

The following blocks, previously declared as `MaxItems: 1` lists or sets in the SDK schema, SHALL be declared as `schema.SingleNestedBlock` in the Plugin Framework schema:

- `data_stream`
- `template`
- `template.lifecycle`
- `template.data_stream_options`
- `template.data_stream_options.failure_store`
- `template.data_stream_options.failure_store.lifecycle`

The `template.alias` block SHALL remain a `schema.SetNestedBlock`.

Plugin Framework `SingleNestedBlock` does not expose a `Required` flag. The previous SDK-side `Required: true` on `template.data_stream_options.failure_store` SHALL be enforced through `ValidateConfig` returning a plan-time error diagnostic when `data_stream_options` is configured and `failure_store` is null.

#### Scenario: Type name and identity unchanged

- WHEN the provider exposes the resource and data source
- THEN both SHALL be registered as `elasticstack_elasticsearch_index_template`
- AND the resource `id` attribute SHALL continue to use the format `<cluster_uuid>/<template_name>`

#### Scenario: HCL block syntax preserved

- GIVEN existing configuration that uses block syntax (e.g. `data_stream { hidden = true }`, `template { settings = "…" }`, `lifecycle { data_retention = "7d" }`)
- WHEN the configuration is parsed against the Plugin Framework schema
- THEN parsing SHALL succeed without modification

#### Scenario: `data_stream_options` without `failure_store` rejected

- GIVEN `template.data_stream_options` is configured and `failure_store` is null
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a plan-time error diagnostic identifying the missing `failure_store` block

### Requirement: Index settings semantic equality (REQ-039)

`template.settings` SHALL be modeled with a custom Plugin Framework string type that implements `basetypes.StringValuableWithSemanticEquals`. The semantic equality comparator SHALL parse both sides as JSON, flatten nested objects to dotted keys, prefix any unprefixed keys with `index.`, stringify all values, and compare the resulting maps. Two `settings` strings SHALL be considered equal whenever they represent the same effective set of index settings, regardless of dotted-vs-nested key form or the presence of an `index.` prefix on individual keys.

#### Scenario: Dotted vs nested keys equivalent

- GIVEN configured settings `{"index": {"number_of_shards": 1}}` and a refreshed value `{"index.number_of_shards": "1"}`
- WHEN plan runs after refresh
- THEN no diff SHALL be reported for `template.settings`

#### Scenario: `index.` prefix normalization

- GIVEN configured settings `{"refresh_interval": "1s"}` and a refreshed value `{"index.refresh_interval": "1s"}`
- WHEN plan runs after refresh
- THEN no diff SHALL be reported for `template.settings`

#### Scenario: Invalid JSON object rejected

- GIVEN `template.settings` configured with a non-object JSON literal (e.g. an array)
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error diagnostic

### Requirement: Client diagnostics use Plugin Framework types (REQ-040)

The internal Elasticsearch client helpers `PutIndexTemplate`, `GetIndexTemplate`, and `DeleteIndexTemplate` SHALL return `terraform-plugin-framework/diag.Diagnostics`. Resource and data source code SHALL append these diagnostics to the response without conversion. The resource SHALL NOT introduce a Plugin SDK ↔ Plugin Framework diagnostics compatibility shim.

#### Scenario: API error surfaced as Plugin Framework diagnostic

- GIVEN an Elasticsearch API call returns a non-success status (other than 404 on read)
- WHEN the resource processes the response
- THEN the response diagnostics SHALL contain a Plugin Framework error diagnostic that surfaces the API error

### Requirement: State schema version 0 → 1 upgrader (REQ-041)

The Plugin Framework resource SHALL declare schema `Version` `1` and SHALL implement `resource.ResourceWithUpgradeState` registering an upgrader for prior schema version `0` (the version under which Plugin SDK v2 wrote state). The upgrader SHALL transform tfstate written by the SDK implementation into the Plugin Framework v1 shape by collapsing each of the following paths from a list-/set-shaped value to a single-object shape:

- `data_stream`
- `template`
- `template.lifecycle`
- `template.data_stream_options`
- `template.data_stream_options.failure_store`
- `template.data_stream_options.failure_store.lifecycle`

For each listed path, after parent paths have already been collapsed, the upgrader SHALL apply the following rule:

- If the value is null or absent, leave it unchanged.
- If the value is an empty array (`[]`), set it to null.
- If the value is a single-element array (`[obj]`), replace it with `obj`.
- If the value is an array with more than one element, return an error diagnostic identifying the path; the upgrader SHALL NOT silently drop elements.

All non-converted attributes (including JSON-string attributes such as `metadata`, `template.mappings`, `template.settings`, and `template.alias.filter`) SHALL be carried through unchanged. After the upgrader runs, Terraform SHALL be able to decode the resulting state against the v1 schema without further migration.

#### Scenario: Upgrade single-element list to object

- GIVEN tfstate written by Plugin SDK v2 with `data_stream = [{"hidden": true, "allow_custom_routing": false}]`
- WHEN the v0 → v1 upgrader runs
- THEN the upgraded state SHALL contain `data_stream = {"hidden": true, "allow_custom_routing": false}`

#### Scenario: Upgrade empty list to null

- GIVEN tfstate written by Plugin SDK v2 with `template.data_stream_options = []`
- WHEN the v0 → v1 upgrader runs
- THEN the upgraded state SHALL contain `template.data_stream_options = null`

#### Scenario: Upgrade nested single-element collections

- GIVEN tfstate written by Plugin SDK v2 with `template = [{"data_stream_options": [{"failure_store": [{"enabled": true, "lifecycle": [{"data_retention": "30d"}]}]}]}]`
- WHEN the v0 → v1 upgrader runs
- THEN the upgraded state SHALL contain `template = {"data_stream_options": {"failure_store": {"enabled": true, "lifecycle": {"data_retention": "30d"}}}}`

#### Scenario: Upgrade preserves non-collapsed attributes

- GIVEN tfstate written by Plugin SDK v2 that includes `metadata`, `composed_of`, `index_patterns`, and `template.alias` populated
- WHEN the v0 → v1 upgrader runs
- THEN those attributes SHALL be carried through byte-equivalent in the upgraded state

#### Scenario: Refuse multi-element arrays at collapsed paths

- GIVEN tfstate at one of the collapsed paths contains an array with two or more elements (a state corruption that should not occur because the SDK enforced `MaxItems: 1`)
- WHEN the v0 → v1 upgrader runs
- THEN it SHALL return an error diagnostic that identifies the offending path
- AND the upgrader SHALL NOT silently discard elements

#### Scenario: End-to-end upgrade from prior SDK release

- GIVEN a resource created by the last Plugin SDK v2 release exercising every collapsed block
- WHEN the same configuration is re-applied with the new Plugin Framework provider
- THEN the v0 → v1 upgrader SHALL run automatically as part of refresh
- AND the subsequent plan SHALL show no diff

### Requirement: SDK → Plugin Framework upgrade acceptance test (REQ-042)

The acceptance test suite for `elasticstack_elasticsearch_index_template` SHALL include a dedicated test (named `TestAccResourceIndexTemplateFromSDK`) that exercises upgrading state created by the last Plugin SDK v2 release of the provider to the new Plugin Framework implementation in a single Terraform run.

The test SHALL be structured as a two-step `resource.TestCase`:

- **Step 1** SHALL pin the prior provider release via `ExternalProviders` with the version constraint set to the last provider release in which `elasticstack_elasticsearch_index_template` was implemented on Plugin SDK v2. It SHALL apply a configuration that exercises every block converted to `SingleNestedBlock`: `data_stream`, `template` (including `template.alias`, `template.lifecycle`, `template.mappings`, `template.settings`), `template.data_stream_options`, `template.data_stream_options.failure_store`, and `template.data_stream_options.failure_store.lifecycle`. Step 1 SHALL also exercise the alias routing scenario (an alias with `routing` set and `index_routing`/`search_routing` unset) so that semantic equality (REQ-031) is validated across the upgrade.
- **Step 2** SHALL switch to `ProtoV6ProviderFactories` pointing at the in-tree Plugin Framework implementation and re-apply the equivalent configuration. Step 2 SHALL assert that the apply is a no-op (no resource replacement, no destructive changes).

The test SHALL be skipped or guarded for stack versions below `9.1.0` only when the test configuration includes `data_stream_options`; equivalent coverage for stack versions below `9.1.0` (without `data_stream_options`) SHALL still run.

#### Scenario: Test exists and runs in CI

- WHEN the acceptance test suite for `elasticstack_elasticsearch_index_template` runs
- THEN `TestAccResourceIndexTemplateFromSDK` SHALL be present and SHALL be discovered by `go test`

#### Scenario: No-op apply after provider switch

- GIVEN Step 1 has applied a configuration that exercises every collapsed block using the last SDK-based provider release
- WHEN Step 2 re-applies the equivalent configuration with the in-tree Plugin Framework provider
- THEN apply SHALL succeed with no resource replacement or destructive change
- AND the post-apply plan SHALL show no diff

#### Scenario: Routing-only alias survives the upgrade

- GIVEN Step 1 configured an alias with `routing` set and `index_routing`/`search_routing` unset
- WHEN Step 2 re-applies the equivalent configuration with the Plugin Framework provider
- THEN no diff SHALL be reported for the alias routing fields after apply

#### Scenario: Collapsed-block state survives the upgrade

- GIVEN Step 1 produced state containing list-shaped values for `data_stream`, `template`, `template.lifecycle`, `template.data_stream_options`, `template.data_stream_options.failure_store`, and `template.data_stream_options.failure_store.lifecycle`
- WHEN Step 2 runs `terraform plan`
- THEN the v0 → v1 upgrader SHALL collapse each of those values to its object/null shape
- AND no diff SHALL be reported for any collapsed path
