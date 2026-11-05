# `elasticstack_kibana_alerting_v2_rule` — Schema and Functional Requirements

Resource implementation: `internal/kibana/alertingv2rule`

## Purpose

Define the Terraform schema and runtime behaviour of the `elasticstack_kibana_alerting_v2_rule` resource, which manages a single Kibana Alerting v2 rule through `/api/alerting/v2/rules`: typed rule configuration, create-or-replace lifecycle, space-aware identity and import, server-owned fields, plan-time validation mirroring the API, and the Elastic Stack 9.6.0 version gate.

This capability is distinct from `kibana-alerting-rule` (classic alerting). The two share no API, no saved object type, and no identifier space.

## Schema

```hcl
resource "elasticstack_kibana_alerting_v2_rule" "example" {
  # Identity
  id       = <computed, string>            # composite "<space_id>/<rule_id>"; UseStateForUnknown
  rule_id  = <optional, computed, string>  # 1-150 chars, letters/digits/underscore/hyphen; RequiresReplace; UseStateForUnknown; generated when omitted
  space_id = <optional, computed, string>  # default "default"; RequiresReplace

  # Rule definition
  kind       = <required, string>            # "alert" | "signal"; RequiresReplace
  enabled    = <optional, computed, bool>    # default true
  time_field = <optional, computed, string>  # default "@timestamp"; 1-256 chars

  metadata = {                                # required
    name         = <required, string>         # 1-256 chars; unique within the space
    description  = <optional, string>         # 1-1024 chars
    tags         = <optional, list(string)>   # 1-20 items, each 1-128 chars
    routing_tags = <optional, list(string)>   # 1-20 items, each 1-128 chars; kind = "alert" only
    builder = {                               # optional
      type = <required, string>               # 1-64 chars
    }
  }

  schedule = {                                # required
    every    = <required, duration string>    # >= 5s, <= 365d
    lookback = <optional, duration string>    # <= 365d
  }

  query = {                                   # required
    base = <required, string>                 # ES|QL with FROM clause, 1-10000 chars
    breach = {                                # optional
      segment = <required, string>            # ES|QL clause appended to base, 1-10000 chars
    }
  }

  recovery = {                                # required when kind = "alert"; not allowed when kind = "signal"
    strategy = <required, string>             # "no_breach" | "condition" | "query" | "manual"
    segment  = <optional, string>             # required when strategy = "condition", not allowed otherwise
    query    = <optional, string>             # required when strategy = "query", not allowed otherwise
  }

  no_data = {                                 # required when kind = "alert"; not allowed when kind = "signal"
    strategy = <required, string>             # "ignore" | "keep_last" | "resolve"
    query    = <optional, string>             # not allowed when strategy = "ignore"
  }

  state_transition = {                        # optional; kind = "alert" only
    pending = {                               # optional
      count     = <optional, int64>           # 0-1000
      timeframe = <optional, duration string> # <= 365d
      operator  = <optional, string>          # "and" | "or"; only when both count and timeframe are set
    }
    recovering = {                            # optional; not allowed when recovery.strategy = "manual"
      count     = <optional, int64>
      timeframe = <optional, duration string>
      operator  = <optional, string>
    }
  }

  grouping = {                                # optional
    fields = <required, list(string)>         # 1-16 items, each 1-256 chars
  }

  artifacts = [                               # optional; 1-100 items; unique id per rule
    {
      id   = <required, string>               # 1-150 chars
      type = <required, string>               # 1-128 chars
      data = <required, normalized JSON object string>  # at most 32 top-level keys, each 1-256 chars
    },
  ]

  # Server-owned
  version    = <computed, int64>   # configuration revision, incremented on every change
  created_at = <computed, string>  # ISO 8601; UseStateForUnknown
  updated_at = <computed, string>  # ISO 8601
}
```

Notes:

- `kibana_connection` and `timeouts` are injected by the Kibana resource envelope and are not part of the resource's own schema.
- A duration string is an unsigned integer followed by one of `ms`, `s`, `m`, `h`, `d`, or `w`, for example `5m` or `250ms`.
- Omitting an optional collection (`tags`, `routing_tags`, `grouping`, `artifacts`) or object (`breach`, `builder`, `state_transition`) omits it from the request. An empty collection is a plan-time error because the API rejects empty arrays.

## ADDED Requirements

### Requirement: Resource type and registration

