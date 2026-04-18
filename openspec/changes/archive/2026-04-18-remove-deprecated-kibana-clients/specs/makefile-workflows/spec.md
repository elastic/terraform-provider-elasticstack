## REMOVED Requirements

### Requirement: SLO client generation (REQ-064–REQ-065)

**Reason:** Kibana SLO and other migrated surfaces use the shared OpenAPI client under `generated/kbapi`; maintaining a separate `generated/slo` generator and Docker-based pipeline is obsolete once all consumers are migrated.

**Migration:** Complete and merge the kbapi migration changes that remove `generated/slo` and `go-kibana-rest` imports; regenerate Kibana clients via the `gen` / `generated/kbapi` workflow documented for the provider. Use `make generate-clients` only after it has been retargeted to kbapi-only codegen per the ADDED requirement below.

## ADDED Requirements

### Requirement: Consolidated Kibana client codegen (`generate-clients`)

The `generate-clients` target SHALL run the repository’s general Kibana/OpenAPI codegen path (`gen`) and SHALL NOT invoke a separate OpenAPI generation step that writes under `generated/slo`. The Makefile SHALL NOT define a `generate-slo-client` target.

#### Scenario: generate-clients does not produce generated/slo

- **WHEN** `make generate-clients` completes successfully on a clean checkout after this change
- **THEN** the recipe SHALL NOT run OpenAPI generation dedicated to a `generated/slo` package path

#### Scenario: Deprecated SLO generator target absent

- **WHEN** a contributor inspects the root Makefile for SLO-specific client generation
- **THEN** there SHALL be no `generate-slo-client` phony target or equivalent recipe that populated `generated/slo`
