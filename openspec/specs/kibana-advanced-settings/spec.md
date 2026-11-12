# kibana-advanced-settings Specification

## Purpose

Define the Terraform behavior of the `elasticstack_kibana_advanced_settings` resource, which manages individual Kibana Advanced Settings (such as `dateFormat:tz` and `defaultRoute`) for a single Kibana space: which keys the resource owns, how values are typed, how drift is detected, and how settings are reset when keys are removed or the resource is destroyed. The resource manages only the setting keys declared in `string_settings` and `json_settings` ("managed keys"); every other setting in the space is left untouched, including settings changed through the Kibana UI.

## Requirements

### Requirement: Space scoping (REQ-001)

The resource SHALL apply, read, and reset settings only in the Kibana space identified by `space_id`. If `space_id` is omitted, it SHALL default to `default`.

#### Scenario: Default space

- **GIVEN** a configuration that does not set `space_id`
- **WHEN** Terraform plans the resource
- **THEN** `space_id` SHALL be `default`
- **AND** applying the configuration SHALL change settings in the `default` space

#### Scenario: Settings in different spaces are independent

- **GIVEN** one resource for space `ops` and one for space `default`, each declaring `dateFormat:tz` with a different value
- **WHEN** both resources are applied
- **THEN** each space SHALL hold its own configured value
- **BUT** neither resource SHALL change the value in the other space

### Requirement: Space change forces replacement (REQ-002)

Changing `space_id` on an existing resource SHALL require resource replacement rather than an in-place update.

#### Scenario: Replace on space change

- **GIVEN** an existing resource recorded for space `ops`
- **WHEN** `space_id` changes to `staging` in configuration
- **THEN** Terraform SHALL plan replacement of the resource

### Requirement: Identity (REQ-003)

The resource SHALL store a computed `id` equal to the value of `space_id`.

#### Scenario: ID equals space ID

- **GIVEN** a successful apply for space `ops`
- **WHEN** Terraform records state
- **THEN** `id` SHALL equal `ops`

### Requirement: String settings (REQ-004)

Each entry in `string_settings` SHALL be applied to Kibana as a string value for the setting named by the entry's key.

#### Scenario: Plain string applied

- **GIVEN** `string_settings = { "dateFormat:tz" = "Europe/Rome" }`
- **WHEN** the configuration is applied
- **THEN** Kibana SHALL hold the string `Europe/Rome` as the value of `dateFormat:tz` in the target space

#### Scenario: JSON text in a string setting stays a string

- **GIVEN** `string_settings` contains a key whose value is the text `["a","b"]`
- **WHEN** the configuration is applied
- **THEN** Kibana SHALL hold that value as the string `["a","b"]`, not as an array

### Requirement: JSON settings (REQ-005)

Each entry in `json_settings` SHALL be applied to Kibana as the JSON value obtained by parsing the entry's value, preserving its JSON type (boolean, number, string, array, or object).

#### Scenario: Boolean applied as a boolean

- **GIVEN** `json_settings = { "csv:quoteValues" = "false" }`
- **WHEN** the configuration is applied
- **THEN** Kibana SHALL hold the boolean `false` as the value of `csv:quoteValues`

#### Scenario: Number applied as a number

- **GIVEN** `json_settings = { "truncate:maxHeight" = "0" }`
- **WHEN** the configuration is applied
- **THEN** Kibana SHALL hold the number `0` as the value of `truncate:maxHeight`

#### Scenario: Array applied as an array

- **GIVEN** `json_settings = { "defaultColumns" = "[\"@timestamp\",\"message\"]" }`
- **WHEN** the configuration is applied
- **THEN** Kibana SHALL hold an array of the two strings as the value of `defaultColumns`

### Requirement: Setting key length (REQ-006)

Every key in `string_settings` and `json_settings` SHALL be between 1 and 1024 characters long.

#### Scenario: Empty key rejected

- **GIVEN** a configuration with an empty-string key in `string_settings`
- **WHEN** Terraform validates the configuration
- **THEN** validation SHALL fail with an error diagnostic
- **AND** no request SHALL be sent to Kibana

#### Scenario: Key longer than 1024 characters rejected

- **GIVEN** a configuration with a 1025-character key in `json_settings`
- **WHEN** Terraform validates the configuration
- **THEN** validation SHALL fail with an error diagnostic

