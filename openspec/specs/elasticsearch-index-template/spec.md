# `elasticstack_elasticsearch_index_template` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/template.go`

## Purpose

Define schema and behavior for the Elasticsearch index template resource: API usage, identity/import, connection, compatibility, mapping, and state refresh semantics including alias routing quirks.

## Schema

```hcl
resource "elasticstack_elasticsearch_index_template" "example" {
  id   = <computed, string> # internal identifier: <cluster_uuid>/<template_name>
  name = <required, string> # force new

  composed_of                         = <optional, computed, list(string)>
  ignore_missing_component_templates  = <optional, computed, list(string)> # requires Elasticsearch >= 8.7.0 when non-empty
  index_patterns                      = <required, set(string)>
  metadata                            = <optional, json string>
  priority                            = <optional, int> # must be >= 0
  version                             = <optional, int>

  elasticsearch_connection {
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    headers                  = <optional, map(string)>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    key_file                 = <optional, string>
    cert_data                = <optional, string>
    key_data                 = <optional, string>
  }

  data_stream {
    hidden               = <optional, bool>
    allow_custom_routing = <optional, bool>
  }

  template {
    mappings = <optional, json object string>
    settings = <optional, json object string>
    alias {
      name           = <required, string>
      filter         = <optional, json string>
      index_routing  = <optional, computed, string>
      is_hidden      = <optional, bool>
      is_write_index = <optional, bool>
      routing        = <optional, string>
      search_routing = <optional, computed, string>
    }
    lifecycle {
      data_retention = <required, string>
    }
  }
}
```
## Requirements
### Requirement: Index template CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Put index template API to create and update index templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-put-template.html)). The resource SHALL use the Elasticsearch Get index template API to read index templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-template.html)). The resource SHALL use the Elasticsearch Delete index template API to delete index templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-delete-template.html)). When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than not found on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API errors surfaced

- GIVEN a failing Elasticsearch response (other than 404 on read)
- WHEN the provider processes the response
- THEN diagnostics SHALL include the API error

### Requirement: Identity and import (REQ-005–REQ-008)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<template_name>`. During create and update, the resource SHALL compute `id` from the current cluster UUID and configured `name`. The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` value directly to state. For imported or stored `id` values, read/delete operations SHALL require the format `<cluster_uuid>/<resource identifier>` and SHALL return an error diagnostic when the format is invalid.

#### Scenario: Invalid id on read

- GIVEN a malformed `id` in state
- WHEN read or delete runs
- THEN the provider SHALL return an error diagnostic

### Requirement: Lifecycle and connection (REQ-009–REQ-011)

Changing `name` SHALL require replacement (`ForceNew`). By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls.

#### Scenario: ForceNew on name change

- GIVEN `name` changes in configuration
- WHEN Terraform plans
- THEN replacement SHALL be required

### Requirement: Compatibility (REQ-012)

When `ignore_missing_component_templates` is configured with one or more values, the resource SHALL require Elasticsearch version >= 8.7.0; otherwise it SHALL return an error diagnostic.

#### Scenario: Feature on old cluster

- GIVEN non-empty `ignore_missing_component_templates` and ES < 8.7.0
- WHEN create or update runs
- THEN the provider SHALL error

### Requirement: Create, update, and read (REQ-013–REQ-016)

On create/update, the resource SHALL construct a `models.IndexTemplate` request body from Terraform state and submit it with the Put index template API. After a successful Put request, the resource SHALL set `id` and perform a read to refresh state. On read, the resource SHALL parse `id`, fetch the index template by name, and remove the resource from state when the template is not found. If the Get index template API returns a result count other than exactly one template, the read path SHALL return an error diagnostic.

#### Scenario: Template not found on refresh

- GIVEN the template was deleted in Elasticsearch
- WHEN read runs
- THEN the resource SHALL be removed from state

### Requirement: Delete (REQ-017)

On delete, the resource SHALL parse `id` and delete the template identified by the parsed resource identifier.

#### Scenario: Destroy deletes by parsed name

- GIVEN destroy
- WHEN delete runs
- THEN Delete index template SHALL be called for the parsed identifier

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

### Requirement: Schema — `template.data_stream_options` block (REQ-032)

