# kibana-alerting-rules-datasource Specification

## Purpose

Define the `elasticstack_kibana_alerting_rules` data source, which reads Kibana alerting rules in a space, either a single rule by ID or a set selected by a KQL filter, and exposes their identity and read-only operational execution fields without placing volatile runtime state on the managed resource.

## Requirements

### Requirement: Data source inputs

The `elasticstack_kibana_alerting_rules` data source SHALL accept `space_id` (Optional string), `rule_id` (Optional string) and `filter` (Optional string, a KQL expression). `rule_id` SHALL be the plain Kibana rule ID, not the composite `<space_id>/<rule_id>` form, and the space SHALL be taken from `space_id`. The data source SHALL also accept the provider's standard optional `kibana_connection` block. `rule_id` and `filter` SHALL be mutually exclusive; configuring both SHALL fail validation with an error diagnostic and SHALL NOT call the Kibana API. An empty string for `rule_id` or `filter` SHALL fail validation. Configuring neither `rule_id` nor `filter` SHALL be valid and means "all rules in the space".

When `space_id` is omitted, the data source SHALL read from the `default` space and SHALL store `default` as the `space_id` value in state, so that state, `id` and the space actually queried are consistent.

#### Scenario: Defaults to the default space and records it in state
- GIVEN `space_id` is not configured
- WHEN the data source is read
- THEN the data source SHALL read rules from the `default` space
- AND `space_id` in state SHALL be `default`

#### Scenario: Configured space is used
- GIVEN `space_id = "ops"` is configured
- WHEN the data source is read
- THEN the data source SHALL read rules from the `ops` space
- AND `space_id` in state SHALL be `ops`

#### Scenario: rule_id and filter are mutually exclusive
- GIVEN both `rule_id` and `filter` are configured
- WHEN the configuration is validated
- THEN the data source SHALL return an error diagnostic
- AND no Kibana API call SHALL be made

#### Scenario: Neither selector is valid
- GIVEN neither `rule_id` nor `filter` is configured
- WHEN the configuration is validated
- THEN validation SHALL succeed

#### Scenario: Empty string selectors are rejected
- GIVEN `rule_id = ""` or `filter = ""` is configured
- WHEN the configuration is validated
- THEN the data source SHALL return an error diagnostic

#### Scenario: Composite rule_id is not interpreted
- GIVEN `rule_id = "other-space/abc"` is configured
- WHEN the data source is read
- THEN the value SHALL be treated as a literal rule ID in the configured space
- AND, if no such rule exists, the data source SHALL return an error diagnostic

### Requirement: Rule lookup by ID

When `rule_id` is configured, the data source SHALL read that rule in the effective space and SHALL set `rules` to a list containing exactly one element for that rule. If the rule does not exist in the space, the data source SHALL return an error diagnostic and SHALL NOT return an empty list. A lookup SHALL NOT be affected by the search behaviour in the next requirement.

#### Scenario: Existing rule returns a single-element list
- GIVEN a rule with the given ID exists in the space
- WHEN the data source is read with that `rule_id`
- THEN `rules` SHALL contain exactly one element
- AND that element's `id` SHALL equal the configured `rule_id`

#### Scenario: Unknown rule ID is an error
- GIVEN no rule with the configured `rule_id` exists in the space
- WHEN the data source is read
- THEN the data source SHALL return an error diagnostic

#### Scenario: Unknown space is an error
- GIVEN the configured `space_id` does not exist
- WHEN the data source is read
- THEN the data source SHALL return an error diagnostic

### Requirement: Rule search by filter

When `rule_id` is not configured, the data source SHALL search the rules in the effective space. If `filter` is configured, the data source SHALL pass it to Kibana unmodified as the KQL filter, so the result contains only the rules Kibana returns for that filter. If `filter` is not configured, the result SHALL contain every rule in the space visible to the provider's credentials.

The data source SHALL return all matching rules across all result pages, not only the first page, provided the set of matching rules does not change while the read is in progress. Reading pages is not a snapshot: if rules are created or deleted during a read, a rule MAY be missing from or duplicated in the result. Kibana cannot page past 10,000 results; if the reported number of matching rules exceeds 10,000, the data source SHALL return an error diagnostic asking the practitioner to narrow `filter`, and SHALL NOT return a truncated list.