### Requirement: A key is declared in at most one map (REQ-007)

A setting key SHALL NOT appear in both `string_settings` and `json_settings`.

#### Scenario: Duplicate key across maps rejected

- **GIVEN** `dateFormat:tz` is a key in both `string_settings` and `json_settings`
- **WHEN** Terraform validates the configuration
- **THEN** validation SHALL fail with an error diagnostic naming `dateFormat:tz`

#### Scenario: Same key in different resources is not checked

- **GIVEN** two resources for the same space that both declare `dateFormat:tz`
- **WHEN** Terraform validates the configuration
- **THEN** validation SHALL NOT fail because of the overlap

### Requirement: JSON value validation (REQ-008)

Each value in `json_settings` SHALL be valid JSON and SHALL NOT be the JSON value `null`. A null entry in either `string_settings` or `json_settings` SHALL also be rejected. A value that is unknown at plan time SHALL be validated once it is known.

#### Scenario: Invalid JSON rejected

- **GIVEN** `json_settings = { "truncate:maxHeight" = "not json" }`
- **WHEN** Terraform validates the configuration
- **THEN** validation SHALL fail with an error diagnostic identifying `truncate:maxHeight`
- **AND** no request SHALL be sent to Kibana

#### Scenario: JSON null rejected

- **GIVEN** `json_settings = { "defaultRoute" = "null" }`
- **WHEN** Terraform validates the configuration
- **THEN** validation SHALL fail with an error diagnostic advising that removing the key resets the setting

#### Scenario: Null entry in string settings rejected

- **GIVEN** `string_settings = { "defaultRoute" = null }`
- **WHEN** Terraform validates the configuration
- **THEN** validation SHALL fail with an error diagnostic advising that removing the key resets the setting

#### Scenario: Value unknown at plan time is validated once known

- **GIVEN** a `json_settings` value taken from another resource's output that is unknown at plan time
- **AND** the value resolves to invalid JSON at apply time
- **WHEN** the configuration is applied
- **THEN** the apply SHALL fail with an error diagnostic identifying the key
- **AND** no setting SHALL be changed in Kibana

#### Scenario: JSON string value accepted

- **GIVEN** `json_settings = { "defaultRoute" = "\"/app/dashboards\"" }`
- **WHEN** Terraform validates the configuration
- **THEN** validation SHALL succeed

### Requirement: Keys are not restricted to a provider-defined list (REQ-009)

The resource SHALL NOT reject a key because the provider does not recognize it. Whether a key and value are accepted SHALL be determined by Kibana.

#### Scenario: Unknown key passed through

- **GIVEN** `string_settings = { "custom:teamNote" = "hello" }` where Kibana has no registered setting named `custom:teamNote`
- **WHEN** the configuration is applied
- **THEN** the apply SHALL succeed
- **AND** Kibana SHALL hold the string `hello` for `custom:teamNote`

### Requirement: Create (REQ-010)

On create, the resource SHALL apply every declared key to the target space and SHALL NOT change any other setting in that space. If both maps are omitted or empty, create SHALL succeed without changing any setting.

#### Scenario: Create applies declared keys

- **GIVEN** a configuration declaring `dateFormat:tz` in `string_settings` and `csv:quoteValues` in `json_settings`
- **WHEN** the resource is created
- **THEN** both settings SHALL hold the configured values in the target space
- **AND** Terraform state SHALL contain both keys

#### Scenario: Unmanaged settings untouched on create

- **GIVEN** the target space already has `defaultRoute` set through the Kibana UI
- **AND** the configuration does not declare `defaultRoute`
- **WHEN** the resource is created
- **THEN** `defaultRoute` SHALL keep its existing value

#### Scenario: Empty configuration is a no-op

- **GIVEN** a configuration that omits both `string_settings` and `json_settings`, or sets both to empty maps
- **WHEN** the resource is created
- **THEN** the create SHALL succeed
- **AND** no setting in Kibana SHALL change

### Requirement: Update (REQ-011)

On update, the resource SHALL apply declared keys that are new or whose value changed, and SHALL reset to its Kibana default every key that was managed before the update and is no longer declared in either map. It SHALL NOT change any other setting.

#### Scenario: Changed value applied