The `template` block SHALL support an optional `data_stream_options` sub-block. The `data_stream_options` block SHALL contain at most one optional `failure_store` sub-block. If `data_stream_options` is configured without a `failure_store` sub-block, the provider SHALL reject the configuration at plan time with an error diagnostic. The `failure_store` block SHALL contain:

- `enabled` — required boolean; activates or deactivates document redirection to the failure store on newly created matching data streams.
- `lifecycle` — optional sub-block containing `data_retention`, a required string specifying how long failure store documents are retained (e.g. `"30d"`).

When `data_stream_options` is omitted from the configuration, the provider SHALL not include the field in API requests and SHALL leave it unset in Terraform state.

**Example HCL:**

```hcl
resource "elasticstack_elasticsearch_index_template" "example" {
  name           = "my-index-template"
  index_patterns = ["my-datastream-*"]

  template {
    data_stream_options {
      failure_store {
        enabled = true
        lifecycle {
          data_retention = "30d"
        }
      }
    }
  }

  data_stream {}
}
```

#### Scenario: `data_stream_options` omitted

- **WHEN** `data_stream_options` is not configured
- **THEN** the provider SHALL not include `data_stream_options` in the Put index template API request body

#### Scenario: `failure_store.enabled` without lifecycle

- **WHEN** `failure_store.enabled = true` is configured and `lifecycle` is omitted
- **THEN** the provider SHALL send `{"failure_store": {"enabled": true}}` inside `template.data_stream_options` in the API request

#### Scenario: `failure_store` with lifecycle retention

- **WHEN** `failure_store.enabled = true` and `failure_store.lifecycle.data_retention = "10d"` are configured
- **THEN** the provider SHALL send `{"failure_store": {"enabled": true, "lifecycle": {"data_retention": "10d"}}}` inside `template.data_stream_options`

---

### Requirement: Compatibility — version gate for `data_stream_options` (REQ-033)

When `data_stream_options` is configured and the Elasticsearch server version is below `9.1.0`, the provider SHALL return an error diagnostic and SHALL not call the Put index template API.

#### Scenario: Feature on unsupported cluster version

- **GIVEN** `data_stream_options` is configured
- **AND** the connected Elasticsearch server version is below `9.1.0`
- **WHEN** create or update runs
- **THEN** the provider SHALL return an error diagnostic without calling the Put index template API

#### Scenario: Feature on supported cluster version

- **GIVEN** `data_stream_options` is configured
- **AND** the connected Elasticsearch server version is `9.1.0` or above
- **WHEN** create or update runs
- **THEN** the provider SHALL include `data_stream_options` in the API request normally

---

### Requirement: Create/update — expand `data_stream_options` into API request (REQ-034)

On create and update, when `template.data_stream_options` is configured, the provider SHALL construct a `DataStreamOptions` model from the Terraform configuration and include it in the `template` field of the Put index template API request body.

#### Scenario: `failure_store.enabled` round-trip on create

- **GIVEN** `failure_store.enabled = true` configured
- **WHEN** create runs and the template is read back
- **THEN** state SHALL contain `template.data_stream_options.failure_store.enabled = true`

#### Scenario: Update changes `enabled` value

- **GIVEN** an existing template with `failure_store.enabled = true`
- **WHEN** configuration changes `failure_store.enabled` to `false` and apply runs
- **THEN** the provider SHALL send `enabled: false` in the updated API request
- **AND** state SHALL reflect `failure_store.enabled = false` after the read-back

---

### Requirement: Read — flatten `data_stream_options` from API response (REQ-035)

On read, when the API response includes `data_stream_options` inside the `template` object, the provider SHALL populate `template.data_stream_options` in Terraform state, including the `failure_store.enabled` value and `failure_store.lifecycle.data_retention` if present.

When the API response does not include `data_stream_options` (or `data_stream_options` is null), the provider SHALL leave `template.data_stream_options` unset in state.

#### Scenario: Read-back with `data_stream_options` present

- **GIVEN** the API response includes `template.data_stream_options.failure_store.enabled = true`
- **WHEN** read runs
- **THEN** state SHALL contain `template.data_stream_options.failure_store.enabled = true`

#### Scenario: Read-back with `data_stream_options` absent

- **GIVEN** the API response does not include `template.data_stream_options`
- **WHEN** read runs
- **THEN** `template.data_stream_options` SHALL be unset in state

