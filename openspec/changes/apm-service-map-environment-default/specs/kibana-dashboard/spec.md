## MODIFIED Requirements

### Requirement: APM service map panel support (REQ-049) — ENVIRONMENT_ALL server-default suppression

The provider SHALL treat `environment = "ENVIRONMENT_ALL"` returned by the Kibana API as a
server-injected default (equivalent to the field being absent) and suppress it to null in state
when the practitioner's configuration does not explicitly set `environment`. This suppression
applies on both the normal read path and the import path.

Background: Kibana 9.5+ injects `environment = "ENVIRONMENT_ALL"` into APM service-map panel
configs when the practitioner omits the field. Because `"ENVIRONMENT_ALL"` means "no environment
filter" — identical in effect to the field being absent — storing it in state against a config
that omits `environment` causes spurious drift on every refresh and import.

#### Read-path suppression

On read (post-create/update refresh and standalone `terraform refresh`), the provider SHALL continue to apply REQ-049/REQ-009 null-preservation semantics for `apm_service_map_config.environment`: when the prior state has `environment` null or unknown, state SHALL keep it null regardless of the API value (including `"ENVIRONMENT_ALL"`).
When the prior state has a known `environment` value, the provider SHALL leave `environment` as returned by the API.

#### Import-path suppression

On import (no prior state), when the API returns `environment = "ENVIRONMENT_ALL"`, the provider
SHALL set `environment` to null in the imported state. Any other `environment` value SHALL be
imported verbatim.

The rationale is that `"ENVIRONMENT_ALL"` is not a meaningful configuration choice — it is the
server's representation of "unfiltered". Importing it would produce an immediate diff against any
configuration that omits `environment`, breaking `ImportStateVerify` and forcing practitioners to
add a redundant `environment = "ENVIRONMENT_ALL"` to their configs.

#### Explicit `environment = "ENVIRONMENT_ALL"` is preserved

When a practitioner explicitly sets `environment = "ENVIRONMENT_ALL"` in their configuration, the
prior state SHALL have a known value for `environment`. The suppression condition (prior
`environment` is null or unknown) SHALL NOT fire, so the value SHALL be preserved in state without
modification. A subsequent plan against that configuration SHALL show no changes.

#### Scenario: Server default suppressed when environment not configured

- GIVEN a panel with `type = "apm_service_map"` and `apm_service_map_config` that does not set
  `environment`
- WHEN Kibana 9.5+ returns `environment = "ENVIRONMENT_ALL"` in the API response
- THEN the provider SHALL set `environment` to null in state
- AND a subsequent plan against a configuration that omits `environment` SHALL show no changes

#### Scenario: Import with server-injected ENVIRONMENT_ALL

- GIVEN an APM service-map dashboard panel whose Terraform config does not set `environment`
- WHEN the panel is imported from Kibana 9.5+ (which returns `environment = "ENVIRONMENT_ALL"`)
- THEN the imported state SHALL have `environment = null`
- AND `ImportStateVerify` SHALL pass against a configuration that omits `environment`

#### Scenario: Explicit ENVIRONMENT_ALL is preserved

- GIVEN a panel with `apm_service_map_config.environment = "ENVIRONMENT_ALL"` explicitly set in
  the configuration
- WHEN apply runs and Kibana returns `environment = "ENVIRONMENT_ALL"` in the API response
- THEN the provider SHALL keep `environment = "ENVIRONMENT_ALL"` in state (suppression SHALL NOT
  apply)
- AND a subsequent plan SHALL show no changes

#### Scenario: Non-default environment values are preserved

- GIVEN a panel with `apm_service_map_config.environment = "production"`
- WHEN apply runs and Kibana returns `environment = "production"` in the API response
- THEN state SHALL contain `environment = "production"`
- AND a subsequent plan SHALL show no changes
