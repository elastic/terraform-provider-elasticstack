# `elasticstack_elasticsearch_index_alias` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/alias`

## Purpose

Define schema and behavior for the Elasticsearch index alias resource: API usage, identity/import, connection, validation, atomic alias management (create, update, delete), read-time state mapping, and lifecycle.

## Schema

```hcl
resource "elasticstack_elasticsearch_index_alias" "example" {
  id   = <computed, string> # internal identifier: <cluster_uuid>/<alias_name>
  name = <required, string> # force new

  write_index = <optional, single nested object> {
    name          = <required, string>
    filter        = <optional, JSON (normalized) string>
    index_routing = <optional, string>
    is_hidden     = <optional+computed, bool> # default false
    routing       = <optional, string>
    search_routing = <optional, string>
  }

  read_indices = <optional, set of nested objects> {
    name          = <required, string>
    concrete_indices = <computed, set(string)>
    filter        = <optional, JSON (normalized) string>
    index_routing = <optional, string>
    is_hidden     = <optional+computed, bool> # default false
    routing       = <optional, string>
    search_routing = <optional, string>
  }

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
}
```

## Requirements

### Requirement: Alias CRUD APIs (REQ-001–REQ-003)

The resource SHALL use the Elasticsearch Get Alias API to read alias-to-index associations ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-alias.html)). The resource SHALL use the Elasticsearch Update Aliases API to create, update, and delete aliases atomically using `add` and `remove` actions ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-aliases.html)). When Elasticsearch returns a non-success response (other than 404 on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create

- GIVEN the Update Aliases API returns a non-success response
- WHEN create runs
- THEN Terraform diagnostics SHALL include the error

#### Scenario: 404 treated as not found on read

- GIVEN the Get Alias API returns HTTP 404
- WHEN read runs
- THEN the resource SHALL remove itself from state without an error

### Requirement: Identity (REQ-004–REQ-005)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<alias_name>`. During create, the resource SHALL compute `id` from the current cluster UUID and the configured `name` attribute.

#### Scenario: ID format

- GIVEN a successful create
- WHEN the provider sets the ID
- THEN `id` SHALL be `<cluster_uuid>/<alias_name>`

### Requirement: Import (REQ-006)

The resource SHALL support import via `ImportStatePassthroughID`, persisting the supplied `id` directly to state. The resource uses `name` from state (not from the parsed id) for read operations; it does not re-derive the alias name from `id` on read.

#### Scenario: Import passthrough

- GIVEN an import command with a valid composite id
- WHEN import completes
- THEN the id SHALL be stored in state for subsequent operations

### Requirement: Lifecycle (REQ-007)

Changing `name` SHALL require replacement of the resource (`RequiresReplace`). The computed `id` SHALL be preserved across plan/apply cycles using `UseStateForUnknown`.

#### Scenario: Name change triggers replace

- GIVEN an existing alias
- WHEN `name` is changed in configuration
- THEN Terraform SHALL plan a replace (destroy + create)

### Requirement: Connection (REQ-008)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls (create, read, update, delete).

#### Scenario: Resource-level client

- GIVEN `elasticsearch_connection` is set
- WHEN API calls run
- THEN they SHALL use the resource-scoped client

### Requirement: Configuration validation (REQ-009)

The resource SHALL validate that the index named in `write_index.name` does not also appear in any entry of `read_indices`, either by literal name equality or, once expressions are resolved during plan and apply, as a concrete target of any `read_indices` expression. If the same index appears in both roles, the resource SHALL return an "Invalid Configuration" error diagnostic and SHALL NOT proceed to create or update. Literal-name validation is applied both at plan time (`ValidateConfig`) and at apply time before API calls.

The resource SHALL reject multi-target selector syntax in `write_index.name`: a name containing `*`, `?` or `,`, a name beginning with `-` or `<` (date math), or the name `_all`. This validation SHALL be syntactic, SHALL be skipped for unknown values, and SHALL NOT reject single-target names that were previously accepted.

#### Scenario: Write index in read indices

- GIVEN `write_index.name` equals one of the `read_indices[*].name` values
- WHEN config is validated
- THEN the provider SHALL return an "Invalid Configuration" error diagnostic

#### Scenario: Write index matched by a read index expression

- GIVEN `write_index.name` is `logs-1` and a `read_indices` expression `logs-*` resolves to include `logs-1`
- WHEN plan or apply resolves the expressions
- THEN the provider SHALL return an "Invalid Configuration" error diagnostic

#### Scenario: Wildcard write index rejected

- GIVEN `write_index.name` is `logs-*`, `a,b`, `-a`, `<logs-{now/d}>` or `_all`
- WHEN config is validated
- THEN the provider SHALL return an error diagnostic before any API call

#### Scenario: Concrete write index accepted

- GIVEN `write_index.name` is `my.index_1-v2`
- WHEN config is validated
- THEN no selector validation error SHALL be returned

### Requirement: Create (REQ-010–REQ-012)

On create, the resource SHALL convert `write_index` (if set) and all entries in `read_indices` (if set) into `add` alias actions. Each `read_indices[*].name` expression SHALL be resolved at apply time (not only at plan time) to its current concrete targets, and one `add` action SHALL be produced per resolved target. The `write_index` entry SHALL be submitted with `is_write_index: true`; `read_indices` entries SHALL be submitted without `is_write_index`. When one or more actions exist, the resource SHALL submit them in a single atomic Update Aliases API call and perform a read to refresh state before storing the final state. When no actions exist because there is no write index and no resolved read target, the resource SHALL retain virtual state without calling Update Aliases.

#### Scenario: Atomic create

- GIVEN both `write_index` and `read_indices` are configured
- WHEN create runs
- THEN a single Update Aliases call SHALL include one `add` action for the write index and one for each resolved concrete read target

#### Scenario: Wildcard create attaches every match

- GIVEN `read_indices` contains `name = "traces-apm*"` and `traces-apm-default` and `traces-apm.rum-default` exist
- WHEN create runs
- THEN a single Update Aliases call SHALL include an `add` action for each of the two concrete indices and none for the pattern string
- AND the apply SHALL complete without an inconsistent-result error

#### Scenario: Index created between plan and apply

- GIVEN a matching index is created after plan but before apply
- WHEN create or update runs
- THEN the alias SHALL also be attached to that index

### Requirement: Update (REQ-013–REQ-015)

On update, the resource SHALL resolve each `read_indices[*].name` expression to concrete targets and compute a diff between the live alias associations (from the Get Alias API, keyed by concrete name) and the desired set of targets. Current members absent from the desired set SHALL be submitted as `remove` actions. Desired targets that are not current members, or are current members with changed settings, SHALL be submitted as `add` actions. Targets in both with identical settings SHALL be skipped. Selector strings SHALL NOT be used as diff keys or sent in `remove` actions. All resulting actions SHALL be submitted in a single atomic Update Aliases API call. If no actions are required, the Update Aliases API SHALL NOT be called. After a successful update (or when no actions were needed), the resource SHALL perform a read to refresh state.

If two expressions resolve to the same target with identical settings, the resource SHALL produce a single action for that target. If they resolve to the same target with different settings, the resource SHALL return a configuration error diagnostic and SHALL NOT call the Update Aliases API. The same check SHALL run during plan (`ModifyPlan`) so the conflict is reported before apply.

#### Scenario: Index removed from plan

- GIVEN an existing alias with two read indices
- WHEN one read index is removed from configuration
- THEN the update SHALL issue a `remove` action for the removed index and an `add` action (if changed) or no action for the unchanged index

#### Scenario: Unchanged wildcard produces no churn

- GIVEN an alias attached to every index matching `traces-apm*` with the configured settings
- WHEN update runs for an unrelated change
- THEN no `remove` or `add` action SHALL be issued for those indices

#### Scenario: Newly matching target is attached

- GIVEN a new index now matches a configured `read_indices` expression but is not attached to the alias
- WHEN update runs
- THEN an `add` action SHALL be issued for that index only

#### Scenario: Overlapping expressions with identical settings

- GIVEN two `read_indices` expressions resolve to the same index with identical settings
- WHEN update runs
- THEN exactly one `add` action SHALL be issued for that index

#### Scenario: Overlapping expressions with conflicting settings

- GIVEN two `read_indices` expressions resolve to the same index with different settings
- WHEN create or update runs
- THEN the resource SHALL return an error diagnostic and SHALL NOT call the Update Aliases API

### Requirement: Read (REQ-016–REQ-018)

On read, the resource SHALL call the Get Alias API with the alias name from state. If the API returns an empty result (no indices) or the alias name is not present in any returned index, the resource SHALL remove itself from state without an error except when it qualifies for virtual empty-alias state. When alias data is returned, the resource SHALL classify each index as write or read based on the `is_write_index` flag from the API response and populate `write_index` and `read_indices` accordingly.

For each `read_indices` element in prior state or plan, the resource SHALL resolve its `name` expression and set the element's `concrete_indices` to the intersection of the resolved targets and the alias's actual members. Alias members that are not covered by any configured `read_indices` expression and are not the write index SHALL be returned as additional `read_indices` entries whose `name` and sole `concrete_indices` value are the concrete index name, so that the next plan removes them. When no prior configuration exists (for example immediately after import), every read member SHALL be returned in this singleton form.

A member is "covered" when it is in the resolved set of at least one configured `read_indices` element. A member covered by more than one element SHALL appear in the `concrete_indices` of every covering element and SHALL NOT also be emitted as a singleton entry. Each retained element SHALL preserve its own configured settings, so read never has to choose between overlapping elements: overlapping expressions that resolve to the same target with different settings are rejected during plan, create and update, and therefore cannot reach read.

When an alias is absent, the resource SHALL retain virtual state only when prior state has no `write_index` and every prior `read_indices` element has empty `concrete_indices`. Read SHALL NOT consider whether the configured expressions currently resolve to targets: a target that begins matching after the virtual state was recorded is membership drift for `ModifyPlan` to report as an in-place update, not evidence that the resource was deleted. The read that follows a successful create or update SHALL instead retain virtual state, with empty `concrete_indices` values, whenever that apply determined the desired alias has no `write_index` and no resolved read target, regardless of prior membership. In every other absent-alias case, the resource SHALL remove itself from state.

#### Scenario: Alias not found on read

- GIVEN the Get Alias API returns an empty map or the alias name is missing
- AND the resource does not qualify for virtual empty-alias state
- WHEN read runs
- THEN the resource SHALL be removed from state

#### Scenario: State classification

- GIVEN the API returns two indices, one with `is_write_index: true`
- WHEN read runs
- THEN `write_index` SHALL hold the write index and `read_indices` SHALL hold the remaining index

#### Scenario: Wildcard element reconciles after read

- GIVEN `read_indices` contains `name = "traces-apm*"` and the alias is attached to both matching indices
- WHEN read runs
- THEN state SHALL contain a single `read_indices` element with `name = "traces-apm*"` and `concrete_indices` containing both indices

#### Scenario: Missing attachment is not hidden

- GIVEN `traces-apm*` resolves to two indices but the alias is attached to only one
- WHEN read runs
- THEN the element's `concrete_indices` SHALL contain only the attached index
- AND the next plan SHALL show an in-place update to attach the other

#### Scenario: Virtual empty alias state

- GIVEN an alias resource has no `write_index`
- AND every prior `concrete_indices` value is empty
- AND Elasticsearch has no alias with the configured name
- WHEN read runs
- THEN the resource SHALL remain in state with empty `concrete_indices` values

#### Scenario: Virtual state retained when a target now matches

- GIVEN the resource is in virtual state with no `write_index` and empty `concrete_indices` values
- AND an index matching a configured expression has since been created
- AND Elasticsearch has no alias with the configured name
- WHEN read runs
- THEN the resource SHALL remain in state with empty `concrete_indices` values
- AND the next plan SHALL show an in-place update, not a create

#### Scenario: Update removes the last association

- GIVEN prior state has a non-empty `concrete_indices` value and no `write_index`
- AND the updated configuration resolves to no targets
- WHEN update removes the last alias association and the post-update read finds no alias
- THEN the resource SHALL remain in state with empty `concrete_indices` values
- AND the next refresh and plan SHALL be clean

#### Scenario: Missing alias with prior membership is deleted drift

- GIVEN prior state has a `write_index` or a non-empty `concrete_indices` value
- AND Elasticsearch has no alias with the configured name
- WHEN read runs
- THEN the resource SHALL be removed from state

#### Scenario: Member covered by overlapping expressions

- GIVEN two `read_indices` elements with identical settings both resolve to `logs-1`, and the alias is attached to `logs-1`
- WHEN read runs
- THEN `logs-1` SHALL appear in `concrete_indices` of both elements
- AND no singleton entry for `logs-1` SHALL be emitted

#### Scenario: Unconfigured member is drift

- GIVEN the alias is attached to an index not covered by any configured `read_indices` expression
- WHEN read runs
- THEN state SHALL include a singleton `read_indices` entry for that index
- AND the next plan SHALL remove it

### Requirement: Delete (REQ-019)

On delete, the resource SHALL read the alias's live associations with the Get Alias API and submit `remove` alias actions, in a single atomic Update Aliases API call, for every index currently associated with the alias. It SHALL NOT submit selector expressions or rely on state element names. If the alias has no current associations, the Update Aliases API SHALL NOT be called.

#### Scenario: Delete removes all indices

- GIVEN an alias with a write index and one read index
- WHEN delete runs
- THEN a single Update Aliases call SHALL include `remove` actions for both indices

#### Scenario: Delete with wildcard configuration

- GIVEN state holds `read_indices` with `name = "traces-apm*"` and the alias is attached to two concrete indices
- WHEN delete runs
- THEN the Update Aliases call SHALL include `remove` actions for the two concrete indices and none for `traces-apm*`

### Requirement: Mapping — filter field (REQ-020–REQ-021)

`filter` in both `write_index` and `read_indices` entries SHALL be declared as a JSON-normalized string and validated as JSON by the schema type. On create and update, if `filter` is set, the resource SHALL unmarshal it into a map and pass it to the Update Aliases API. On read, if the API response contains a non-nil filter, the resource SHALL marshal it back to a JSON string and store it in state, except that a `read_indices` element retained from prior configuration SHALL preserve its configured filter value.

#### Scenario: Filter round-trip

- GIVEN a `filter` JSON value is configured
- WHEN create runs
- THEN the filter SHALL be sent as a map in the API payload and stored back as JSON on read

#### Scenario: Configured selector filter drift

- GIVEN a configured read-index selector is attached to a concrete target
- AND the target's alias filter is changed outside Terraform
- WHEN refresh and plan run without a membership or configuration change
- THEN the configured filter SHALL remain in state and no change SHALL be planned

### Requirement: Mapping — routing and hidden fields (REQ-022)

On create and update, `index_routing`, `routing`, and `search_routing` SHALL be omitted from the API payload when their values are null or empty. `is_hidden` SHALL be included in the API payload only when its value is `true`. On read, string routing fields that are empty in the API response SHALL be stored as null, except that a `read_indices` element retained from prior configuration SHALL preserve its configured routing values.

#### Scenario: Empty routing omitted

- GIVEN `routing` is null in configuration
- WHEN the alias action is built
- THEN the `routing` key SHALL NOT appear in the API payload

### Requirement: Mapping — is_hidden default (REQ-023)

`is_hidden` in both `write_index` and `read_indices` entries SHALL default to `false` when not explicitly configured. On read, the API-returned boolean value for `is_hidden` SHALL be written to state, except that a `read_indices` element retained from prior configuration SHALL preserve its configured `is_hidden` value.

#### Scenario: Default is_hidden

- GIVEN `is_hidden` is not set in configuration
- WHEN plan is computed
- THEN `is_hidden` SHALL default to `false`

### Requirement: Read indices selector expressions

`read_indices[*].name` SHALL accept any Elasticsearch multi-target expression, including `*` and `?` wildcards, comma-separated lists and `-` exclusions, and SHALL retain the user's original expression in state. The resource SHALL resolve expressions during plan (`ModifyPlan`), create, read and update using an explicit `expand_wildcards=all` policy (open, closed and hidden targets) and `allow_no_indices=true`. An expression with no current matches SHALL be valid and SHALL resolve to an empty set; it SHALL be re-resolved on later plans. If the desired alias has no write index and no resolved read target, the resource SHALL retain virtual state until a later plan resolves a target or the resource is deleted. A resolution result containing the resource's own alias SHALL exclude that alias from the result; any other alias entry SHALL return an error diagnostic naming the alias, and SHALL NOT be treated as a no-match. Remote-cluster targets SHALL also return an error diagnostic. The remaining attachable targets SHALL be either regular indices or data streams, but not both; data-stream backing indices SHALL be excluded only when their data stream is also resolved.

`_all` and bare `*` are not given special treatment. They are passed to Elasticsearch like any other expression and are subject to the same checks as every expression, including the mixed-kind rejection and the REQ-009 check that no resolved target is the `write_index`. Consequently they are only usable when the cluster has a single attachable target kind and the expression does not resolve to the `write_index`; otherwise the resource SHALL return an error diagnostic, and practitioners narrow the expression with a prefix or `-` exclusions (for example `logs-*,-logs-current`).

#### Scenario: No-match expression

- GIVEN a `read_indices` expression matches no indices
- WHEN create runs
- THEN the apply SHALL succeed and the element's `concrete_indices` SHALL be an empty set
- AND the resource SHALL remain in virtual state when no write index or other resolved read target exists

#### Scenario: Later-created match is attached on next apply

- GIVEN an expression had no matches at create time and the resource is in virtual state
- AND a matching index is created later
- WHEN the next plan runs
- THEN the plan SHALL show an in-place update, and applying it SHALL attach the alias to the new index

#### Scenario: Hidden and closed targets included

- GIVEN a closed index and a hidden index match an expression
- WHEN the expression is resolved
- THEN both SHALL be included as targets

#### Scenario: Comma and exclusion expressions

- GIVEN expressions such as `a-*,-a-skip` and `a,b`
- WHEN they are resolved
- THEN the resolved targets SHALL follow Elasticsearch multi-target semantics

#### Scenario: Broad expression covering the write index rejected

- GIVEN `write_index.name` is `logs-current` and a `read_indices` expression is `_all` or `logs-*`, both resolving to include `logs-current`
- WHEN plan or apply resolves the expressions
- THEN the provider SHALL return an "Invalid Configuration" error diagnostic

#### Scenario: Exclusion removes the write index from a broad expression

- GIVEN `write_index.name` is `logs-current` and a `read_indices` expression is `logs-*,-logs-current`
- WHEN plan or apply resolves the expressions
- THEN `logs-current` SHALL NOT be a resolved read target
- AND no write-index collision error SHALL be returned

#### Scenario: Broad expression on a cluster with indices and data streams

- GIVEN a cluster has both regular indices and data streams and a `read_indices` expression is `_all` or `*`
- WHEN plan, create or update runs
- THEN the resource SHALL return the mixed-kind error diagnostic before any alias action

#### Scenario: Alias results rejected

- GIVEN an expression resolves to an alias, with or without regular indices
- WHEN plan, create or update runs
- THEN the resource SHALL return an error diagnostic naming the alias before any alias action
- AND the expression SHALL NOT be treated as an empty match or produce virtual state

#### Scenario: Remote target rejected

- GIVEN an expression resolves to a remote-cluster target
- WHEN plan, create or update runs
- THEN the resource SHALL return an error diagnostic naming that target

#### Scenario: Mixed target kinds rejected

- GIVEN an expression resolves to both regular indices and data streams
- WHEN plan, create or update runs
- THEN the resource SHALL return an error diagnostic before any alias action

### Requirement: Computed concrete_indices

Each `read_indices` element SHALL expose a computed `concrete_indices` attribute of type `set(string)` containing the concrete indices and data streams currently attached to the alias for that element's `name` expression. During plan the resource SHALL resolve the expression and compare the desired membership to state. When create or update is required, the resource SHALL plan the affected `concrete_indices` value as unknown so apply-time resolution can determine the final membership. When no change is required, the resource SHALL preserve the current concrete membership in the plan. During read it SHALL contain only the intersection of the resolved targets and the alias's actual members. When `name` is unknown at plan time, `concrete_indices` SHALL be unknown.

State created before this attribute existed SHALL be handled without error: the initial refresh and plan SHALL succeed and converge to the same result as for newly created state.

#### Scenario: Plan populates concrete indices

- GIVEN `read_indices` contains `name = "traces-apm*"` and two indices match
- WHEN plan runs on create
- THEN the plan SHALL create the alias resource
- AND the planned `concrete_indices` SHALL be unknown

#### Scenario: Unattached resolved target plans an update

- GIVEN state records one attached index for an expression and a second matching index exists
- WHEN plan runs
- THEN the plan SHALL show an in-place update for that `read_indices` element
- AND the planned `concrete_indices` SHALL be unknown

#### Scenario: Target created between plan and apply

- GIVEN a create or update plan has unknown `concrete_indices` for a read expression
- AND another target begins matching that expression after plan completes
- WHEN apply re-resolves the expression
- THEN the alias SHALL be attached to the newly matching target
- AND the apply SHALL complete without an inconsistent-result error

#### Scenario: Existing state without concrete_indices

- GIVEN state saved by a provider version without `concrete_indices`
- WHEN the new provider version refreshes and plans
- THEN no error SHALL occur and the plan SHALL be empty when membership matches the configuration

### Requirement: Out-of-band setting drift scope

Drift detection for `read_indices` SHALL be membership-based. An out-of-band change to `filter`, routing or hidden settings on an already attached concrete target SHALL NOT, by itself, produce a plan when membership is unchanged. Whenever an update is otherwise required, the resource SHALL compare the live alias associations and apply the configured settings to every resolved target whose live settings differ.

#### Scenario: Settings drift without membership change

- GIVEN an attached target's filter is changed out-of-band and membership is unchanged
- WHEN plan runs
- THEN no change SHALL be planned for that reason alone

#### Scenario: Settings repaired during another update

- GIVEN an attached target's filter was changed out-of-band
- WHEN an update is applied for another reason
- THEN an `add` action restoring the configured settings SHALL be issued for that target