#### Scenario: Read-back with `lifecycle.data_retention` present

- **GIVEN** the API response includes `template.data_stream_options.failure_store.lifecycle.data_retention = "10d"`
- **WHEN** read runs
- **THEN** state SHALL contain `template.data_stream_options.failure_store.lifecycle.data_retention = "10d"`

---

### Requirement: Model — `DataStreamOptions` struct in `models.Template` (REQ-036)

The internal `models.Template` struct SHALL include a `DataStreamOptions` field typed as `*DataStreamOptions` and serialized as `"data_stream_options"` in JSON. The field SHALL be `omitempty` so that templates without `data_stream_options` serialize correctly. New structs `DataStreamOptions`, `FailureStoreOptions`, and `FailureStoreLifecycle` SHALL be added to `internal/models/models.go`.

The `DataStreamOptions` struct SHALL contain:
- `FailureStore *FailureStoreOptions json:"failure_store,omitempty"`

The `FailureStoreOptions` struct SHALL contain:
- `Enabled bool json:"enabled"`
- `Lifecycle *FailureStoreLifecycle json:"lifecycle,omitempty"`

The `FailureStoreLifecycle` struct SHALL contain:
- `DataRetention string json:"data_retention,omitempty"`

Adding these fields to the shared `models.Template` struct SHALL not affect the `elasticstack_elasticsearch_component_template` resource, because the `data_stream_options` field will never be populated by that resource and `omitempty` ensures the key is absent from component template API payloads.

#### Scenario: Component template payloads are not affected

- **GIVEN** a component template create or update operation where `data_stream_options` is not configured
- **WHEN** the `models.Template` struct is serialized to JSON
- **THEN** the JSON payload SHALL NOT include the `data_stream_options` key

#### Scenario: Index template payload includes `data_stream_options`

- **GIVEN** an index template create or update operation where `data_stream_options` is configured
- **WHEN** the `models.Template` struct is serialized to JSON
- **THEN** the JSON payload SHALL include `"data_stream_options": { "failure_store": { "enabled": true } }`

---

### Requirement: Acceptance tests — `data_stream_options` coverage (REQ-037)

Acceptance tests for `elasticstack_elasticsearch_index_template` SHALL include coverage for:

- Creating a template with `data_stream_options.failure_store.enabled = true` and verifying state after create.
- Updating the template to change `failure_store.enabled` and verifying state after update.
- Creating a template with `failure_store.lifecycle.data_retention` set and verifying state after create.
- Verifying that omitting `data_stream_options` produces no drift in plan after apply.

These tests SHALL only run against Elasticsearch >= 9.1.0 and SHALL be skipped or guarded appropriately when a lower version is detected.

A unit test (not an acceptance test) SHALL verify the version-gate logic for the error path: when `data_stream_options` is configured and the detected Elasticsearch version is below `9.1.0`, the provider function under test SHALL return an error diagnostic without invoking the Put index template API.

#### Scenario: Acceptance test create with failure store enabled

- **GIVEN** an acceptance test configuration with `failure_store.enabled = true`
- **WHEN** the test creates the template and refreshes state
- **THEN** the acceptance test SHALL assert `template.0.data_stream_options.0.failure_store.0.enabled` equals `true` in state

#### Scenario: Acceptance test update failure store enabled value

- **GIVEN** an acceptance test that first creates a template with `failure_store.enabled = true`
- **WHEN** the configuration is updated to `failure_store.enabled = false` and applied
- **THEN** the acceptance test SHALL assert that the state reflects `failure_store.enabled = false` after the update

#### Scenario: Acceptance test with data_retention

- **GIVEN** an acceptance test configuration with `failure_store.lifecycle.data_retention = "14d"`
- **WHEN** the test creates the template and refreshes state
- **THEN** the acceptance test SHALL assert `template.0.data_stream_options.0.failure_store.0.lifecycle.0.data_retention` equals `"14d"` in state

#### Scenario: Unit test — version-gate error path

- **GIVEN** a unit test that simulates `data_stream_options` configured with Elasticsearch version `9.0.0`
- **WHEN** the create or update function is invoked
- **THEN** the function SHALL return an error diagnostic containing the minimum version requirement
- **AND** the Put index template API SHALL NOT be called

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

