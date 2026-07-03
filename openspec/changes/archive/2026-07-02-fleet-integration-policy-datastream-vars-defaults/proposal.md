## Why

Fleet 9.5 changed package-policy enrichment in two ways that break `elasticstack_fleet_integration_policy` for input-type packages:

1. **`data_stream.*` key injection into stream `vars`** — Fleet now synthesises `data_stream.type` and `data_stream.dataset` into compiled stream `vars` for input packages (driven by Kibana PRs #214216, #258143). The provider's semantic-equality logic for stream `vars` does not strip these server-managed keys before comparing, so any plan/apply cycle sees them as a user-introduced change and raises a `"Provider produced inconsistent result after apply"` error.

2. **`defaults` block going from `null` to a populated object** — Fleet now hydrates the policy with package defaults on create/GET (Kibana #119739). The `defaults` attribute is a plain `types.Object` with no semantic-equality implementation; its `null` (plan) → populated object (API) transition is seen as a raw leaf-level inequality, triggering the same inconsistency error.

Both failures were reproduced against a running 9.5.0-SNAPSHOT stack via `TestAccResourceIntegrationPolicyGCPPubSub` and `TestAccResourceIntegrationPolicySecrets` (including the `multi-valued_secrets` sub-test). On 9.4 these enrichments did not occur, so the existing machinery was sufficient. On 9.5 neither gap in the defaults machinery is reconciled.

This change closes both gaps so the resource works correctly on Elastic Stack 9.5+.

## What Changes

**Two targeted fixes to `internal/fleet/integration_policy/`:**

1. **`vars` semantic equality strips `data_stream.*` keys** — Extend `compareStreams` (and/or the underlying `jsontypes.Normalized.StringSemanticEquals` call path) so that both sides of the comparison have the server-managed `data_stream.type` and `data_stream.dataset` keys stripped before comparison. The stripping applies to stream-level `vars` only (not input-level vars, not the top-level `vars_json`). Keys are stripped from the API-returned value before equality; they are never written back to the plan or stored in state on their own.

2. **`defaults` attribute treated as computed-when-unset** — Make the `defaults` attribute `Computed: true` on the schema (which it effectively already is, since it is read-only and populated from package info). Additionally, introduce a semantic-equality mechanism (either a custom `DefaultsValue` type implementing `ObjectSemanticEquals`, or a `UseStateForUnknown`-style plan modifier) so that when the user omits `defaults` the planned value is treated as unknown/computed — allowing the API-returned value to flow through without triggering the inconsistency check. Alternatively, ensure that `InputValue.ObjectSemanticEquals` absorbs the `null` ⇄ populated-defaults transition, treating a null planned defaults that matches a populated API defaults as semantically equal.

The existing `applyDefaultsToVars` logic already handles the `data_stream.dataset` key because it appears in `defaults.vars`; it does not handle `data_stream.type` because that key is absent from `defaults.vars`. The fix is targeted at the comparison step, not at back-filling.

## Capabilities

### Modified Capabilities

- `fleet-integration-policy`: Extended to handle Fleet 9.5 enrichment of stream `vars` with server-managed `data_stream.*` keys and the new `defaults` block, eliminating post-apply inconsistency errors for input-type packages on Elastic Stack 9.5+.

## Impact

- **Modified code**: `internal/fleet/integration_policy/input_value.go` (stream vars comparison logic), `internal/fleet/integration_policy/schema.go` or `models.go` (`defaults` attribute / plan modifier).
- **New or updated unit tests**: `input_value_test.go` / `vars_json_value_test.go` for `data_stream.*` key stripping in `compareStreams`; `models_defaults_test.go` or a new file for the `defaults` null ⇄ populated-object transition.
- **Updated acceptance tests**: `TestAccResourceIntegrationPolicyGCPPubSub` and `TestAccResourceIntegrationPolicySecrets` SHALL pass without modification on 9.5.0-SNAPSHOT.
- **No schema version bump required**: the `defaults` attribute is already `Computed: true` in the schema (read-only, populated from package info). Adjusting its plan handling does not change the persisted state format.
- **Backward compatibility**: the stripping of `data_stream.*` keys is invisible on 9.4 (the keys are not present, so stripping them is a no-op). The `defaults` fix is also invisible on 9.4 (defaults remain null if not returned by Fleet). No breaking changes to configuration or state.
