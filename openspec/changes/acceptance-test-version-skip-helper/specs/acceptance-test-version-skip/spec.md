## ADDED Requirements

### Requirement: Skip tests when Elasticsearch version is below a minimum
The `versionutils.SkipIfUnsupported` helper SHALL skip the current test when the connected Elasticsearch server version is strictly less than the supplied minimum version.

#### Scenario: Stateful cluster below minimum version
- **WHEN** `SkipIfUnsupported(t, v8_11_0, FlavorAny)` is called
- **AND** the connected cluster reports `build_flavor` of `"default"` and version `8.10.0`
- **THEN** the helper calls `t.Skip` with a message indicating the version mismatch

#### Scenario: Stateful cluster at or above minimum version
- **WHEN** `SkipIfUnsupported(t, v8_11_0, FlavorAny)` is called
- **AND** the connected cluster reports `build_flavor` of `"default"` and version `8.11.0`
- **THEN** the helper returns without calling `t.Skip`

#### Scenario: Serverless cluster always passes version check
- **WHEN** `SkipIfUnsupported(t, v8_11_0, FlavorAny)` is called
- **AND** the connected cluster reports `build_flavor` of `"serverless"`
- **THEN** the helper returns without calling `t.Skip` regardless of the reported version number

### Requirement: Skip tests when Elasticsearch version does not meet constraints
The `versionutils.SkipIfUnsupportedConstraints` helper SHALL skip the current test when the connected Elasticsearch server version does not satisfy the supplied `version.Constraints`.

#### Scenario: Version outside constraint range
- **WHEN** `SkipIfUnsupportedConstraints(t, ">=8.9.0,!=8.11.0", FlavorAny)` is called
- **AND** the connected cluster reports version `8.11.0`
- **THEN** the helper calls `t.Skip` with a message indicating the constraint mismatch

#### Scenario: Version inside constraint range
- **WHEN** `SkipIfUnsupportedConstraints(t, ">=8.9.0,!=8.11.0", FlavorAny)` is called
- **AND** the connected cluster reports version `8.12.0`
- **THEN** the helper returns without calling `t.Skip`

### Requirement: Skip tests based on deployment flavor
Both `SkipIfUnsupported` and `SkipIfUnsupportedConstraints` SHALL evaluate the `Flavor` parameter and skip the test when the connected cluster's flavor does not match the requirement.

#### Scenario: Stateful-only test on serverless cluster
- **WHEN** `SkipIfUnsupported(t, v8_11_0, FlavorStateful)` is called
- **AND** the connected cluster reports `build_flavor` of `"serverless"`
- **THEN** the helper calls `t.Skip` with a message indicating serverless is not supported

#### Scenario: Serverless-only test on stateful cluster
- **WHEN** `SkipIfUnsupported(t, nil, FlavorServerless)` is called
- **AND** the connected cluster reports `build_flavor` of `"default"`
- **THEN** the helper calls `t.Skip` with a message indicating stateful is not supported

#### Scenario: FlavorAny passes regardless of flavor
- **WHEN** any helper is called with `FlavorAny`
- **THEN** the flavor check is bypassed and only the version check (if any) is evaluated

### Requirement: Fail fast on infrastructure errors
Both helpers SHALL call `t.Fatal` when the version or flavor check itself cannot be completed (e.g. cannot create acceptance-testing client, cannot parse server version, cannot reach Elasticsearch).

#### Scenario: Elasticsearch client cannot be created
- **WHEN** `SkipIfUnsupported(t, v8_11_0, FlavorAny)` is called
- **AND** `clients.NewAcceptanceTestingElasticsearchScopedClient()` returns an error
- **THEN** the helper calls `t.Fatal` with the error message

#### Scenario: Server version cannot be parsed
- **WHEN** `SkipIfUnsupported(t, v8_11_0, FlavorAny)` is called
- **AND** the server info response contains an unparseable version string
- **THEN** the helper calls `t.Fatal` with a message describing the parsing failure
