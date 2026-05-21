## MODIFIED Requirements

### Requirement: Read — preserve configured start and end (REQ-017)

On read, the resource SHALL NOT overwrite the `start` or `end` attribute with values returned by the Get Datafeed Stats API. The `start` and `end` attributes SHALL round-trip from configuration (or from prior state when no configuration value is supplied) so that practitioner-declared values are preserved verbatim — even when Elasticsearch reports a different effective search interval (e.g. after bucket alignment or first-document snap-forward). Read SHALL still call the Get Datafeed Stats API to populate `state` and the computed `effective_search_start` / `effective_search_end` attributes (see new requirement below).

#### Scenario: Explicit start is preserved across apply

- GIVEN a `started` datafeed configured with `start = "2022-01-01T00:07:30Z"`
- AND Elasticsearch reports `running_state.search_interval.start_ms = "2022-01-01T00:10:00Z"` after the datafeed begins searching
- WHEN create or update runs
- THEN the `start` attribute in state SHALL equal `"2022-01-01T00:07:30Z"` (the configured value)
- AND the apply SHALL NOT produce a "Provider produced inconsistent result after apply" diagnostic

#### Scenario: Explicit start is preserved across bucket alignment

- GIVEN a `started` datafeed for a job with `bucket_span = "15m"` configured with `start = "2025-07-13T02:23:23.935Z"`
- AND Elasticsearch reports `running_state.search_interval.start_ms = "2025-07-13T02:26:42.000Z"` after bucket alignment
- WHEN create runs
- THEN the `start` attribute in state SHALL equal `"2025-07-13T02:23:23.935Z"`
- AND the apply SHALL succeed

#### Scenario: Explicit end is preserved across apply

- GIVEN a `started` datafeed configured with `end = "2024-12-31T23:59:59Z"`
- AND Elasticsearch reports a different `running_state.search_interval.end_ms`
- WHEN create or update runs
- THEN the `end` attribute in state SHALL equal `"2024-12-31T23:59:59Z"` (the configured value)

#### Scenario: Omitted start remains null on read

- GIVEN a `started` datafeed where `start` is not set in configuration
- WHEN read runs
- THEN the `start` attribute in state SHALL be null
- AND the `effective_search_start` attribute SHALL be populated from `running_state.search_interval.start_ms`

## ADDED Requirements

### Requirement: Computed effective search interval attributes (REQ-022)

The resource SHALL expose two computed read-only attributes that report Elasticsearch's view of the active search interval:

- `effective_search_start` (RFC 3339 datetime): the value of `running_state.search_interval.start_ms` from the Get Datafeed Stats API.
- `effective_search_end` (RFC 3339 datetime): the value of `running_state.search_interval.end_ms` from the Get Datafeed Stats API.

Both attributes SHALL be `Computed` only (never settable by configuration). On read, when the datafeed is in the `started` state and `running_state.search_interval` is present, the resource SHALL populate the attributes from the corresponding `start_ms` and `end_ms` values, preserving the timezone of any previously-configured `start` / `end` values for display. When `running_state.real_time_configured` is `true`, the resource SHALL set `effective_search_end` to null. When the datafeed is in the `stopped` state, or `running_state` / `search_interval` is absent (including the "datafeed started and stopped too quickly" path covered by REQ-014), the resource SHALL set both attributes to null.

#### Scenario: Effective search interval populated for a started datafeed

- GIVEN a `started` datafeed whose `running_state.search_interval` reports `start_ms = "2022-01-01T00:10:00Z"` and `end_ms = "2022-01-01T01:00:00Z"`
- WHEN read runs
- THEN `effective_search_start` SHALL equal `"2022-01-01T00:10:00Z"`
- AND `effective_search_end` SHALL equal `"2022-01-01T01:00:00Z"`

#### Scenario: Effective end is null when real-time

- GIVEN a `started` datafeed where `running_state.real_time_configured = true`
- WHEN read runs
- THEN `effective_search_end` SHALL be null in state

#### Scenario: Effective attributes are null for a stopped datafeed

- GIVEN a datafeed in the `stopped` state
- WHEN read runs
- THEN both `effective_search_start` and `effective_search_end` SHALL be null in state

#### Scenario: Effective attributes are null when running_state is missing

- GIVEN a `started` datafeed whose `running_state` is absent or whose `search_interval` is absent
- WHEN read runs
- THEN both `effective_search_start` and `effective_search_end` SHALL be null in state

#### Scenario: Effective attributes round-trip without drift

- GIVEN a `started` datafeed with `effective_search_start` and `effective_search_end` already in state
- WHEN a subsequent `terraform plan` runs against an unchanged configuration
- THEN no plan diff SHALL be produced for either attribute

## REMOVED Requirements

### Requirement: Plan modifier — start becomes unknown when state changes (REQ-018)

**Reason**: With REQ-017 modified so that `start` is no longer overwritten by reads, and the `start` schema attribute no longer being `Computed`, this plan modifier is unnecessary. The `start` attribute is now a pure user input: it round-trips from configuration to state without any provider-side rewriting, so there is never a need to mark it unknown to "make room" for an API-supplied value. Elasticsearch's effective search start is now surfaced separately via the new computed attribute `effective_search_start` (REQ-022).

**Migration**:

- Practitioners who omit `start`: no action required. State will hold `start = null`; Elasticsearch's effective start is observable via the new `effective_search_start` attribute.
- Practitioners who set `start` explicitly (the case that previously errored with "Provider produced inconsistent result after apply"): no action required. The configured value is now preserved verbatim in state, and the new `effective_search_start` attribute exposes the ES-reported value for observability.
- Existing state files where `start` was previously set to the ES-reported value (because the bug was masked by `state` changing on a previous apply) will show a one-time plan diff on the next apply, reverting `start` to the configured (or null) value. The apply is a no-op against Elasticsearch because the datafeed `state` does not change.
- Code: delete `internal/elasticsearch/ml/datafeed_state/set_unknown_if_state_has_changes.go` and its corresponding plan modifier wiring in the schema.
