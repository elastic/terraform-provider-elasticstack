# `elasticstack_kibana_security_detection_rule` — Schema and Functional Requirements

Resource implementation: `internal/kibana/security_detection_rule`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_security_detection_rule` resource: Kibana Security Detection Rules API, composite identity and import, provider-level Kibana connection only, type-dispatched rule creation and update (query, eql, esql, machine_learning, new_terms, saved_query, threat_match, threshold), version-gated response actions, config-time validation of index/data_view_id exclusivity, and stable mapping between Terraform state and API payloads for all supported rule types.

## Schema

```hcl
resource "elasticstack_kibana_security_detection_rule" "example" {
  # Identity
  id       = <computed, string> # composite: "<space_id>/<rule_uuid>"; UseStateForUnknown
  space_id = <optional, computed, string> # default "default"; RequiresReplace
  rule_id  = <optional, computed, string> # Kibana rule_id (stable signature id); RequiresReplaceIfConfigured

  # Rule definition — required
  name        = <required, string> # 1–255 characters
  type        = <required, string> # one of: query | eql | esql | machine_learning | new_terms | saved_query | threat_match | threshold; RequiresReplace
  description = <required, string>

  # Scheduling
  enabled  = <optional, computed, bool>   # default true
  from     = <optional, computed, string> # default "now-6m"; regex now-\d+[smhd]
  to       = <optional, computed, string> # default "now"
  interval = <optional, computed, string> # default "5m"; regex \d+[smhd]

  # Severity / risk
  risk_score = <optional, computed, int64> # 0–100; default 50
  severity   = <optional, computed, string> # one of: low | medium | high | critical; default "medium"

  risk_score_mapping {            # optional list
    field      = <required, string>
    operator   = <required, string> # "equals"
    value      = <required, string>
    risk_score = <optional, int64>  # 0–100
  }

  severity_mapping {              # optional list
    field    = <required, string>
    operator = <required, string> # "equals"
    value    = <required, string>
    severity = <required, string> # one of: low | medium | high | critical
  }

  # Common optional metadata
  author               = <optional, computed, list(string)> # default []
  tags                 = <optional, computed, list(string)> # default []
  license              = <optional, string>
  false_positives      = <optional, computed, list(string)> # default []
  references           = <optional, computed, list(string)> # default []
  note                 = <optional, string>
  setup                = <optional, string>
  max_signals          = <optional, computed, int64> # ≥1; default 100
  version              = <optional, computed, int64> # ≥1; default 1
  namespace            = <optional, string>
  rule_name_override   = <optional, string>
  timestamp_override   = <optional, string>
  timestamp_override_fallback_disabled = <optional, bool>
  investigation_fields = <optional, list(string)>
  building_block_type  = <optional, string> # "default"

  # Query / filter — applicable to most rule types (forbidden for machine_learning and esql)
  query    = <optional, computed, string>
  language = <optional, computed, string> # one of: kuery | lucene | eql | esql
  index    = <optional, computed, list(string)> # forbidden for machine_learning and esql
  filters  = <optional, JSON-normalized string> # forbidden for machine_learning and esql
  data_view_id = <optional, string> # forbidden for machine_learning and esql

  # EQL-specific
  tiebreaker_field = <optional, string>

  # Machine Learning-specific
  anomaly_threshold     = <optional, int64> # 0–100; required when type == "machine_learning"
  machine_learning_job_id = <optional, list(string)>

  # New Terms-specific
  new_terms_fields     = <optional, list(string)>
  history_window_start = <optional, string>

  # Saved Query-specific
  saved_id = <optional, string>

  # Threat Match-specific
  threat_index         = <optional, list(string)>
  threat_query         = <optional, computed, string>
  threat_filters       = <optional, list(string)>
  threat_indicator_path = <optional, computed, string>
  concurrent_searches  = <optional, int64> # ≥1
  items_per_search     = <optional, int64> # ≥1

  threat_mapping {              # optional list; required for threat_match
    entries {                   # required list
      field = <required, string>
      type  = <required, string> # "mapping"
      value = <required, string>
    }
  }

  # Threshold-specific
  threshold {                   # optional; required for threshold rules
    field = <optional, list(string)>
    value = <required, int64>   # ≥1

    cardinality {               # optional list
      field = <required, string>
      value = <required, int64> # ≥1
    }
  }

  # MITRE ATT&CK
  threat {                      # optional list
    framework = <required, string>
    tactic {
      id        = <required, string>
      name      = <required, string>
      reference = <required, string>
    }
    technique {                 # optional list
      id        = <required, string>
      name      = <required, string>
      reference = <required, string>
      subtechnique {            # optional list
        id        = <required, string>
        name      = <required, string>
        reference = <required, string>
      }
    }
  }

  # Timeline
  timeline_id    = <optional, string>
  timeline_title = <optional, string>

  # Alert suppression
  alert_suppression {           # optional
    group_by               = <optional, list(string)>
    duration               = <optional, string> # custom duration type
    missing_fields_strategy = <optional, string> # one of: suppress | doNotSuppress
  }

  # Exception containers
  exceptions_list {             # optional list
    id             = <required, string>
    list_id        = <required, string>
    namespace_type = <required, string> # one of: single | agnostic
    type           = <required, string> # one of: detection | endpoint | endpoint_events | endpoint_host_isolation_exceptions | endpoint_blocklists | endpoint_trusted_apps
  }

  # Related integrations / required fields
  related_integrations {        # optional list
    package     = <required, string>
    version     = <required, string>
    integration = <optional, string>
  }

  required_fields {             # optional list
    name = <required, string>
    type = <required, string>
    ecs  = <computed, bool>     # populated by backend
  }

  # Actions
  actions = [{                  # optional list of objects
    action_type_id = <required, string>
    id             = <required, string>  # connector ID
    params         = <required, JSON-normalized string> # jsonencode() of the action params object
    group          = <optional, string>
    uuid           = <optional, computed, string>

    alerts_filter = {             # optional nested attribute
      query = {                   # optional nested attribute
        kql          = <optional, string>
        filters_json = <optional, computed, JSON-normalized string>  # jsonencode([]) for empty
      }
      timeframe = {               # optional nested attribute
        days        = <optional, list(int64)>  # 1–7
        timezone    = <optional, string>
        hours_start = <optional, string>       # HH:MM
        hours_end   = <optional, string>       # HH:MM
      }
    }

    frequency = {               # optional+computed nested attribute
      notify_when = <required, string> # one of: onActionGroupChange | onActiveAlert | onThrottleInterval
      summary     = <required, bool>
      throttle    = <required, string>
    }
  }]

  # Response actions (requires server ≥ 8.16.0)
  response_actions {            # optional list
    action_type_id = <required, string> # one of: .osquery | .endpoint

    params {
      # Osquery params
      query          = <optional, string>
      pack_id        = <optional, string>
      saved_query_id = <optional, string>
      timeout        = <optional, int64>  # 60–900 seconds
      ecs_mapping    = <optional, map(string)>
      queries {                 # optional list
        id          = <required, string>
        query       = <required, string>
        platform    = <optional, string>
        version     = <optional, string>
        removed     = <optional, bool>
        snapshot    = <optional, bool>
        ecs_mapping = <optional, map(string)>
      }

      # Endpoint params
      command = <optional, string> # one of: isolate | kill-process | suspend-process
      comment = <optional, string>
      config {                  # optional; for kill-process and suspend-process
        field     = <required, string>
        overwrite = <optional, computed, bool> # default true
      }
    }
  }

  # Read-only
  created_at = <computed, string>
  created_by = <computed, string>
  updated_at = <computed, string>
  updated_by = <computed, string>
  revision   = <computed, int64>
}
```

Notes:

- `required_fields[*].ecs` is computed by the Kibana backend based on field name and type.
- `data_view_id`, `index`, and `filters` are forbidden when `type` is `machine_learning` or `esql` (enforced by schema-level validators and `ValidateConfig`).
- For all rule types other than `machine_learning` and `esql`, exactly one of `index` or `data_view_id` must be set (enforced by `ValidateConfig`).
- `anomaly_threshold` is required when `type == "machine_learning"` (schema-level validator).
- `response_actions` require server version ≥ **8.16.0** (version-gated at create/update time).
- `rule_id` triggers replacement only when the practitioner configures a value (`RequiresReplaceIfConfigured`); if omitted, Kibana assigns a UUID and the resource tracks it as computed.

## Requirements

### Requirement: Kibana security detection rule APIs (REQ-001–REQ-004)

The resource SHALL manage detection rules through Kibana's Security Detections API: create rule, read rule, update rule, and delete rule. The generated Kibana OAPI client (`kbapi`) SHALL be used for all API interactions. Reference: [Kibana Security Detection Rules API](https://www.elastic.co/guide/en/kibana/current/rule-api-overview.html).

#### Scenario: Create then authoritative read

- GIVEN a successful create API response (HTTP 200)
- WHEN create completes
- THEN the provider SHALL re-fetch the rule with the read API and SHALL fail with an error diagnostic if the rule cannot be read back

#### Scenario: Update then authoritative read

- GIVEN a successful update API response (HTTP 200)
- WHEN update completes
- THEN the provider SHALL re-fetch the rule with the read API and SHALL fail with an error diagnostic if the rule cannot be read back

#### Scenario: Delete idempotency

- GIVEN delete is called
- WHEN the API returns 404 for the rule
- THEN the provider SHALL treat delete as successful (no error diagnostic)

#### Scenario: Read removes missing rules

- GIVEN a read/refresh
- WHEN the read API returns 404
- THEN the provider SHALL remove the resource from Terraform state

### Requirement: API error surfacing (REQ-005)

For create and update, when the API returns a non-200 status code or a transport error occurs, the resource SHALL surface an error diagnostic containing the HTTP status code and response body. For read, non-200 and non-404 responses SHALL produce an error diagnostic. For delete, non-200 and non-404 responses SHALL produce an error diagnostic; a 404 SHALL be treated as success.

#### Scenario: Non-success create

- GIVEN a create API call that returns HTTP 4xx or 5xx
- WHEN create runs
- THEN the provider SHALL return an error diagnostic with the status code and body

#### Scenario: Non-success update

- GIVEN an update API call that returns HTTP 4xx or 5xx
- WHEN update runs
- THEN the provider SHALL return an error diagnostic with the status code and body

### Requirement: Provider configuration and Kibana client (REQ-006)

On create, read, update, and delete, the resource SHALL obtain the Kibana OAPI client from the provider-configured API client by default. If the provider did not supply a usable client, the resource SHALL return a configuration error diagnostic and not proceed. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OAPI client for all operations.

#### Scenario: Unconfigured provider

- GIVEN the resource has no provider-supplied API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with a provider configuration error

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the provider-configured Kibana OAPI client

#### Scenario: Scoped Kibana connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the scoped Kibana OAPI client derived from that block

### Requirement: Identity and composite `id` (REQ-007–REQ-009)

After a successful create, the resource SHALL set `id` to the composite string `<space_id>/<rule_uuid>`, where `space_id` is the Kibana space and `rule_uuid` is the UUID of the created rule as returned by the API. The `id` SHALL be computed (UseStateForUnknown) so it is stable across plans. The `rule_id` (Kibana rule signature id) SHALL be set from the API response and preserved across reads.

#### Scenario: Composite id after create

- GIVEN a successful create that returns a rule with UUID `abc-123` and space_id `"default"`
- WHEN create writes state
- THEN `id` SHALL equal `"default/abc-123"`

#### Scenario: ID preserved across read

- GIVEN prior state with a known `id` `"my-space/abc-123"`
- WHEN read runs
- THEN the provider SHALL set the same `id` value in updated state and SHALL not alter `space_id`

### Requirement: Import (REQ-010–REQ-011)

The resource SHALL support Terraform import by passing the composite `id` (in the format `<space_id>/<rule_uuid>`) directly to state via `ImportStatePassthroughID`. The full import string becomes `id` in state; on the subsequent read, `space_id` and the rule UUID are derived from parsing the composite id.

#### Scenario: Valid import id

- GIVEN an import id `"my-space/abc-123-def"`
- WHEN import runs
- THEN state SHALL hold `id = "my-space/abc-123-def"` and the following read SHALL fetch the rule from space `my-space` using UUID `abc-123-def`

### Requirement: Lifecycle — force replacement (REQ-012)

Changing `space_id` SHALL require destroying and recreating the resource. Changing `type` SHALL require destroying and recreating the resource. If `rule_id` is configured by the practitioner and the plan changes its value, that SHALL also require replacement (`RequiresReplaceIfConfigured`).

#### Scenario: Replace on space_id change

- GIVEN an in-place plan that changes only `space_id`
- WHEN Terraform evaluates the plan
- THEN the plan SHALL indicate replace (destroy/create) for the resource

#### Scenario: Replace on type change

- GIVEN an in-place plan that changes only `type`
- WHEN Terraform evaluates the plan
- THEN the plan SHALL indicate replace (destroy/create) for the resource

### Requirement: Rule type dispatch (REQ-013)

The resource SHALL dispatch create and update API requests to the appropriate type-specific builder based on the `type` attribute. Each supported rule type SHALL use its corresponding `ruleProcessor` (QueryRuleProcessor, EqlRuleProcessor, EsqlRuleProcessor, MachineLearningRuleProcessor, NewTermsRuleProcessor, SavedQueryRuleProcessor, ThreatMatchRuleProcessor, ThresholdRuleProcessor). An unsupported `type` value SHALL cause the provider to return an error diagnostic at create/update time.

#### Scenario: Query rule dispatch

- GIVEN `type = "query"` in configuration
- WHEN create runs
- THEN the provider SHALL build and send a query-type rule create request

#### Scenario: Unsupported rule type

- GIVEN `type = "unsupported_type"` in configuration
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic indicating the rule type is unsupported

### Requirement: Config validation — index and data_view_id exclusivity (REQ-014)

For rule types other than `machine_learning` and `esql`, exactly one of `index` or `data_view_id` SHALL be set. If both are set, the provider SHALL return an error diagnostic `Invalid Configuration: Both 'index' and 'data_view_id' cannot be set at the same time.` If neither is set, the provider SHALL return an error diagnostic `Invalid Configuration: One of 'index' or 'data_view_id' must be set.`

#### Scenario: Both index and data_view_id set

- GIVEN `type = "query"` and both `index` and `data_view_id` are configured
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error

#### Scenario: Neither index nor data_view_id set for non-ml/esql rule

- GIVEN `type = "query"` with neither `index` nor `data_view_id` configured
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error

#### Scenario: esql rule skips index/data_view_id validation

- GIVEN `type = "esql"`
- WHEN Terraform validates the configuration
- THEN the provider SHALL NOT enforce the index/data_view_id exclusivity check

#### Scenario: machine_learning rule skips index/data_view_id validation

- GIVEN `type = "machine_learning"`
- WHEN Terraform validates the configuration
- THEN the provider SHALL NOT enforce the index/data_view_id exclusivity check

### Requirement: Config validation — index and data_view_id forbidden for machine_learning and esql (REQ-015)

The `index`, `data_view_id`, and `filters` attributes SHALL be forbidden when `type` is `machine_learning` or `esql`. The provider SHALL enforce this via schema-level `ForbiddenIfDependentPathOneOf` validators on those attributes.

#### Scenario: index forbidden for machine_learning

- GIVEN `type = "machine_learning"` and `index` is set
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error on `index`

### Requirement: Config validation — anomaly_threshold required for machine_learning (REQ-016)

The `anomaly_threshold` attribute SHALL be required when `type == "machine_learning"`. The provider SHALL enforce this via a `RequiredIfDependentPathEquals` schema-level validator on `anomaly_threshold`.

#### Scenario: Missing anomaly_threshold for ML rule

- GIVEN `type = "machine_learning"` and `anomaly_threshold` is not set
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error on `anomaly_threshold`

### Requirement: Compatibility — response_actions (REQ-017)

When `response_actions` is configured with at least one entry, the resource SHALL verify the server version is at least **8.16.0** before including response actions in the create or update request. If the server version is below 8.16.0, the provider SHALL fail with an error diagnostic stating `Response actions are unsupported` and the minimum required version.

#### Scenario: Response actions on old stack

- GIVEN server version &lt; 8.16.0 and `response_actions` configured with at least one entry
- WHEN create or update runs
- THEN the provider SHALL return a response actions unsupported error

### Requirement: Read — rule UUID validation (REQ-018)

When reading a rule, the resource SHALL parse the `ResourceID` portion of the composite `id` as a UUID. If parsing fails, the provider SHALL return an error diagnostic `ID was not a valid UUID`.

#### Scenario: Non-UUID resource ID

- GIVEN state with a composite `id` whose resource-id segment is not a valid UUID
- WHEN read runs
- THEN the provider SHALL return an error diagnostic

### Requirement: Read — API discriminator dispatch (REQ-019)

After fetching a rule from the API, the provider SHALL use the API response discriminator to determine the rule type and dispatch to the corresponding `ruleProcessor.UpdateFromResponse`. If no processor matches the API response type, the provider SHALL return an error diagnostic.

#### Scenario: Unknown API rule type in response

- GIVEN the API returns a rule type not recognized by any registered processor
- WHEN read or post-create/update read runs
- THEN the provider SHALL return an error diagnostic indicating no processor was found

### Requirement: Delete — UUID validation (REQ-020)

When deleting a rule, the resource SHALL parse the `ResourceID` portion of the composite `id` as a UUID. If parsing fails, the provider SHALL return an error diagnostic `ID was not a valid UUID`.

#### Scenario: Non-UUID resource ID on delete

- GIVEN state with a non-UUID resource-id segment
- WHEN delete runs
- THEN the provider SHALL return an error diagnostic

### Requirement: Mapping — read-only fields from API (REQ-021)

After a successful read, the provider SHALL store the following computed fields from the API response in state: `created_at`, `created_by`, `updated_at`, `updated_by`, `revision`. These fields SHALL be stored as returned by the API.

#### Scenario: Audit fields populated after read

- GIVEN a rule that was created via the API
- WHEN read maps the API response to state
- THEN `created_at`, `created_by`, `updated_at`, `updated_by`, and `revision` SHALL reflect the values returned by Kibana

### Requirement: Mapping — required_fields.ecs computed by backend (REQ-022)

The `required_fields[*].ecs` attribute is computed by the Kibana backend. The provider SHALL store the value returned by the API in state and SHALL NOT send `ecs` in create or update requests for `required_fields` entries.

#### Scenario: ecs populated from API response

- GIVEN a rule with `required_fields` entries configured
- WHEN read maps the API response
- THEN each `required_fields[*].ecs` SHALL hold the value returned by the Kibana API

### Requirement: Mapping — actions from API (REQ-023)

When reading actions from Kibana, each action SHALL be mapped to state with its `action_type_id`, `id`, `params` (as JSON-normalized string), `group`, `uuid`, structured `alerts_filter` attribute (when present), and `frequency` attribute. If the API returns no actions, the actions list in state SHALL be null. The `params` object returned by the API (`map[string]any`) SHALL be marshaled to a JSON-normalized string using `jsontypes.NewNormalizedValue()`. This preserves nested object structures that cannot be represented as `map(string)`. The `alertsFilter` API object SHALL be mapped to `alerts_filter.query` and `alerts_filter.timeframe` without producing Go map-literal strings in state.

#### Scenario: Empty actions from API

- GIVEN a rule returned by the API with no actions
- WHEN the provider maps the response to state
- THEN `actions` SHALL be null in state

#### Scenario: Nested params preserved as JSON

- GIVEN an action whose `params` object contains nested keys (e.g. a Slack `message` block with `attachments`)
- WHEN the provider maps the API response to state
- THEN `params` SHALL be a JSON-normalized string that round-trips the full nested structure without data loss

### Requirement: Mapping — response_actions to API (REQ-024)

When `response_actions` is configured and the server version check passes (REQ-017), the provider SHALL build the appropriate `SecurityDetectionsAPIResponseAction` union for each entry based on `action_type_id`. For `.osquery`, the provider SHALL send the configured osquery params. For `.endpoint`, the provider SHALL send the configured endpoint params. An unsupported `action_type_id` within `response_actions` SHALL produce an error diagnostic.

#### Scenario: Osquery response action sent correctly

- GIVEN a `response_actions` entry with `action_type_id = ".osquery"` and osquery params
- WHEN create or update runs
- THEN the API request SHALL include an osquery response action with the configured params

### Requirement: Mapping — interval format validation (REQ-025)

The `interval` attribute SHALL accept only strings matching the regex `^\d+[smhd]$`. The `from` attribute SHALL accept only strings matching `^now-\d+[smhd]$`. These are enforced by schema-level regex validators and the provider SHALL return a validation diagnostic if either is invalid.

#### Scenario: Invalid interval rejected

- GIVEN `interval = "5 minutes"` (not matching the regex)
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error on `interval`

### Requirement: Mapping — risk_score range (REQ-026)

The `risk_score` attribute SHALL accept only integers from **0** to **100** inclusive. The schema-level `int64validator.Between(0, 100)` SHALL enforce this and return a validation error if violated.

#### Scenario: Risk score out of range

- GIVEN `risk_score = 150`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error on `risk_score`

### Requirement: Mapping — severity enum (REQ-027)

The `severity` attribute SHALL accept only one of `low`, `medium`, `high`, `critical`. The schema-level `stringvalidator.OneOf(...)` SHALL enforce this.

#### Scenario: Invalid severity rejected

- GIVEN `severity = "extreme"` (not in the allowed set)
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error on `severity`

### Requirement: Mapping — common optional fields sent when known (REQ-028)

When building create or update requests, the resource SHALL send optional fields (e.g., `license`, `note`, `setup`, `tags`, `author`, `false_positives`, `references`, `version`, `max_signals`, `interval`, `from`, `to`, `enabled`, `timeline_id`, `timeline_title`, `rule_name_override`, `timestamp_override`, `namespace`, `building_block_type`) only when they are known and non-null. Unknown or null values SHALL be omitted from the request body.

#### Scenario: Optional fields omitted when null

- GIVEN a configuration where `license` and `note` are not set
- WHEN the provider builds the create or update request
- THEN `license` and `note` SHALL be absent from the request body

### Requirement: Mapping — space_id defaults to "default" (REQ-029)

When `space_id` is not explicitly configured, the resource SHALL use `"default"` as the space identifier for all API calls. This default SHALL be applied by the schema's `stringdefault.StaticString("default")` on the `space_id` attribute.

#### Scenario: Default space used when space_id omitted

- GIVEN `space_id` is not set in configuration
- WHEN create runs
- THEN the API call SHALL target the `"default"` space

### Requirement: Mapping — KQL language mapping (REQ-030)

When mapping the `language` attribute to the API, if `language` is `"kuery"` or `"lucene"`, the provider SHALL send it as the corresponding `SecurityDetectionsAPIKqlQueryLanguage` value. Any other value SHALL default to `"kuery"` in the API request for the KQL language field.

#### Scenario: kuery language maps correctly

- GIVEN `language = "kuery"`
- WHEN the provider builds the create or update request
- THEN the API request SHALL include `language: "kuery"`

### Requirement: Mapping — filters JSON normalization (REQ-031)

The `filters` attribute uses a normalized JSON type (`jsontypes.NormalizedType`). The provider SHALL accept valid JSON for `filters` and send it to the API. The `filters` attribute SHALL be forbidden for `machine_learning` and `esql` rule types (REQ-015).

#### Scenario: Valid JSON filters accepted

- GIVEN `filters = "[{\"match_all\": {}}]"` as a valid JSON string on a query rule
- WHEN Terraform validates the configuration
- THEN the provider SHALL NOT return a validation error on `filters`

#### Scenario: filters forbidden for esql rule

- GIVEN `type = "esql"` and `filters` is configured
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error on `filters`

### Requirement: State upgrade — actions.params map(string) → JSON string (REQ-032)

When Terraform reads prior state written by schema version **0** (where `actions[*].params` was stored as `map(string)`), the resource SHALL perform an in-place state upgrade to schema version **1** by JSON-encoding each entry's `params` map into a JSON-normalized string. The upgrade SHALL NOT require destroying and recreating the resource. The state upgrade logic SHALL be registered via `ResourceWithUpgradeState`. (Schema version **2** and the `alerts_filter` state upgrade are defined in REQ-084.)

#### Scenario: Upgrade from v0 state with map params

- GIVEN persisted state at schema version 0 where `actions[0].params` is stored as a `map(string)` (e.g. `{"body": "hello"}`)
- WHEN Terraform refreshes or plans against the resource
- THEN the provider SHALL upgrade the state to schema version 1, converting `params` to the JSON string `"{\"body\":\"hello\"}"` without triggering a destroy/create

#### Scenario: v1 state requires no upgrade

- GIVEN persisted state already at schema version 1
- WHEN Terraform refreshes or plans against the resource
- THEN no state upgrade SHALL be applied

### Requirement: Empty-list consistency for optional nested list attributes (REQ-033)

When a practitioner explicitly configures any of the following `elasticstack_kibana_security_detection_rule` attributes as an empty list (`[]`), the provider SHALL return an empty list — not `null` — for that attribute in state after `Create`, `Read`, and `Update`. This preserves the Terraform Plugin Framework invariant for `Optional`-only list attributes: the provider MUST return the planned value unchanged when the planned value is a known, non-null empty list.

Affected attributes:

| Attribute | Schema type |
|---|---|
| `actions` | `ListNestedAttribute` |
| `exceptions_list` | `ListNestedAttribute` |
| `severity_mapping` | `ListNestedAttribute` |
| `risk_score_mapping` | `ListNestedAttribute` |
| `related_integrations` | `ListNestedAttribute` |
| `threat` | `ListNestedAttribute` |
| `threat_mapping` | `ListNestedAttribute` |

#### Scenario: Apply with all affected attributes set to empty list

- GIVEN a resource configuration with `actions = []`, `exceptions_list = []`, `severity_mapping = []`, `risk_score_mapping = []`, `related_integrations = []`, `threat = []`, and `threat_mapping = []`
- WHEN `terraform apply` runs
- THEN the provider SHALL succeed without a "Provider produced inconsistent result after apply" diagnostic
- AND each of the seven attributes SHALL be stored as an empty list (`[]`) in Terraform state

#### Scenario: Subsequent plan shows no diff for empty-list attributes

- GIVEN a successfully applied resource with any of the seven attributes stored as `[]` in state
- WHEN `terraform plan` runs without any configuration change
- THEN the plan SHALL be empty (no changes) for those attributes

#### Scenario: Null configuration is preserved

- GIVEN a resource configuration where one or more of the seven attributes is absent or explicitly `null`
- WHEN `terraform apply` runs
- THEN the provider SHALL store `null` (not `[]`) for those attributes in state

### Requirement: Empty-list consistency for nested `threat` sub-lists (REQ-034)

When a `threat` block is configured with one or more entries, and a practitioner explicitly configures `technique = []` for a threat entry, the provider SHALL return an empty list (`[]`) — not `null` — for `technique` in state. The same rule applies when a practitioner explicitly configures `subtechnique = []` within a technique entry.

If `technique` or `subtechnique` is absent from configuration or explicitly `null`, the provider SHALL preserve `null` for that attribute in state and SHALL NOT normalize it to `[]`.

#### Scenario: Threat entry with explicitly empty techniques preserves empty list

- GIVEN a resource configuration with one `threat` entry and `technique = []`
- WHEN `terraform apply` runs
- THEN the provider SHALL store `[]` for `technique` in state for that threat entry
- AND the provider SHALL NOT produce a "Provider produced inconsistent result after apply" diagnostic

#### Scenario: Threat entry with omitted or null techniques preserves null

- GIVEN a resource configuration with one `threat` entry and `technique` absent or explicitly `null`
- WHEN `terraform apply` runs
- THEN the provider SHALL store `null` for `technique` in state for that threat entry

#### Scenario: Technique entry with explicitly empty subtechniques preserves empty list

- GIVEN a resource configuration with a threat entry containing a technique entry and `subtechnique = []`
- WHEN `terraform apply` runs
- THEN the provider SHALL store `[]` for `subtechnique` in state for that technique entry
- AND the provider SHALL NOT produce a "Provider produced inconsistent result after apply" diagnostic

#### Scenario: Technique entry with omitted or null subtechniques preserves null

- GIVEN a resource configuration with a threat entry containing a technique entry and `subtechnique` absent or explicitly `null`
- WHEN `terraform apply` runs
- THEN the provider SHALL store `null` for `subtechnique` in state for that technique entry

### Requirement: Reconciliation helper for plan/state alignment (REQ-035)

The provider SHALL implement a `reconcileEmptyListsFromPlan` function (or equivalent logic) in the `securitydetectionrule` package. For each of the seven affected attributes, this function SHALL: if the reference value (plan for Create/Update, prior state for Read) is a known, non-null empty list AND the post-read value is null, replace the post-read null with the reference empty list.

This function SHALL be called after each `r.read()` invocation in `Create`, `Read`, and `Update`.

#### Scenario: Null in post-read is overwritten when reference has empty list

- GIVEN a reference `Data` where `Actions` is a known empty list and `target.Actions` is null
- WHEN `reconcileEmptyListsFromPlan` is called
- THEN `target.Actions` SHALL be set to an empty list identical to `reference.Actions`

#### Scenario: Non-null target is not overwritten

- GIVEN a reference `Data` where `Actions` is a known empty list and `target.Actions` is a non-empty list with items
- WHEN `reconcileEmptyListsFromPlan` is called
- THEN `target.Actions` SHALL remain unchanged

#### Scenario: Null reference does not overwrite null target

- GIVEN a reference `Data` where `Actions` is null and `target.Actions` is null
- WHEN `reconcileEmptyListsFromPlan` is called
- THEN `target.Actions` SHALL remain null

### Requirement: Acceptance test — empty-list round-trip (REQ-036)

The acceptance test suite SHALL include a test that exercises the empty-list scenario for the seven list attributes in REQ-033 in a single resource configuration. The test SHALL apply a configuration with those attributes set to `[]`, assert that `terraform apply` succeeds without "inconsistent result" diagnostics, assert that those attributes are stored as empty lists in state, and assert that a subsequent `terraform plan` produces an empty plan.

#### Scenario: Acceptance test apply with empty lists succeeds

- GIVEN a resource configuration with `actions = []`, `exceptions_list = []`, `severity_mapping = []`, `risk_score_mapping = []`, `related_integrations = []`, `threat = []`, and `threat_mapping = []`
- WHEN the acceptance test runs `terraform apply`
- THEN `terraform apply` SHALL succeed without any "Provider produced inconsistent result after apply" diagnostics
- AND the test SHALL verify that each of the seven attributes is stored as an empty list in state

#### Scenario: No-op plan after empty-list apply

- GIVEN a successfully applied rule with the seven attributes stored as empty lists
- WHEN the acceptance test runs a second `terraform plan`
- THEN the plan SHALL be empty (no proposed changes)

### Requirement: Schema — `actions.alerts_filter` structured attribute (REQ-080)

The `alerts_filter` attribute on `actions` SHALL be a `SingleNestedAttribute` (not a `MapAttribute`). It SHALL contain:

- A `query` nested attribute (optional) with:
  - `kql` — optional string. Defines a KQL query filter that determines whether the action runs.
  - `filters_json` — optional + computed `jsontypes.Normalized` JSON string. Encodes the Kibana filter DSL array (same type as `params`). Use `jsonencode([])` for an empty filter list. Marked computed because the provider populates `[]` from the API when the user omits the attribute, keeping round-trip plans clean. Named `filters_json` to allow a future typed `filters` list attribute to be added without conflict.
- A `timeframe` nested attribute (optional) with:
  - `days` — optional list of int64 (values 1–7, where 1=Monday, 7=Sunday).
  - `timezone` — optional string. ISO time zone name (e.g., `"UTC"`, `"Europe/London"`).
  - `hours_start` — optional string. Start of active hours in 24-hour `HH:MM` notation.
  - `hours_end` — optional string. End of active hours in 24-hour `HH:MM` notation.

The old `MapAttribute(string)` form of `alerts_filter` is removed.

#### Scenario: alerts_filter with query and no timeframe

- GIVEN a detection rule action with `alerts_filter.query.kql = "event.action : \"test\""` and `alerts_filter.query.filters_json = jsonencode([])`
- WHEN the resource is created or updated
- THEN the Kibana API request SHALL include `alertsFilter.query.kql` and `alertsFilter.query.filters` in the expected nested JSON shape

#### Scenario: alerts_filter with timeframe

- GIVEN a detection rule action with `alerts_filter.query.kql` set and `alerts_filter.timeframe` with all four attributes
- WHEN the resource is created or updated
- THEN the Kibana API request SHALL include `alertsFilter.timeframe.days`, `alertsFilter.timeframe.timezone`, and `alertsFilter.timeframe.hours.start` / `alertsFilter.timeframe.hours.end`

#### Scenario: no alerts_filter

- GIVEN a detection rule action without an `alerts_filter` attribute set
- WHEN the resource is created or updated
- THEN the Kibana API request SHALL omit `alertsFilter` for that action

### Requirement: Validation — `timeframe` attributes required together (REQ-081)

When the `alerts_filter.timeframe` attribute is present, all four attributes (`days`, `timezone`, `hours_start`, `hours_end`) SHALL be required. The provider SHALL enforce this via `objectvalidator.AlsoRequires` (or equivalent). Omitting any one of the four attributes while `timeframe` is present SHALL be a validation error.

#### Scenario: timeframe with missing timezone

- GIVEN an `alerts_filter.timeframe` value with `days` and `hours_start` and `hours_end` set but `timezone` absent
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic for the missing attribute

### Requirement: Write path — `alerts_filter` serialization (REQ-082)

When `alerts_filter` is configured, the provider SHALL serialize it to the Kibana API payload as:

```json
{
  "query": {
    "kql": "<kql value>",
    "filters": <parsed filters_json array>
  },
  "timeframe": {
    "days": [...],
    "timezone": "<tz>",
    "hours": {
      "start": "<hours_start>",
      "end": "<hours_end>"
    }
  }
}
```

- `query` SHALL be included when the `query` attribute is present with at least one of `kql` or `filters_json` set.
- `timeframe` SHALL be included when the `timeframe` attribute is present.
- `filters_json` SHALL be unmarshaled from a JSON string into a native array before serialization (not sent as a raw string).

#### Scenario: filters_json marshaled correctly

- GIVEN `alerts_filter.query.filters_json = jsonencode([{"meta": {"alias": null}}])`
- WHEN the provider serializes the action
- THEN `alertsFilter.query.filters` in the API request SHALL be the parsed array, not a JSON string

### Requirement: Read path — `alerts_filter` deserialization (REQ-083)

When the Kibana API returns an `alertsFilter` object on a detection rule action, the provider SHALL:

- Map `alertsFilter.query.kql` to `alerts_filter.query.kql` (string).
- Marshal `alertsFilter.query.filters` (array) to a normalized JSON string for `alerts_filter.query.filters_json`.
- Map `alertsFilter.timeframe.days`, `.timezone`, `.hours.start`, `.hours.end` to the corresponding `alerts_filter.timeframe` attributes.
- Set `alerts_filter` to null when the API response omits `alertsFilter` for that action.

The provider SHALL NOT produce a Go map-literal string (e.g., `"map[filters:[] kql:...]"`) in any state attribute.

#### Scenario: round-trip correctness

- GIVEN a detection rule with `alerts_filter.query.kql = "event.action : \"test\""` and `alerts_filter.query.filters_json = jsonencode([])`
- WHEN the resource is applied and then refreshed
- THEN `terraform plan` SHALL show no diff for `alerts_filter`

### Requirement: Schema version bump and state migration (REQ-084)

The `elasticstack_kibana_security_detection_rule` resource schema version SHALL be bumped from 1 to 2.

A `StateUpgraders` entry for version 1 → 2 SHALL be registered. Because the previous `alerts_filter` implementation was functionally broken (invalid read and write paths), no valid `alerts_filter` state can exist in practice. The upgrade function SHALL discard any stored `alerts_filter` map data without error.

#### Scenario: state upgrade from v1

- GIVEN existing Terraform state at schema version 1 (no valid `alerts_filter` value)
- WHEN the provider is upgraded to the version that introduces schema version 2
- THEN Terraform state is upgraded to version 2 without error

### Requirement: Acceptance tests — `alerts_filter` (REQ-085)

The acceptance test suite for `elasticstack_kibana_security_detection_rule` SHALL include:

1. A test case that creates a detection rule with an action using `alerts_filter.query.kql` and `alerts_filter.query.filters_json = jsonencode([])`, applies, then asserts no plan diff (round-trip correctness).
2. A test step that updates `alerts_filter.query.kql` to a different value and asserts apply succeeds and state reflects the change.
3. A test step that includes `alerts_filter.timeframe` with all four attributes and asserts round-trip correctness.
4. A test case (or step) that uses an action without `alerts_filter` and asserts no regression in create/update/read behavior.

#### Scenario: create with alerts_filter — no diff on refresh

- GIVEN a detection rule configured with `alerts_filter.query.kql` and `alerts_filter.query.filters_json`
- WHEN the resource is applied and then `terraform plan` is re-run
- THEN no diff is shown for `alerts_filter`

#### Scenario: update kql

- GIVEN an applied detection rule with `alerts_filter.query.kql = "old_value"`
- WHEN the configuration is updated to `alerts_filter.query.kql = "new_value"` and applied
- THEN the apply succeeds and state shows the new kql value

#### Scenario: timeframe round-trip

- GIVEN a detection rule with `alerts_filter.timeframe = { days = [1,2,3,4,5], timezone = "UTC", hours_start = "08:00", hours_end = "17:00" }`
- WHEN the resource is applied and refreshed
- THEN all four timeframe attributes in state match configuration

### Requirement: No regression — actions without alerts_filter (REQ-086)

Existing detection rule actions that do not include `alerts_filter` SHALL continue to work without changes. Create, update, and read operations for such actions SHALL produce no errors and no unexpected plan diffs.

#### Scenario: action without alerts_filter creates successfully

- GIVEN a detection rule action with `id`, `action_type_id`, and `params` set but no `alerts_filter` attribute
- WHEN the resource is created and then refreshed
- THEN apply succeeds and `terraform plan` shows no diff for the action

### Requirement: Documentation — `alerts_filter` (REQ-087)

The provider documentation for `elasticstack_kibana_security_detection_rule` SHALL document:

- The `alerts_filter` nested attribute and its sub-attributes (`query.kql`, `query.filters_json`, `timeframe.*`).
- A usage example showing `jsonencode([])` for `filters_json`.
- The `filters_json` attribute name rationale is not required in docs; clear examples suffice.

#### Scenario: docs include alerts_filter example

- GIVEN the generated resource documentation
- WHEN a practitioner reads the `actions.alerts_filter` section
- THEN the docs show a `query.filters_json = jsonencode([])` example and describe each sub-attribute

## Traceability (implementation index)

| Area | Primary files |
|------|----------------|
| Metadata / Configure / Import | `resource.go` |
| Schema | `schema.go` |
| Config validation (index/data_view_id) | `schema.go` (`ValidateConfig`) |
| Create | `create.go` |
| Read | `read.go` |
| Update | `update.go` |
| Delete | `delete.go` |
| Rule type dispatch | `rule_processor.go` |
| Data model | `models.go` |
| Query rule | `models_query.go` |
| EQL rule | `models_eql.go` |
| ESQL rule | `models_esql.go` |
| Machine Learning rule | `models_machine_learning.go` |
| New Terms rule | `models_new_terms.go` |
| Saved Query rule | `models_saved_query.go` |
| Threat Match rule | `models_threat_match.go` |
| Threshold rule | `models_threshold.go` |
| API → model utilities | `models_from_api_type_utils.go` |
| Model → API utilities (response actions, actions, common props) | `models_to_api_type_utils.go` |
| State upgrade (v0 → v1 params map→JSON; v1 → v2 alerts_filter) | `state_upgrade.go` |
| `alerts_filter` expand/flatten | `alerts_filter_utils.go` |
| Composite id parsing | `internal/clients/api_client.go` (`CompositeID`, `CompositeIDFromStrFw`) |