The provider SHALL expose a managed resource with type name `elasticstack_kibana_alerting_v2_rule`, registered in the provider's standard resource set, with generated registry documentation under `docs/resources/` and a usage example under `examples/resources/`. The resource SHALL NOT reuse or alter the classic `elasticstack_kibana_alerting_rule` type.

#### Scenario: Resource is available without opt-in

- **GIVEN** a provider configured with Kibana credentials and no experimental environment variable set
- **WHEN** a configuration declares `elasticstack_kibana_alerting_v2_rule`
- **THEN** Terraform SHALL recognise the resource type
- **AND** `make docs-generate` SHALL produce `docs/resources/kibana_alerting_v2_rule.md`

### Requirement: Minimum Elastic Stack version

The resource SHALL require Elastic Stack 9.6.0 or later and SHALL fail every lifecycle operation with a diagnostic naming the required version when the target stack is older. Serverless targets SHALL bypass the version check.

#### Scenario: Older stack is rejected before any API call

- **GIVEN** a target Elastic Stack at version 9.5.x
- **WHEN** create, read, update, or delete runs
- **THEN** the provider SHALL return an error diagnostic stating that Alerting v2 rules require Elastic Stack 9.6.0 or later
- **AND** SHALL NOT call any `/api/alerting/v2/` endpoint

#### Scenario: Supported stack proceeds

- **GIVEN** a target Elastic Stack at version 9.6.0 or later, or a serverless project
- **WHEN** any lifecycle operation runs
- **THEN** the provider SHALL proceed with the operation

### Requirement: CRUD endpoint mapping

The resource SHALL manage a rule with `PUT /api/alerting/v2/rules/{rule_id}` for both create and update, `GET /api/alerting/v2/rules/{rule_id}` for read, and `DELETE /api/alerting/v2/rules/{rule_id}` for delete, each addressed to the space in `space_id`. It SHALL NOT call `POST /api/alerting/v2/rules`, `PATCH`, bulk, by-query, `_run`, `_enable`, `_disable`, or history endpoints. Final state after create and update SHALL come from a subsequent read.

#### Scenario: Create issues a PUT with the full body and reads back

- **GIVEN** a new resource configuration
- **WHEN** create runs
- **THEN** the provider SHALL issue one `PUT /api/alerting/v2/rules/{rule_id}` whose body is derived entirely from the plan
- **AND** SHALL then issue `GET /api/alerting/v2/rules/{rule_id}` and derive state from that response

#### Scenario: Update replaces rather than patches

- **GIVEN** an existing managed rule with a changed attribute
- **WHEN** update runs
- **THEN** the provider SHALL issue one `PUT /api/alerting/v2/rules/{rule_id}` with the full replacement body
- **AND** SHALL NOT issue a `PATCH` request

#### Scenario: Requests target the configured space

- **GIVEN** `space_id = "team-a"`
- **WHEN** any lifecycle operation runs
- **THEN** the request path SHALL be prefixed with `/s/team-a`

#### Scenario: Default space uses the unprefixed path

- **GIVEN** `space_id` omitted or `"default"`
- **WHEN** any lifecycle operation runs
- **THEN** the request path SHALL NOT carry an `/s/` prefix

### Requirement: Create refuses to adopt an existing rule

Before the first write, create SHALL check whether a rule with the target id already exists in the space (`GET` answering 200) and, if so, SHALL return an error diagnostic naming the id and pointing at `terraform import` without writing. If the `PUT` nevertheless answers HTTP 200 (replaced) instead of 201 (created), the provider SHALL return the same diagnostic and state that the replacement has already happened server-side.

#### Scenario: Create against a pre-existing id

- **GIVEN** a rule `my-rule` already exists in the target space outside Terraform
- **AND** a new resource with `rule_id = "my-rule"`
- **WHEN** create runs
- **THEN** the provider SHALL return an error diagnostic naming the id and recommending import
- **AND** SHALL NOT issue a `PUT`
- **AND** the resource SHALL NOT be recorded in state

#### Scenario: Race between existence check and write

- **GIVEN** a new resource whose id was free at the existence check
- **WHEN** the `PUT` answers HTTP 200 because the rule appeared in between
- **THEN** the provider SHALL return an error diagnostic naming the id, recommending import, and stating the rule was replaced

#### Scenario: Normal create

- **GIVEN** a new resource whose id is not in use
- **WHEN** create runs and the API answers HTTP 201
- **THEN** the provider SHALL continue to the read-after-write and record the resource in state

### Requirement: Identity and `rule_id` generation

