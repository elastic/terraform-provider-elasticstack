## Why

The Elastic Security Entity Store enables advanced entity analytics by indexing and enriching entity
data (users, hosts, services, generic entities) from Elasticsearch. Currently there is no Terraform
resource to install, configure, or manage the Entity Store lifecycle, forcing operators to rely on
manual Kibana UI steps or raw API calls. This gap makes it impossible to use Terraform to provision
a complete Elastic Security environment declaratively.

## What Changes

Add two new Terraform entities targeting the Kibana Security Entity Store API
(`/api/security/entity_store/*`):

**New resource: `elasticstack_kibana_security_entity_store`**

Manages Entity Store installation, log-extraction configuration, desired engine start/stop state,
and entity-type membership within a Kibana space. On create, calls `POST
/api/security/entity_store/install`. On update, calls `PUT /api/security/entity_store` for
log-extraction changes and start/stop endpoints to reconcile engine running state. On delete, calls
`POST /api/security/entity_store/uninstall`. The resource enforces `EnforceMinVersion` at Elastic
Stack 9.1.0.

**New data source: `elasticstack_kibana_security_entity_store_status`**

Read-only data source that calls `GET /api/security/entity_store/status` and exposes overall and
per-engine status, including optional component-level detail.

Both entities use the generated Kibana OpenAPI client (`generated/kbapi/kibana.gen.go`) and follow
the Plugin Framework patterns established in `internal/kibana/connectors/` and
`internal/kibana/security_role/`.

## Capabilities

### New Capabilities

- `kibana-security-entity-store`: new resource `elasticstack_kibana_security_entity_store`
  implementing install, update (log-extraction + start/stop), read, delete (uninstall), and import.
- `kibana-security-entity-store-status`: new data source
  `elasticstack_kibana_security_entity_store_status` implementing status reads with optional
  component detail.

### Modified Capabilities

None.

## Impact

- New package: `internal/kibana/security_entity_store/` containing resource and data source
  implementations.
- No changes to generated clients, provider schema registration is the only provider-level change.
- No schema version bump.
- Min stack version: 9.1.0 (enforced via `EnforceMinVersion`).
- Acceptance tests required for install, update, import, data source, and entity-type shrink guard.
