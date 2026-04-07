# `elasticstack_fleet_elastic_defend_integration_policy` — Schema and Functional Requirements

Resource implementation: `internal/fleet/elastic_defend_integration_policy`

## Purpose

Manage Elastic Defend Fleet integration policies (package policies for the `endpoint` package). The resource provides a typed Terraform schema for Defend-specific configuration, including operating-system protection settings, event collection, and popups. It uses the Kibana Fleet package policy API with a two-phase create (bootstrap then finalize) and preserves opaque server-managed payloads such as `artifact_manifest` and package policy `version` in private state rather than exposing them in the public schema.

## Schema

```hcl
resource "elasticstack_fleet_elastic_defend_integration_policy" "example" {
  id                  = <computed, string>             # same as policy_id; UseStateForUnknown
  policy_id           = <optional+computed, string>    # force new; UseStateForUnknown; import key
  name                = <required, string>
  namespace           = <required, string>
  agent_policy_id     = <required, string>
  description         = <optional, string>
  enabled             = <optional+computed, bool>      # default true
  force               = <optional, bool>
  integration_version = <required, string>
  space_ids           = <optional+computed, set(string)>  # UseStateForUnknown

  preset = <optional, string>

  policy = <required, single nested attribute> {
    windows = <optional, single nested attribute> {
      events = <optional, single nested attribute> {
        process             = <optional, bool>
        network             = <optional, bool>
        file                = <optional, bool>
        dll_and_driver_load = <optional, bool>
        dns                 = <optional, bool>
        registry            = <optional, bool>
        security            = <optional, bool>
        authentication      = <optional, bool>
      }
      malware = <optional, single nested attribute> {
        mode          = <optional, string>
        blocklist     = <optional, bool>
        on_write_scan = <optional, bool>
        notify_user   = <optional, bool>
      }
      ransomware = <optional, single nested attribute> {
        mode      = <optional+computed, string>  # default "off"
        supported = <optional+computed, bool>    # default true
      }
      memory_protection = <optional, single nested attribute> {
        mode      = <optional+computed, string>  # default "off"
        supported = <optional+computed, bool>    # default true
      }
      behavior_protection = <optional, single nested attribute> {
        mode               = <optional+computed, string>  # default "off"
        supported          = <optional+computed, bool>    # default true
        reputation_service = <optional+computed, bool>    # default false
      }
      popup = <optional+computed, single nested attribute> {
        malware = <optional, single nested attribute> {
          message = <optional+computed, string>  # default ""
          enabled = <optional+computed, bool>    # default false
        }
        ransomware = <optional, single nested attribute> {
          message = <optional+computed, string>  # default ""
          enabled = <optional+computed, bool>    # default false
        }
        memory_protection = <optional, single nested attribute> {
          message = <optional+computed, string>  # default ""
          enabled = <optional+computed, bool>    # default false
        }
        behavior_protection = <optional, single nested attribute> {
          message = <optional+computed, string>  # default ""
          enabled = <optional+computed, bool>    # default false
        }
      }
      logging = <optional, single nested attribute> {
        file = <optional, string>
      }
      antivirus_registration = <optional, single nested attribute> {
        mode    = <optional+computed, string>  # default "disabled"
        enabled = <optional+computed, bool>    # default false
      }
      attack_surface_reduction = <optional, single nested attribute> {
        credential_hardening = <optional, single nested attribute> {
          enabled = <optional+computed, bool>  # default false
        }
      }
    }
    mac = <optional, single nested attribute> {
      events = <optional, single nested attribute> {
        process = <optional, bool>
        network = <optional, bool>
        file    = <optional, bool>
      }
      malware = <optional, single nested attribute> {
        mode          = <optional, string>
        blocklist     = <optional, bool>
        on_write_scan = <optional, bool>
        notify_user   = <optional, bool>
      }
      memory_protection = <optional, single nested attribute> {
        mode      = <optional+computed, string>  # default "off"
        supported = <optional+computed, bool>    # default true
      }
      behavior_protection = <optional, single nested attribute> {
        mode               = <optional+computed, string>  # default "off"
        supported          = <optional+computed, bool>    # default true
        reputation_service = <optional+computed, bool>    # default false
      }
      popup = <optional, single nested attribute> {
        malware = <optional, single nested attribute> {
          message = <optional+computed, string>  # default ""
          enabled = <optional+computed, bool>    # default false
        }
        memory_protection = <optional, single nested attribute> {
          message = <optional+computed, string>  # default ""
          enabled = <optional+computed, bool>    # default false
        }
        behavior_protection = <optional, single nested attribute> {
          message = <optional+computed, string>  # default ""
          enabled = <optional+computed, bool>    # default false
        }
      }
      logging = <optional, single nested attribute> {
        file = <optional, string>
      }
    }
    linux = <optional, single nested attribute> {
      events = <optional, single nested attribute> {
        process      = <optional, bool>
        network      = <optional, bool>
        file         = <optional, bool>
        session_data = <optional, bool>
        tty_io       = <optional, bool>
      }
      malware = <optional, single nested attribute> {
        mode      = <optional, string>
        blocklist = <optional, bool>
      }
      memory_protection = <optional, single nested attribute> {
        mode      = <optional+computed, string>  # default "off"
        supported = <optional+computed, bool>    # default true
      }
      behavior_protection = <optional, single nested attribute> {
        mode               = <optional+computed, string>  # default "off"
        supported          = <optional+computed, bool>    # default true
        reputation_service = <optional+computed, bool>    # default false
      }
      popup = <optional, single nested attribute> {
        malware = <optional, single nested attribute> {
          message = <optional+computed, string>  # default ""
          enabled = <optional+computed, bool>    # default false
        }
        memory_protection = <optional, single nested attribute> {
          message = <optional+computed, string>  # default ""
          enabled = <optional+computed, bool>    # default false
        }
        behavior_protection = <optional, single nested attribute> {
          message = <optional+computed, string>  # default ""
          enabled = <optional+computed, bool>    # default false
        }
      }
      logging = <optional, single nested attribute> {
        file = <optional, string>
      }
    }
  }
}
```

