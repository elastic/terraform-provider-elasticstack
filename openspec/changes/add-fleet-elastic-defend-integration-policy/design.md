## Context

`elasticstack_fleet_integration_policy` currently models Fleet package policies through the mapped request/response shape used by most integrations: `inputs` is a map keyed by input ID, nested `streams` are also maps, and the resource exposes generic `vars_json` and per-input/per-stream JSON configuration. Today the generated Fleet package policy types and transform layer are also map-oriented, so the current client surface does not faithfully represent the typed Defend payload.

Elastic Defend does not fit that model cleanly. Its documented API flow uses a typed `inputs` list, with request and response payloads carrying Defend-specific `config` objects such as `integration_config`, `artifact_manifest`, and `policy`. The Defend policy surface is also conceptually different from most integrations: users reason about operating-system policy settings and presets, not arbitrary package-policy inputs and streams.

The aim of this change is to add Defend support without degrading the generic integration-policy resource. The shared Fleet client should understand both mapped and typed package policy inputs, but each Terraform resource should still own only one shape.

## Goals / Non-Goals

**Goals:**

- Add a dedicated `elasticstack_fleet_elastic_defend_integration_policy` resource with a Terraform-first schema for Elastic Defend package policies.
- Keep the resource envelope familiar to users of `elasticstack_fleet_integration_policy` by reusing common concepts such as `policy_id`, `name`, `namespace`, `agent_policy_id`, `description`, `enabled`, `force`, and `integration_version`.
- Model Defend policy settings as typed Terraform attributes and nested attributes where the API structure is stable and meaningful to users.
- Hide server-managed payloads such as `artifact_manifest` from the public schema while still preserving them for later updates.
- Add shared `generated/kbapi` support for both mapped and typed Fleet package policy input encodings.
- Keep the generic `integration_policy` implementation mapped-only, and keep the Defend resource typed-only.
- Preserve space-aware lifecycle behavior comparable to the existing integration policy resource.

**Non-Goals:**

- Expanding `elasticstack_fleet_integration_policy` to accept both mapped and typed package policy shapes in its Terraform schema or public behavior.
- Exposing generic package-policy surfaces such as `integration_name`, `vars_json`, generic `inputs`, generic `streams`, `output_id`, `agent_policy_ids`, or arbitrary raw request JSON on the Defend resource in v1.
- Modeling every possible Elastic Defend setting as raw passthrough JSON just to mirror the API exactly.
- Reworking unrelated Fleet package policy resources or changing how non-Defend integrations are handled.

## Decisions

Create a dedicated capability and resource for Elastic Defend.
Elastic Defend is a package policy under the hood, but its API shape and user-facing semantics are specialized enough that a dedicated resource is the simpler long-term design. This keeps the generic integration-policy capability focused on mapped inputs while allowing the Defend resource to own the typed-input shape it actually needs.

Alternative considered: extend `elasticstack_fleet_integration_policy`.
Rejected because it would spread Defend-specific typed-input handling into generic schema, state mapping, secrets/defaults logic, and acceptance expectations.

Expose a familiar package-policy envelope, but only one Defend-specific configuration surface.
The new resource should feel like a sibling of `elasticstack_fleet_integration_policy`, not a totally different provider experience. It should therefore retain familiar envelope fields for identity and placement, while omitting generic package-policy knobs that do not belong in a Defend-focused resource.

The proposed public schema is:

