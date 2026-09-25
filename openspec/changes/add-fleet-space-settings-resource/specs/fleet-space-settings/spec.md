## Purpose

Defines the `elasticstack_fleet_space_settings` resource, which manages Fleet's per-space settings (currently `allowed_namespace_prefixes`) for a Kibana space so the restriction on writable data stream namespaces can be reviewed as code and drift-detected.

## Schema

```hcl
resource "elasticstack_fleet_space_settings" "example" {
  space_id                   = string       # required, forces replacement
  allowed_namespace_prefixes = set(string)  # required, max 10 elements; empty set allowed

  # Read-only
  id         = <computed, string>           # equals space_id
  managed_by = <computed, string>           # reported by Fleet; may be null

  kibana_connection {                       # optional scoped-connection override
    # ...
  }
}
```

## ADDED Requirements

### Requirement: Fleet space settings API

The resource SHALL read a space's Fleet settings with `GET /api/fleet/space_settings` and write them with `PUT /api/fleet/space_settings`, both routed through the space given by `space_id` (`/s/<space_id>/api/fleet/space_settings`). Create and update SHALL both use the write call, and the settings object is treated as always existing for a space.
Reference: https://www.elastic.co/docs/api/doc/kibana/operation/operation-put-fleet-space-settings

#### Scenario: Create writes the configured prefixes

- GIVEN a configuration with `space_id = "team-a"` and `allowed_namespace_prefixes = ["team_a"]`
- WHEN apply creates the resource
- THEN the resource SHALL write the settings for space `team-a` with `allowed_namespace_prefixes = ["team_a"]`
- AND the resulting state SHALL contain `allowed_namespace_prefixes = ["team_a"]`

#### Scenario: Update writes changed prefixes

- GIVEN the resource exists with `allowed_namespace_prefixes = ["team_a"]`
- WHEN the configuration changes to `["team_a", "shared"]` and apply runs
- THEN the resource SHALL write the settings for the same space with the new set
- AND the state SHALL contain `["team_a", "shared"]`

### Requirement: Schema and validation

`space_id` SHALL be required and SHALL force replacement when changed. `allowed_namespace_prefixes` SHALL be a required set of strings that MAY be empty and SHALL contain at most 10 elements. Because it is a set, a duplicate value in configuration collapses to one element rather than being rejected. `managed_by` SHALL be computed and read-only. Validation SHALL fail at plan time with a clear error when the size constraint is violated.

#### Scenario: More than 10 prefixes is rejected

- GIVEN a configuration with 11 entries in `allowed_namespace_prefixes`
- WHEN Terraform validates or plans the configuration
- THEN the provider SHALL return a validation error
- AND no API call SHALL be made

#### Scenario: Empty set is accepted

- GIVEN a configuration with `allowed_namespace_prefixes = []`
- WHEN apply runs
- THEN the resource SHALL write an empty set for the space
- AND the state SHALL contain an empty set, not null

#### Scenario: Changing space_id replaces the resource

- GIVEN the resource exists for `space_id = "team-a"`
- WHEN the configuration changes `space_id` to `team-b`
- THEN the plan SHALL show the resource being replaced

### Requirement: Identity and import

The resource `id` SHALL equal `space_id`. The resource SHALL support import by `space_id`, and after import a refresh SHALL populate `space_id`, `allowed_namespace_prefixes` and `managed_by` from the Fleet API.

#### Scenario: Import by space ID

- GIVEN Fleet settings exist for space `team-a` with `allowed_namespace_prefixes = ["team_a"]`
- WHEN `terraform import elasticstack_fleet_space_settings.example team-a` runs
- THEN the state SHALL contain `id = "team-a"`, `space_id = "team-a"` and `allowed_namespace_prefixes = ["team_a"]`

#### Scenario: Import of the default space

- GIVEN the default Kibana space
- WHEN the resource is imported with ID `default`
- THEN the resource SHALL read the settings of the default space

### Requirement: Read and state mapping

On read, the resource SHALL set `allowed_namespace_prefixes` and `managed_by` from the API response. As a set, `allowed_namespace_prefixes` carries no meaningful order. A missing or null `allowed_namespace_prefixes` in the response SHALL map to an empty set, never null. When `managed_by` is absent in the response it SHALL be null in state. Values changed outside Terraform SHALL appear as drift in the next plan.

#### Scenario: Prefixes changed outside Terraform

