## Why

Practitioners cannot manage **alert flapping detection** on Kibana alerting rules as code today ([terraform-provider-elasticstack#1425](https://github.com/elastic/terraform-provider-elasticstack/issues/1425)). Kibana exposes rule-level `flapping` on create/update rule APIs (from **8.16.0** onward per [elastic/kibana#190019](https://github.com/elastic/kibana/pull/190019)). The provider’s generated `kbapi` already models this field; the Terraform resource does not yet surface or map it.

## What Changes

- Add OpenSpec requirements (delta) for `elasticstack_kibana_alerting_rule`: optional **`flapping`** configuration, stack version gate (**≥ 8.16.0**), validation rules, create/update/read mapping, and acceptance-test expectations.
- **Out of scope for this proposal artifact**: editing `openspec/specs/kibana-alerting-rule/spec.md` directly; that happens when the change is synced or archived.

### Schema sketch (to merge into canonical `## Schema` on sync)

Add an optional single nested block at rule level:

```hcl
  flapping {
    enabled                  = <optional, bool>   # only supported from Elastic Stack 9.3+; see version rules below
    look_back_window         = <required-with-block, int64>
    status_change_threshold  = <required-with-block, int64>
  }
```

When the `flapping` block is **absent**, the provider does not send `flapping` on **update** (server-side rule flapping is left unchanged). When the block is **present**, **`look_back_window`** and **`status_change_threshold`** are **required** (integers). **`enabled`** alone is not a valid configuration without both integers.

### Version rules

- **Flapping block** (the two integer attributes): supported from stack **≥ 8.16.0** (unchanged from the Kibana API introduction).
- **`flapping.enabled`**: supported only from stack **≥ 9.3.0**. If `enabled` is set in configuration and the target stack is **below 9.3.0**, the provider **must** return a clear error on create/update (not silently ignore or send an unsupported field).

### Acceptance tests

- Any acceptance test step that sets **`flapping.enabled`** must be **skipped** unless the stack is **9.3.0 or newer** (for example via `SkipFunc` aligned with that minimum).
- The suite must still exercise **`flapping`** without **`enabled`** on stacks **from 8.16.0 up** so that integer-only flapping remains covered when `enabled` is not available. If the only flapping test today sets `enabled`, **add** a separate test (or steps) that omit `enabled` and remain gated at **8.16.0+**.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-alerting-rule`: Add rule-level flapping detection configuration and behavior requirements (REQ-036–REQ-040), including **`flapping.enabled` compatibility (9.3.0+)** and related acceptance-test gating.

## Impact

- **Specs**: Delta under `openspec/changes/kibana-alerting-rule-flapping/specs/kibana-alerting-rule/spec.md` until merged into canonical spec.
- **Implementation** (future): `internal/kibana/alertingrule` (schema, model, validation), `internal/models`, `internal/clients/kibanaoapi` request/response mapping, docs/descriptions, acceptance tests: **flapping integers** gated at **8.16.0+**, **`flapping.enabled`** gated at **9.3.0+**, plus integer-only coverage as described above.
