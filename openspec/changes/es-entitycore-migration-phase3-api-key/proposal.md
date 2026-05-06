## Why

Migrate `elasticstack_elasticsearch_security_api_key` to the entitycore envelope. This is the most complex remaining PF resource due to private provider data usage and version-gated behavior. It does NOT fit the simple callback contract because Create/Update need to cache the cluster version in private state.

## What Changes

- Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[Data]` in `internal/elasticsearch/security/api_key/resource.go`
- Use placeholder callbacks for Create/Update (concrete type overrides them)
- Envelope owns Delete, Schema, Configure, and Metadata
- Concrete type overrides Create, Read, and Update to preserve version-check + private-data flow
- Keep `UpgradeState` on concrete type
- No import support (unchanged)

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `elasticsearch-security-api-key`: implementation migrates to the entitycore envelope while preserving private-state version caching.

## Impact

- `internal/elasticsearch/security/api_key/`. No public interface changes.
