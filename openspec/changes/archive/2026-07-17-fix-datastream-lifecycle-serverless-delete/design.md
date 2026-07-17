## Context

`elasticstack_elasticsearch_data_stream_lifecycle` delegates destroy to `DeleteDataStreamLifecycle`. On Elastic Cloud Serverless, the Delete Data Lifecycle API returns an unavailable-API response because Elasticsearch manages lifecycle and retention. Treating that response as a normal delete failure blocks `terraform destroy` even though there is no lifecycle configuration the provider can remove.

The scoped Elasticsearch client already exposes `IsServerless(ctx)` and returns Plugin Framework diagnostics. EntityCore removes state after a delete callback returns no error diagnostics, including when warning diagnostics are present.

## Goals / Non-Goals

**Goals:**

- Preserve stateful Delete Data Lifecycle API behavior.
- Avoid the unavailable Delete Data Lifecycle API on serverless.
- Make the serverless destroy outcome visible through a warning diagnostic.
- Keep the canonical resource requirements and tests aligned with the behavior.

**Non-Goals:**

- Change data stream lifecycle create, read, or update behavior on serverless.
- Suppress errors when serverless detection fails.
- Add a schema attribute or alter serverless lifecycle ownership.

## Decisions

### Detect deployment flavor before the Delete API call

`DeleteDataStreamLifecycle` calls `IsServerless(ctx)` before constructing the Delete request. This reuses the provider's existing server-flavor detection rather than attempting to recognize the Delete API's unavailable response after the request.

The alternative—special-casing the HTTP 410 response—would still issue an API request that is known to be unavailable and ties behavior to a particular server response.

### Return a warning on serverless and let EntityCore remove state

For a serverless deployment the delete helper returns a warning diagnostic and skips the API request. Since the diagnostic contains no error, the EntityCore delete envelope completes deletion and removes the Terraform resource from state. This accurately represents that the Terraform-managed lifecycle settings cannot be independently removed from Elastic-managed serverless data streams.

The alternative—returning no diagnostic—would hide the meaningful distinction that no server-side deletion occurred.

### Treat flavor-detection failures as delete failures

If `IsServerless` returns errors, propagate them without sending the Delete request. Proceeding without knowing deployment flavor could reintroduce the serverless destroy failure.

## Risks / Trade-offs

- **Serverless destruction does not change server-side lifecycle settings** → The warning explicitly tells users that Elastic manages those settings and Terraform only removes its state entry.
- **Flavor detection adds an Info API dependency to deletion** → Detection errors are surfaced and no potentially invalid Delete request is sent.
