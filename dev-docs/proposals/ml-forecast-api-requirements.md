# Proposal: Elasticsearch ML Forecast API Enhancements for Terraform Support

## Summary

The current Elasticsearch ML Forecast API is designed for imperative, pipeline-style workflows (create → poll → consume results → delete). It lacks the read-back and idempotency semantics required by declarative infrastructure-as-code tools like Terraform. This document outlines the minimum API changes needed to make forecasts manageable as a Terraform resource.

## Background

### Current API surface

| Operation | Endpoint | Notes |
|-----------|----------|-------|
| Create | `POST /_ml/anomaly_detectors/{job_id}/_forecast` | Returns server-generated `forecast_id`. Not idempotent — every call creates a new forecast. |
| Delete | `DELETE /_ml/anomaly_detectors/{job_id}/_forecast/{forecast_id}` | Deletes forecast results. Succeeds if forecast already expired. |
| Read results | `GET /_ml/anomaly_detectors/{job_id}/results/forecasts/{forecast_id}` | Returns *result data* (predicted values, bounds), not forecast *configuration*. |
| Job stats | `GET /_ml/anomaly_detectors/{job_id}/_stats` | Returns aggregate forecast counts/memory, not individual forecast details. |

### Gaps for Terraform

Terraform resources follow a declarative lifecycle: **Create → Read → Update → Delete**, where Read is called on every plan/apply to detect drift between desired and actual state. The current forecast API has the following gaps:

1. **No configuration read endpoint** — There is no API to retrieve a forecast's input parameters (`duration`, `expires_in`, `max_model_memory`) or its lifecycle status by `forecast_id`. Terraform cannot verify that a forecast exists or compare its configuration against the desired state.

2. **No idempotent create** — Every POST creates a new forecast with a new server-generated ID. If `terraform apply` runs twice with no config changes, two forecasts are created. Terraform requires that applying the same configuration repeatedly produces the same result.

3. **Silent expiry** — Forecasts auto-delete when `expires_in` elapses. Terraform state becomes stale with no way to detect this, leading to phantom resources in state that no longer exist on the server.

4. **No update path** — Forecast parameters are immutable after creation. This is acceptable for Terraform (all attributes can use `RequiresReplace`), but combined with the other gaps it compounds the problem.

## Proposed Changes

### 1. GET forecast by ID (Required)

Add an endpoint to retrieve a single forecast's configuration and status:

```
GET /_ml/anomaly_detectors/{job_id}/_forecast/{forecast_id}
```

**Response:**

```json
{
  "forecast_id": "wkCWa2IB2lF8nSE_TzZo",
  "job_id": "my-job",
  "status": "finished",
  "duration": "7d",
  "expires_in": "30d",
  "max_model_memory": "20mb",
  "create_time": 1711152000000,
  "expiry_time": 1713744000000
}
```

**Status values:**

| Status | Meaning |
|--------|---------|
| `scheduled` | Forecast has been requested but not yet started |
| `running` | Forecast computation is in progress |
| `finished` | Forecast completed successfully |
| `failed` | Forecast computation failed |
| `expired` | Forecast results have been deleted due to expiry (see proposal 4) |

**Not found behaviour:** Return HTTP 404 when the `forecast_id` does not exist (or has been fully purged), consistent with other ML GET endpoints.

**Why required:** Without this endpoint, Terraform has no way to implement the Read operation. This is the single most critical gap.

### 2. User-specified forecast ID (Highly Desirable)

Allow the create endpoint to accept an optional, user-specified `forecast_id`:

```
POST /_ml/anomaly_detectors/{job_id}/_forecast
{
  "forecast_id": "my-weekly-forecast",
  "duration": "7d",
  "expires_in": "30d"
}
```

**Behaviour:**
- If `forecast_id` is provided and a forecast with that ID already exists for the job, return the existing forecast (idempotent). Do not create a duplicate.
- If `forecast_id` is not provided, generate one server-side (current behaviour, backward compatible).
- Validate `forecast_id` format (e.g. lowercase alphanumeric, hyphens, underscores, 1-64 characters).

**Why highly desirable:** Idempotent creates are fundamental to Terraform's model. Without this, the provider would need to implement workarounds (e.g. checking if a forecast exists before creating, which is racy) or accept that re-applies create duplicates.

**Alternative:** If user-specified IDs are not feasible, a deterministic ID generation scheme (e.g. hash of `job_id` + `duration` + `expires_in` + `create_time_bucket`) could work, but is harder to reason about.

### 3. Status field in create response (Required if proposal 1 is implemented)

The current create response returns only `acknowledged` and `forecast_id`. Enhance it to include `status`:

```json
{
  "acknowledged": true,
  "forecast_id": "wkCWa2IB2lF8nSE_TzZo",
  "status": "scheduled"
}
```

This allows the Terraform provider to poll via the GET endpoint (proposal 1) until the forecast reaches `finished` or `failed`, similar to the existing patterns for job state and datafeed state transitions.

### 4. Forecast metadata survives expiry (Nice to Have)

Currently, when a forecast expires, its results are deleted and the forecast effectively ceases to exist. This causes silent drift in Terraform state.

**Proposal:** When a forecast expires, delete the result data but retain a lightweight metadata record with `status: "expired"`. This record would be returned by the GET endpoint (proposal 1) so that Terraform can detect the expiry and present it to the user as a state change.

**Cleanup:** The metadata record could be purged when:
- The forecast is explicitly deleted via the DELETE API
- The parent job is deleted
- A configurable metadata retention period elapses (e.g. 90 days)

**Why nice to have:** This prevents silent state drift but is not strictly required. If this is not implemented, the GET endpoint would return 404 for expired forecasts, and the Terraform provider would handle it by removing the resource from state (standard drift detection). The user would see the resource recreated on the next apply, which is acceptable if `expires_in: 0` (never expire) is the recommended pattern for Terraform-managed forecasts.

## Minimum Viable Set

For basic Terraform support, only **proposals 1 and 3** are strictly required. The provider would:

- Create forecasts via POST (current API)
- Read forecast status and detect expiry/deletion via GET (proposal 1)
- Use `expires_in: 0` as the recommended default for Terraform-managed forecasts to avoid automatic expiry
- Accept that re-applies without config changes are no-ops (the read sees the forecast exists and skips creation)

With **proposal 2** additionally, the provider would also have idempotent creates, making it robust against interrupted applies and state corruption.

## Impact on Existing Clients

All proposed changes are **additive and backward compatible**:
- The GET endpoint is new
- The `forecast_id` field in the create request is optional
- The `status` field in responses is a new field
- Metadata retention (proposal 4) changes deletion semantics but only for the metadata record, not the result data

## References

- [Forecast API documentation](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-forecast)
- [Delete forecast API documentation](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-delete-forecast)
- [go-elasticsearch client](https://github.com/elastic/go-elasticsearch) — `MLForecast`, `MLDeleteForecast`
- Existing Terraform provider patterns: `job_state` (polling for state transitions), `calendar_event` (server-generated IDs)
