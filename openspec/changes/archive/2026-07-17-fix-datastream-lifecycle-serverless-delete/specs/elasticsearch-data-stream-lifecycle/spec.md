## MODIFIED Requirements

### Requirement: Delete (REQ-019–REQ-020)

On delete, the resource SHALL parse the composite `id` from state to extract the data stream name, then determine whether the connected Elasticsearch deployment is Elastic Cloud Serverless before deleting lifecycle settings.

On a stateful deployment, the resource SHALL call `DeleteDataStreamLifecycle` with that name and the `expand_wildcards` value. On a successful Delete API response, the resource SHALL complete deletion so that the EntityCore resource envelope removes it from Terraform state. If serverless detection fails, or if the Delete API returns an error on a stateful deployment, the resource SHALL return the error diagnostic and SHALL not remove the resource from state.

On Elastic Cloud Serverless, where data stream lifecycle and retention are Elastic-managed, the resource SHALL NOT call the Delete Data Lifecycle API. Instead, it SHALL return a warning diagnostic explaining that lifecycle removal was skipped and allow the EntityCore resource envelope to remove the resource from Terraform state without modifying the data stream on the server.

#### Scenario: Successful delete on a stateful deployment

- GIVEN a resource with a valid `id` on a stateful Elasticsearch deployment
- WHEN delete runs and the Delete Data Lifecycle API succeeds
- THEN the resource SHALL be removed from state

#### Scenario: Delete on Elastic Cloud Serverless

- GIVEN a resource with a valid `id` on Elastic Cloud Serverless
- WHEN delete runs
- THEN the provider SHALL NOT call the Delete Data Lifecycle API
- AND Terraform diagnostics SHALL contain a warning that lifecycle removal was skipped
- AND the resource SHALL be removed from Terraform state

#### Scenario: Serverless detection fails during delete

- GIVEN the provider cannot determine whether Elasticsearch is serverless
- WHEN delete runs
- THEN the provider SHALL return the detection error diagnostic
- AND the Delete Data Lifecycle API SHALL NOT be called
- AND the resource SHALL remain in Terraform state
