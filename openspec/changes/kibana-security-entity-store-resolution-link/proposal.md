## Why

Terraform practitioners managing Elastic Security entity analytics cannot currently express entity resolution relationships as code. The Kibana Entity Store API (available from Elastic Stack 9.1.0) supports linking alias entities to a single "golden" target entity (a resolution group), but no Terraform resource or data source exposes this capability.

Teams building entity-analytics pipelines need to:
- Declaratively link one or more alias entity IDs to a target entity ID (creating a resolution group).
- Read the current resolution group for any entity as a data source for downstream reference or verification.

## What Changes

- **New resource** `elasticstack_kibana_security_entity_store_entity_link`: manages a set of alias entities linked to one target (golden) entity via `POST /api/security/entity_store/resolution/link` and `POST /api/security/entity_store/resolution/unlink`.
- **New data source** `elasticstack_kibana_security_entity_store_resolution_group`: reads the resolution group for a given entity via `GET /api/security/entity_store/resolution/group`.

Both entities use the generated Kibana OpenAPI client (`generated/kbapi/`) operations:
- `GetSecurityEntityStoreResolutionGroup` / `GetSecurityEntityStoreResolutionGroupWithResponse`
- `PostSecurityEntityStoreResolutionLink` / `PostSecurityEntityStoreResolutionLinkWithResponse`
- `PostSecurityEntityStoreResolutionUnlink` / `PostSecurityEntityStoreResolutionUnlinkWithResponse`

Space routing uses `kibanautil.SpaceAwarePathRequestEditor(spaceID)`.

### Resource schema sketch — `elasticstack_kibana_security_entity_store_entity_link`

```hcl
resource "elasticstack_kibana_security_entity_store_entity_link" "example" {
  space_id    = "default"                            # optional/computed, RequiresReplace
  target_id   = "user-123"                          # required, RequiresReplace
  entity_ids  = ["user-456", "user-789"]            # required set(string), 1–1000 items

  resolution_group_json = <computed>                 # normalized JSON of the resolution group

  kibana_connection {}                               # optional, follows existing PF convention
}
```

`id` is a computed composite: `<space_id>/<target_id>`.

### Data source schema sketch — `elasticstack_kibana_security_entity_store_resolution_group`

```hcl
data "elasticstack_kibana_security_entity_store_resolution_group" "example" {
  space_id  = "default"   # optional/computed
  entity_id = "user-123"  # required

  # Computed outputs
  resolution_group_json = <computed>

  kibana_connection {}    # optional, follows existing PF convention
}
```

### Version gate

Both entities require `EnforceMinVersion` at Elastic Stack `9.1.0`.

### License and privileges

Both entities require an enterprise license and route privileges `securitySolution` AND `securitySolution-entity-analytics`. The acceptance test suite MUST skip gracefully when these are unavailable.

## Capabilities

### New Capabilities

- `kibana-security-entity-store-entity-link`: Resource for managing entity resolution links in the Kibana Entity Store. Supports create (link), read (get resolution group), update (set-diff link/unlink), delete (unlink managed IDs), import (`<space_id>/<target_id>`), and schema validation (1–1000 entity IDs; target_id not in entity_ids).
- `kibana-security-entity-store-resolution-group`: Data source for reading the resolution group for a given entity ID.

### Modified Capabilities

- _(none)_

## Impact

- **Specs**: Delta specs under `openspec/changes/kibana-security-entity-store-resolution-link/specs/`.
- **Implementation** (future):
  - `internal/kibana/security_entity_store_entity_link/` — resource package (schema, models, create, read, update, delete, resource.go, acc_test.go, testdata/).
  - `internal/kibana/security_entity_store_resolution_group/` — data source package.
  - `provider/plugin_framework.go` — register resource and data source.
  - `templates/` and `docs/` — generated documentation.