```hcl
resource "elasticstack_fleet_elastic_defend_integration_policy" "example" {
  id                  = <computed, string>
  policy_id           = <optional+computed, string> # force new; import key
  name                = <required, string>
  namespace           = <required, string>
  agent_policy_id     = <required, string>
  description         = <optional, string>
  enabled             = <optional+computed, bool>   # default true
  force               = <optional, bool>
  integration_version = <required, string>
  space_ids           = <optional+computed, set(string)>

  preset = <optional, string> # maps to config.integration_config.value.endpointConfig.preset

  policy = <required, single nested attribute> {
    windows = <optional, single nested attribute> {
      events = <optional, single nested attribute> {
        # Windows-specific event collection flags
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
        mode          = <optional, string>  # "off" | "detect" | "prevent"
        blocklist     = <optional, bool>
        on_write_scan = <optional, bool>
        notify_user   = <optional, bool>
      }
      ransomware = <optional, single nested attribute> {
        mode      = <optional, string>  # "off" | "detect" | "prevent"
        supported = <optional, bool>
      }
      memory_protection = <optional, single nested attribute> {
        mode      = <optional, string>  # "off" | "detect" | "prevent"
        supported = <optional, bool>
      }
      behavior_protection = <optional, single nested attribute> {
        mode               = <optional, string>  # "off" | "detect" | "prevent"
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
        file = <optional, string>  # "info" | "debug" | "warning" | "error" | "critical"
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
        # Mac-specific event collection flags
        process = <optional, bool>
        network = <optional, bool>
        file    = <optional, bool>
      }
      malware = <optional, single nested attribute> {
        mode          = <optional, string>  # "off" | "detect" | "prevent"
        blocklist     = <optional, bool>
        on_write_scan = <optional, bool>
        notify_user   = <optional, bool>
      }
      memory_protection = <optional, single nested attribute> {
        mode      = <optional, string>  # "off" | "detect" | "prevent"
        supported = <optional, bool>
      }
      behavior_protection = <optional, single nested attribute> {
        mode               = <optional, string>  # "off" | "detect" | "prevent"
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
        file = <optional, string>  # "info" | "debug" | "warning" | "error" | "critical"
      }
    }
    linux = <optional, single nested attribute> {
      events = <optional, single nested attribute> {
        # Linux-specific event collection flags
        process      = <optional, bool>
        network      = <optional, bool>
        file         = <optional, bool>
        session_data = <optional, bool>
        tty_io       = <optional, bool>
      }
      malware = <optional, single nested attribute> {
        mode      = <optional, string>  # "off" | "detect" | "prevent"
        blocklist = <optional, bool>
      }
      memory_protection = <optional, single nested attribute> {
        mode      = <optional, string>  # "off" | "detect" | "prevent"
        supported = <optional, bool>
      }
      behavior_protection = <optional, single nested attribute> {
        mode               = <optional, string>  # "off" | "detect" | "prevent"
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
        file = <optional, string>  # "info" | "debug" | "warning" | "error" | "critical"
      }
    }
  }
}
```

Each OS block (`windows`, `mac`, `linux`) uses a **distinct nested attribute schema** containing only the fields applicable to that operating system. This makes structurally invalid combinations (such as `policy.linux.ransomware` or `policy.mac.antivirus_registration`) impossible at plan time without requiring custom validation. Each OS block's schema is defined by the attributes documented above.

Leaf fields such as booleans, mode strings, notification enablement, and message text are modeled as typed attributes rather than as raw JSON blobs. Event collection includes the OS-specific leaves exposed in the documented API examples, with Linux carrying `session_data` and `tty_io` that do not appear in the Windows or Mac schemas.

The public schema should not expose the typed package policy input itself. The resource owns that translation internally, including:

- bootstrap create input `type = "ENDPOINT_INTEGRATION_CONFIG"`
- bootstrap preset path `config._config.value.endpointConfig.preset`
- persisted/read/update input `type = "endpoint"`
- persisted/read/update preset path `config.integration_config.value.endpointConfig.preset`
- fixed input `enabled = true`
- fixed `streams = []`

Alternative considered: a single `policy_json` string.
Rejected for the main configuration surface because it would throw away most of the value of a dedicated resource and make drift, validation, and docs much weaker. A raw JSON escape hatch could be reconsidered later if the API surface proves too volatile, but it is not the preferred first design.

Treat `artifact_manifest` and update concurrency tokens as opaque provider-managed state.
The documented Defend API flow shows that create and update require server-managed payloads that users should not author directly. The resource should therefore preserve the latest `artifact_manifest` and the package policy `version` token used on update in provider-managed private state (or equivalent internal state), then echo them back on update without exposing them as schema.

Alternative considered: expose `artifact_manifest_json` as computed state.
Rejected because it is operational noise for Terraform users, encourages accidental coupling to a server-managed implementation detail, and would make plans noisier without improving ergonomics.

Use a two-phase create flow.
The documented Defend flow first creates the package policy using the bootstrap Defend input shape, then customizes it with a full typed policy payload. The provider should mirror that lifecycle internally:

1. Create a minimal Elastic Defend package policy attached to the target agent policy.
2. Read the server-managed fields returned by Kibana, including `artifact_manifest` and the top-level package policy `version`.
3. Immediately update the new package policy with the desired typed Defend configuration.

This keeps the public schema simple while matching the API's behavior.

Alternative considered: submit the full desired policy in a single create request.
Rejected for the proposal because the documented API flow suggests the bootstrap response is the safe source of truth for server-managed Defend payloads.