- **GIVEN** a managed `dateFormat:tz` set to `Europe/Rome`
- **WHEN** configuration changes it to `America/New_York` and is applied
- **THEN** Kibana SHALL hold `America/New_York` for `dateFormat:tz`

#### Scenario: Added key applied

- **GIVEN** a resource managing only `dateFormat:tz`
- **WHEN** configuration adds `defaultRoute` and is applied
- **THEN** Kibana SHALL hold the configured `defaultRoute`
- **AND** `dateFormat:tz` SHALL be unchanged

#### Scenario: Removed key reset to default

- **GIVEN** a resource managing `dateFormat:tz` and `defaultRoute`
- **WHEN** configuration removes `defaultRoute` and is applied
- **THEN** `defaultRoute` SHALL be reset to its Kibana default
- **AND** `dateFormat:tz` SHALL keep its configured value

#### Scenario: Key moved between maps

- **GIVEN** `custom:teamNote` declared in `string_settings` with the text `5`
- **WHEN** configuration moves it to `json_settings` as `5` and is applied
- **THEN** Kibana SHALL hold the number `5` for `custom:teamNote`
- **BUT** the setting SHALL NOT be left reset to its default

#### Scenario: Unmanaged settings untouched on update

- **GIVEN** a setting in the target space that the resource does not manage
- **WHEN** any update is applied
- **THEN** that setting SHALL keep its value

### Requirement: Read refreshes managed keys (REQ-012)

On read, the resource SHALL refresh the value of every managed key from Kibana. A managed key for which Kibana holds no user-set value SHALL be dropped from the refreshed state. The resource SHALL NOT add, report, or compare keys that it does not manage.

#### Scenario: Out-of-band change detected

- **GIVEN** a managed `dateFormat:tz` applied as `Europe/Rome`
- **AND** the value in Kibana is changed to `UTC` outside Terraform
- **WHEN** Terraform refreshes and plans
- **THEN** the plan SHALL show `dateFormat:tz` changing back to `Europe/Rome`

#### Scenario: Managed key reset outside Terraform

- **GIVEN** a managed `defaultRoute` applied by Terraform
- **AND** the setting is reset to its default in Kibana outside Terraform
- **WHEN** Terraform refreshes and plans
- **THEN** the plan SHALL show `defaultRoute` being applied again

#### Scenario: Unmanaged change causes no drift

- **GIVEN** a resource not declaring `truncate:maxHeight`
- **AND** `truncate:maxHeight` is changed in Kibana outside Terraform
- **WHEN** Terraform refreshes and plans
- **THEN** the plan SHALL show no changes

#### Scenario: No diff after apply

- **GIVEN** a successful apply of any valid configuration
- **WHEN** Terraform plans again with no configuration change
- **THEN** the plan SHALL show no changes

### Requirement: Representation of refreshed values (REQ-013)

On read, a `string_settings` key SHALL be recorded as Kibana's string value, or as the JSON text of the value if Kibana holds a non-string. A `json_settings` key SHALL be recorded as the JSON text of Kibana's value, keeping the previously recorded text when the two are equal, so that insignificant whitespace, object key order, and equivalent number formatting produce no diff. Numbers SHALL be compared by value (`1.50` equals `1.5`).

#### Scenario: JSON formatting differences cause no diff

- **GIVEN** `json_settings = { "defaultColumns" = "[ \"@timestamp\",  \"message\" ]" }` applied
- **WHEN** Terraform refreshes and plans
- **THEN** the plan SHALL show no changes

#### Scenario: String setting now holds a non-string

- **GIVEN** a managed key `custom:teamNote` declared in `string_settings` as `hello`
- **AND** the value in Kibana is changed to the number `5` outside Terraform
- **WHEN** Terraform refreshes and plans
- **THEN** the plan SHALL show `custom:teamNote` changing from the recorded JSON text `5` to the configured string `hello`

#### Scenario: Object key order irrelevant

- **GIVEN** a `json_settings` object value applied with keys in one order
- **AND** Kibana returns the same object with keys in a different order
- **WHEN** Terraform refreshes and plans
- **THEN** the plan SHALL show no changes

#### Scenario: Number formatting differences cause no diff