## Requirements

### Requirement: Dedicated Elastic Defend integration policy resource (REQ-001)

The provider SHALL expose a dedicated `elasticstack_fleet_elastic_defend_integration_policy` resource for managing Fleet package policies whose package name is `endpoint` (Elastic Defend). The resource SHALL own the full package policy lifecycle for that Defend policy rather than layering additional behavior into `elasticstack_fleet_integration_policy`.

#### Scenario: Elastic Defend uses the dedicated resource

- GIVEN a Terraform configuration that manages an Elastic Defend package policy
- WHEN the provider plans and applies the configuration
- THEN the package policy SHALL be managed through `elasticstack_fleet_elastic_defend_integration_policy`
- AND the generic `elasticstack_fleet_integration_policy` capability SHALL remain unchanged

### Requirement: Shared Fleet client supports both package policy input encodings (REQ-002)

The provider implementation backing this resource SHALL use a shared Kibana Fleet package policy client that supports both mapped and typed input encodings. That shared client support SHALL be available to provider code without requiring a Defend-specific transport or duplicate package policy model outside `generated/kbapi`. The shared client support SHALL preserve the fields needed for the Defend typed path, including typed input `type`, typed input `config`, and the top-level package policy `version` used on Defend updates. The shared Fleet helper layer SHALL also expose the package policy query-format behavior needed for mapped and typed workflows so the generic and Defend resources can choose the correct Fleet API behavior explicitly.

#### Scenario: Shared client can represent typed and mapped inputs

- GIVEN provider code interacting with Fleet package policies
- WHEN it needs to work with mapped inputs for generic integrations or typed inputs for Elastic Defend
- THEN the shared Fleet package policy client SHALL support both encodings
- AND the generated package policy model SHALL preserve the typed input `type`, typed input `config`, and top-level package policy `version` fields needed by the Defend flow
- AND the Fleet helper layer SHALL allow mapped and typed workflows to select the correct package policy query-format behavior
- AND provider code SHALL NOT need a separate Defend-only package policy client model

### Requirement: Focused package-policy envelope and fixed package identity (REQ-003)

The resource SHALL expose a familiar package-policy envelope with `id`, `policy_id`, `name`, `namespace`, `agent_policy_id`, `description`, `enabled`, `force`, `integration_version`, and `space_ids`. The resource SHALL always target package name `endpoint` and SHALL NOT expose a user-configurable `integration_name`. The resource SHALL NOT expose the generic `vars_json`, generic `inputs`, generic `streams`, `output_id`, or `agent_policy_ids` surfaces from `elasticstack_fleet_integration_policy` in v1.

#### Scenario: Package name is fixed to Elastic Defend

- GIVEN a valid `elasticstack_fleet_elastic_defend_integration_policy` configuration
- WHEN create or update builds the API request
- THEN the request body SHALL target package name `endpoint`
- AND there SHALL be no user-configurable `integration_name` in the Terraform schema

### Requirement: Identity and import (REQ-004)

The resource SHALL expose computed `id` and `policy_id` attributes whose values are set from the Kibana package policy id returned by the API. `policy_id` SHALL be the import key and SHALL use import passthrough semantics. Changes to a configured `policy_id` SHALL require replacement.

