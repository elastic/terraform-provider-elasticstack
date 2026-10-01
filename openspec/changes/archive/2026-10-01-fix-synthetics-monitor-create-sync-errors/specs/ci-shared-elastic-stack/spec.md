## ADDED Requirements

### Requirement: Docker Compose stack SHALL provide a stub Synthetics Service for Kibana 8.14 and later

When `STACK_VERSION` is 8.14.0 or later, the Docker Compose stack started through the Makefile SHALL configure Kibana with a stub Synthetics Service. The stub SHALL expose exactly one Elastic-managed location, `us_west`, whose service URL is `http://127.0.0.1:1` (unreachable), so every push to the service fails. The location manifest SHALL be served from a document in a hidden Elasticsearch index seeded before Kibana starts. The stub SHALL NOT require an additional container. For earlier stack versions, Kibana SHALL start without any Synthetics Service settings. A failure to seed the manifest SHALL NOT prevent Kibana from starting.

#### Scenario: Stub enabled for supported versions

- **WHEN** a developer or CI runs `make docker-fleet` with `STACK_VERSION=9.4.0`
- **THEN** Kibana SHALL list `us_west` as an Elastic-managed location with service URL `http://127.0.0.1:1`

#### Scenario: Stub disabled for older versions

- **WHEN** `make docker-fleet` runs with `STACK_VERSION=8.13.4`
- **THEN** Kibana SHALL start without `xpack.uptime.service` settings

#### Scenario: Manifest seed failure is non-fatal

- **WHEN** seeding the manifest document fails
- **THEN** the `kibana_settings` service SHALL still complete successfully and Kibana SHALL start
