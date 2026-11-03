## ADDED Requirements

### Requirement: Kibana connections accept custom request headers
The provider SHALL expose an optional sensitive `headers` map of string values on both the provider-level `kibana` block and the entity-local `kibana_connection` block. The attribute SHALL follow the existing Elasticsearch connection header contract: each header name maps to one string value, and no dedicated environment-variable override is provided.

#### Scenario: Provider-level Kibana headers are configured
GIVEN a provider `kibana` block with one or more `headers` entries
WHEN Terraform evaluates the provider configuration
THEN the provider SHALL accept the entries as a sensitive map of strings

#### Scenario: Entity-local Kibana headers are configured
GIVEN an entity-local `kibana_connection` block with one or more `headers` entries
WHEN Terraform evaluates the entity configuration
THEN the provider SHALL accept the entries using the same schema contract as provider-level Kibana headers

#### Scenario: Repeated header values are not part of the contract
GIVEN a user configures Kibana custom request headers
WHEN the header configuration is evaluated
THEN each header name SHALL correspond to exactly one string value

### Requirement: Effective Kibana headers are sent on Kibana-derived requests
The provider SHALL send every custom header from the effective Kibana connection on each request made by its Kibana-derived clients. This SHALL include Kibana OpenAPI requests and Fleet API requests that use provider-level or entity-local Kibana connection configuration. Header names and values SHALL have surrounding whitespace removed, matching Elasticsearch connection header behavior.

#### Scenario: Provider-level headers reach Kibana requests
GIVEN provider-level Kibana configuration containing a custom request header
WHEN a Kibana API request uses the provider-configured client
THEN the request SHALL contain the configured header and value

#### Scenario: Entity-local headers reach scoped requests
GIVEN an entity-local `kibana_connection` containing a custom request header
WHEN the entity makes a Kibana API request through its scoped client
THEN the request SHALL contain the entity-local header and value

#### Scenario: Headers reach Fleet requests derived from Kibana configuration
GIVEN an effective Kibana connection containing a custom request header
WHEN a Fleet API request uses the Fleet client derived from that Kibana connection
THEN the request SHALL contain the configured header and value

#### Scenario: Header whitespace matches Elasticsearch behavior
GIVEN a Kibana custom header whose name or value has surrounding whitespace
WHEN the effective client configuration is built
THEN the transmitted header name and value SHALL have surrounding whitespace removed

#### Scenario: Omitted headers preserve existing request behavior
GIVEN an effective Kibana connection with no custom headers
WHEN a Kibana-derived client sends a request
THEN the request SHALL retain the provider's existing request headers and authentication behavior

### Requirement: Native Kibana authentication takes precedence
The provider SHALL give native Kibana authentication precedence over a custom `Authorization` header. A custom `Authorization` header SHALL be sent when no native Kibana authentication is configured. When username and password, API key, or bearer token authentication is configured, the resulting native `Authorization` value SHALL replace a custom `Authorization` value.

#### Scenario: Custom Authorization is used without native authentication
GIVEN an effective Kibana connection with a custom `Authorization` header
AND no username and password, API key, or bearer token is configured
WHEN a Kibana-derived client sends a request
THEN the request SHALL contain the custom `Authorization` value

#### Scenario: Native authentication replaces custom Authorization
GIVEN an effective Kibana connection with a custom `Authorization` header
AND native Kibana authentication is configured
WHEN a Kibana-derived client sends a request
THEN the request SHALL contain the native authentication value
AND SHALL NOT contain the custom `Authorization` value
