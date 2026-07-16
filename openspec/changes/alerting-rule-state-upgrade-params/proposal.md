## Why

`elasticstack_kibana_alerting_rule` contains two `jsontypes.Normalized` fields —
`params` at the rule level and `params` inside each `actions[]` entry — that are not
handled by the resource's v0 → v1 state upgrader
(`internal/kibana/alertingrule/state_upgrade.go`).

When the old SDKv2 provider persisted an empty string (`""`) for either of those
fields (e.g. an action connector type that carries no configured params), upgrading
the provider to any Plugin Framework build (≥ 0.14.0) and then running
`terraform plan` produces an `Invalid JSON String Value` error because the Plugin
Framework's `jsontypes.Normalized` type rejects `""` as invalid JSON.

The same class of bug was previously fixed for the ILM resource in #3914 and for
index/component templates. This change brings `elasticstack_kibana_alerting_rule`
into parity.

## What Changes

- In `migrateV0ToV1` (`internal/kibana/alertingrule/state_upgrade.go`):
  - Call `stateutil.NullifyEmptyString(stateMap, "params")` to normalize the
    rule-level `params` field.
  - Inside the existing `actions` loop, call
    `stateutil.NullifyEmptyString(action, "params")` for each action entry's
    `params` field.
- Add unit tests in `internal/kibana/alertingrule/` that exercise the upgrader with
  empty-string `params` at both the rule level and the action level.

No schema changes, no API changes, and no changes to the canonical spec are
required — this is a purely internal state-migration correctness fix.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-alerting-rule`: Fix state upgrader (v0 → v1) to nullify empty-string
  `params` at the rule level and within each action entry, preventing
  `Invalid JSON String Value` errors when upgrading from the SDKv2 provider to the
  Plugin Framework provider.

## Impact

- **Specs**: Delta under
  `openspec/changes/alerting-rule-state-upgrade-params/specs/kibana-alerting-rule/spec.md`
  capturing the new state-upgrade invariant.
- **Implementation**: `internal/kibana/alertingrule/state_upgrade.go` (two
  `NullifyEmptyString` additions); unit test file in the same package.
