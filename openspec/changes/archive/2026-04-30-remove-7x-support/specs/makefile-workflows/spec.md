## MODIFIED Requirements

### Requirement: Fleet Server image for older stack versions (REQ-017)

When `STACK_VERSION` matches `8.0.%` or `8.1.%`, the Makefile SHALL set the Fleet agent image to **`elastic/elastic-agent` on Docker Hub** so Compose can pull an image that is not published to `docker.elastic.co` for those lines. For other versions, Compose SHALL use the default image source from the Compose files unless overridden elsewhere.

#### Scenario: Older 8.0 / 8.1 line

- GIVEN `STACK_VERSION` matches `8.0.%` or `8.1.%`
- WHEN Compose runs Fleet
- THEN `FLEET_IMAGE` SHALL resolve to Docker Hub's `elastic/elastic-agent` so pulls can succeed

#### Scenario: Unsupported 7.x line has no special fallback

- GIVEN `STACK_VERSION` matches `7.%`
- WHEN Compose runs Fleet
- THEN the Makefile SHALL NOT select Docker Hub's `elastic/elastic-agent` because of the 7.x version alone
