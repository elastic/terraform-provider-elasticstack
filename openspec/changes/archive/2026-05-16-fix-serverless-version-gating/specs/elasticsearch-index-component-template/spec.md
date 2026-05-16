## MODIFIED Requirements

### Requirement: Data stream options support (REQ-027–REQ-031)

The resource SHALL support an optional `template.data_stream_options` block with nested `failure_store`
and `failure_store.lifecycle` blocks. During create and update, when `template.data_stream_options` is
configured, the provider SHALL map `failure_store.enabled` and `failure_store.lifecycle.data_retention`
into the Elasticsearch component template request body. During read, when Elasticsearch returns
`data_stream_options.failure_store`, the provider SHALL flatten those values back into Terraform state.
The `template.data_stream_options` block SHALL require `failure_store` when the block is present.

The `componenttemplate.Data` model SHALL implement the `entitycore.WithVersionRequirements` interface
via a `GetVersionRequirements()` method. That method SHALL delegate to
`datastreamoptions.GetVersionRequirements(d.Template)` and SHALL return a version requirement
(minimum ES 9.1.0) when `template.data_stream_options` is configured and non-null. When the template
object is null or unknown, or when `data_stream_options` is absent or null, the method SHALL return
`nil` (no requirements).

The entitycore resource envelope SHALL enforce these requirements automatically before every write
operation and during Read by calling `client.EnforceMinVersion` for each returned requirement.
`client.EnforceMinVersion` correctly handles Serverless clusters by short-circuiting to `true`
regardless of the reported server version. As a result, `data_stream_options` SHALL be usable on
Serverless clusters without error.

The `datastreamoptions` package SHALL be the single authoritative source for the `data_stream_options`
minimum version constant (`MinSupportedVersion = 9.1.0`) and the `GetVersionRequirements` helper.
The write callback (`writeComponentTemplate`) SHALL NOT contain a manual server version fetch or call
`EnforceMinServerVersion`; version enforcement is delegated to the envelope.

#### Scenario: Unsupported server version on stateful cluster

- GIVEN `template.data_stream_options` is configured
- AND the target Elasticsearch cluster is stateful and its version is below `9.1.0`
- WHEN create, update, or refresh runs
- THEN the provider SHALL return an error diagnostic
- AND it SHALL not call the Put API (on create/update)

#### Scenario: Serverless cluster is always supported

- GIVEN `template.data_stream_options` is configured
- AND the target Elasticsearch cluster flavour is `"serverless"`
- WHEN create, update, or refresh runs
- THEN the provider SHALL NOT return a version-gate error
- AND it SHALL include `data_stream_options` in the API request normally (on create/update)

#### Scenario: Read-time enforcement

- GIVEN `template.data_stream_options` is present in Terraform state
- AND the target Elasticsearch cluster is stateful and its version is below `9.1.0`
- WHEN `terraform refresh` runs
- THEN the provider SHALL return an error diagnostic (consistent with Write-time behavior)