#### Scenario: Import by policy id

- GIVEN an existing Elastic Defend package policy id
- WHEN `terraform import` is run for `elasticstack_fleet_elastic_defend_integration_policy`
- THEN the imported id SHALL populate `policy_id`
- AND a subsequent read SHALL populate the modeled schema fields from the API response

### Requirement: Read and import validate package identity (REQ-005)

On read and import, the resource SHALL validate that the resolved package policy belongs to package name `endpoint`. If the resolved package policy does not belong to the `endpoint` package, the provider SHALL return an error diagnostic rather than attempting to map it into Defend resource state.

#### Scenario: Importing a non-endpoint package policy fails clearly

- GIVEN a package policy id that belongs to a package other than `endpoint`
- WHEN that id is imported into `elasticstack_fleet_elastic_defend_integration_policy`
- THEN the provider SHALL return an error diagnostic stating that the package policy is not an Elastic Defend policy
- AND the provider SHALL NOT write Defend resource state for it

### Requirement: Typed Defend configuration schema (REQ-006)

The resource SHALL model Defend-owned configuration through typed Terraform attributes and nested attributes. The `preset` attribute SHALL map to `config.integration_config.value.endpointConfig.preset` in read/update payloads. The `policy` attribute SHALL contain optional `windows`, `mac`, and `linux` nested attributes, each with a distinct schema containing only the fields applicable to that operating system. Structurally invalid combinations (such as `policy.linux.ransomware`) SHALL be impossible at plan time without requiring custom validation. The typed schema SHALL also model the provider-known Defend defaults for popup entries, protection mode objects, behavior protection, Windows antivirus registration, and Windows attack surface reduction credential hardening so omitted or empty nested blocks plan with the same effective values the Defend policy uses.

#### Scenario: Policy settings are modeled as typed attributes

- GIVEN a configuration that enables or disables Defend protections and event collection
- WHEN Terraform validates the configuration
- THEN those settings SHALL be represented by typed resource attributes and nested attributes
- AND the configuration SHALL NOT require users to provide raw `policy` JSON

#### Scenario: Linux event settings include documented Linux-specific leaves

- GIVEN a configuration for Linux event collection on the Defend resource
- WHEN Terraform maps the `policy.linux.events` schema to and from the API
- THEN the typed schema SHALL include the documented Linux-specific event flags
- AND those flags SHALL include `session_data` and `tty_io`

#### Scenario: Omitted nested policy settings use modeled defaults

- GIVEN a configuration that omits or leaves empty a popup item, protection-mode object, behavior-protection object, Windows antivirus registration object, or Windows attack-surface-reduction credential hardening object
- WHEN Terraform plans the Defend resource
- THEN the omitted or empty nested object SHALL use the modeled defaults
- AND `policy.windows.popup` SHALL default to an object whose `malware`, `ransomware`, `memory_protection`, and `behavior_protection` entries each use the popup-item defaults
- AND popup items SHALL default to `{ message = "", enabled = false }`
- AND protection-mode objects SHALL default to `{ mode = "off", supported = true }`
- AND behavior-protection objects SHALL default to `{ mode = "off", supported = true, reputation_service = false }`
- AND `policy.windows.antivirus_registration` SHALL default to `{ mode = "disabled", enabled = false }`
- AND `policy.windows.attack_surface_reduction.credential_hardening` SHALL default to `{ enabled = false }`

### Requirement: Resource boundary — Defend is typed-only (REQ-007)

`elasticstack_fleet_elastic_defend_integration_policy` SHALL use only the typed-input package policy encoding in its create, read, and update behavior. It SHALL NOT expose or depend on the mapped-input encoding used by `elasticstack_fleet_integration_policy`.

#### Scenario: Defend resource uses typed inputs only

- GIVEN a Defend resource create or update operation
- WHEN the provider builds the Fleet package policy request
- THEN the request SHALL use the typed-input encoding
- AND the resource's public behavior SHALL not depend on mapped input IDs or mapped stream IDs

### Requirement: Create uses the documented Defend bootstrap flow (REQ-008)

On create, the resource SHALL create the Elastic Defend package policy using the Defend-specific bootstrap request shape, attached to the configured `agent_policy_id`. The bootstrap request SHALL use package name `endpoint`, the configured `integration_version`, and the configured `preset`. The bootstrap request SHALL use the typed input shape with input `type = "endpoint"`, input `enabled = true`, input `streams = []`, and preset mapped under `config.integration_config.value.endpointConfig.preset`. After the bootstrap call succeeds, the resource SHALL use the API response as the source of truth for server-managed Defend data required for subsequent operations.

