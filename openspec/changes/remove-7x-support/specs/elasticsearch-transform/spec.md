## MODIFIED Requirements

### Requirement: Timeout parameter (REQ-016–REQ-017)

The `timeout` attribute SHALL accept a Go duration string and SHALL default to `"30s"`. The resource SHALL pass the parsed `timeout` value as the API operation timeout parameter to the Put Transform, Update Transform, Start Transform, and Stop Transform APIs.

#### Scenario: Timeout passed to API

- GIVEN `timeout = "60s"`
- WHEN create or update runs
- THEN the API call SHALL include a 60-second timeout parameter

### Requirement: Version-gated settings (REQ-020–REQ-032)

Settings and capabilities that require a minimum supported Elasticsearch version later than `8.0.0` SHALL be silently omitted from API calls (with a log warning) when the server version is below the minimum. The version requirements are:

- `destination.aliases`: requires Elasticsearch >= `8.8.0`
- `deduce_mappings`: requires Elasticsearch >= `8.1.0`
- `num_failure_retries`: requires Elasticsearch >= `8.4.0`
- `unattended`: requires Elasticsearch >= `8.5.0`

Transform settings and capabilities that are available throughout the supported `8.x` and later range SHALL NOT have pre-8.0 compatibility gates.

#### Scenario: Version-gated setting silently omitted

- GIVEN `deduce_mappings = true` and an Elasticsearch server version below `8.1.0`
- WHEN create or update runs
- THEN `deduce_mappings` SHALL be omitted from the API request body and a warning SHALL be logged

#### Scenario: Supported-range setting is always sent

- GIVEN `align_checkpoints = true`
- WHEN create or update runs against a supported Elasticsearch server version
- THEN `align_checkpoints` SHALL be included in the API request body

### Requirement: JSON field mapping — pivot, latest, metadata (REQ-038–REQ-040)

The `pivot` and `latest` fields SHALL be validated as JSON strings and SHALL apply JSON-normalized diff suppression. On create, the resource SHALL decode `pivot` or `latest` (whichever is set) into an `any` value for the API request. The `metadata` field SHALL be validated as a JSON string and SHALL apply JSON-normalized diff suppression. On create and update, when `metadata` is set, the resource SHALL decode it into a `map[string]any` for the API request.

#### Scenario: Invalid pivot JSON rejected

- GIVEN `pivot` contains invalid JSON
- WHEN create runs
- THEN the provider SHALL return an error and SHALL not call the Put Transform API

#### Scenario: Metadata decoded on supported versions

- GIVEN `metadata` is configured with a valid JSON object
- WHEN create or update runs against a supported Elasticsearch server version
- THEN the provider SHALL decode `metadata` into the API request body

## REMOVED Requirements

### Requirement: Minimum server version for transforms (REQ-006)

**Reason**: The provider's supported Elastic Stack floor is now `8.0.0`, so an explicit transform feature gate for Elasticsearch versions below `7.2.0` is redundant.

**Migration**: Users running Elastic Stack 7.x should upgrade to Elastic Stack 8.0 or higher before relying on supported transform resource behavior.
