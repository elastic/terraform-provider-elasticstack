## Why

`elasticstack_kibana_data_view` can be left unmanaged when Kibana persists a create request but the provider receives an error response instead of a usable success payload, as reported in issue [#620](https://github.com/elastic/terraform-provider-elasticstack/issues/620). That produces a bad practitioner experience: the first apply fails, the resource is missing from state, and the next apply can fail again with a duplicate data view error.

## What Changes

- Add deterministic recovery for managed data view creates when the request includes an explicit `data_view.id` and Kibana returns an error after persisting the data view.
- Update create handling so the provider can reconcile state from a follow-up read instead of unconditionally failing on the create response alone.
- Add a targeted regression test that injects a post-create Kibana error while still forwarding the write to a real Kibana instance.
- Add the small acceptance-test wiring needed for one test to use a proxy Kibana endpoint without depending on global `KIBANA_ENDPOINT` overrides.
- Out of scope: heuristic recovery for create requests that do not include a stable, caller-supplied data view id.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-data-view`: extend create-time behavior so a managed data view can still converge when Kibana persists the create but returns an error response to the provider.

## Impact

- Specs: delta spec under `openspec/changes/kibana-data-view-create-error-recovery/specs/kibana-data-view/spec.md`.
- Provider behavior: `internal/kibana/dataview/create.go` and `internal/clients/kibanaoapi/data_views.go`.
- Acceptance test wiring: `internal/acctest` and `internal/kibana/dataview/acc_test.go`.
- Test infrastructure: a proxy-backed acceptance regression plus a narrower HTTP-level test for the create error path.
