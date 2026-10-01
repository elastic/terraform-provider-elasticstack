# `elasticstack_kibana_advanced_settings` — Schema and Functional Requirements

Resource implementation: `internal/kibana/advancedsettings/`

## Purpose

Manage Kibana advanced settings (uiSettings) for a single space, or the global advanced settings shared by every space, so that changes made outside Terraform (for example in the Kibana UI) are detected as drift. The resource manages only the settings declared in its configuration and leaves every other setting in the same scope untouched.

## Schema

```hcl
resource "elasticstack_kibana_advanced_settings" "example" {
  id = <computed, string> # the space ID, or "global" for the global scope

  space_id = <optional, computed, string> # default "default"; force new; must not be set when global = true
  global   = <optional, computed, bool>   # default false; force new; requires Kibana >= 8.7.0 when true
  settings = <required, map(string)>      # setting name => JSON-encoded value; at least one entry

  kibana_connection {
    endpoints    = <optional, list(string)>
    username     = <optional, string>
    password     = <optional, string>
    api_key      = <optional, string>
    bearer_token = <optional, string>
    ca_certs     = <optional, list(string)>
    insecure     = <optional, bool>
  }
}
```

## Requirements

### Requirement: Advanced settings APIs (REQ-001–REQ-003)

For the space scope, the resource SHALL read settings with `GET /s/{space_id}/api/kibana/settings` and write them with `POST /s/{space_id}/api/kibana/settings`. For the global scope, the resource SHALL read settings with `GET /api/kibana/global_settings` and write them with `POST /api/kibana/global_settings`. Every request SHALL carry the `x-elastic-internal-origin: Kibana` header, because Kibana registers these routes without public access and rejects requests to them from Kibana 9.0 onwards unless that header is present. When Kibana returns a non-success response, the resource SHALL surface the error to Terraform diagnostics.

#### Scenario: API failure surfaces to diagnostics

- GIVEN a non-success Kibana response on read or write
- WHEN the provider handles the response
- THEN the error SHALL appear in Terraform diagnostics

#### Scenario: Unregistered global setting

- GIVEN `global = true` and a setting that Kibana does not register as a global setting
- WHEN the write returns HTTP 500
- THEN the diagnostics SHALL explain that the setting may not be a global advanced setting

### Requirement: Identity and import (REQ-004–REQ-006)

The resource SHALL set `id` to the `space_id` for the space scope and to `global` for the global scope. The resource SHALL support import with an ID that is either a space ID or `global`. Import SHALL set only `id`, `space_id` and `global`; it SHALL NOT import settings, so that a subsequent apply never resets settings the configuration does not declare. An empty import ID SHALL be rejected.

#### Scenario: Import a space

- GIVEN the import ID `team-a`
- WHEN import runs
- THEN `id` and `space_id` SHALL be `team-a`, `global` SHALL be false and `settings` SHALL be null

#### Scenario: Import the global scope

- GIVEN the import ID `global`
- WHEN import runs
- THEN `id` SHALL be `global` and `global` SHALL be true

### Requirement: Configuration validation (REQ-007–REQ-009)

The resource SHALL reject a configuration that sets `space_id` while `global` is true. The resource SHALL reject a setting whose value is null or the JSON literal `null`, because Kibana interprets null as a reset. Each setting value SHALL be valid JSON and `settings` SHALL contain at least one entry with non-empty keys.

#### Scenario: space_id with global

- GIVEN `global = true` and `space_id` set in configuration
- WHEN the configuration is validated
- THEN the provider SHALL return an error diagnostic on `space_id`

#### Scenario: Null setting value

- GIVEN a setting whose value is `jsonencode(null)`
- WHEN the configuration is validated
- THEN the provider SHALL return an error diagnostic for that setting

### Requirement: Version requirements (REQ-010)

When `global` is true, the resource SHALL require Kibana 8.7.0 or later, the first version that serves the global settings API (the first global settings are registered in 8.8.0), and SHALL return an error diagnostic on `global` for older versions.

#### Scenario: Global settings on an old stack

- GIVEN `global = true` and Kibana older than 8.7.0
- WHEN create, read or update runs
- THEN the provider SHALL return an error diagnostic on `global`

### Requirement: Create and update (REQ-011–REQ-013)

On create and update, the resource SHALL send every configured setting, decoded from its JSON value, in a single write request. On update, the resource SHALL also send `null` for every setting present in the prior state but absent from the plan, resetting it to its Kibana default. The resource SHALL build the resulting state from the settings returned by the write response instead of a separate read, because Kibana caches settings per instance and only invalidates the cache of the instance that handled the write.

#### Scenario: Setting removed from configuration

- GIVEN a setting tracked in state
- AND the setting is removed from configuration
- WHEN update runs
- THEN the write request SHALL send the setting with a `null` value

### Requirement: Read and drift detection (REQ-014–REQ-016)

On read for the space scope, the resource SHALL first check that the space exists using the Kibana Get Space API, and SHALL remove itself from state when the space is not found, because Kibana serves settings for a nonexistent space without error. On read, for each setting tracked in state, the resource SHALL store the value Kibana reports, JSON-encoded. When the reported value is semantically equal to the tracked value, the resource SHALL keep the tracked string. A tracked setting that Kibana no longer reports SHALL be removed from state so the next plan sets it again. Settings Kibana reports but the state does not track SHALL be ignored. When `settings` is null in state (after import), read SHALL leave it null.

#### Scenario: Space deleted outside Terraform

- GIVEN the space of a space-scoped resource is deleted outside Terraform
- WHEN read runs
- THEN the resource SHALL be removed from state

#### Scenario: Setting changed outside Terraform

- GIVEN a tracked setting changed in the Kibana UI
- WHEN read runs
- THEN state SHALL hold the changed value and the next plan SHALL update the resource

#### Scenario: Setting reset outside Terraform

- GIVEN a tracked setting reset to its default in the Kibana UI
- WHEN read runs
- THEN the setting SHALL be absent from state and the next plan SHALL update the resource

### Requirement: Delete (REQ-017)

On delete, the resource SHALL send `null` for every setting tracked in state in a single write request, resetting those settings to their Kibana defaults, and SHALL leave every other setting unchanged.

#### Scenario: Destroy resets managed settings

- GIVEN a resource managing `xpackCustomBranding:pageTitle` in the global scope
- WHEN the resource is destroyed
- THEN Kibana SHALL no longer report a user value for `xpackCustomBranding:pageTitle`

### Requirement: Connection (REQ-018)

By default, the resource SHALL use the provider-level Kibana client. When `kibana_connection` is configured, the resource SHALL use a resource-scoped Kibana client for all API calls.

#### Scenario: Resource-level connection override

- GIVEN `kibana_connection` is set on the resource
- WHEN API calls run
- THEN they SHALL use the resource-scoped client
