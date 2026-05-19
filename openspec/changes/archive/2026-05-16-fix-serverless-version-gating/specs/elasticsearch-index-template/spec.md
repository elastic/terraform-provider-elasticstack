## MODIFIED Requirements

### Requirement: Compatibility (REQ-012)

When `ignore_missing_component_templates` is configured with one or more values, the resource SHALL
require Elasticsearch version >= 8.7.0; otherwise it SHALL return an error diagnostic and SHALL not
call the Put API.

This requirement is implemented via `template.Model.GetVersionRequirements()`, which returns an
`entitycore.VersionRequirement` for `ignore_missing_component_templates` (minimum ES 8.7.0) when
the attribute is non-null, non-unknown, and contains at least one element. On Create and Update, the
provider SHALL iterate over all returned requirements and call `client.EnforceMinVersion` for each.
`client.EnforceMinVersion` correctly handles Serverless clusters by short-circuiting to `true`
regardless of the reported server version. As a result, `ignore_missing_component_templates` SHALL
be usable on Serverless clusters without error.

#### Scenario: Feature on stateful old cluster

- GIVEN non-empty `ignore_missing_component_templates` and the cluster is stateful with ES < 8.7.0
- WHEN create or update runs
- THEN the provider SHALL error and SHALL not call the Put API

#### Scenario: Feature on Serverless cluster

- GIVEN non-empty `ignore_missing_component_templates`
- AND the target Elasticsearch cluster flavour is `"serverless"`
- WHEN create or update runs
- THEN the provider SHALL NOT return a version-gate error
- AND it SHALL include `ignore_missing_component_templates` in the API request normally

---

### Requirement: Compatibility — version gate for `data_stream_options` (REQ-033)

The provider SHALL enforce a minimum Elasticsearch version of `9.1.0` for `data_stream_options` and
SHALL return an error diagnostic (without calling the Put index template API) when the configured
cluster is stateful and reports a version below `9.1.0`. On Serverless clusters, the provider SHALL
treat the version requirement as satisfied regardless of the reported server version.

This requirement is implemented via `template.Model.GetVersionRequirements()`. The method delegates
to `datastreamoptions.GetVersionRequirements(m.Template)` and returns a version requirement (minimum
ES 9.1.0) when `data_stream_options` is configured and non-null. On Create and Update, the provider
SHALL iterate over all requirements returned by `plan.GetVersionRequirements()` and call
`client.EnforceMinVersion` for each — replacing the prior explicit `serverVersion` fetch and
`EnforceMinServerVersion` call. The `serverVersion` variable SHALL be removed from Create and Update
entirely; no other code in those methods requires it.

The entitycore envelope additionally calls `enforceVersionRequirements` during Read. As a result,
`data_stream_options` version enforcement also applies at refresh time (consistent with the component
template resource and Kibana resource envelopes).

#### Scenario: Feature on unsupported stateful cluster version

- GIVEN `data_stream_options` is configured
- AND the connected Elasticsearch server is stateful with version below `9.1.0`
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic without calling the Put index template API

#### Scenario: Feature on Serverless cluster

- GIVEN `data_stream_options` is configured
- AND the connected Elasticsearch cluster flavour is `"serverless"`
- WHEN create or update runs
- THEN the provider SHALL NOT return a version-gate error
- AND it SHALL include `data_stream_options` in the API request normally

#### Scenario: Feature on supported stateful cluster version

- GIVEN `data_stream_options` is configured
- AND the connected Elasticsearch server version is `9.1.0` or above
- WHEN create or update runs
- THEN the provider SHALL include `data_stream_options` in the API request normally

#### Scenario: Read-time enforcement

- GIVEN `data_stream_options` is present in Terraform state
- AND the target Elasticsearch cluster is stateful with version below `9.1.0`
- WHEN `terraform refresh` runs
- THEN the provider SHALL return an error diagnostic (consistent with Write-time behavior)
