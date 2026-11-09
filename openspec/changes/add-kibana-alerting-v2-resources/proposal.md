# Proposal

## Why

Kibana Alerting v2 (the `alerting_v2` plugin) ships as GA in Kibana 9.6.0 with a public HTTP surface under `/api/alerting/v2/`. It introduces two new objects: **rules**, which are ES|QL-first and evaluated on a schedule, and **action policies**, which route alerts to notification destinations and are configured separately from the rules that produce them. The provider today manages only classic alerting rules through `elasticstack_kibana_alerting_rule`. Classic and v2 share no API, no saved object type, and no identifier space, so users adopting v2 have no way to declare rules or action policies as code, review them, or detect drift. Landing Terraform support alongside the 9.6.0 GA release means v2 users get infrastructure-as-code from day one instead of after the fact.

Program tracking: [elastic/rna-program#655](https://github.com/elastic/rna-program/issues/655) (epic), with implementation issues [#676](https://github.com/elastic/rna-program/issues/676) (rule) and [#677](https://github.com/elastic/rna-program/issues/677) (action policy). API reference: the **Alerting V2** group of the [Kibana API documentation](https://www.elastic.co/docs/api/doc/kibana/).

## What Changes

- Add a new Terraform resource **`elasticstack_kibana_alerting_v2_rule`** that manages a single Alerting v2 rule: create, read, update, delete, and import, scoped to a Kibana space.
- Add a new Terraform resource **`elasticstack_kibana_alerting_v2_action_policy`** that manages a single Alerting v2 action policy with the same lifecycle.
- Register both as **standard provider resources** (not behind the experimental opt-in), gated to Elastic Stack **9.6.0 or later** for rules and **9.7.0 or later** for action policies (the first release whose action policy create-or-replace request accepts the enabled state), with generated registry documentation and examples. The resource type names keep the `alerting` qualifier and the `v2` marker, mirroring the API path, per the Alerting team's naming preference.
- Deliver in two steps: the **rule resource is implemented first**. The **action policy resource follows** once Kibana accepts the enabled state in the action policy create-or-replace request ([elastic/rna-program#1266](https://github.com/elastic/rna-program/issues/1266), targeted at 9.7.0), so that the action policy's enabled state is fully managed by Terraform from the first release of that resource. The design and specs in this change cover both resources so the second implementation needs no new proposal.

**Not in scope** (explicit non-goals):

- Read-only lookups of rules or action policies that already exist in Kibana but are not managed by Terraform (Terraform "data sources"). This change only adds the managed resources that create, update, and delete these objects.
- Bulk endpoints, by-query endpoints, the manual `_run` endpoint, execution history, rule change history, rule templates, and tag or routing-tag lookups.
- Alert actions (acknowledge, assign, activate, deactivate, tag) and alert or series snooze state.
- Snooze state on action policies: it is runtime operational state, not configuration, and is not managed by Terraform.
- Any change to the classic `elasticstack_kibana_alerting_rule` resource, and any migration path from classic rules to v2 rules.
- Hand-made changes to the generated Kibana API client. The pinned OpenAPI bundle already contains every Alerting v2 operation. The action policy create-or-replace body still lacks `enabled`, and that arrives through the regular Renovate `kibana-openapi-spec` bump once Kibana ships it. The action policy resource waits for that bump.

## Capabilities

### New Capabilities

- `kibana-alerting-v2-rule`: Terraform resource `elasticstack_kibana_alerting_v2_rule` for Kibana Alerting v2 rules: typed schema for the rule configuration, CRUD and import lifecycle, space-aware identity, server-owned field handling, plan-time validation mirroring the API's constraints, version gating, and acceptance-test expectations.
- `kibana-alerting-v2-action-policy`: Terraform resource `elasticstack_kibana_alerting_v2_action_policy` for Kibana Alerting v2 action policies: typed schema for destinations, matcher, grouping, and throttling, CRUD and import lifecycle, space-aware identity, enabled-state handling, server-owned field handling, plan-time validation, version gating, and acceptance-test expectations.

### Modified Capabilities

- _(none)_ — the classic `kibana-alerting-rule` capability is untouched.

## Impact

- **Provider surface**: two new resource types registered in the provider's standard resource list and two new generated documentation pages under `docs/resources/` with examples under `examples/resources/`.
- **Code**: two new resource packages under `internal/kibana/` and new Kibana API client wrappers under `internal/clients/kibanaoapi/`, following the existing Plugin Framework resource conventions.
- **Generated client**: no regeneration required. The `generated/kbapi` already exposes the Alerting v2 operations and named schema components.
- **Testing**: acceptance tests require an Elastic Stack 9.6.0+ environment for rules and 9.7.0+ for action policies. Tests are skipped on older stacks.
- **Upstream dependency**: the action policy resource's implementation (not its design) waits on Kibana accepting `enabled` in `PUT /api/alerting/v2/action_policies/{id}`.
- **Existing users**: no impact on existing resources, state, or behaviour.
