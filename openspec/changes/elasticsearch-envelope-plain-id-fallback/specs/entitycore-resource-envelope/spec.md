## MODIFIED Requirements

### Requirement: Envelope owns the Read prelude (updated)

The `ElasticsearchResource` envelope SHALL resolve the read identity using a lenient three-step fallback when the decoded state model does not implement `WithReadResourceID` (or `GetReadResourceID` returns empty) and no write fallback is supplied: (1) attempt to parse `model.GetID().ValueString()` as a composite ID with `clients.CompositeIDFromStr`; if successful (non-nil result), use `compID.ResourceID`; (2) otherwise fall back to `model.GetResourceID().ValueString()` if it is non-empty; (3) otherwise use the raw `model.GetID().ValueString()` string as the resource ID. The envelope SHALL NOT return an error diagnostic solely because the state `id` is not in composite format. If all three fallback steps resolve to an empty string the envelope SHALL return an error diagnostic with summary "Invalid resource identifier" and SHALL NOT invoke the concrete read function.

This replaces the "Read falls back to composite ID resource segment" scenario in the canonical spec, which previously stated that a non-composite ID always fails.

#### Scenario: Read succeeds with plain (non-composite) ID — GetResourceID available

- GIVEN a state `id` attribute that is not in `<cluster_uuid>/<resource_id>` format (e.g. `my-job-id`)
- AND `model.GetResourceID()` returns a non-empty string (e.g. `my-job-id`)
- WHEN `Read` runs
- THEN the envelope SHALL NOT return an error diagnostic for the ID format
- AND the concrete `readFunc` SHALL be invoked with `my-job-id` as the `resourceID` argument

#### Scenario: Read succeeds with plain (non-composite) ID — GetResourceID empty

- GIVEN a state `id` attribute that is not in `<cluster_uuid>/<resource_id>` format (e.g. `my-job-id`)
- AND `model.GetResourceID()` returns an empty string
- WHEN `Read` runs
- THEN the envelope SHALL NOT return an error diagnostic for the ID format
- AND the concrete `readFunc` SHALL be invoked with `my-job-id` (the raw ID string) as the `resourceID` argument

#### Scenario: Read still fails when all fallbacks produce empty string

- GIVEN a state `id` attribute that is empty
- AND `model.GetResourceID()` also returns an empty string
- WHEN `Read` runs
- THEN the envelope SHALL return an error diagnostic with summary "Invalid resource identifier"
- AND the concrete `readFunc` SHALL NOT be invoked

#### Scenario: Read with composite ID continues to work

- GIVEN a state `id` attribute in `<cluster_uuid>/<resource_id>` format (e.g. `abc123/my-job-id`)
- WHEN `Read` runs
- THEN the envelope SHALL parse the composite ID and invoke `readFunc` with `my-job-id` as the `resourceID` argument
- AND behavior SHALL be unchanged from the prior implementation

### Requirement: Envelope owns the Delete prelude (updated)

The `ElasticsearchResource` envelope SHALL implement `Delete` by deserializing the prior state into the generic model `T`, resolving the resource ID using the same lenient three-step fallback as the updated Read prelude (composite parse → `GetResourceID()` → raw ID), resolving the scoped Elasticsearch client from the model's connection block via `GetElasticsearchClient`, and invoking the concrete delete function with `(context, *clients.ElasticsearchScopedClient, resourceID string, T)`. The envelope SHALL NOT return an error diagnostic solely because the state `id` is not in composite format. If all three fallback steps produce an empty string the envelope SHALL return an error diagnostic and SHALL NOT invoke the concrete delete function.

This replaces the "Composite ID parse failure short-circuits delete" scenario in the canonical spec, which previously stated that a non-composite ID always fails.

#### Scenario: Delete succeeds with plain (non-composite) ID — GetResourceID available

- GIVEN a state `id` attribute that is not in `<cluster_uuid>/<resource_id>` format (e.g. `my-job-id`)
- AND `model.GetResourceID()` returns a non-empty string (e.g. `my-job-id`)
- WHEN `Delete` runs
- THEN the envelope SHALL NOT return an error diagnostic for the ID format
- AND the concrete `deleteFunc` SHALL be invoked with `my-job-id` as the `resourceID` argument

#### Scenario: Delete succeeds with plain (non-composite) ID — GetResourceID empty

- GIVEN a state `id` attribute that is not in `<cluster_uuid>/<resource_id>` format (e.g. `my-job-id`)
- AND `model.GetResourceID()` returns an empty string
- WHEN `Delete` runs
- THEN the envelope SHALL NOT return an error diagnostic for the ID format
- AND the concrete `deleteFunc` SHALL be invoked with `my-job-id` (the raw ID string) as the `resourceID` argument

#### Scenario: Delete with composite ID continues to work

- GIVEN a state `id` attribute in `<cluster_uuid>/<resource_id>` format (e.g. `abc123/my-job-id`)
- WHEN `Delete` runs
- THEN the envelope SHALL parse the composite ID and invoke `deleteFunc` with `my-job-id` as the `resourceID` argument
- AND behavior SHALL be unchanged from the prior implementation

#### Scenario: Client resolution failure still short-circuits delete

- GIVEN `GetElasticsearchClient` returns error diagnostics
- WHEN `Delete` runs
- THEN the diagnostics SHALL be appended to `resp.Diagnostics`
- AND the concrete delete function SHALL NOT be invoked

## ADDED Requirements

### Requirement: Acceptance test for plain-ID import compatibility (REQ-ES-ENV-001)

The acceptance test suite for `elasticstack_elasticsearch_ml_anomaly_detection_job` SHALL include a test named `TestAccResourceAnomalyDetectionJobFrom0_12_2` that simulates a job originally stored with a plain `job_id` as the state `id` (as produced by provider ≤ 0.12.2) and verifies that the current provider can successfully refresh and apply using that state without returning a "Wrong resource ID" diagnostic.

#### Scenario: Plain-ID state from old provider can be refreshed by current provider

- GIVEN an existing anomaly detection job whose Terraform state stores `id = "<job_id>"` (plain, non-composite)
- AND the current provider implements the lenient-ID fallback
- WHEN `terraform plan` or `terraform apply` is run with the current provider
- THEN the provider SHALL successfully read the job from Elasticsearch using the plain `job_id` as the resource identifier
- AND the plan SHALL NOT include an error diagnostic about wrong resource ID format
