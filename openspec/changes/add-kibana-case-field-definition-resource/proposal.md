## Why

In Elastic Stack 9.5, the Kibana Cases UI moved (the public `/api/cases/field_definitions` API followed in 9.6.0, elastic/kibana#289790). Kibana Cases moved custom fields out of the per-space case configuration and into a new, reusable field library defined as YAML (`POST /api/cases/field_definitions` and friends). Today the provider has no way to manage field library entries, so teams that declare their case configuration as code either fall back to manual UI changes or `terraform_data`/`local-exec`, neither of which gives state management or drift detection. This change is split out from [#5191](https://github.com/elastic/terraform-provider-elasticstack/issues/5191), which tracked the pre-9.5 case-configuration custom fields that the field library replaces; migration of those existing fields happens automatically server-side and needs no Terraform action.

Source: [elastic/terraform-provider-elasticstack#5192](https://github.com/elastic/terraform-provider-elasticstack/issues/5192).

## What Changes

- Add a new Terraform resource `elasticstack_kibana_case_field_definition` that manages a single field in the Kibana Cases field library through `POST`/`GET`/`PUT`/`DELETE /api/cases/field_definitions{,/{id}}`.
- The field's shape (type, label, control, metadata, validation) is declared as a YAML string in `definition`; `owner` scopes the field to a solution (Security/Observability/Stack), and `name` (optional, immutable, extracted from the YAML when omitted), `description`, and `is_global` are additional top-level attributes.
- `name` and the YAML-internal `type` are identity attributes enforced by the API as immutable after creation (400/409 on mismatch); the provider relies on those API errors rather than adding its own YAML-parsing plan modifier.
- Built on the repo's `entitycore.KibanaResource` envelope (matching `connectors`, `agentbuilderworkflow`), with the existing `customtypes.NormalizedYamlType` for semantic YAML equality, space-aware routing, and a minimum Elastic Stack version of 9.6.0.
- Add documentation, an example, and acceptance tests for the resource, with a Technical Preview note (the API's own `x-state`) alongside the 9.5+ GA badge on the feature itself (UI availability; the public API ships in 9.6.0).
- `dry_run` is explicitly not exposed as a resource attribute in v1 (see design.md) (live 9.6 probing showed the server preserves omitted `description`/`is_global` on update, so the implementation relies on preserve-on-omit).

Non-goals:

- No data source for listing/searching field definitions (`GET /api/cases/field_definitions` with paging/sorting).
- No "case template" resource/data source (`/api/cases/templates`).
- No migration tooling for pre-9.5 case-configuration custom fields (#5191); Kibana migrates those automatically.
- No `dry_run` pre-flight validation exposed on the resource.
- No custom plan modifier that parses `definition`'s YAML to pre-empt the server's 400/409 identity checks.

## Capabilities

### New Capabilities
- `kibana-case-field-definition`: Schema, identity/import, CRUD, version gating, and error-mapping requirements for the `elasticstack_kibana_case_field_definition` resource.

### Modified Capabilities
<!-- None. This is a net-new, additive resource with no interaction with existing specs. -->

## Impact

- New Kibana resource package under `internal/kibana/`; the required API operations and types already exist in the generated Kibana client (`generated/kbapi`), so no client regeneration is expected.
- New thin `internal/clients/kibanaoapi` wrapper for the field-definition CRUD calls.
- User-facing documentation and an example.
- Acceptance tests requiring a 9.6+ Kibana; some scenarios (partial-update semantics for `description`/`is_global`, named-space path routing) were resolved by probing a live 9.6 cluster during implementation (see design.md).