The `rules` list SHALL be sorted by rule `name` in ascending order, as ordered by Kibana; the relative order of rules with identical names is unspecified. A search that matches no rules SHALL succeed with an empty `rules` list and SHALL NOT be an error, including for a search with no selector in a space that has no rules.

#### Scenario: Filter narrows the result
- GIVEN a space contains rules with differing enabled states
- WHEN the data source is read with a `filter` that selects only enabled rules
- THEN `rules` SHALL contain only rules that Kibana returns for that filter

#### Scenario: No selector returns all rules in the space
- GIVEN a space contains several rules
- WHEN the data source is read with neither `rule_id` nor `filter`
- THEN `rules` SHALL contain every rule in the space

#### Scenario: Results spanning multiple pages are all returned
- GIVEN a search matches more rules than fit in a single page of results
- AND the matching rules do not change during the read
- WHEN the data source is read
- THEN `rules` SHALL contain all matching rules, not only the first page

#### Scenario: Reading is not a snapshot
- GIVEN a search spans multiple pages
- AND a matching rule is created or deleted between page requests
- WHEN the data source is read
- THEN the result MAY omit or duplicate a rule
- BUT the data source SHALL NOT return an error diagnostic solely because of the change

#### Scenario: More than 10,000 matches is an error
- GIVEN Kibana reports more than 10,000 rules matching the search
- WHEN the data source is read
- THEN the data source SHALL return an error diagnostic that tells the practitioner to narrow `filter`
- AND `rules` SHALL NOT be populated with a truncated list

#### Scenario: Results are ordered by name
- GIVEN a search matches rules named "b-rule", "a-rule" and "c-rule"
- WHEN the data source is read
- THEN `rules` SHALL list "a-rule", then "b-rule", then "c-rule"

#### Scenario: No matches is an empty list
- GIVEN no rule in the space matches the configured `filter`
- WHEN the data source is read
- THEN `rules` SHALL be an empty list
- BUT the data source SHALL NOT return an error diagnostic

#### Scenario: Search with no rules in the space is an empty list
- GIVEN a space contains no rules
- WHEN the data source is read with neither `rule_id` nor `filter`
- THEN `rules` SHALL be an empty list
- BUT the data source SHALL NOT return an error diagnostic

#### Scenario: Invalid filter surfaces the API error
- GIVEN the configured `filter` is rejected by Kibana
- WHEN the data source is read
- THEN the data source SHALL return an error diagnostic containing the API error

### Requirement: Rule element attributes

The data source SHALL expose `rules` as a Computed list of objects. Each element of `rules` SHALL expose the following Computed attributes:

- `id` — string, the rule ID
- `name` — string
- `rule_type_id` — string
- `consumer` — string
- `enabled` — bool
- `tags` — set of strings
- `scheduled_task_id` — string
- `last_execution_status` — string, the status of the rule's last execution
- `last_execution_date` — string, the time of the rule's last execution

`last_execution_date` SHALL be formatted as an RFC3339 timestamp in UTC with exactly three fractional-second digits (for example `2026-09-18T16:15:41.450Z`), so that the same instant always produces the same string. If Kibana reports no last execution status for a rule, `last_execution_status` SHALL be null. If Kibana reports no last execution time, or the time cannot be parsed, `last_execution_date` SHALL be null. If Kibana omits `scheduled_task_id`, it SHALL be null. A rule with no tags SHALL have `tags` as an empty set, not null.

#### Scenario: Execution fields populated from the API
- GIVEN a rule whose execution status is "ok" and whose last execution time is "2026-09-18T16:15:41.450Z"
- WHEN the data source is read
- THEN that rule's `last_execution_status` SHALL be "ok"
- AND its `last_execution_date` SHALL be exactly "2026-09-18T16:15:41.450Z"

#### Scenario: Non-UTC input is normalised to UTC
- GIVEN a rule whose last execution time is reported as "2026-09-18T18:15:41.450+02:00"
- WHEN the data source is read
- THEN that rule's `last_execution_date` SHALL be exactly "2026-09-18T16:15:41.450Z"

#### Scenario: Rule that has never executed
- GIVEN a rule for which Kibana reports no last execution time
- WHEN the data source is read
- THEN that rule's `last_execution_date` SHALL be null
- AND the data source SHALL NOT return an error diagnostic

#### Scenario: Absent execution status
- GIVEN a rule for which Kibana reports no execution status
- WHEN the data source is read
- THEN that rule's `last_execution_status` SHALL be null
- AND the data source SHALL NOT return an error diagnostic