`id` SHALL be `<space_id>/<rule_id>`. `rule_id` SHALL be optional and computed: when supplied it SHALL be used as the rule's identifier. When omitted the provider SHALL generate a UUID before the first write and persist it. A configured `rule_id` SHALL be 1-150 characters of letters, digits, underscores, or hyphens. Changing `rule_id` or `space_id` SHALL require replacement.

#### Scenario: Generated rule_id is stable

- **GIVEN** a configuration omitting `rule_id`
- **WHEN** the resource is applied and then re-planned without configuration changes
- **THEN** `rule_id` in state SHALL be unchanged
- **AND** the plan SHALL be empty

#### Scenario: Supplied rule_id is used verbatim

- **GIVEN** `rule_id = "cpu-high_prod"`
- **WHEN** create runs
- **THEN** the `PUT` path SHALL end with `/rules/cpu-high_prod`
- **AND** state SHALL hold `id = "default/cpu-high_prod"`

#### Scenario: Invalid rule_id rejected at plan time

- **GIVEN** `rule_id = "my rule!"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic describing the permitted characters and length

#### Scenario: Changing space or id replaces the resource

- **GIVEN** a managed rule
- **WHEN** `space_id` or `rule_id` changes in configuration
- **THEN** the plan SHALL show destroy-and-create

### Requirement: Import

The resource SHALL support `terraform import` with an id of the form `<space_id>/<rule_id>`. On success it SHALL populate `space_id`, `rule_id`, and `id`, and the following read SHALL populate every other attribute. A malformed import id SHALL produce an error diagnostic naming the required format.

#### Scenario: Import round-trip

- **GIVEN** a rule `my-rule` in space `team-a`
- **WHEN** `terraform import` runs with id `team-a/my-rule`
- **THEN** state SHALL hold `space_id = "team-a"`, `rule_id = "my-rule"`, `id = "team-a/my-rule"`
- **AND** a subsequent plan against a matching configuration SHALL be empty

#### Scenario: Malformed import id

- **GIVEN** an import id without a `/` separator or with an empty segment
- **WHEN** import runs
- **THEN** the provider SHALL return an error diagnostic describing the `<space_id>/<rule_id>` format

### Requirement: Immutable `kind`

`kind` SHALL be required, accept only `alert` or `signal`, and require replacement when changed, because the API treats it as immutable.

#### Scenario: Changing kind plans a replacement

- **GIVEN** a managed rule with `kind = "alert"`
- **WHEN** configuration changes it to `"signal"`
- **THEN** the plan SHALL show destroy-and-create

#### Scenario: Unknown kind rejected

- **GIVEN** `kind = "metric"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic listing `alert` and `signal`

### Requirement: Typed rule configuration

Every field of the rule write body SHALL be a typed attribute or nested attribute as laid out in the Schema section, with the sole exception of `artifacts[].data`, which SHALL be a normalized JSON object string. Enum membership, string lengths, integer ranges, collection sizes, and duration format SHALL be validated at plan time.

#### Scenario: Bound violations are caught before apply

- **GIVEN** `grouping.fields` with 17 entries, or `metadata.name` of 300 characters, or `state_transition.pending.count = 1001`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic
- **AND** no API request SHALL be sent

#### Scenario: Empty collections are rejected

- **GIVEN** `metadata.tags = []` or `artifacts = []` or `grouping.fields = []`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating at least one item is required, or the attribute must be omitted

### Requirement: `enabled` is written in the create-or-replace body

`enabled` SHALL be optional, computed, and default to `true`. The provider SHALL always send `enabled` in the `PUT` body so that a rule is created in, or replaced into, the configured state in a single request, with no follow-up enable or disable call.

#### Scenario: Create a disabled rule

- **GIVEN** a new resource with `enabled = false`
- **WHEN** create runs
- **THEN** the `PUT` body SHALL contain `"enabled": false`
- **AND** the read-after-write SHALL show `enabled = false`
- **AND** no `_disable` request SHALL be sent

#### Scenario: Toggle enabled on update

- **GIVEN** a managed rule with `enabled = false`
- **WHEN** configuration sets `enabled = true` and update runs
- **THEN** the `PUT` body SHALL contain `"enabled": true`
- **AND** state SHALL hold `enabled = true`

#### Scenario: Omitted enabled defaults to true

- **GIVEN** a configuration omitting `enabled`
- **WHEN** create runs
- **THEN** the plan SHALL show `enabled = true` and the `PUT` body SHALL contain `"enabled": true`

