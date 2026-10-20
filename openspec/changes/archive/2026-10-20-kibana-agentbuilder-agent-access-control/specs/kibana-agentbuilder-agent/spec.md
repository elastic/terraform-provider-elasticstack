# Delta spec: `kibana-agentbuilder-agent` — `access_control` attribute

Extends the main spec at `openspec/specs/kibana-agentbuilder-agent/spec.md`.

## Schema refinements

Add an `access_control` attribute to the canonical schema:

```hcl
resource "elasticstack_kibana_agentbuilder_agent" "example" {
  # ... existing attributes unchanged ...

  access_control = {                      # optional, computed, single nested object; requires Elastic Stack 9.5.0+
    access_mode = "public"                # required (within the object); one of "public", "shared", "private"
  }
}
```

## ADDED Requirements

### Requirement: access_control schema

The resource SHALL expose an optional, computed `access_control` single-nested attribute. When present, its nested `access_mode` attribute SHALL be required and SHALL accept only `"public"`, `"shared"`, or `"private"`.

#### Scenario: access_control attribute present in schema

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be an `access_control` attribute exposing a required `access_mode` string

#### Scenario: Invalid access_mode rejected

- **GIVEN** `access_control = { access_mode = "invalid" }` is configured
- **WHEN** Terraform validates the configuration
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: access_control with no access_mode rejected

- **GIVEN** `access_control = {}` is configured
- **WHEN** Terraform validates the configuration
- **THEN** Terraform SHALL report a validation error before any API call, since `access_mode` is required within the object

### Requirement: access_control version gate

An agent create or update that sends `access_control` SHALL require Elastic Stack 9.5.0 or later. If the server is below 9.5.0, the provider SHALL return a version-requirement error diagnostic before the write API call. This attribute-specific requirement SHALL NOT block refresh or deletion, even if state carries a prior `access_control` value; the existing 9.3.0 lifecycle baseline continues to apply. Omitting `access_control` from a write SHALL be unaffected by this gate on any Elastic Stack version supported by this resource (9.3.0+).

#### Scenario: access_control sent to an older stack

- **GIVEN** the Kibana server is below version 9.5.0
- **WHEN** create or update would send `access_control`
- **THEN** the provider SHALL return a version-requirement error diagnostic naming the 9.5.0 minimum
- **AND** the provider SHALL NOT perform the write API call

#### Scenario: Refresh and deletion with carried-forward mode against an older stack

- **GIVEN** the Kibana server is below version 9.5.0
- **AND** state contains an `access_control` value from a prior operation against a supported server
- **WHEN** refresh and deletion run
- **THEN** refresh SHALL proceed without a version-requirement error and leave `access_control` null in state
- **AND** deletion SHALL proceed subject only to the existing lifecycle baseline

#### Scenario: access_control omitted from a write against an older stack

- **GIVEN** the Kibana server is below version 9.5.0
- **WHEN** a create or update does not send `access_control`
- **THEN** the provider SHALL proceed without an attribute-specific version-requirement error

### Requirement: access_control create mapping

On create, when `access_control` is known and non-null on the plan, the resource SHALL include `access_control.access_mode` in the `POST` request body; when null or unknown, it SHALL be omitted, leaving the server to apply its own default.

#### Scenario: access_control set on create

- **GIVEN** `access_control = { access_mode = "private" }` is configured
- **WHEN** the resource is created
- **THEN** the `POST` request body SHALL include `access_control.access_mode = "private"`

#### Scenario: access_control omitted on create

- **GIVEN** `access_control` is not configured
- **WHEN** the resource is created
- **THEN** the `POST` request body SHALL NOT include `access_control`

### Requirement: access_control update mapping

On update, the resource SHALL include `access_control.access_mode` in the `PUT` request body only when `access_control` is non-null in the **Terraform configuration** (not merely on the plan). When `access_control` is absent from configuration, the resource SHALL omit `access_control` from the update body entirely, regardless of any computed value carried forward in state or plan.

#### Scenario: access_control changed on update

- **GIVEN** an agent exists with `access_control.access_mode = "private"`
- **WHEN** the resource is updated with `access_control = { access_mode = "shared" }`
- **THEN** the `PUT` request body SHALL include `access_control.access_mode = "shared"`

#### Scenario: access_control removed from config on update relinquishes management

- **GIVEN** an agent exists with `access_control.access_mode = "shared"` previously configured and present in state
- **AND** another mutable attribute is changed so that an update is required
- **WHEN** the resource is updated with `access_control` removed from configuration
- **THEN** the `PUT` request body SHALL NOT include `access_control`
- **AND** the plan and prior state SHALL still contain the carried-forward mode before the update
- **AND** a subsequent read SHALL report whatever `access_mode` the server currently has, without the provider asserting a client-side reset

### Requirement: access_control read-back via dedicated endpoint

After a successful create, update, or refresh (read) of an agent — and only when the configured Kibana server supports the Agent Builder access-control sub-resource (9.5.0+) — the resource SHALL call the dedicated `GET /api/agent_builder/agents/{id}/access_control` endpoint and populate `access_control.access_mode` in state from its response. When the server is below that version, `access_control` SHALL be left null in state and this additional call SHALL NOT be attempted.

#### Scenario: State populated from dedicated endpoint on supported stack

- **GIVEN** the Kibana server is 9.5.0 or later and an agent exists with an assigned `access_mode`
- **WHEN** the resource is created, updated, or refreshed
- **THEN** the provider SHALL call the dedicated access-control `GET` endpoint and `access_control.access_mode` in state SHALL match its response

#### Scenario: access_control left null on unsupported stack

- **GIVEN** the Kibana server is below version 9.5.0
- **WHEN** the resource is created, updated, or refreshed
- **THEN** `access_control` SHALL be null in state and the provider SHALL NOT call the dedicated access-control endpoint

#### Scenario: Configured mode drift detected when changed out-of-band

- **GIVEN** the Kibana server is 9.5.0 or later and an agent has an explicitly configured `access_mode`
- **AND** its server mode was changed outside Terraform to a different value since the last apply
- **WHEN** `terraform plan` runs
- **THEN** the plan SHALL show a diff on `access_control.access_mode` to restore the explicitly configured value

#### Scenario: Unmanaged mode adopted on refresh

- **GIVEN** the Kibana server is 9.5.0 or later and `access_control` is omitted from configuration or management has been relinquished
- **AND** the server has an `access_mode` value
- **WHEN** refresh runs
- **THEN** state SHALL adopt the server's current `access_mode`
- **AND** a subsequent plan SHALL NOT propose a corrective update for that mode

### Requirement: No read-back for a missing agent

The dedicated access-control read-back SHALL NOT be attempted when the main agent `GET` for the same refresh returned HTTP 404 (agent not found); the existing not-found-removes-from-state handling SHALL take precedence.

#### Scenario: No read-back attempted for a deleted agent

- **GIVEN** an agent no longer exists in Kibana
- **WHEN** refresh runs
- **THEN** the provider SHALL remove the resource from state per the existing not-found handling and SHALL NOT call the dedicated access-control endpoint