- **GIVEN** `json_settings = { "histogram:barTarget" = "1.50" }` applied
- **AND** Kibana returns the number as `1.5`
- **WHEN** Terraform refreshes and plans
- **THEN** the plan SHALL show no changes

#### Scenario: Reformatting a value in configuration changes nothing in Kibana

- **GIVEN** a `json_settings` value applied
- **WHEN** the configuration reformats that value without changing its meaning, and is applied
- **THEN** the apply SHALL succeed
- **AND** no change SHALL be sent to Kibana for that key

### Requirement: Destroy (REQ-014)

On destroy, the resource SHALL reset every managed key in the target space to its Kibana default and SHALL NOT change any other setting.

#### Scenario: Destroy resets managed keys only

- **GIVEN** a resource managing `dateFormat:tz`
- **AND** the space also has `defaultRoute` set through the Kibana UI
- **WHEN** the resource is destroyed
- **THEN** `dateFormat:tz` SHALL be reset to its Kibana default
- **AND** `defaultRoute` SHALL keep its value

#### Scenario: Destroy with no managed keys

- **GIVEN** a resource that manages no keys
- **WHEN** the resource is destroyed
- **THEN** the destroy SHALL succeed
- **AND** no setting in Kibana SHALL change

#### Scenario: Destroy when a key is already at its default

- **GIVEN** a managed key that was reset in Kibana outside Terraform
- **WHEN** the resource is destroyed
- **THEN** the destroy SHALL succeed

### Requirement: Space no longer exists (REQ-015)

For a space other than `default`, the resource SHALL check that the space exists before reading or destroying its settings. If the space does not exist, read SHALL remove the resource from Terraform state and destroy SHALL succeed without changing any setting. A failed space lookup SHALL be reported as an error and SHALL NOT be treated as a missing space. This takes precedence over REQ-012 and REQ-014.

#### Scenario: Space deleted outside Terraform, then refresh

- **GIVEN** a resource for space `ops` and that space is deleted outside Terraform
- **AND** Kibana still serves settings that were previously written for that space
- **WHEN** Terraform refreshes
- **THEN** the resource SHALL be removed from state
- **AND** the refresh SHALL NOT fail

#### Scenario: Space deleted outside Terraform, then destroy

- **GIVEN** a resource for space `ops` and that space is deleted outside Terraform
- **WHEN** the resource is destroyed
- **THEN** the destroy SHALL succeed
- **AND** no setting SHALL be changed in Kibana

#### Scenario: Default space is never reported missing

- **GIVEN** a resource for the `default` space
- **WHEN** Terraform refreshes
- **THEN** the resource SHALL remain in state

#### Scenario: A failed space lookup is not treated as a missing space

- **GIVEN** a resource for space `ops` recorded in state
- **AND** the spaces lookup fails for a reason other than the space not existing, such as a 403
- **WHEN** Terraform refreshes
- **THEN** the refresh SHALL fail with an error diagnostic
- **AND** the resource SHALL remain in state
- **AND** a destroy in the same situation SHALL fail without changing any setting

### Requirement: Create and update require an existing space (REQ-019)

For a space other than `default`, create and update SHALL fail with an error diagnostic stating that the space must exist when it does not, and SHALL NOT write any setting to it. This takes precedence over REQ-010 and REQ-011.

#### Scenario: Create in a space that does not exist

- **GIVEN** a configuration whose `space_id` names a space that does not exist
- **WHEN** the configuration is applied
- **THEN** the apply SHALL fail with an error diagnostic stating that the space must exist
- **AND** no setting SHALL be written for that space
- **AND** the resource SHALL NOT be in Terraform state

#### Scenario: Update after the space was deleted outside Terraform

- **GIVEN** a resource for space `ops` recorded in state
- **AND** that space is deleted outside Terraform without a refresh
- **WHEN** a changed configuration is applied
- **THEN** the apply SHALL fail with an error diagnostic stating that the space must exist
- **AND** Terraform state SHALL keep the values recorded before the update

### Requirement: Kibana errors are surfaced (REQ-016)

When a Kibana request fails or is rejected, the operation SHALL fail with an error diagnostic that includes Kibana's error message. When the failed or rejected request sent `string_settings` entries, each such diagnostic SHALL also advise that settings with non-string values belong in `json_settings`. The error for a missing settings route (REQ-020) carries no such advice.

