## MODIFIED Requirements

### Requirement: API and client error surfacing (REQ-002)

When the provider cannot obtain the Kibana client, create and update operations SHALL return an
error diagnostic. Transport or client errors from the Import API SHALL also be surfaced as error
diagnostics.

When the Kibana Import API returns a non-200, non-400 HTTP error with a JSON response body
conforming to the Kibana Boom format (`{"statusCode": N, "error": "...", "message": "..."}`), the
provider SHALL extract the `message` field and use it as the diagnostic detail. When the response
body is not a valid Boom envelope or `message` is empty, the provider SHALL fall back to a generic
diagnostic containing the raw response body.

#### Scenario: Missing Kibana client

- GIVEN the resource cannot obtain a Kibana client from provider configuration
- WHEN create or update runs
- THEN the operation SHALL fail with an error diagnostic

#### Scenario: Import API call failure

- GIVEN a network or server error when calling the Import API
- WHEN create or update runs
- THEN the provider SHALL surface an error diagnostic

#### Scenario: Kibana Boom error detail surfaced for non-200 non-400 responses

- GIVEN the Kibana Saved Objects Import API returns a non-200, non-400 HTTP response (e.g. HTTP 422 Unprocessable Entity)
- AND the response body is a valid Kibana Boom JSON envelope with a non-empty `message` field
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic with summary `"failed to import saved objects"` and detail equal to the `message` field value from the Boom envelope

#### Scenario: Fallback for non-Boom error bodies

- GIVEN the Kibana Saved Objects Import API returns a non-200, non-400 HTTP response
- AND the response body is not a valid Kibana Boom envelope (invalid JSON or empty `message` field)
- WHEN create or update runs
- THEN the provider SHALL return a generic error diagnostic containing the raw response body as the detail