- GIVEN the resource is in state with `["team_a"]`
- AND the prefixes are changed to `["team_b"]` directly in Kibana
- WHEN refresh runs
- THEN the state SHALL contain `["team_b"]`
- AND the next plan SHALL show a change back to `["team_a"]`

#### Scenario: managed_by is reported

- GIVEN Fleet returns a `managed_by` value for the space
- WHEN read runs
- THEN `managed_by` in state SHALL equal that value

#### Scenario: managed_by is absent

- GIVEN Fleet returns no `managed_by` value
- WHEN read runs
- THEN `managed_by` in state SHALL be null

### Requirement: Not-found handling on read

When the space or its settings cannot be found on read (HTTP 404), the resource SHALL remove itself from state. Any other API error SHALL be surfaced as a Terraform diagnostic.

#### Scenario: Space deleted outside Terraform

- GIVEN the resource is in state for space `team-a`
- AND the space `team-a` has been deleted in Kibana
- WHEN refresh runs
- THEN the resource SHALL be removed from state

#### Scenario: API error on read

- GIVEN the Fleet API returns a non-404 error on read
- WHEN refresh runs
- THEN the resource SHALL return the error as a diagnostic
- AND SHALL NOT remove the resource from state

### Requirement: Destroy resets prefixes

On destroy, the resource SHALL write an empty `allowed_namespace_prefixes` set for the space and then remove itself from state. The resource SHALL NOT delete the Kibana space. If the space no longer exists when destroy runs (HTTP 404), destroy SHALL succeed.

#### Scenario: Destroy clears the restriction

- GIVEN the resource exists with `allowed_namespace_prefixes = ["team_a"]`
- WHEN `terraform destroy` runs
- THEN the resource SHALL write an empty set for the space
- AND the resource SHALL be removed from state
- AND the Kibana space SHALL still exist

#### Scenario: Destroy when the space is already gone

- GIVEN the space has been deleted outside Terraform
- WHEN `terraform destroy` runs
- THEN destroy SHALL succeed and the resource SHALL be removed from state

### Requirement: Space is neither created nor verified

The resource SHALL NOT create Kibana spaces and SHALL NOT itself verify that `space_id` refers to an existing space. Whether Fleet accepts settings for a non-existent space is decided by the Fleet API (Kibana 9.4 accepts them without error); any error the API returns SHALL be surfaced as a diagnostic.

#### Scenario: Space is not created

- GIVEN no Kibana space `missing` exists
- WHEN apply creates the resource with `space_id = "missing"`
- THEN the resource SHALL NOT create the space `missing`
- AND the resource SHALL NOT perform a space-existence check of its own

#### Scenario: API rejects the space

- GIVEN the Fleet API responds with an error for space `missing`
- WHEN apply creates the resource with `space_id = "missing"`
- THEN apply SHALL fail with a diagnostic containing the API error

### Requirement: Minimum stack version

The resource SHALL require Elastic Stack 9.1.0 or newer. When the target server is older, any lifecycle operation SHALL fail with an unsupported-feature error stating the minimum version, and no settings API call SHALL be made.

#### Scenario: Server older than 9.1.0

- GIVEN the target stack version is 9.0.0
- WHEN apply creates the resource
- THEN the provider SHALL return an unsupported-feature error mentioning `9.1.0`
- AND no settings API call SHALL be made

#### Scenario: Refresh against a server older than 9.1.0

- GIVEN the resource is in state and the target stack version is 9.0.0
- WHEN refresh runs
- THEN the provider SHALL return an unsupported-feature error mentioning `9.1.0`

#### Scenario: Server at 9.1.0 or newer

- GIVEN the target stack version is 9.1.0 or newer
- WHEN apply creates the resource
- THEN the resource SHALL proceed with the settings write

### Requirement: Provider-level Fleet client by default with optional scoped override

The resource SHALL use the provider-configured Kibana/Fleet client by default. When `kibana_connection` is configured on the resource, the resource SHALL use a client scoped from that block for all API calls.

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN an API call runs
- THEN the resource SHALL use the provider-configured client

#### Scenario: Scoped connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN an API call runs
- THEN the resource SHALL use the client derived from that block

### Requirement: API errors are surfaced

When a create, update or destroy call returns an error response (including authorization failures due to missing `fleet-settings-all` privileges), the resource SHALL surface the API error as a Terraform diagnostic and SHALL NOT report success.

#### Scenario: Insufficient privileges

- GIVEN the configured credentials lack `fleet-settings-all`
- WHEN apply creates the resource
- THEN apply SHALL fail with a diagnostic containing the API error
