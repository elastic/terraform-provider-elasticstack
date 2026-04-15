## MODIFIED Requirements

### Requirement: Provider configuration and Kibana client (REQ-003)

On create, read, update, and delete, if the provider did not supply a usable API client, the resource SHALL return an "Unconfigured Client" error diagnostic and SHALL NOT proceed to the Kibana API. The resource SHALL use the provider-configured Kibana client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana client for all operations.

For synthetics monitor HTTP operations (create, read, update, delete against Kibana’s `/api/synthetics/monitors` endpoints), the resource SHALL use the OpenAPI-generated Kibana client in `generated/kbapi`, invoked through monitor-specific helper functions in `internal/clients/kibanaoapi`, rather than the legacy `go-kibana-rest` synthetics monitor client. Ancillary operations that are not synthetics monitor HTTP calls (for example, server version checks required for `labels` compatibility) MAY continue to use `*clients.KibanaScopedClient` as they do today.

#### Scenario: Unconfigured provider

- GIVEN the resource has no provider-supplied API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an "Unconfigured Client" error diagnostic

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the provider-configured Kibana client

#### Scenario: Scoped Kibana connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the scoped Kibana client derived from that block

#### Scenario: Synthetics monitor HTTP uses OpenAPI client path

- GIVEN a fully configured provider and a synthetics monitor CRUD operation
- WHEN the provider issues the synthetics monitor HTTP request
- THEN the request SHALL be executed via `generated/kbapi` through `internal/clients/kibanaoapi` monitor helpers and SHALL NOT use `go-kibana-rest` synthetics monitor methods for that HTTP request

### Requirement: Write path — type-specific API request fields (REQ-022)

On create and update, the provider SHALL build the API request from the one configured type block. The HTTP type SHALL send `url`, `ssl` config, `max_redirects`, `mode`, `ipv4`, `ipv6`, `username`, `password`, `proxy_header`, `proxy_url`, `response`, and `check`. The TCP type SHALL send `host`, `check_send`, `check_receive`, `proxy_url`, `proxy_use_local_resolver`, and `ssl` config. The ICMP type SHALL send `host` and `wait` (as a string). The browser type SHALL send `inline_script`, `screenshots`, `synthetics_args`, `ignore_https_errors`, and `playwright_options`. If the configured type block is absent (none of the four), the provider SHALL fail with an "Unsupported monitor type config" error diagnostic.

#### Scenario: HTTP fields sent to API

- GIVEN a configuration with an `http` block
- WHEN create or update runs
- THEN the provider SHALL build an HTTP-type synthetics monitor request payload including the configured `url` and optional HTTP-specific fields that conform to Kibana’s synthetics HTTP monitor schema

### Requirement: Write path — common config fields (REQ-023)

On create and update, the provider SHALL send common monitor configuration fields: `name`, `schedule`, `locations` (as managed location strings), `private_locations` (as label strings), `enabled`, `tags`, `labels`, `alert`, `service_name` (as `apm_service_name`), `timeout`, `namespace`, `params`, and `retest_on_failure`. When `labels` is null or unknown, the provider SHALL send an empty map rather than null.

#### Scenario: Common fields in create request

- GIVEN a monitor config with `name`, `schedule`, `locations`, and `tags`
- WHEN create runs
- THEN the API request SHALL include `name`, `schedule`, `locations`, and `tags` matching the plan values

## ADDED Requirements

### Requirement: OpenAPI-backed implementation placement (REQ-024)

The implementation SHALL define synthetics monitor HTTP client operations in `internal/clients/kibanaoapi` (for example `synthetics_monitor.go`), using types and clients from `github.com/elastic/terraform-provider-elasticstack/generated/kbapi`. The package `internal/kibana/synthetics/monitor` SHALL call those helpers for CRUD and SHALL NOT import `github.com/disaster37/go-kibana-rest/v8/kbapi` for monitor request or response modeling after this migration is complete.

#### Scenario: Code ownership

- GIVEN a maintainer searches for synthetics monitor HTTP calls
- WHEN they inspect the implementation
- THEN monitor HTTP calls SHALL be routed through `internal/clients/kibanaoapi` and SHALL rely on `generated/kbapi` for wire types
