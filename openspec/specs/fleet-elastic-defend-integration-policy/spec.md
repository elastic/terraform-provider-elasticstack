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
        mode      = <optional, string>
        supported = <optional, bool>
      }
      memory_protection = <optional, single nested attribute> {
        mode      = <optional, string>
        supported = <optional, bool>
      }
      behavior_protection = <optional, single nested attribute> {
        mode               = <optional, string>
        supported          = <optional, bool>
        reputation_service = <optional, bool>
      }
      popup = <optional, single nested attribute> {
        malware = <optional, single nested attribute> {
          message = <optional, string>
          enabled = <optional, bool>
        }
        ransomware = <optional, single nested attribute> {
          message = <optional, string>
          enabled = <optional, bool>
        }
        memory_protection = <optional, single nested attribute> {
          message = <optional, string>
          enabled = <optional, bool>
        }
        behavior_protection = <optional, single nested attribute> {
          message = <optional, string>
          enabled = <optional, bool>
        }
      }
      logging = <optional, single nested attribute> {
        file = <optional, string>
      }
      antivirus_registration = <optional, single nested attribute> {
        enabled = <optional, bool>
      }
      attack_surface_reduction = <optional, single nested attribute> {
        credential_hardening = <optional, single nested attribute> {
          enabled = <optional, bool>
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
        mode      = <optional, string>
        supported = <optional, bool>
      }
      behavior_protection = <optional, single nested attribute> {
        mode               = <optional, string>
        supported          = <optional, bool>
        reputation_service = <optional, bool>
      }
      popup = <optional, single nested attribute> {
        malware = <optional, single nested attribute> {
          message = <optional, string>
          enabled = <optional, bool>
        }
        memory_protection = <optional, single nested attribute> {
          message = <optional, string>
          enabled = <optional, bool>
        }
        behavior_protection = <optional, single nested attribute> {
          message = <optional, string>
          enabled = <optional, bool>
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
        mode      = <optional, string>
        supported = <optional, bool>
      }
      behavior_protection = <optional, single nested attribute> {
        mode               = <optional, string>
        supported          = <optional, bool>
        reputation_service = <optional, bool>
      }
      popup = <optional, single nested attribute> {
        malware = <optional, single nested attribute> {
          message = <optional, string>
          enabled = <optional, bool>
        }
        memory_protection = <optional, single nested attribute> {
          message = <optional, string>
          enabled = <optional, bool>
        }
        behavior_protection = <optional, single nested attribute> {
          message = <optional, string>
          enabled = <optional, bool>
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

### Requirement: Shared Fleet client supports both package policy input encodings (REQ-002)

The provider implementation backing this resource SHALL use a shared Kibana Fleet package policy client that supports both mapped and typed input encodings. That shared client support SHALL be available to provider code without requiring a Defend-specific transport or duplicate package policy model outside `generated/kbapi`. The shared client support SHALL preserve the fields needed for the Defend typed path, including typed input `type`, typed input `config`, and the top-level package policy `version` used on Defend updates. The shared Fleet helper layer SHALL also expose the package policy query-format behavior needed for mapped and typed workflows so the generic and Defend resources can choose the correct Fleet API behavior explicitly.

### Requirement: Focused package-policy envelope and fixed package identity (REQ-003)

The resource SHALL expose a familiar package-policy envelope with `id`, `policy_id`, `name`, `namespace`, `agent_policy_id`, `description`, `enabled`, `force`, `integration_version`, and `space_ids`. The resource SHALL always target package name `endpoint` and SHALL NOT expose a user-configurable `integration_name`. The resource SHALL NOT expose the generic `vars_json`, generic `inputs`, generic `streams`, `output_id`, or `agent_policy_ids` surfaces from `elasticstack_fleet_integration_policy` in v1.

### Requirement: Identity and import (REQ-004)

The resource SHALL expose computed `id` and `policy_id` attributes whose values are set from the Kibana package policy id returned by the API. `policy_id` SHALL be the import key and SHALL use import passthrough semantics. Changes to a configured `policy_id` SHALL require replacement.

### Requirement: Read and import validate package identity (REQ-005)

On read and import, the resource SHALL validate that the resolved package policy belongs to package name `endpoint`. If the resolved package policy does not belong to the `endpoint` package, the provider SHALL return an error diagnostic rather than attempting to map it into Defend resource state.

### Requirement: Typed Defend configuration schema (REQ-006)

The resource SHALL model Defend-owned configuration through typed Terraform attributes and nested attributes. The `preset` attribute SHALL map to `config.integration_config.value.endpointConfig.preset` in read/update payloads. The `policy` attribute SHALL contain optional `windows`, `mac`, and `linux` nested attributes, each with a distinct schema containing only the fields applicable to that operating system.

### Requirement: Resource boundary — Defend is typed-only (REQ-007)

`elasticstack_fleet_elastic_defend_integration_policy` SHALL use only the typed-input package policy encoding. It SHALL NOT expose or depend on the mapped-input encoding used by `elasticstack_fleet_integration_policy`.

### Requirement: Create uses the documented Defend bootstrap flow (REQ-008)

On create, the resource SHALL create the Elastic Defend package policy using the Defend-specific bootstrap request shape with:
- package name `endpoint`, the configured `integration_version`, and the configured `preset`
- typed input shape with `type = "ENDPOINT_INTEGRATION_CONFIG"`, `enabled = true`, `streams = []`
- preset mapped under `config._config.value.endpointConfig.preset`

### Requirement: Create finalizes the modeled policy after bootstrap (REQ-009)

After the bootstrap create succeeds, the resource SHALL submit a Defend-specific update request applying the configured typed `policy` settings, including `preset` under `config.integration_config.value.endpointConfig.preset`, and the server-managed `artifact_manifest` and `version` from the bootstrap response.

### Requirement: Update preserves opaque server-managed Defend payloads (REQ-010)

On update, the resource SHALL include the stored server-managed Defend payloads (`artifact_manifest` and `version`) in the update request without exposing those values in the public Terraform schema.

### Requirement: Read and import map only modeled fields to state (REQ-011)

On read and import, the resource SHALL parse the Defend-specific package policy response and populate only the modeled Terraform schema fields. The provider SHALL ignore unmodeled server-managed fields in Terraform state, except for preserving opaque values required for future updates in internal private state.

### Requirement: Provider-managed internal state for update prerequisites (REQ-012)

The resource SHALL maintain internal private state for opaque Defend data including at least the latest `artifact_manifest` and package policy `version`. This SHALL be refreshed from successful create, read, update, and import responses.

### Requirement: Fleet package policy CRUD, space awareness, and diagnostics (REQ-013)

The resource SHALL use the Kibana Fleet package policy APIs for CRUD. When `space_ids` is configured or returned, the resource SHALL preserve the operational space for subsequent operations. A not-found response on read SHALL remove the resource from state.
