## Why

A `reroute` processor in an integration's `@custom` ingest pipeline sends documents to a data stream the integration package does not own. Fleet scopes each integration's Elasticsearch API key to the package's own data streams, so those documents are rejected with `security_exception` on `indices:admin/auto_create` unless the destination is granted explicitly. Kibana 9.1.0 added `additional_datastreams_permissions` on the Fleet package policy API for exactly this case (surfaced in the Kibana UI as "Add a reroute processor permission" under a policy's advanced options), but `elasticstack_fleet_integration_policy` does not expose it. Practitioners who manage integration policies in Terraform must therefore edit the policy by hand in Kibana or call the Fleet API out of band, which drifts from state on the next apply.

`elasticstack_fleet_managed_integration` already exposes `additional_datastreams_permissions`, so the gap is limited to the classic integration policy resource, and the generated Fleet client already carries the field on both the request and response bodies.

## What Changes

- **Add** an optional `additional_datastreams_permissions` attribute (`list(string)`) to `elasticstack_fleet_integration_policy`, named to match the Fleet API field and the equivalent attribute on `elasticstack_fleet_managed_integration`.
- **Send** the configured values on create and update, and **populate** them back into state on read and import, so the attribute round-trips.
- **Clear** permissions server-side when the attribute is removed from configuration, by sending an explicit empty array rather than omitting the field.
- **Gate** the attribute behind Kibana 9.1.0 using the resource's existing version-requirement mechanism, alongside `agent_policy_ids` (8.15.0) and `output_id` (8.16.0). Kibana PR elastic/kibana#210452 is labelled `backport:skip`, so there is no 8.x support to accommodate.
- **Document** the attribute, including the Kibana UI label so users searching for "reroute processor permission" find it, and note that Kibana validates entries against the space's allowed namespace prefix.

Not breaking: the attribute is optional and additive, and existing configurations that omit it keep their current behaviour.

Out of scope: `elasticstack_fleet_elastic_defend_integration_policy` (a separate resource with its own typed schema) and any change to the `elasticstack_elasticsearch_ingest_processor_reroute` data source, which builds pipeline JSON and grants no privileges.

## Capabilities

### New Capabilities

_None — this extends an existing resource._

### Modified Capabilities

- `fleet-integration-policy`: adds `additional_datastreams_permissions` to the schema, with requirements for the create/update request body, read/import state population, clearing semantics, and the Kibana 9.1.0 compatibility gate.

## Impact

- `internal/fleet/integration_policy/`: new schema attribute on schema V3 (`schema.go`, `constants.go`), new model field and conversion in `models.go` (`toAPIModel`, `populateFromAPI`, `GetVersionRequirements`), new minimum-version constant in `resource.go`.
- No generated client changes: `PackagePolicyRequestMappedInputs` (create/update) and `PackagePolicy` (read) already declare `AdditionalDatastreamsPermissions`.
- No state upgrade: the attribute is additive and optional, so schema V3 gains an attribute without a version bump.
- Tests: unit coverage for conversion and version gating, plus acceptance coverage for setting, updating, and clearing the list.
- Docs: `docs/resources/fleet_integration_policy.md` regenerated via `make docs-generate`; examples updated.