### Requirement: Duration attributes

`schedule.every`, `schedule.lookback`, `state_transition.pending.timeframe`, and `state_transition.recovering.timeframe` SHALL accept only an unsigned integer followed by `ms`, `s`, `m`, `h`, `d`, or `w`, at most 365 days in total. `schedule.every` SHALL additionally be at least 5 seconds. Values SHALL be stored and sent exactly as written, without normalization.

#### Scenario: Malformed duration rejected at plan time

- **GIVEN** `schedule.every = "5 minutes"` or `"1y"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic for that attribute

#### Scenario: Below the minimum interval

- **GIVEN** `schedule.every = "1s"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic naming the `5s` minimum

#### Scenario: Deployment minimum surfaced from the server

- **GIVEN** `schedule.every = "10s"` against a deployment whose configured minimum interval is `1m`
- **WHEN** apply runs
- **THEN** the provider SHALL surface the server's rejection as an error diagnostic

### Requirement: Query mapping

`query.base` SHALL be required and `query.breach.segment` SHALL be optional, each 1-10000 characters. The provider SHALL NOT parse or validate ES|QL. Server-side ES|QL errors SHALL be surfaced verbatim.

#### Scenario: Base-only query

- **GIVEN** `query = { base = "FROM metrics-* | STATS avg_cpu = AVG(cpu) BY host.name | WHERE avg_cpu > 0.9" }`
- **WHEN** create runs
- **THEN** the `PUT` body SHALL contain `query.base` and SHALL NOT contain `query.breach`

#### Scenario: Server-side ES|QL error is surfaced

- **GIVEN** a syntactically invalid `query.base` within the length limit
- **WHEN** apply runs and Kibana answers HTTP 400
- **THEN** the provider SHALL return an error diagnostic carrying the server's message

### Requirement: Recovery mapping

`recovery` SHALL be a single object with a required `strategy` of `no_breach`, `condition`, `query`, or `manual`. `segment` SHALL be required when `strategy = "condition"` and not allowed otherwise. `query` SHALL be required when `strategy = "query"` and not allowed otherwise. The object SHALL map to the API's strategy-discriminated `recovery` body and back.

#### Scenario: Condition strategy round-trip

- **GIVEN** `query.breach.segment = "WHERE avg_cpu > 0.9"`
- **AND** `recovery = { strategy = "condition", segment = "WHERE avg_cpu < 0.6" }`
- **WHEN** the rule is applied and read back
- **THEN** state SHALL hold the same `strategy` and `segment`, with `query` null

#### Scenario: Strategy-field mismatch rejected at plan time

- **GIVEN** `recovery = { strategy = "no_breach", segment = "WHERE x < 1" }`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating `segment` is only allowed with `strategy = "condition"`

#### Scenario: Missing variant field rejected at plan time

- **GIVEN** `recovery = { strategy = "query" }`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating `query` is required for `strategy = "query"`

### Requirement: No-data mapping

`no_data` SHALL be a single object with a required `strategy` of `ignore`, `keep_last`, or `resolve`, and an optional `query` that SHALL NOT be set when `strategy = "ignore"`. The API's fourth stored value `alert` SHALL be rejected in configuration but tolerated on read.

#### Scenario: Keep-last with presence query round-trip

- **GIVEN** `no_data = { strategy = "keep_last", query = "FROM metrics-* | STATS c = COUNT() BY host.name" }`
- **WHEN** the rule is applied and read back
- **THEN** state SHALL hold the same `strategy` and `query`

#### Scenario: Query with ignore rejected at plan time

- **GIVEN** `no_data = { strategy = "ignore", query = "FROM x" }`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

#### Scenario: alert strategy rejected in configuration

- **GIVEN** `no_data.strategy = "alert"` in configuration
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic listing the three writable values

#### Scenario: Stored alert strategy tolerated on read

- **GIVEN** an imported rule whose stored `no_data.strategy` is `alert`
- **WHEN** refresh runs
- **THEN** the provider SHALL store `"alert"` without error
- **AND** the rule SHALL show a planned change until the configuration selects a writable strategy, which is an accepted limitation

### Requirement: Cross-field validation mirroring the API

The provider SHALL reject at plan time every configuration the API rejects through a cross-field rule, so these failures never first appear as an HTTP 400 during apply. Each diagnostic SHALL name the attributes involved. A check SHALL be skipped when any participating value is unknown at plan time.

