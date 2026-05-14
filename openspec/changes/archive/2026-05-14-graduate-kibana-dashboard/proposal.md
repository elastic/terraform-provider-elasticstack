## Why

The `elasticstack_kibana_dashboard` resource has been registered as experimental since it was introduced, requiring practitioners to set `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` before they can use it. The resource has accumulated broad panel-type coverage, a comprehensive acceptance test suite, and a complete OpenSpec capability (`openspec/specs/kibana-dashboard/spec.md`). It is mature enough that practitioners should be able to use it from the default provider surface without an opt-in flag, and it should appear in the generated provider documentation alongside the other Kibana resources.

## What Changes

- Promote `elasticstack_kibana_dashboard` from the conditional experimental Plugin Framework resource set into the always-registered resource set in `provider/plugin_framework.go`.
- Update the `kibana-dashboard` capability so its registration requirement reflects standard (non-experimental) Plugin Framework registration.
- Regenerate provider documentation so `docs/resources/kibana_dashboard.md` is produced from the resource schema and is available to practitioners reading the registry docs.
- Add a CHANGELOG entry announcing that the dashboard resource is no longer experimental.

The Plugin Framework still registers `elasticstack_kibana_stream` as experimental; that capability is intentionally out of scope for this change.

## Capabilities

### New Capabilities

<!-- None -->

### Modified Capabilities

- `kibana-dashboard`: Add a provider registration requirement that documents the resource as part of the standard (non-experimental) Plugin Framework resource set, replacing the previous experimental opt-in registration.

## Impact

- `provider/plugin_framework.go`: `dashboard.NewResource` moves out of `experimentalResources()` into `resources()`; `experimentalResources()` continues to register `streams.NewResource` only.
- `openspec/specs/kibana-dashboard/spec.md`: gains a registration requirement section consistent with how other graduated capabilities describe their provider registration.
- `docs/resources/kibana_dashboard.md`: new generated documentation page produced by `make docs-generate`.
- `CHANGELOG.md`: practitioner-facing entry under the unreleased section noting the graduation.
- No schema, API behavior, or state changes. Practitioners who were already using the resource via `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` will continue to work; the env var is no longer required for this resource.