#### Scenario: Invalid value for a registered setting

- **GIVEN** a configuration with a value that Kibana rejects for a registered setting
- **WHEN** the configuration is applied
- **THEN** the apply SHALL fail with an error diagnostic containing Kibana's message

#### Scenario: Setting pinned by Kibana server configuration

- **GIVEN** a setting that Kibana's server configuration pins to a fixed value
- **AND** the configuration declares that setting
- **WHEN** the configuration is applied
- **THEN** the apply SHALL fail with an error diagnostic containing Kibana's message

#### Scenario: Missing privilege

- **GIVEN** credentials without Kibana's privilege to manage advanced settings
- **WHEN** the configuration is applied
- **THEN** the apply SHALL fail with an error diagnostic containing Kibana's message

#### Scenario: Non-string setting declared as a string

- **GIVEN** `string_settings = { "csv:quoteValues" = "maybe" }` where `csv:quoteValues` is a boolean setting and `maybe` is not a value Kibana accepts for it
- **WHEN** the configuration is applied
- **THEN** the apply SHALL fail with an error diagnostic containing Kibana's message
- **AND** the diagnostic SHALL advise using `json_settings` for non-string values

#### Scenario: Rejected JSON setting carries no string hint

- **GIVEN** a configuration declaring only `json_settings = { "histogram:barTarget" = "\"tall\"" }` where `histogram:barTarget` is a number setting
- **WHEN** the configuration is applied
- **THEN** the apply SHALL fail with an error diagnostic containing Kibana's message
- **BUT** the diagnostic SHALL NOT advise using `json_settings`

#### Scenario: Other failures of a request that sent string settings carry the advice

- **GIVEN** a configuration declaring a `string_settings` entry
- **AND** the request fails for a reason other than an invalid value, such as missing privilege or a connection failure
- **WHEN** the configuration is applied
- **THEN** the apply SHALL fail with an error diagnostic containing the failure's message
- **AND** the diagnostic SHALL advise using `json_settings` for non-string values

### Requirement: Failed operations do not record partial changes (REQ-017)

If create fails, the resource SHALL NOT be recorded in Terraform state. If update fails, Terraform state SHALL keep the values recorded before the update.

#### Scenario: Failed create

- **GIVEN** a configuration that Kibana rejects
- **WHEN** the resource is created
- **THEN** the create SHALL fail
- **AND** the resource SHALL NOT be in Terraform state

#### Scenario: Failed update

- **GIVEN** a managed `dateFormat:tz` applied as `Europe/Rome`
- **AND** an update to a value that Kibana rejects
- **WHEN** the update is applied
- **THEN** the update SHALL fail
- **AND** Terraform state SHALL still record `Europe/Rome`

### Requirement: Missing settings route is an error (REQ-020)

If Kibana answers HTTP 404 on the advanced settings route itself, for example on a deployment that does not serve it, the operation SHALL fail with an error diagnostic naming the missing route, and read SHALL NOT drop any managed key from state.

#### Scenario: Create against a deployment without the settings route

- **GIVEN** a Kibana deployment whose advanced settings route answers HTTP 404
- **WHEN** a configuration declaring settings is applied
- **THEN** the apply SHALL fail with an error diagnostic naming `POST /api/kibana/settings`
- **AND** the resource SHALL NOT be in Terraform state

#### Scenario: Refresh against a deployment without the settings route

- **GIVEN** a resource with managed keys recorded in state
- **AND** the Kibana deployment's advanced settings route answers HTTP 404
- **WHEN** Terraform refreshes
- **THEN** the refresh SHALL fail with an error diagnostic naming `GET /api/kibana/settings`
- **AND** every managed key SHALL remain in state

### Requirement: Provider connection with optional scoped override (REQ-018)

The resource SHALL use the provider's configured Kibana connection by default. When `kibana_connection` is configured on the resource, it SHALL use the connection defined by that block for all create, read, update, and destroy operations.

#### Scenario: Provider connection

- **GIVEN** a resource without `kibana_connection`
- **WHEN** any operation runs
- **THEN** the operation SHALL use the provider's Kibana connection

#### Scenario: Scoped connection

- **GIVEN** a resource with `kibana_connection` configured
- **WHEN** any operation runs
- **THEN** the operation SHALL use the Kibana connection defined by that block
