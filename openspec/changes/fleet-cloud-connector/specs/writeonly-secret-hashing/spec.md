## ADDED Requirements

### Requirement: Reusable write-only secret hash helper

The provider SHALL provide a reusable utility package at `internal/utils/writeonlyhash` that resources can use to detect drift on write-only secret attributes by storing bcrypt hashes of those secrets in resource-private state. The helper SHALL expose a constructor parameterised by a resource-type-stable salt string, and methods to compute a hash, compare a value against a stored hash, and derive a stable private-state key from an attribute path string. The helper SHALL NOT log, expose, or otherwise leak secret values.

#### Scenario: Constructor binds a per-resource-type salt
- **WHEN** a value is hashed with a `Hasher` created by `writeonlyhash.New("fleet_cloud_connector")`
- **THEN** `Matches(value, hash)` SHALL return `true` on that `Hasher`
- **AND** `Matches(value, hash)` SHALL return `false` on a `Hasher` created with a different resource type string

#### Scenario: Hash comparison roundtrip
- **WHEN** a value is hashed and the resulting bytes are passed back to `Matches(value, hash)` on the same `Hasher`
- **THEN** `Matches` SHALL return `true`

#### Scenario: Hash mismatch on different value
- **WHEN** a value is hashed and `Matches(differentValue, hash)` is called
- **THEN** `Matches` SHALL return `false`

#### Scenario: Stable private-state key derivation
- **WHEN** `PrivateStateKey("aws.external_id")` is called
- **THEN** the helper SHALL return a stable key whose format is reproducible across runs of the same provider build

### Requirement: bcrypt with per-resource-type salt

The helper SHALL use bcrypt (not SHA-256 or other fast hashes) for hashing. The cost parameter SHALL default to 10 and SHALL be configurable by the caller (for example, via a `Hasher.Cost` field set before calling `Compute`). The salt SHALL be derived from the resource-type-stable identifier passed to the constructor, ensuring that the same secret value produces different hashes across different resource types — protecting against rainbow-table attacks across state files.

#### Scenario: bcrypt is the chosen algorithm
- **WHEN** the helper hashes any value
- **THEN** the resulting hash SHALL be a valid bcrypt hash recognisable by `golang.org/x/crypto/bcrypt`

#### Scenario: Salt depends on resource type
- **WHEN** the same value is hashed by `Hasher`s constructed with different resource type strings
- **THEN** the resulting hashes SHALL NOT match each other

### Requirement: No leakage of secret material

The helper SHALL NOT write secret values to logs, error messages, or any returned diagnostic. Errors raised by the helper SHALL describe the failure (e.g. "hash mismatch", "bcrypt cost out of range") without including the input value.

#### Scenario: Hash failure does not include the input
- **WHEN** the helper is asked to hash with an out-of-range bcrypt cost and returns an error
- **THEN** the error message SHALL NOT contain the input value
