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
  actions {                     # optional list
    action_type_id = <required, string>
    id             = <required, string>  # connector ID
    params         = <required, map(string)>
    group          = <optional, string>
    uuid           = <optional, computed, string>
    alerts_filter  = <optional, map(string)>

    frequency {                 # optional+computed
      notify_when = <required, string> # one of: onActionGroupChange | onActiveAlert | onThrottleInterval
      summary     = <required, bool>
      throttle    = <required, string>
    }
  }

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

On create, read, update, and delete, the resource SHALL obtain the Kibana OAPI client from the provider-configured API client. If the provider did not supply a usable client, the resource SHALL return a configuration error diagnostic and not proceed.

#### Scenario: Unconfigured provider

- GIVEN the resource has no provider-supplied API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with a provider configuration error

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

When reading actions from Kibana, each action SHALL be mapped to state with its `action_type_id`, `id`, `params` (as map(string)), `group`, `uuid`, `alerts_filter` (as map(string)), and `frequency` block. If the API returns no actions, the actions list in state SHALL be null. The `params` values SHALL be coerced to strings from the API's `map[string]any` representation.

#### Scenario: Empty actions from API

- GIVEN a rule returned by the API with no actions
- WHEN the provider maps the response to state
- THEN `actions` SHALL be null in state

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
| Composite id parsing | `internal/clients/api_client.go` (`CompositeID`, `CompositeIDFromStrFw`) |
