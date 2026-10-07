## MODIFIED Requirements

### Requirement: status_json computed field (REQ-007)

The resource SHALL expose a `status_json` Computed `string` attribute containing the normalized
JSON representation of the most recent `GET /api/security/entity_store/status` response, for use
in `output` blocks or external tooling.

The value SHALL be refreshed on every Read.

Before the response body is stored as `status_json`, the provider SHALL normalize the order of the
`engines` array within the response by stable-sorting its elements by each element's `type` field.
This ensures that two reads returning the same logical engine set, but with `engines` in a
different order, produce byte-identical (and therefore semantically-equal) `status_json` values.

#### Scenario: status_json reflects current status on read

- GIVEN an installed Entity Store resource in state
- WHEN Terraform refreshes the resource
- THEN the provider SHALL call `GET /api/security/entity_store/status`
- AND `status_json` in state SHALL contain the normalized JSON of the full response body
- AND the value SHALL differ from a previous read if the logical response content changed, excluding engine-array order and insignificant JSON formatting

#### Scenario: status_json is stable across reads despite engine reordering

- GIVEN an installed Entity Store with engines for entity types `generic` and `user`, both in
  status `running`
- AND an apply-time `GET /api/security/entity_store/status` response returns `engines` in the order
  `[generic, user]`
- AND a subsequent `GET /api/security/entity_store/status` response (e.g. during
  `terraform import` or a later refresh) returns the same two engines but in the order
  `[user, generic]`
- WHEN the provider builds `status_json` from each response
- THEN both resulting `status_json` values SHALL be equal
- AND `terraform plan`/`ImportStateVerify` SHALL NOT report a `status_json` difference solely due to
  the engine order change