Teach the shared generated `kbapi` client to support both mapped and typed package policy inputs.
The right boundary is not a bespoke Defend-only transport model. The shared Fleet package policy client should represent both input encodings so package-policy consumers can rely on one generated client surface. This change should incorporate the same general direction explored in [PR #1500](https://github.com/elastic/terraform-provider-elasticstack/pull/1500): the client understands both mapped and typed input bodies, and resource code chooses the encoding it owns.

That work needs to be broader than just the `inputs` union. The generated and transformed client surface must preserve:

- mapped and typed `inputs`
- typed input `type`
- typed input `config`
- top-level package policy `version` on update requests

The current package policy transform and generated types are not sufficient for this, so the proposal should treat `generated/kbapi` and `generated/kbapi/transform_schema.go` as first-class implementation scope.

Alternative considered: keep typed-input support out of `kbapi` and localize it inside the Defend resource.
Rejected because it would duplicate package-policy modeling outside the shared client, create two independent representations of the same API, and make Defend support harder to maintain.

Preserve a strict resource boundary on top of the shared client.
Even though `kbapi` should support both encodings, the Terraform resources should not become polymorphic. `elasticstack_fleet_integration_policy` should continue to use only mapped inputs, matching its existing schema and canonical spec. `elasticstack_fleet_elastic_defend_integration_policy` should use only typed inputs, matching the Defend API shape.

Alternative considered: let the generic resource accept either encoding once `kbapi` supports both.
Rejected because that would reintroduce the same complexity this change is trying to avoid, just one layer higher.

Make Fleet `format` an explicit wrapper-level concern.
The current Fleet helper layer hardcodes `format=simplified`. Because the Defend typed path may require different Fleet query-format handling from the mapped path, `internal/clients/fleet` should expose that distinction explicitly instead of burying it inside resource code. The proposal should require either parameterized helpers or separate wrapper methods so the generic and Defend resources can choose the correct Fleet behavior without becoming transport-aware.

Alternative considered: leave `format` hardcoded in `internal/clients/fleet` and work around it in the Defend resource.
Rejected because it would blur the client/resource boundary and make Defend support fragile.

Preserve mapped-only secrets behavior until Defend actually needs secret-backed attributes.
The existing generic resource has complex secret-reference handling that walks mapped `inputs` and `streams`. The Defend resource should not inherit that complexity speculatively. In v1, the typed Defend schema should not expose secret-backed attributes unless the API requirements clearly demand them. If Defend later needs secret-backed fields, that should be added deliberately with typed-input-aware secret handling.

## Risks / Trade-offs

- The Elastic Defend policy surface may evolve faster than a typed Terraform schema -> Scope the initial schema to the stable, user-facing settings already represented in the documented policy payload, and leave unmodeled server-managed fields internal.
- A two-phase create flow means create is more complex than typical package-policy CRUD -> Keep that complexity inside the resource and specify it clearly so tests cover both bootstrap and final update paths.
- Shared `kbapi` changes could accidentally broaden generic resource behavior -> Keep `fleet_integration_policy` mapped-only at the resource layer and cover that boundary with tests.
- Typed-input support in `kbapi` could complicate helper logic such as secret handling -> Reuse the lessons from [PR #1500](https://github.com/elastic/terraform-provider-elasticstack/pull/1500) and add focused unit coverage for both encodings.
- The Fleet query `format` parameter may differ between mapped and typed workflows -> Make wrapper-level format selection explicit and verify it against real Defend API behavior before implementation is considered done.
- Importing a non-`endpoint` package policy into the Defend resource would produce misleading state unless validated -> Require read/import to verify the package name and fail clearly when it is not `endpoint`.
- Omitting generic package-policy attributes might frustrate power users who want more control -> That trade-off is intentional; the generic `elasticstack_fleet_integration_policy` remains available for integrations that fit its simplified model.

## Migration Plan

- Add a new `fleet-elastic-defend-integration-policy` delta spec defining the schema and lifecycle of the dedicated Defend resource.
- Update `generated/kbapi` and `generated/kbapi/transform_schema.go` so Fleet package policies support both mapped and typed input encodings, typed input `type` and `config`, and the top-level package policy `version` used for Defend updates.
- Extend `internal/clients/fleet` so mapped and typed package policy workflows can choose the correct Fleet query-format behavior.
- Keep `internal/fleet/integration_policy` on mapped inputs only, even after the shared client supports both encodings.
- Implement a new `internal/fleet/elastic_defend_integration_policy` package that uses the typed-input path from `kbapi`, validates `package.name == "endpoint"` on read/import, and preserves the opaque Defend update prerequisites in private state.
- Register the new resource in the provider and generate resource documentation.
- Add targeted unit coverage for request construction, response parsing, private-state preservation of opaque server-managed data, and import/readback.
- Add acceptance coverage for create, update, import, and deletion against a Defend-capable stack.

## Open Questions

- Should `preset` remain a free-form string attribute or be validated against a known set of Defend presets? The initial proposal prefers a plain string unless the supported preset set is clearly stable and documented.
- Should the first iteration model only the policy fields present in the documented API example, or should it also include additional advanced settings surfaced in the UI if they map cleanly to the API? The proposal assumes the documented policy example is the minimum stable baseline and that advanced settings can be added incrementally.