**Note:** Kibana rejects `"ENDPOINT_INTEGRATION_CONFIG"` as the input type and `"_config"` as the config key. Both bootstrap and finalize/update paths use `"endpoint"` and `"integration_config"` respectively.

#### Scenario: Create bootstraps a new Defend package policy

- GIVEN a new `elasticstack_fleet_elastic_defend_integration_policy` resource
- WHEN create runs
- THEN the provider SHALL first create the underlying package policy through the Defend bootstrap request flow
- AND the provider SHALL capture the returned package policy id and server-managed Defend payloads from the response

### Requirement: Create finalizes the modeled policy after bootstrap (REQ-009)

After the bootstrap create succeeds, the resource SHALL submit a Defend-specific update request that applies the configured typed `policy` settings to the new package policy. The finalized request SHALL include the provider-modeled Defend `policy` payload, the configured `preset` mapped under `config.integration_config.value.endpointConfig.preset`, and the top-level package policy `version`. The `artifact_manifest` SHALL NOT be included — Kibana manages it server-side and rejects it when present in typed input config.

#### Scenario: Create applies modeled policy settings after bootstrap

- GIVEN a Defend resource configuration with non-default policy settings
- WHEN create completes
- THEN the provider SHALL apply those settings through a follow-up Defend package policy update
- AND Terraform users SHALL NOT need to supply server-managed Defend payloads directly

### Requirement: Update preserves opaque server-managed Defend payloads (REQ-010)

On update, the resource SHALL send the Defend-specific typed package policy shape required by Kibana, including the latest provider-modeled `preset` and `policy` values. The provider SHALL preserve and resend the top-level package policy `version` (for optimistic concurrency control) without exposing it in the public Terraform schema. The `artifact_manifest` SHALL NOT be included in update requests — Kibana manages it server-side and rejects it when present. The provider captures `artifact_manifest` from API responses into private state but does not send it on update.

#### Scenario: Update succeeds without exposing artifact manifest

- GIVEN an existing Defend resource that was previously created or imported
- WHEN a user changes a modeled policy setting
- THEN the provider SHALL include the package policy `version` token in the update request for optimistic concurrency control
- AND the Terraform schema SHALL not expose `artifact_manifest` as a configurable field
- AND the provider SHALL NOT send `artifact_manifest` in the update request body

### Requirement: Read and import map only modeled fields to state (REQ-011)

On read and import, the resource SHALL parse the Defend-specific package policy response and populate only the modeled Terraform schema fields. The provider SHALL map `preset` from the Defend `integration_config` payload and SHALL map the typed `policy` payload into the corresponding operating-system nested attributes. The provider SHALL ignore unmodeled server-managed Defend payloads in Terraform state, except for preserving any opaque values required for future updates in internal provider-managed state.

#### Scenario: Read ignores unmodeled server-managed Defend fields

- GIVEN a Defend package policy response that includes `artifact_manifest` and other server-managed Defend data
- WHEN the resource reads or imports that package policy
- THEN Terraform state SHALL include only the modeled schema fields
- AND the provider SHALL preserve any required opaque update data internally

### Requirement: Provider-managed internal state for update prerequisites (REQ-012)

The resource SHALL maintain internal provider-managed state for opaque Defend data that must survive between operations but does not belong in the public schema. This internal state SHALL include the latest package policy `version` (used for optimistic concurrency control on updates) and the latest `artifact_manifest` (captured for observability, even though it is not sent in requests). Both SHALL be refreshed from successful create, read, update, and import responses.

#### Scenario: Import captures opaque update prerequisites

- GIVEN a Defend package policy imported into Terraform
- WHEN the import-triggered read runs
- THEN the provider SHALL capture the current opaque Defend update prerequisites from the API response
- AND a subsequent update SHALL be able to reuse them without additional user input

### Requirement: Fleet package policy CRUD, space awareness, and diagnostics (REQ-013)

The resource SHALL use the Kibana Fleet package policy APIs to create, read, update, and delete the underlying package policy. The resource SHALL obtain its Fleet client from provider configuration. When `space_ids` is configured or returned, the resource SHALL preserve the operational space needed for subsequent read, update, and delete operations, following the same space-aware lifecycle pattern as the existing Fleet integration policy resource. Transport failures, unexpected response shapes, and API errors SHALL be surfaced as Terraform diagnostics. On read, a not-found response SHALL remove the resource from state.

#### Scenario: Read removes missing Defend policy from state

- GIVEN a Defend package policy that has been deleted outside Terraform
- WHEN the resource refreshes state
- THEN the provider SHALL remove the Terraform resource from state instead of returning a persistent error
