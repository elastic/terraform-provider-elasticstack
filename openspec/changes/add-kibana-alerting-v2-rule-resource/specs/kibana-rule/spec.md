# `elasticstack_kibana_rule` — Schema and Functional Requirements

> **Status.** Ships as an **experimental** technical-preview resource: registered via `experimentalResources()`, gated by `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`, no generated registry docs. Graduation is [rna-program#678](https://github.com/elastic/rna-program/issues/678). Normative detail: REQ-001.

Resource implementation (future): `internal/kibana/rule`

## Purpose

Define the Terraform schema and runtime behaviour for the `elasticstack_kibana_rule` resource, which manages **Kibana Alerting v2** rules through `/api/alerting/v2/rules`.

In scope:

- **Schema** — typed attributes for the closed v2 rule body (ES|QL query union, schedule, recovery / no-data / state-transition, grouping, artifacts)
- **Lifecycle** — CRUD endpoints, space-aware identity and import, server-owned and volatile fields
- **Gates** — plan-time cross-field validation, disabled-engine diagnostics, experimental registration

This capability is distinct from `kibana-alerting-rule` (classic / v1). The two share no schema, no saved object type, and no identifier space.

## Schema

```hcl
resource "elasticstack_kibana_rule" "example" {
  # Identity
  id       = <computed, string>            # composite "<space_id>/<rule_id>"; UseStateForUnknown
  rule_id  = <optional, computed, string>  # client-supplied saved object id, <=150 chars; RequiresReplace; UseStateForUnknown
  space_id = <optional, computed, string>  # default "default"; RequiresReplace

  # Rule definition
  kind       = <required, string>            # "alert" | "signal"; RequiresReplace
  time_field = <optional, computed, string>  # default "@timestamp"; 1..128 chars
  enabled    = <optional, computed, bool>    # default true

  metadata {                                  # required, single nested block
    name         = <required, string>         # 1..256
    description  = <optional, string>         # <=1024
    owner        = <optional, string>         # <=256
    tags         = <optional, set(string)>    # non-empty when set
    builder_type = <optional, string>         # <=64
    version      = <computed, int64>          # server-managed revision; response-only, not accepted on create/update
  }

  schedule {                                  # required, single nested block
    every    = <required, string>             # duration; >= 5s and <= 365d
    lookback = <optional, string>             # duration; <= 365d
  }

  # Query — exactly one of composed_query / standalone_query
  composed_query {                            # optional, single nested block
    base = <required, string>                 # ES|QL, 1..10000 chars
    breach   { segment = <required, string> } # required nested block
    recovery { segment = <required, string> } # optional nested block
  }

  standalone_query {                          # optional, single nested block
    breach   { query = <required, string> }   # required nested block
    recovery { query = <required, string> }   # optional nested block
    no_data  { query = <required, string> }   # optional nested block
  }

  recovery_strategy = <optional, string>      # "no_breach" | "query" | "none"
  no_data_strategy  = <optional, string>      # "last_known_status" | "recover" | "none"

  state_transition {                          # optional, single nested block; kind = "alert" only
    pending_operator     = <optional, string> # "AND" | "OR"
    pending_count        = <optional, int64>  # 0..1000
    pending_timeframe    = <optional, string> # duration
    recovering_operator  = <optional, string> # "AND" | "OR"
    recovering_count     = <optional, int64>  # 0..1000
    recovering_timeframe = <optional, string> # duration
  }

  grouping {                                  # optional, single nested block
    fields = <required, list(string)>         # <=16 entries, each 1..256 chars
  }

  artifacts {                                 # optional, list nested block, <=100 entries
    id   = <required, string>                 # 1..256
    type = <required, string>                 # 1..128
    data = <required, json object string>     # Record<string, unknown>; <=32 keys, each key 1..256
  }

  # Server-owned
  version    = <computed, string>  # saved object OCC token
  created_by = <computed, string>  # nullable in API
  created_at = <computed, string>  # UseStateForUnknown
  updated_by = <computed, string>  # nullable in API
  updated_at = <computed, string>
}
```

Notes:

- `kibana_connection` and `timeouts` are injected by the `entitycore` Kibana resource envelope and are not declared in the resource's own schema factory.
- The API's `query` object is a discriminated union on `format`. The Terraform schema does not expose `format`; it is derived from which of `composed_query` / `standalone_query` is configured. See REQ-006.
- `no_data_strategy` deliberately omits the API's `"emit"` value: it is a valid stored value but is rejected by the create and update APIs. See REQ-014.
- `artifacts[].data` is an open record in the API ([kibana#281751](https://github.com/elastic/kibana/pull/281751)); Terraform exposes it as a JSON object string. See REQ-012.

## ADDED Requirements

### Requirement: Resource type and experimental registration (REQ-001)

The provider SHALL expose a managed resource with type name **`elasticstack_kibana_rule`**, registered through `Provider.experimentalResources(...)` in `provider/plugin_framework.go`. It SHALL NOT be returned from `Provider.resources(...)`, so practitioners MUST set `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` (or use an `AccTestVersion` provider build) to reach it. The resource SHALL embed the `entitycore` Kibana resource envelope so it satisfies the provider's `*entitycore.ResourceBase` registration contract.

Because `make docs-generate` runs `tfplugindocs` with `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=false`, the change SHALL NOT add a generated page under `docs/resources/`.

The resource SHALL NOT reuse or reclaim the type name `elasticstack_kibana_alerting_rule`, which belongs to the classic alerting resource, has an incompatible schema, and is not being renamed (a classic type rename would be a breaking change for customers).

#### Scenario: Resource is present only on the experimental surface

- **GIVEN** a provider built with a released version string
- **WHEN** `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` is unset and Terraform requests the provider's resource set
- **THEN** `elasticstack_kibana_rule` SHALL NOT be present
- **AND WHEN** the variable is set to `true`
- **THEN** `elasticstack_kibana_rule` SHALL be present

#### Scenario: No generated registry documentation

- **GIVEN** the resource is registered as experimental
- **WHEN** `make docs-generate` runs
- **THEN** no `docs/resources/kibana_rule.md` page SHALL be produced, and `make check-docs` SHALL remain green

### Requirement: CRUD endpoint mapping (REQ-002)

The resource SHALL manage rules through the following Kibana Alerting v2 endpoints, addressed space-aware (`/s/{space_id}` prefix when `space_id` is not `default`):

| Operation | Endpoint |
|-----------|----------|
| Create | `PUT /api/alerting/v2/rules/{rule_id}` |
| Read | `GET /api/alerting/v2/rules/{rule_id}` |
| Update | `PUT /api/alerting/v2/rules/{rule_id}` |
| Delete | `DELETE /api/alerting/v2/rules/{rule_id}` |
| Enable / disable reconciliation | `POST /api/alerting/v2/rules/{rule_id}/_enable` and `.../_disable` |

Create and Update SHALL both use `PUT` (create-or-replace). The resource SHALL NOT use `POST /api/alerting/v2/rules` and SHALL NOT use `PATCH /api/alerting/v2/rules/{id}`. Bulk endpoints, `_run`, `_tags`, the find endpoint, and `GET /api/alerting/v2/execution_history/rules` SHALL NOT be called by this resource.

After a successful create or update, the resource SHALL re-read the rule with `GET` and SHALL fail with an error diagnostic if the rule cannot be read back.

On `GET` returning HTTP 404, Read SHALL remove the resource from state without an error. On `DELETE` returning HTTP 404, Delete SHALL be treated as success.

#### Scenario: Create issues a PUT with a known id

- **GIVEN** a new `elasticstack_kibana_rule` resource
- **WHEN** create runs
- **THEN** the provider SHALL issue `PUT /api/alerting/v2/rules/{rule_id}` with a body derived entirely from the plan
- **AND** SHALL NOT issue `POST /api/alerting/v2/rules`

#### Scenario: Update replaces rather than patches

- **GIVEN** an existing managed rule with a changed attribute
- **WHEN** update runs
- **THEN** the provider SHALL issue `PUT /api/alerting/v2/rules/{rule_id}` with a full-replace body derived entirely from the plan
- **AND** SHALL NOT issue `PATCH /api/alerting/v2/rules/{rule_id}`

#### Scenario: Attribute removed from configuration is removed server side

- **GIVEN** a managed rule that was applied with a `grouping` block, and a new configuration that omits `grouping`
- **WHEN** update runs
- **THEN** the PUT body SHALL omit `grouping`
- **AND** the subsequent read SHALL show `grouping` absent, producing an empty plan on the next run

#### Scenario: Read removes a deleted rule from state

- **GIVEN** a managed rule that has been deleted outside Terraform
- **WHEN** refresh runs and `GET` returns 404
- **THEN** the provider SHALL remove the resource from state without an error diagnostic

#### Scenario: Delete tolerates an already-absent rule

- **GIVEN** a managed rule that has already been deleted outside Terraform
- **WHEN** destroy runs and `DELETE` returns 404
- **THEN** the provider SHALL treat the operation as successful

### Requirement: Identity, `rule_id` generation, and import (REQ-003)

The resource SHALL set `id` to the composite string `<space_id>/<rule_id>` after a successful read, matching the convention used by other space-scoped Kibana resources (the v2 rule saved object is registered with `namespaceType: 'multiple-isolated'`, so a rule id is only unique within a space).

`rule_id` SHALL be optional and computed. When the practitioner supplies it, the provider SHALL use that value as the saved object id in the `PUT` path. When the practitioner omits it, the provider SHALL generate a value at apply time (before the `PUT`) and persist it in state, so that every write is addressed by a known id. `rule_id` SHALL be at most 150 characters.

Changing `rule_id` or `space_id` SHALL require replacement.

The resource SHALL support `terraform import` with an id of the form `<space_id>/<rule_id>` — exactly one `/` separating two non-empty segments. On success it SHALL populate `space_id` and `rule_id` from those segments and set `id` to the full import string. A malformed import id SHALL produce an error diagnostic naming the required format.

#### Scenario: Generated rule_id is stable across applies

- **GIVEN** a configuration that omits `rule_id`
- **WHEN** the resource is applied and then re-planned with no configuration change
- **THEN** `rule_id` in state SHALL be unchanged and the plan SHALL be empty

#### Scenario: Re-apply with an explicit rule_id converges

- **GIVEN** a configuration with an explicit `rule_id` whose rule already exists in Kibana with different attributes
- **WHEN** apply runs
- **THEN** the `PUT` SHALL replace the existing rule rather than fail with a conflict

#### Scenario: Import round-trip

- **GIVEN** an import id `my-space/my-rule`
- **WHEN** import runs
- **THEN** state SHALL hold `space_id = "my-space"`, `rule_id = "my-rule"`, and `id = "my-space/my-rule"`

#### Scenario: Malformed import id is rejected

- **GIVEN** an import id with no `/` separator, or with an empty segment
- **WHEN** import runs
- **THEN** the provider SHALL return an error diagnostic describing the `<space_id>/<rule_id>` format

### Requirement: Immutable `kind` (REQ-004)

`kind` SHALL be a required string accepting only `"alert"` or `"signal"`, and changing it SHALL require replacement. The API classifies `kind` as its only immutable rule field and answers a `PUT` that changes it with HTTP 409; marking the attribute `RequiresReplace` means the provider never sends such a request.

#### Scenario: Changing kind plans a replacement

- **GIVEN** a managed rule with `kind = "alert"` and a configuration changing it to `"signal"`
- **WHEN** Terraform plans
- **THEN** the plan SHALL indicate destroy-and-create for the resource
- **AND** no `PUT` carrying a changed `kind` against the existing rule SHALL be issued

#### Scenario: Unknown kind rejected at plan time

- **GIVEN** `kind` set to a value other than `"alert"` or `"signal"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic listing the permitted values

### Requirement: Typed rule configuration, not an opaque JSON blob (REQ-005)

Every field of the rule create body SHALL be expressed as a typed Terraform attribute or nested block, as laid out in the **Schema** section. The resource SHALL NOT expose a single JSON-string attribute (an equivalent of the v1 resource's `params`) as the means of configuring the whole rule.

The sole exception is `artifacts[].data`, which SHALL be a JSON object string because the API models it as an open `Record<string, unknown>` (see REQ-012 and `design.md` D10). All other rule fields SHALL be ordinary typed attributes or blocks and SHALL NOT require `jsonencode`.

Constraints the API declares — enum membership, string length bounds, integer ranges, collection size limits, and duration format — SHALL be enforced by schema-level validators so they surface during `terraform validate` and `terraform plan` rather than as an HTTP 400 during apply.

#### Scenario: Rule body is typed outside of artifact data

- **GIVEN** a configuration exercising `metadata`, `schedule`, a query block, `state_transition`, `grouping`, and `artifacts`
- **WHEN** the configuration is written
- **THEN** every rule field other than `artifacts[].data` SHALL be a typed attribute or block, with no `jsonencode` call
- **AND** each `artifacts` entry's `data` MAY be supplied via `jsonencode` (or an equivalent JSON object string)

#### Scenario: Bound violation caught before apply

- **GIVEN** `grouping.fields` with more than 16 entries, or `artifacts` with more than 100 entries, or `state_transition.pending_count` above 1000
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic
- **AND** no request SHALL be sent to Kibana

### Requirement: Query union modelled as mutually exclusive typed blocks (REQ-006)

The API's `query` discriminated union SHALL be modelled as two mutually exclusive top-level blocks, and the provider SHALL derive the API's `query.format` discriminator from which one is configured:

| Terraform | API |
|-----------|-----|
| `composed_query.base` | `query.base` with `query.format = "composed"` |
| `composed_query.breach.segment` | `query.breach.segment` |
| `composed_query.recovery.segment` | `query.recovery.segment` |
| `standalone_query.breach.query` | `query.breach.query` with `query.format = "standalone"` |
| `standalone_query.recovery.query` | `query.recovery.query` |
| `standalone_query.no_data.query` | `query.no_data.query` |

**Exactly one** of `composed_query` or `standalone_query` SHALL be configured. Configuring both, or neither, SHALL be a plan-time validation error.

Within `composed_query`, `base` and the `breach` block SHALL be required. Within `standalone_query`, the `breach` block SHALL be required. `segment` and `query` leaves SHALL each be required within their enclosing block, so an empty `breach {}` is a plan-time error rather than an HTTP 400.

ES|QL strings SHALL be limited to 10000 characters. The provider SHALL NOT attempt to parse or semantically validate ES|QL; that validation belongs to the server and its diagnostics SHALL be surfaced verbatim.

#### Scenario: Both query blocks configured

- **GIVEN** a configuration with both `composed_query` and `standalone_query`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating that exactly one query block is permitted

#### Scenario: Neither query block configured

- **GIVEN** a configuration with neither `composed_query` nor `standalone_query`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

#### Scenario: Format derived from the configured block

- **GIVEN** a configuration with only `composed_query`
- **WHEN** create runs
- **THEN** the request body SHALL contain `query.format = "composed"` and `query.base`
- **AND** SHALL NOT contain `query.no_data`

#### Scenario: Server-side ES|QL error is surfaced

- **GIVEN** a syntactically invalid ES|QL query that passes the length check
- **WHEN** apply runs and Kibana responds 400
- **THEN** the provider SHALL surface an error diagnostic carrying the server's message rather than a generic failure

### Requirement: Cross-field validation mirroring the API's refinements (REQ-007)

The provider SHALL reject at plan or validate time every configuration that the API's rule schema rejects through a cross-field refinement, so these failures do not first appear as an HTTP 400 mid-apply:

1. `state_transition` SHALL only be configured when `kind = "alert"`.
2. When `kind = "signal"`, the query SHALL be `standalone_query`; `composed_query` SHALL be invalid.
3. When `kind = "signal"`, `recovery_strategy` and `no_data_strategy` SHALL be unset or `"none"`.
4. A `recovery` block SHALL only be configured when `recovery_strategy = "query"`.
5. When `recovery_strategy = "query"`, a `recovery` block SHALL be configured.
6. A `standalone_query.no_data` block SHALL only be configured when `no_data_strategy` is set to a value other than `"none"`.
7. When `no_data_strategy` is set to a value other than `"none"` and the query is `standalone_query`, a `no_data` block SHALL be configured. Composed-format rules use `base` as the data-presence query and SHALL NOT require one.

Each diagnostic SHALL name the attributes involved and state the rule that was violated.

When any attribute participating in one of these rules is unknown at plan time, the provider SHALL skip that check rather than reporting a false positive.

#### Scenario: state_transition on a signal rule

- **GIVEN** `kind = "signal"` and a `state_transition` block
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating `state_transition` is permitted only when `kind` is `"alert"`

#### Scenario: Signal rule with a composed query

- **GIVEN** `kind = "signal"` and a `composed_query` block
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic requiring `standalone_query`

#### Scenario: Signal rule with a recovery strategy

- **GIVEN** `kind = "signal"` and `recovery_strategy = "no_breach"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

#### Scenario: Recovery query without the matching strategy

- **GIVEN** a `recovery` block and `recovery_strategy` unset or set to `"no_breach"` or `"none"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

#### Scenario: Recovery strategy without a recovery query

- **GIVEN** `recovery_strategy = "query"` and no `recovery` block in the configured query
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

#### Scenario: Standalone no-data strategy without a no_data query

- **GIVEN** `standalone_query` configured, `no_data_strategy = "recover"`, and no `no_data` block
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

#### Scenario: Composed no-data strategy needs no no_data block

- **GIVEN** `composed_query` configured and `no_data_strategy = "last_known_status"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL NOT return a validation diagnostic under this requirement

#### Scenario: Unknown values suppress the check

- **GIVEN** `kind` is unknown at plan time because it derives from another resource
- **WHEN** the cross-field checks run
- **THEN** the provider SHALL skip the checks that depend on `kind` rather than reporting an error

### Requirement: `metadata` mapping and tag handling (REQ-008)

`metadata` SHALL be a required single nested block mapping one-to-one onto the API's `metadata` object: `name` (required, 1–256 characters), `description` (optional, ≤1024), `owner` (optional, ≤256), `builder_type` (optional, ≤64), and `tags`.

`tags` SHALL be an optional set of strings. The API rejects an empty tags array, so when `tags` is null or an empty set the provider SHALL omit the `tags` key from the request body rather than sending `[]`. On read, an absent or empty API tag list SHALL become a null set in state, so a rule with no tags produces an empty plan.

`metadata.version` SHALL be a computed integer; see REQ-010.

#### Scenario: Empty tag set is omitted from the request

- **GIVEN** `metadata.tags` configured as an empty set, or omitted
- **WHEN** create or update runs
- **THEN** the request body SHALL NOT contain a `metadata.tags` key

#### Scenario: Absent tags read back as null

- **GIVEN** a rule that Kibana returns with no tags
- **WHEN** the response is mapped to state
- **THEN** `metadata.tags` SHALL be null, not an empty set

#### Scenario: Name length enforced at plan time

- **GIVEN** `metadata.name` longer than 256 characters or empty
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

### Requirement: Duration attributes (REQ-009)

`schedule.every`, `schedule.lookback`, `state_transition.pending_timeframe`, and `state_transition.recovering_timeframe` SHALL accept only the duration strings the API accepts, and SHALL be validated at plan time. All four SHALL be at most `365d`. `schedule.every` SHALL additionally be at least `5s`.

The provider SHALL NOT enforce the deployment-configurable minimum schedule interval (`xpack.alerting_v2.rules.minimumScheduleInterval`, default `1m`), which is a server setting the provider cannot read; a violation SHALL surface as the server's error.

#### Scenario: Malformed duration rejected at plan time

- **GIVEN** `schedule.every = "5 minutes"` or any string that is not a valid duration
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic for that attribute

#### Scenario: Below the absolute minimum

- **GIVEN** `schedule.every = "1s"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic naming the `5s` minimum

#### Scenario: Deployment minimum surfaced from the server

- **GIVEN** `schedule.every = "10s"` against a deployment whose configured minimum is `1m`
- **WHEN** apply runs
- **THEN** the provider SHALL surface the server's rejection as an error diagnostic rather than silently succeeding

### Requirement: Server-owned and volatile fields (REQ-010)

The following attributes SHALL be computed and SHALL NOT be settable by practitioners. Their plan behaviour SHALL reflect whether the server changes them on every write:

| Attribute | API source | Plan behaviour |
|-----------|-----------|----------------|
| `metadata.version` | `metadata.version` (integer revision counter) | computed, **no** `UseStateForUnknown` — the server increments it on every mutation, including enable and disable |
| `version` | top-level `version` (saved object OCC token, string) | computed, **no** `UseStateForUnknown` |
| `created_at` | `createdAt` | computed with `UseStateForUnknown` |
| `created_by` | `createdBy` (nullable) | computed with `UseStateForUnknown` |
| `updated_at` | `updatedAt` | computed, **no** `UseStateForUnknown` |
| `updated_by` | `updatedBy` (nullable) | computed, **no** `UseStateForUnknown` |

Because `metadata.version`, `version`, `updated_at`, and `updated_by` change on every successful write, they SHALL be planned as unknown whenever the resource has any other change, and SHALL be left at their prior state value when the plan is otherwise empty. They SHALL NOT be sent in any request body.

`created_by` and `updated_by` are nullable in the API and SHALL map to a null Terraform string when the API returns null.

The provider SHALL NOT send the top-level `version` token back to the API. `PUT` does not accept it, and the resource does not use `PATCH`.

#### Scenario: Volatile fields do not create a perpetual diff

- **GIVEN** a rule that has been applied and is unchanged in configuration
- **WHEN** Terraform re-plans after a refresh
- **THEN** the plan SHALL be empty, including for `metadata.version`, `version`, `updated_at`, and `updated_by`

#### Scenario: Volatile fields planned unknown on a real change

- **GIVEN** a managed rule with a changed `metadata.name`
- **WHEN** Terraform plans
- **THEN** `metadata.version`, `version`, `updated_at`, and `updated_by` SHALL be shown as known-after-apply

#### Scenario: created_at survives an update

- **GIVEN** a managed rule that is updated
- **WHEN** the update is applied
- **THEN** `created_at` and `created_by` in state SHALL be unchanged from before the update

#### Scenario: Null createdBy maps to a null attribute

- **GIVEN** a rule whose API response has `createdBy: null`
- **WHEN** the response is mapped to state
- **THEN** `created_by` SHALL be null and SHALL NOT be the string `"null"` or an empty string

### Requirement: `enabled` reconciliation (REQ-011)

`enabled` SHALL be an optional, computed boolean defaulting to `true`.

The v2 create and replace APIs do not accept `enabled`: `PUT` on a new id always produces an enabled rule, and `PUT` on an existing id preserves whatever the current value is. The provider SHALL therefore reconcile after every write — comparing the desired `enabled` against the value returned by the API and, when they differ, issuing `POST /rules/{rule_id}/_enable` or `POST /rules/{rule_id}/_disable` before the authoritative read.

The reconciliation call SHALL be skipped when the desired and actual values already agree.

When the reconciliation call fails, the provider SHALL return an error diagnostic that states the rule was written but its enabled state was not applied, so the practitioner understands the partially applied state.

If the API later accepts `enabled` in the `PUT` body (see `proposal.md`, contingent change 1), the provider SHALL send it in the body and SHALL drop the follow-up call; this requirement's observable behaviour for practitioners is unchanged by that switch.

#### Scenario: Create a disabled rule

- **GIVEN** a new resource with `enabled = false`
- **WHEN** create runs
- **THEN** the provider SHALL issue the `PUT`, observe the rule came back enabled, issue `POST .../_disable`, and only then perform the authoritative read
- **AND** state SHALL hold `enabled = false`

#### Scenario: No reconciliation when already correct

- **GIVEN** an existing enabled rule with `enabled = true` and an unrelated attribute change
- **WHEN** update runs
- **THEN** the provider SHALL NOT issue `_enable` or `_disable`

#### Scenario: Disabled rule stays disabled across an unrelated update

- **GIVEN** a managed rule with `enabled = false` and a changed `metadata.description`
- **WHEN** update runs
- **THEN** the `PUT` SHALL preserve the disabled state and state SHALL still hold `enabled = false` after the read

#### Scenario: Reconciliation failure is reported explicitly

- **GIVEN** a create where the `PUT` succeeds and the subsequent `_disable` call fails
- **WHEN** the operation completes
- **THEN** the provider SHALL return an error diagnostic stating that the rule was created but could not be disabled

### Requirement: `artifacts` mapping (REQ-012)

`artifacts` SHALL be an optional list nested block of at most 100 entries. Each entry SHALL have required `id` (1–256 characters), `type` (1–128), and `data` attributes, mapping onto the API's `{ id, type, data }[]` shape from [kibana#281751](https://github.com/elastic/kibana/pull/281751) (`value: string` → `data: Record<string, unknown>`).

`data` SHALL be a JSON object string (normalized on read). Plan-time validation SHALL reject non-objects, more than 32 keys (`MAX_ARTIFACT_DATA_FIELDS`), and key names outside 1–256 characters (`MAX_FIELD_NAME_LENGTH`). The provider SHALL NOT treat `data` as an opaque string blob that skips JSON parsing.

For types registered in the API's `ARTIFACT_DATA_SCHEMAS`, plan-time validation SHALL mirror the server rules:

- `runbook` — `data.content` is a required non-blank string of at most 50000 characters (`RUNBOOK_CONTENT_LIMIT`)
- `dashboard` — `data.dashboardId` is a required non-blank string of at most 1024 characters (`DEFAULT_ARTIFACT_DATA_FIELD_LIMIT`)

Unregistered types SHALL accept any `data` object at plan time; per-field size limits (default 1024 characters for string fields, or for the JSON serialization of structured values) SHALL surface as the server's rejection rather than a hardcoded provider table, so new artifact types do not require a provider schema change.

Because the resource uses full-replace `PUT`, omitting the `artifacts` block SHALL remove all artifacts from the rule. The provider SHALL NOT implement the "omit the key so the server preserves the existing value" behaviour that the classic resource needs for `flapping` and `artifacts` on `PUT`.

This shape is unrelated to the classic alerting rule's `artifacts` object, which is a `{dashboards, investigation_guide}` structure rather than a typed list.

#### Scenario: Artifacts round-trip

- **GIVEN** a rule applied with a `runbook` artifact (`data = jsonencode({ content = "# Steps" })`) and a `dashboard` artifact (`data = jsonencode({ dashboardId = "abc" })`)
- **WHEN** state is read after apply
- **THEN** both entries SHALL be present with the configured `id`, `type`, and equivalent `data` JSON

#### Scenario: Removing the block clears artifacts

- **GIVEN** a managed rule with `artifacts` configured, and a new configuration omitting the block
- **WHEN** update runs and the rule is read back
- **THEN** the rule SHALL have no artifacts and the following plan SHALL be empty

#### Scenario: Known-type required field rejected at plan time

- **GIVEN** an artifact with `type = "runbook"` and `data = jsonencode({})`
- **WHEN** plan runs
- **THEN** the provider SHALL return a plan-time validation error naming `data.content`

#### Scenario: Unregistered-type field limit surfaced from the server

- **GIVEN** an artifact with `type = "host"` and a `data` field whose string value exceeds 1024 characters
- **WHEN** apply runs
- **THEN** the provider SHALL surface the server's rejection as an error diagnostic naming the field and limit

### Requirement: Engine-disabled diagnostic (REQ-013)

Every v2 route responds HTTP 503 with error code `ALERTING_DISABLED` while the Kibana advanced setting `alerting:v2:enabled` is off (the setting defaults on in the next Kibana release, but can still be disabled). On receiving a 503 with that code from any operation, the provider SHALL return an error diagnostic that names the `alerting:v2:enabled` advanced setting and states that it must be turned on in the target space's Kibana, rather than surfacing a bare HTTP status.

#### Scenario: Actionable diagnostic when the engine is off

- **GIVEN** a target Kibana with `alerting:v2:enabled` set to false
- **WHEN** any CRUD operation on the resource runs
- **THEN** the provider SHALL return an error diagnostic naming `alerting:v2:enabled`

### Requirement: `no_data_strategy` write-path enum (REQ-014)

`no_data_strategy` SHALL accept only `"last_known_status"`, `"recover"`, and `"none"`. The API's fourth value, `"emit"`, is a valid stored value but is explicitly rejected by both the create and update APIs, so accepting it in configuration would guarantee an apply-time 400.

The read path SHALL tolerate `"emit"` arriving from the API and SHALL store it in state without error, so refreshing a rule created outside Terraform does not fail.

#### Scenario: emit rejected in configuration

- **GIVEN** `no_data_strategy = "emit"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic listing the three writable values

#### Scenario: emit tolerated on read

- **GIVEN** an imported rule whose stored `no_data_strategy` is `"emit"`
- **WHEN** refresh maps the response to state
- **THEN** the provider SHALL store `"emit"` without an error diagnostic

### Requirement: Error surfacing and connection handling (REQ-015)

For create, read, update, and delete, transport-layer failures, unexpected HTTP statuses, and successful responses with an empty body where rule data is required SHALL produce clear error diagnostics. The exceptions are the two not-found cases in REQ-002 and the 503 case in REQ-013.

The resource SHALL use the provider's configured Kibana client by default, and the scoped client derived from a resource-level `kibana_connection` block when one is present. If no usable client is available, every operation SHALL fail with a provider configuration error.

#### Scenario: Unexpected status produces a diagnostic

- **GIVEN** any operation whose response status is not a documented success, and is not one of the tolerated not-found or engine-disabled cases
- **WHEN** the operation completes
- **THEN** Terraform SHALL receive an error diagnostic describing the status and the server's message

#### Scenario: Scoped connection is honoured

- **GIVEN** a `kibana_connection` block on the resource
- **WHEN** any CRUD operation runs
- **THEN** the request SHALL be issued through the scoped Kibana client derived from that block

### Requirement: Acceptance tests (REQ-016)

The acceptance test suite for `elasticstack_kibana_rule` SHALL cover:

1. A `kind = "alert"` rule with `composed_query`: create, update an attribute, and verify state round-trip including `metadata`, `schedule`, and the query fields.
2. A `kind = "signal"` rule with `standalone_query`, verifying the signal-specific constraints hold against a live stack.
3. `recovery_strategy = "query"` with a `recovery` block, and a no-data configuration with `no_data_strategy` plus a `no_data` block on a standalone rule.
4. `state_transition`, `grouping`, and `artifacts`, including removing each block and confirming the value is cleared server side (REQ-012).
5. `enabled = false` on create, and toggling `enabled` on update, asserting the reconciliation in REQ-011.
6. Import via `<space_id>/<rule_id>`, and a non-default `space_id`.
7. A no-op re-plan after apply asserting an empty plan, which exercises the volatile-field handling in REQ-010.

Every test SHALL be skipped when the target stack does not meet the minimum version determined for this resource. Tests SHALL follow the repository's acceptance-test isolation conventions so parallel runs do not collide on rule ids or tags.

Plan-time validation requirements (REQ-006, REQ-007, REQ-014) SHALL additionally be covered by unit tests or `ExpectError` plan-only test steps, so they do not depend on a live stack.

#### Scenario: Empty plan after apply

- **GIVEN** any of the acceptance configurations applied successfully
- **WHEN** a plan is run immediately afterwards with no configuration change
- **THEN** the plan SHALL be empty

#### Scenario: Validation cases run without a stack

- **GIVEN** the cross-field validation test cases
- **WHEN** they run with `TF_ACC` unset or against no live stack
- **THEN** they SHALL still execute and assert the expected diagnostics

## Traceability (planned implementation index)

| Area | Planned files |
|------|---------------|
| Registration | `provider/plugin_framework.go` (`experimentalResources`) |
| Envelope wiring, import | `internal/kibana/rule/resource.go` |
| Schema and validators | `internal/kibana/rule/schema.go`, `internal/kibana/rule/validate.go` |
| Model mapping, version gate | `internal/kibana/rule/models.go` |
| CRUD callbacks | `internal/kibana/rule/{create,read,update,delete}.go` |
| HTTP client, enable/disable reconciliation, nullable unwrapping | `internal/clients/kibanaoapi/rule_v2.go` |
| Composite id parsing | `internal/clients/api_client.go` (`CompositeID`) |
| Generated types | `generated/kbapi/kibana.gen.go` (`alerting_v2_*` components) |
