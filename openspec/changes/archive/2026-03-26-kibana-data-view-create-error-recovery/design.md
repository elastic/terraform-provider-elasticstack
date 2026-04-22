## Context

`elasticstack_kibana_data_view` currently treats create as successful only when the Kibana create response can be parsed as a normal success result. The resource `Create` path then derives state from that response and does not perform a follow-up read. When Kibana persists the data view but the provider sees an error response instead, Terraform records no state and a later apply can hit a duplicate error instead of converging.

The desired regression coverage also needs a real Kibana side effect plus a synthetic error returned to the provider. Existing acceptance tests are wired to a shared `KIBANA_ENDPOINT`, and current config loading lets that environment variable override explicit provider config, which makes per-test endpoint injection awkward in the parallel acceptance suite.

## Goals / Non-Goals

**Goals:**

- Recover deterministically when create targets a managed data view with an explicit `data_view.id` and Kibana returns an error after persisting the object.
- Keep the final Terraform state derived from a read of the created data view instead of trusting the mutating response.
- Add a reliable regression test that forwards the create request to a real Kibana instance but returns a synthetic error to the provider for that request.
- Add minimal acceptance-test wiring so this regression test can point only its provider instance at a proxy endpoint.

**Non-Goals:**

- Generic recovery for data view creates that do not include a stable caller-supplied id.
- Broad changes to provider-wide environment override behavior beyond what is needed to support this targeted acceptance test.
- Replacing existing data view CRUD semantics outside the create reconciliation path.

## Decisions

| Topic | Decision | Alternatives considered |
|--------|----------|-------------------------|
| Recovery scope | Limit recovery to creates where Terraform config supplies `data_view.id`, so the provider has a deterministic identifier to read back after the create error. | Searching by title or other mutable fields was rejected because duplicate or ambiguous matches could reconcile the wrong resource. |
| Recovery location | Perform reconciliation in `internal/kibana/dataview/create.go`, where the resource still has access to plan data such as `space_id` and `data_view.id`. | Hiding the fallback entirely inside `kibanaoapi.CreateDataView` would make it harder to use plan-specific inputs and to preserve the original create error when recovery fails. |
| Success criteria after error | If the create call returns an error or unexpected status, attempt a read of the configured id in the target space; when that read succeeds, populate state from the read result and treat create as successful. | Retrying the create request was rejected because it can amplify duplicate creation conflicts after the first write already succeeded server-side. |
| Failure behavior when recovery misses | If the follow-up read fails or returns not found, surface the original create failure and do not write state. | Silently ignoring the error would hide genuine create failures and produce confusing drift. |
| Acceptance test shape | Use an `httptest` reverse proxy that forwards all Kibana traffic to the real stack, but intercepts only the first matching data view create request and returns a synthetic `400` response after forwarding upstream. | A pure mock server was rejected because the provider needs a real Kibana API surface for the rest of the acceptance flow. |
| Test isolation | Give the regression test its own Kibana space and clean it up with `t.Cleanup`, so leaked data views from failed runs do not pollute the shared stack. | Reusing the default space would make the test flaky and leave harder-to-clean duplicates behind. |
| Per-test endpoint wiring | Add a small acceptance-test helper so this test can use an explicit Kibana endpoint without relying on process-wide endpoint mutation during a parallel acceptance run. | Using `t.Setenv("KIBANA_ENDPOINT", ...)` for the test was rejected because the suite runs with parallelism and the env override currently wins over explicit provider config. |

## Risks / Trade-offs

- [Risk] Recovery only covers managed creates with an explicit `data_view.id`. -> Mitigation: document the scope in the proposal/spec and keep error surfacing unchanged for anonymous creates.
- [Risk] The proxy-backed acceptance test could become brittle if it matches the wrong request path or intercepts more than one request. -> Mitigation: match only the create endpoint variants, fail only once, and pass all other traffic through unchanged.
- [Risk] Adjusting endpoint precedence for the targeted test helper could accidentally affect non-test provider behavior. -> Mitigation: keep the change isolated to acceptance-test wiring or an explicit helper path rather than changing unrelated runtime defaults.
- [Risk] A read immediately after the synthetic error could race with Kibana persistence timing. -> Mitigation: keep the injected failure as a post-forward response rewrite, so the upstream write has already completed before the provider starts reconciliation.

## Migration Plan

- No user-facing migration is required for existing state.
- Implementation rolls out behind normal provider behavior: managed data view creates with explicit ids become more robust, while other create failures keep their current error path.
- Rollback is a normal code revert; no schema version or state upgrade is needed.

## Open Questions

- None blocking the proposal. The exact helper API for the per-test Kibana endpoint can be finalized during implementation as long as it avoids global endpoint races in acceptance tests.
