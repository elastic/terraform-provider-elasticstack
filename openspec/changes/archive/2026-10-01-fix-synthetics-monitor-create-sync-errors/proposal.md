## Why

Creating an `elasticstack_kibana_synthetics_monitor` with Elastic-managed `locations` can fail with `unsupported monitor type: ` even though Kibana created the monitor ([#4986](https://github.com/elastic/terraform-provider-elasticstack/issues/4986)). The monitor is left in Kibana without being recorded in Terraform state, and the next apply fails with a uniqueness conflict, so a person has to delete the monitor in Kibana by hand before Terraform can recover.

Kibana's add-monitor route returns HTTP 200 with `{"message":"error pushing monitor to the service","attributes":{"errors":[...]},"id":"..."}` instead of the monitor whenever the Synthetics Service reports push errors. Those errors are left over from the previous push, so this happens intermittently. The edit-monitor route returns the same shape (without `id`) when the push for the edited monitor fails. The provider builds state from the create response, so it fails on this body. For update it throws the body away, so push errors are silently lost.

## What Changes

- Create SHALL take only the monitor identity from the add-monitor response and derive every other state value from the read that runs after the write, as update already does.
- When the add-monitor or edit-monitor response reports Synthetics Service push errors, the resource SHALL add a warning diagnostic that lists the reported locations and reasons, instead of failing or dropping them. The create warning SHALL say that the errors may come from an earlier sync and relate to other monitors.
- The local and CI acceptance-test stack SHALL provide a stub Synthetics Service for Kibana 8.14 and later: a single Elastic-managed location whose service endpoint is unreachable, so the push-error path can be exercised against a live Kibana.
- `TestAccReproduceIssue4986` SHALL exercise the failure against that live stack without an HTTP proxy, and SHALL assert that create and update succeed with correct state.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-synthetics-monitor`: REQ-001 create and update scenarios take state from the post-write read rather than from the mutation response; a new requirement covers warnings for Synthetics Service push errors.
- `ci-shared-elastic-stack`: a new requirement for the stub Synthetics Service in the Docker Compose stack used by local development and CI.

## Impact

- `internal/kibana/synthetics/monitor` (create, update, a new push-error helper, unit tests, and the acceptance test for issue 4986).
- `internal/clients/kibanaoapi/synthetics_monitor.go` (update returns push errors reported by the edit-monitor response).
- `docker-compose.yml` and `Makefile` (stub service configuration for Kibana 8.14 and later). Every stack started through `make docker-fleet`, including CI and agent workflows, picks this up.
- Users see a warning instead of an error, and no orphaned monitor, when Kibana reports push errors. No schema changes.
