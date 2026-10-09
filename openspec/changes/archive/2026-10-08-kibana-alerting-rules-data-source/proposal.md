## Why

`last_execution_status` and `last_execution_date` on `elasticstack_kibana_alerting_rule` are Computed-only attributes that Kibana updates on every rule run. Every `terraform plan` therefore shows an in-place update (`-> (known after apply)`) even when no configuration changed, which breaks CI pipelines that gate on `terraform plan -detailed-exitcode`. Because the attributes are not Optional, `lifecycle { ignore_changes }` cannot suppress them. This is reported in [#4981](https://github.com/elastic/terraform-provider-elasticstack/issues/4981) and the earlier [#3342](https://github.com/elastic/terraform-provider-elasticstack/issues/3342). The maintainer direction in #3342 is to remove volatile runtime state from the resource and offer it through a data source.

## What Changes

- **BREAKING**: Remove `last_execution_status` and `last_execution_date` from the `elasticstack_kibana_alerting_rule` resource. Configurations that reference these attributes (for example in outputs or conditions) must migrate to the new data source. The schema version does not change: version-1 state is read with the attributes ignored, and the existing SDK-format (version 0) upgrade drops them.
- Add a new data source, `elasticstack_kibana_alerting_rules`, that exposes read-only operational information about alerting rules:
  - Inputs: `space_id` (defaults to the default space), and optionally either `rule_id` or a raw KQL `filter`. `rule_id` and `filter` are mutually exclusive; with neither set, all rules in the space are returned.
  - Output: a `rules` list. Each element includes the rule's `id`, `name`, `rule_type_id`, `consumer`, `enabled`, `tags`, `scheduled_task_id`, `last_execution_status` and `last_execution_date`.
  - `last_execution_date` is an RFC3339 timestamp.
  - A `rule_id` lookup returns a single-element `rules` list, and an unknown `rule_id` is an error.
  - A `filter` lookup returns every matching rule across all result pages, up to Kibana's 10,000-result paging limit, beyond which the read fails with a request to narrow the filter.
- Update the resource documentation to note the removal and point to the data source as the replacement.
- Keep `scheduled_task_id` on the resource; it is stable and is not part of this removal.

## Capabilities

### New Capabilities
- `kibana-alerting-rules-datasource`: Data source that reads alerting rules in a Kibana space, by ID or by KQL filter, and exposes their identity and operational execution fields.

### Modified Capabilities
- `kibana-alerting-rule`: Remove the requirement that maps `last_execution_status` and `last_execution_date` into state (REQ-025) and remove both attributes from the resource schema.

## Non-Goals

- No typed filter attributes (such as `tags`, `enabled`, `rule_type_ids` or a consumer filter) in this change; the raw `filter` covers these cases, and typed inputs can be added later without breaking existing configurations.
- The data source does not expose full rule configuration (`params`, `actions`, `flapping`, `artifacts`, and so on); it is limited to identity and operational fields.
- No `UseStateForUnknown` workaround on the existing attributes; the attributes are removed instead.
- No changes to other rule-related resources or data sources.

## Impact

- `internal/kibana/alertingrule` — remove the two attributes from the resource schema and model, and the corresponding state mapping; the existing SDK-format state upgrade drops the removed keys.
- A new data source package for alerting rules under `internal/kibana`, registered with the provider.
- `internal/clients/kibanaoapi` — a client wrapper for the Kibana rules find endpoint, alongside the existing single-rule get.
- `openspec/specs/kibana-alerting-rule` — delta spec; new `kibana-alerting-rules-datasource` spec.
- Generated docs and examples — resource docs updated and a new data source page generated.
- Acceptance tests that assert the removed attributes are updated; new tests cover the data source.
- Users: breaking for any configuration that reads the removed attributes, so the PR description must flag the breaking change and give migration guidance (the changelog is generated from merged PRs).
