## MODIFIED Requirements

### Requirement: Compatibility — minimum Elasticsearch version (REQ-011)

On create and update, the resource SHALL enforce a minimum Elasticsearch server version of 8.2.0. If the server version is less than 8.2.0, the resource SHALL return an "Unsupported server version" error diagnostic (summary, emitted by the entitycore envelope) whose detail explains that this resource requires Elasticsearch 8.2.0 or later, and SHALL NOT proceed to call the Put Component Template API.

This requirement SHALL be satisfied by implementing `WithVersionRequirements` on the resource model (`GetVersionRequirements()` returning a single requirement for ES ≥ 8.2.0), delegating version enforcement to the entitycore envelope. The resource SHALL NOT call `client.ServerVersion()` directly in Create or Update.

The resource SHALL use the `WriteFunc[T]` callback contract via the entitycore envelope for both Create and Update. The resource SHALL NOT override the envelope's `Create` or `Update` method receivers.

#### Scenario: Version below minimum

- **WHEN** create or update runs against an Elasticsearch server with version below 8.2.0
- **THEN** the provider SHALL return an "Unsupported server version" error diagnostic (from the entitycore envelope) and SHALL NOT call the Put Component Template API

#### Scenario: Version at minimum

- **WHEN** create or update runs against an Elasticsearch server with version 8.2.0 or later
- **THEN** the provider SHALL proceed normally

#### Scenario: Version enforced by envelope, not by write callback

- **WHEN** the resource model implements `WithVersionRequirements` returning ES ≥ 8.2.0
- **THEN** the entitycore envelope SHALL enforce the version requirement during Create and Update before invoking the write callback
- **AND** the write callback SHALL NOT contain any explicit server version check
