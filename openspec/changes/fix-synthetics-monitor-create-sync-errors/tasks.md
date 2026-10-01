## 1. Push error parsing

- [x] 1.1 Add `SyncError` and `syncErrorsFromBody` to `internal/clients/kibanaoapi/synthetics_monitor.go`, with table-driven unit tests (captured Kibana body, entry with reason/status, normal monitor body, malformed `attributes`)
- [x] 1.2 Add create and update warning builders in `internal/kibana/synthetics/monitor`, with unit tests for wording

## 2. Create

- [x] 2.1 `createMonitor` sets composite `id` and `space_id` from the add-monitor response and returns the plan; errors when `id` is missing
- [x] 2.2 Append the create push-error warning
- [x] 2.3 Unit tests against an `httptest` Kibana stub: push-error body, normal body, missing `id`

## 3. Update

- [x] 3.1 `kibanaoapi.UpdateMonitor` decodes the PUT 200 body and returns push errors with the monitor
- [x] 3.2 `updateMonitor` appends the update push-error warning
- [x] 3.3 Unit test for the update warning

## 4. Stub Synthetics Service

- [x] 4.1 Makefile computes Kibana stub flags for `STACK_VERSION` 8.14.0 and later
- [x] 4.2 `docker-compose.yml` passes the flags to Kibana and seeds the hidden manifest index in `kibana_settings` (non-fatal)
- [x] 4.3 Verify `make docker-fleet` on 9.4.0 and 8.14.3 lists `us_west`

## 5. Acceptance test

- [x] 5.1 Replace the proxy-based `TestAccReproduceIssue4986` with a live-stack test: detect the stub, enable the service, seed a failing push, create, update, plan-only
- [x] 5.2 Confirm it fails without the fix and passes with it

## 6. Verification

- [x] 6.1 `make build`, unit tests, `make check-openspec`
- [x] 6.2 Run the synthetics monitor acceptance suite with the stub enabled
