## Context

- `createMonitor` maps the add-monitor response with `toModelV0`. When Kibana returns the push-error body, `type` is absent and mapping fails after the monitor was saved server-side.
- The `entitycore` Kibana envelope already reads the resource after every create and update (`SkipReadAfterWrite` is false). `readMonitor` calls `toModelV0` on the GET response, using the written model as the prior value. So the write callback only needs to supply the identity.
- `kibanaoapi.UpdateMonitor` already ignores the PUT body and re-reads with GET.
- Kibana source (`add_monitor.ts`, `synthetics_service.ts`): `addConfigs` does not await the push and returns `this.syncErrors` from the previous push. `editConfig` awaits the push, so edit errors describe the monitor being edited.
- The generated `kbapi.SyntheticsMonitor` keeps unknown keys (`message`, `attributes`) in `AdditionalProperties`.

## Goals / Non-Goals

**Goals:**
- No orphaned monitors when the add-monitor response is not a monitor.
- Surface Synthetics Service push errors as clearly worded warnings on create and update.
- Reproduce the Kibana behaviour against a live stack in acceptance tests.

**Non-Goals:**
- Changing the resource schema or adding a `type` attribute.
- Fixing the Kibana behaviour itself (to be reported upstream).
- Retrying pushes, or failing the apply when pushes fail.
- Supporting the stub service in the TLS compose variant.

## Decisions

- **Create returns identity only.** `createMonitor` sets `id` (composite `<space>/<monitor id>`) and `space_id` on the plan and returns it. Setting `space_id` matters because the envelope derives the read space from it. A missing `id` in the response is an error, because the monitor cannot be tracked.
- **Push errors are parsed from `AdditionalProperties["attributes"].errors`.** Each entry may carry `locationId`, `error.reason` and `error.status`. All are optional (a refused connection yields only `locationId`). Malformed shapes produce no warnings rather than errors.
- **Separate wording for create and update.** Create: errors are "reported by Kibana's most recent sync" and may concern other monitors. Update: the push of this monitor failed. Both state that the monitor was saved, and that Kibana retries syncing periodically.
- **`UpdateMonitor` returns push errors alongside the monitor.** It decodes the PUT 200 body before the GET and returns `([]SyncError, *kbapi.SyntheticsMonitor, diag.Diagnostics)`. The parsing helper lives in `kibanaoapi` so both callers share it. Warning text is built in the resource package.
- **Stub Synthetics Service.**
  - Kibana is started with `--xpack.uptime.service.manifestUrl`, `username` and `password` CLI flags. The Makefile computes the flags for 8.14 and later only (the resource minimum), so older Kibana versions never see the settings. The Kibana Docker entrypoint forwards extra arguments.
  - The manifest is a document in a hidden Elasticsearch index (`tf-stub-synthetics-manifest`) seeded by the existing `kibana_settings` service before Kibana starts. Kibana fetches it via `_source` with credentials in the URL. This needs no extra container. `data:` URLs are rejected by Kibana.
  - The manifest defines `us_west` with service URL `http://127.0.0.1:1`, so every push fails. `username`/`password` make Kibana treat the service as allowed without calling `/allowed`.
- **The acceptance test detects the stub itself.** It lists `/internal/uptime/service/locations` (internal-origin header) and skips unless a location has the stub service URL. It enables the service via the internal enablement route, then creates probe monitors until Kibana returns the push-error body, so the provider's create deterministically receives it.

## Risks / Trade-offs

- [Kibana internal routes change] → the test skips when detection fails. Enablement tries `/internal/synthetics/service/enablement`, then `/internal/uptime/service/enablement`.
- [Stub affects other synthetics tests] → pushes happen only for managed locations, and existing tests use private locations only. Verified by running the synthetics suite with the stub enabled.
- [Extra index visible to wildcard index queries] → the index is created with `index.hidden: true`.
- [Credentials on Kibana's command line] → limited to the disposable test stack's existing default credentials.
