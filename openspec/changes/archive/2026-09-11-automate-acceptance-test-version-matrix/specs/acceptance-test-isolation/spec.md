## REMOVED Requirements

### Requirement: Elastic Stack 9.4.2 included in acceptance-test CI matrix (REQ-ACC-002)

The `version` list in `.github/workflows/provider.yml` SHALL include `"9.4.2"` immediately before `"9.5.0-SNAPSHOT"` and SHALL NOT include `"9.4.0"`, so the matrix tests only one 9.4.x release.

#### Scenario: Entity-store tests pass in CI against 9.4.2

- GIVEN `"9.4.2"` added to the CI matrix and per-test space isolation applied to all five packages
- WHEN the acceptance-test jobs run for the `9.4.2` matrix entry
- THEN all entity-store family tests pass
- AND the issue #3952 closure gate ("tests run successfully against that version") is satisfied

#### Scenario: 9.4.2 replaces 9.4.0

- GIVEN the matrix includes `"9.4.0"` followed by `"9.4.2"` and `"9.5.0-SNAPSHOT"`
- WHEN the redundant 9.4.x matrix entry is removed and CI runs the full matrix
- THEN `"9.4.2"` remains immediately before `"9.5.0-SNAPSHOT"`
- AND `"9.4.0"` is no longer included
- AND all other version entries continue to behave as before

**Reason:** The exact patch strings `"9.4.2"` and `"9.5.0-SNAPSHOT"` are a one-off matrix snapshot. Automated patch bumps and SNAPSHOT promotion would violate this requirement on the next generator run. Replaced by a living "exactly one latest 9.4.x GA" invariant against the pinned versions artifact.

## ADDED Requirements

### Requirement: Exactly one 9.4.x GA entry in the acceptance-test CI matrix (REQ-ACC-002)

The stack-version list pinned at `.github/versions/acceptance-test-matrix.json` (consumed by `.github/workflows/provider.yml`'s acceptance-test matrix) SHALL include exactly one `9.4.x` GA entry — the latest pullable `9.4` patch — and SHALL NOT include more than one `9.4.x` patch at a time. The latest pullable patch is the latest published `9.4` patch whose Elasticsearch, Kibana, and applicable Agent image manifests all resolve. When those manifests are unavailable, the generator's image-probe fallback retains the previously pinned `9.4` patch without failing the run; that retained pin still satisfies this requirement. This requirement is a version-range invariant, not a pin to a specific patch string, so an automated `9.4` patch bump satisfies it without a spec change.

#### Scenario: Pinned list has exactly one latest 9.4.x GA entry

- **GIVEN** the pinned versions artifact after a successful version-matrix computation
- **WHEN** the `9.4.x` entries are inspected
- **THEN** there SHALL be exactly one `9.4.x` GA version
- **AND** that version SHALL be the latest pullable `9.4` patch from the generator's GA catalog (the latest published patch whose compose-stack manifests resolve, or the previously pinned `9.4` patch when image-probe fallback retains it)

#### Scenario: Automated patch bump keeps exactly one 9.4.x entry

- **GIVEN** the pinned versions artifact contains a `9.4.x` GA entry
- **WHEN** the version-matrix generator (`ci-version-matrix-generation` capability) updates that entry to a newer published patch of the same minor
- **THEN** the pinned versions artifact SHALL still contain exactly one `9.4.x` entry
- **AND** all other version entries SHALL continue to behave as before

#### Scenario: Two 9.4.x patches cannot both be present

- **GIVEN** the pinned versions artifact is inspected at any point in time
- **WHEN** more than one `9.4.x` version string is present
- **THEN** this requirement is violated, regardless of which specific patch numbers are involved
