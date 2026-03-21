## Why

Practitioners cannot manage **alert flapping detection** on Kibana alerting rules as code today ([terraform-provider-elasticstack#1425](https://github.com/elastic/terraform-provider-elasticstack/issues/1425)). Kibana exposes rule-level `flapping` on create/update rule APIs (from **8.16.0** onward per [elastic/kibana#190019](https://github.com/elastic/kibana/pull/190019)). The provider’s generated `kbapi` already models this field; the Terraform resource does not yet surface or map it.

## What Changes

- Add OpenSpec requirements (delta) for `elasticstack_kibana_alerting_rule`: optional **`flapping`** configuration, stack version gate (**≥ 8.16.0**), validation rules, create/update/read mapping, and acceptance-test expectations.
- **Out of scope for this proposal artifact**: editing `openspec/specs/kibana-alerting-rule/spec.md` directly; that happens when the change is synced or archived.

### Schema sketch (to merge into canonical `## Schema` on sync)

Add an optional single nested block at rule level:

```hcl
  flapping {
    enabled                  = <optional, bool>   # whether the rule may enter flapping state; API default applies when unset
    look_back_window         = <required-with-block, int64>
    status_change_threshold  = <required-with-block, int64>
  }
```

When the `flapping` block is **absent**, the provider does not send `flapping` on **update** (server-side rule flapping is left unchanged). When the block is **present**, **`look_back_window`** and **`status_change_threshold`** are **required** (integers). **`enabled`** alone is not a valid configuration without both integers.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-alerting-rule`: Add rule-level flapping detection configuration and behavior requirements (REQ-036–REQ-039).

## Impact

- **Specs**: Delta under `openspec/changes/kibana-alerting-rule-flapping/specs/kibana-alerting-rule/spec.md` until merged into canonical spec.
- **Implementation** (future): `internal/kibana/alertingrule` (schema, model, validation), `internal/models`, `internal/clients/kibanaoapi` request/response mapping, docs/descriptions, acceptance tests gated at **8.16.0+**.