The rules are:

1. `state_transition` only when `kind = "alert"`.
2. `metadata.routing_tags` only when `kind = "alert"`.
3. `kind = "alert"` requires both `recovery` and `no_data`. `kind = "signal"` forbids both.
4. `recovery.strategy = "condition"` requires `query.breach`.
5. `no_data.strategy` other than `ignore` requires `query.breach` or `no_data.query`.
6. `recovery.strategy = "manual"` forbids `state_transition.recovering`.
7. Each `state_transition` phase requires at least one of `count` and `timeframe`, and allows `operator` only when both are set.

#### Scenario: Signal rule with lifecycle configuration

- **GIVEN** `kind = "signal"` and a `recovery` object
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating signal rules cannot set `recovery` or `no_data`

#### Scenario: Alert rule missing lifecycle configuration

- **GIVEN** `kind = "alert"` with `recovery` set and `no_data` omitted
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic naming `no_data` as required for alert rules

#### Scenario: Condition recovery without a breach clause

- **GIVEN** `recovery.strategy = "condition"` and `query.breach` omitted
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

#### Scenario: Indistinguishable absence

- **GIVEN** `no_data.strategy = "resolve"`, `no_data.query` omitted, and `query.breach` omitted
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating a no-data strategy other than `ignore` requires `query.breach` or `no_data.query`

#### Scenario: Recovering phase with manual recovery

- **GIVEN** `recovery.strategy = "manual"` and `state_transition.recovering` set
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

#### Scenario: Operator without both thresholds

- **GIVEN** `state_transition.pending = { count = 2, operator = "and" }`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating `operator` requires both `count` and `timeframe`

#### Scenario: Routing tags on a signal rule

- **GIVEN** `kind = "signal"` and `metadata.routing_tags = ["prod"]`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

#### Scenario: Unknown values suppress the check

- **GIVEN** `kind` is unknown at plan time because it derives from another resource
- **WHEN** the cross-field checks run
- **THEN** the checks that depend on `kind` SHALL be skipped

### Requirement: Collections and optional objects round-trip

Optional collections and objects SHALL map null in configuration to absent in the request, and absent in the response to null in state. `tags`, `routing_tags`, and `grouping.fields` SHALL be ordered lists stored exactly as returned. Removing an optional object or collection from configuration SHALL remove it from the rule on the next apply because the write is a full replacement.

#### Scenario: Absent tags read back as null

- **GIVEN** a rule applied without `metadata.tags`
- **WHEN** state is refreshed
- **THEN** `metadata.tags` SHALL be null, not an empty list

#### Scenario: Removing grouping clears it server-side

- **GIVEN** a managed rule applied with `grouping`, and a new configuration omitting `grouping`
- **WHEN** update runs
- **THEN** the `PUT` body SHALL NOT contain `grouping`
- **AND** the read-after-write SHALL show `grouping` null
- **AND** the following plan SHALL be empty

#### Scenario: Tag order is preserved

- **GIVEN** `metadata.tags = ["z", "a"]`
- **WHEN** the rule is applied and read back
- **THEN** state SHALL hold `["z", "a"]` in that order

### Requirement: Artifacts mapping

`artifacts` SHALL be an optional list of 1-100 objects with required `id` (1-150 chars, unique within the rule), `type` (1-128 chars), and `data` (normalized JSON object string with at most 32 top-level keys, each 1-256 chars). Semantically equal JSON SHALL NOT produce a diff. Type-specific `data` rules SHALL be left to the server and its rejection surfaced.

#### Scenario: Artifacts round-trip

- **GIVEN** `artifacts = [{ id = "rb", type = "runbook", data = jsonencode({ content = "# Steps" }) }]`
- **WHEN** the rule is applied and read back
- **THEN** state SHALL hold one artifact with the same `id`, `type`, and semantically equal `data`

#### Scenario: Reformatted JSON produces an empty plan

- **GIVEN** an applied artifact whose `data` is re-written in configuration with different whitespace or key order
- **WHEN** Terraform plans
- **THEN** the plan SHALL be empty

#### Scenario: Duplicate artifact ids rejected at plan time

- **GIVEN** two artifacts with `id = "rb"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

#### Scenario: Non-object data rejected at plan time

- **GIVEN** `data = jsonencode(["a"])` or `data = "not json"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating `data` must be a JSON object

#### Scenario: Server-side type validation surfaced