#### Scenario: Unparseable execution date
- GIVEN a rule whose reported last execution time is not a valid timestamp
- WHEN the data source is read
- THEN that rule's `last_execution_date` SHALL be null
- AND the data source SHALL NOT return an error diagnostic

#### Scenario: Absent scheduled task ID
- GIVEN a rule for which Kibana omits `scheduled_task_id`
- WHEN the data source is read
- THEN that rule's `scheduled_task_id` SHALL be null

#### Scenario: Rule with no tags
- GIVEN a rule that has no tags
- WHEN the data source is read
- THEN that rule's `tags` SHALL be an empty set
- BUT `tags` SHALL NOT be null

#### Scenario: Identity and configuration fields populated
- GIVEN a rule with a name, rule type, consumer, enabled state and tags
- WHEN the data source is read
- THEN the corresponding element's `id`, `name`, `rule_type_id`, `consumer`, `enabled` and `tags` SHALL match the rule in Kibana

### Requirement: Data source identity attribute

The data source SHALL expose a Computed `id` string. When `rule_id` is configured, `id` SHALL be `<space_id>/<rule_id>`; otherwise `id` SHALL be `<space_id>`, where `space_id` is the effective space.

#### Scenario: id for a single-rule lookup
- GIVEN `space_id = "default"` and `rule_id = "abc"` are configured
- WHEN the data source is read
- THEN `id` SHALL be `default/abc`

#### Scenario: id for a search
- GIVEN `space_id = "ops"` is configured and `rule_id` is not
- WHEN the data source is read
- THEN `id` SHALL be `ops`

#### Scenario: id when space_id is omitted
- GIVEN `space_id` and `rule_id` are not configured
- WHEN the data source is read
- THEN `id` SHALL be `default`

### Requirement: Read-only and refreshed on every read

The data source SHALL NOT create, modify, enable, disable or delete any rule. Each read SHALL fetch current values from Kibana, so `last_execution_status` and `last_execution_date` reflect the state at read time.

#### Scenario: Reading does not change rules
- GIVEN an existing rule
- WHEN the data source is read
- THEN no create, update, enable, disable or delete request SHALL be sent for any rule

#### Scenario: Values reflect the latest execution
- GIVEN a rule has executed again since the previous read
- WHEN the data source is read again
- THEN `last_execution_date` SHALL reflect the more recent execution

### Requirement: API and permission errors

If the Kibana API returns an error (including authorization failures) for any request the read makes, the data source SHALL return an error diagnostic that includes the API error, and SHALL NOT return partial results, including when the failure occurs on a later page of a multi-page search. The data source SHALL NOT enforce a minimum Kibana version; if the connected Kibana does not provide the rule search endpoint, the resulting API error SHALL be surfaced as an error diagnostic.

#### Scenario: Authorization failure
- GIVEN the provider's credentials are not permitted to read rules in the space
- WHEN the data source is read
- THEN the data source SHALL return an error diagnostic
- BUT `rules` SHALL NOT be populated with partial data

#### Scenario: Failure on a later page
- GIVEN a search spans several pages
- AND a request for a page after the first fails
- WHEN the data source is read
- THEN the data source SHALL return an error diagnostic
- AND `rules` SHALL NOT contain the rules from the pages that succeeded

#### Scenario: Search endpoint unavailable
- GIVEN the connected Kibana responds that the rule search endpoint does not exist
- WHEN the data source is read with `filter` or with no selector
- THEN the data source SHALL return an error diagnostic containing the API error

### Requirement: Data source documentation

The generated documentation for `elasticstack_kibana_alerting_rules` SHALL describe the `space_id`, `rule_id` and `filter` inputs and the `rules` attributes, including that `last_execution_status`, `last_execution_date` and `scheduled_task_id` can be null. It SHALL include an example that looks up a single rule by `rule_id` and reads its `last_execution_status`, and SHALL recommend setting `filter` for spaces with many rules. It SHALL state that the data source replaces the removed `last_execution_status` and `last_execution_date` resource attributes.

#### Scenario: Documentation content
- GIVEN the generated documentation page for the data source
- WHEN a practitioner reads it
- THEN it SHALL describe the inputs and the `rules` attributes and their null behaviour
- AND it SHALL include a single-rule lookup example
- AND it SHALL recommend `filter` for large spaces
- AND it SHALL state that it replaces the removed resource attributes
