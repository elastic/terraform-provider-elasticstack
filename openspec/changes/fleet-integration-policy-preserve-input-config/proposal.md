## Why

`elasticstack_fleet_integration_policy` updates package policies through the Fleet "simplified" (mapped-inputs) API. That request shape only carries `enabled`, `vars`, `streams` and `condition` for each input; Kibana rebuilds every input from the package template plus those values and discards whatever was stored in the input-level (and stream-level) `config` object.

`config` is not user configuration managed by this resource. It is written by Kibana itself or by other Kibana features, and losing it silently breaks those features:

- **Elastic APM**: when the APM integration is created, Kibana's APM plugin stores two API keys in `inputs[].config["apm-server"].value` (`rum.source_mapping.elasticsearch.api_key` and `agent.config.elasticsearch.api_key`). After any Terraform update of the policy the keys are gone, APM Server falls back to the Fleet output API key (which cannot read `.apm-source-map` or `.apm-agent-configuration`), and RUM source mapping and central agent configuration stop working with a `403 security_exception`. Nothing in Terraform or in the plan indicates the change.
- **Osquery Manager**: packs assigned to an agent policy are stored in `inputs[].config.osquery.value.packs`; every update removes all packs from the agent policy (elastic/terraform-provider-elasticstack#5480).

Because the attribute is not modeled, users cannot work around the problem from configuration: no value of `vars` can reproduce data that lives in `config`.

## What Changes

- **Preserve** input-level and stream-level `config` across updates. Before sending the simplified update, the resource reads the current policy in the typed (non-simplified) format and records each input's and stream's `config`. After the simplified update, it re-reads the policy and, if any previously stored `config` value is now missing, writes it back with a single typed update that carries only the `version` token and the `inputs` array.
- **Merge, do not overwrite.** Restoration is a deep fill: keys that the update (or a Kibana callback that ran during it) wrote are kept as-is; only keys that were present before and are absent after are restored. Lists and scalar values are treated as leaves.
- **No extra writes** when nothing was lost: policies with no `config`, or whose `config` survived the update, are updated exactly as before (one additional read only).
- **Document** the behaviour on the resource page.

Not breaking: no schema change, no state change, and policies without input `config` behave exactly as before.

Out of scope: modeling `config` as a Terraform attribute (it is owned by Kibana, not by the user), and `elasticstack_fleet_elastic_defend_integration_policy`, which already uses the typed API and manages its `config` explicitly.

## Capabilities

### New Capabilities

_None — this changes the update behaviour of an existing resource._

### Modified Capabilities

- `fleet-integration-policy`: adds a requirement that updates preserve input-level and stream-level `config` written outside Terraform.

## Impact

- `internal/clients/fleet/package_policy.go`: generic typed-format `GetPackagePolicyTyped` / `UpdatePackagePolicyTyped` helpers; the existing Defend helpers delegate to them.
- `internal/fleet/integration_policy/`: new `input_config.go` (snapshot, deep-fill merge, typed request construction) and wiring in `update.go`.
- Tests: unit tests for snapshot/merge/request construction; an acceptance test that injects input `config` out of band, applies a Terraform update, and asserts the `config` survived.
- Docs: `docs/resources/fleet_integration_policy.md` regenerated via `make docs-generate`.