- **GIVEN** an artifact with `type = "runbook"` and `data = jsonencode({})`
- **WHEN** apply runs and Kibana answers HTTP 400
- **THEN** the provider SHALL surface the server's message as an error diagnostic

### Requirement: Server-owned fields

`version`, `created_at`, and `updated_at` SHALL be computed and never sent. `created_at` SHALL always keep its prior state value. `version` and `updated_at` SHALL be planned as unknown whenever another attribute changes and SHALL keep their prior values when the plan is otherwise empty. Actor fields (`created_by`, `updated_by`) and the rule-template link SHALL NOT be exposed.

#### Scenario: No perpetual diff

- **GIVEN** an applied rule with unchanged configuration
- **WHEN** Terraform re-plans after refresh
- **THEN** the plan SHALL be empty, including for `version` and `updated_at`

#### Scenario: Version advances on update

- **GIVEN** a managed rule at `version = 1`
- **WHEN** `metadata.name` changes and update runs
- **THEN** state SHALL hold `version = 2`
- **AND** `created_at` SHALL be unchanged

### Requirement: Not-found and engine-disabled handling

On read, HTTP 404 SHALL remove the resource from state without error. On delete, HTTP 404 SHALL be treated as success. On any operation, HTTP 503 whose body carries error code `ALERTING_DISABLED` SHALL produce an error diagnostic naming the `alerting:v2:enabled` Kibana advanced setting.

#### Scenario: Externally deleted rule is dropped from state

- **GIVEN** a managed rule deleted in Kibana outside Terraform
- **WHEN** refresh runs and `GET` returns 404
- **THEN** the provider SHALL remove the resource from state without an error diagnostic
- **AND** the next plan SHALL propose creating it

#### Scenario: Delete tolerates an already-absent rule

- **GIVEN** a managed rule deleted outside Terraform
- **WHEN** destroy runs and `DELETE` returns 404
- **THEN** the provider SHALL treat the operation as successful

#### Scenario: Engine disabled

- **GIVEN** a target Kibana space with `alerting:v2:enabled` set to false
- **WHEN** any lifecycle operation runs and the API answers 503 with code `ALERTING_DISABLED`
- **THEN** the provider SHALL return an error diagnostic naming `alerting:v2:enabled`

### Requirement: Error surfacing and connection handling

Transport failures, unexpected HTTP statuses, and successful responses with an empty body SHALL produce error diagnostics that include the server's message when one is present. The resource SHALL use the provider-level Kibana client by default and the scoped client from a resource-level `kibana_connection` block when present.

#### Scenario: Validation error from the server

- **GIVEN** a request the server rejects with HTTP 400 and a JSON error body
- **WHEN** the operation completes
- **THEN** the diagnostic SHALL include the server's `message`

#### Scenario: Duplicate rule name surfaced from the server

- **GIVEN** a rule whose `metadata.name` already belongs to another rule in the same space
- **WHEN** apply runs and the server answers HTTP 409
- **THEN** the provider SHALL return an error diagnostic including the server's `code` and `message`

#### Scenario: Scoped connection is honoured

- **GIVEN** a resource with a `kibana_connection` block
- **WHEN** any lifecycle operation runs
- **THEN** the request SHALL be issued through the client derived from that block

### Requirement: Acceptance tests

The acceptance test suite SHALL be skipped below Elastic Stack 9.6.0, SHALL follow the repository's acceptance-test isolation conventions, and SHALL cover:

1. An alert rule with `time_field`, `schedule.lookback`, `metadata.builder`, `query.breach`, `recovery`, `no_data`, `state_transition`, `grouping`, `tags`, `routing_tags`, and `artifacts`: create, update, import, and an empty re-plan.
2. A signal rule with only `metadata`, `schedule`, and `query`.
3. `enabled = false` on create and toggling it on update.
4. Removing `grouping`, `state_transition`, and `artifacts` and asserting they are cleared.
5. A rule in a non-default space, with composite-id import.
6. `rule_id` omitted (generated) and supplied.
7. Plan-only `ExpectError` steps for the cross-field rules and schema bounds, runnable without a live stack.

#### Scenario: Empty plan after every apply step

- **GIVEN** any acceptance configuration applied successfully
- **WHEN** a plan runs with no configuration change
- **THEN** the plan SHALL be empty

#### Scenario: Validation cases run without a stack

- **GIVEN** the plan-only validation test cases
- **WHEN** they run without `TF_ACC` or a live stack
- **THEN** they SHALL still execute and assert the expected diagnostics
