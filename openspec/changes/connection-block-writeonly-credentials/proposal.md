## Why

Practitioners sourcing credentials from ephemeral secret stores (such as Vault KV2) need the sensitive attributes in the provider's per-resource `elasticsearch_connection` and `kibana_connection` blocks to be write-only — that is, accepted by Terraform but never stored in state. Currently the provider marks these attributes `Sensitive: true`, which redacts them in plan output but still writes the plaintext value to the Terraform state file. That state-file storage is incompatible with ephemeral-secret workflows and creates unnecessary secret exposure.

The eponymous feature request ([#3706](https://github.com/elastic/terraform-provider-elasticstack/issues/3706)) asks for a write-only `password` companion on the `elasticsearch_connection` block. A comment from @tobio extends the ask to `api_key`, `bearer_token`, `es_client_authentication`, and `key_data` on `elasticsearch_connection`, and to `password`, `api_key`, and `bearer_token` on `kibana_connection`. The existing `PreferWriteOnlyAttribute` pattern established in `elasticstack_elasticsearch_security_user` should be adopted.

## What Changes

- Add `GetEsResourceConnectionBlock()` in `internal/schema/connection.go` using `resource/schema` (not `provider/schema`). This block mirrors the existing `GetEsFWConnectionBlock()` but adds write-only companion attributes (`password_wo`, `api_key_wo`, `bearer_token_wo`, `es_client_authentication_wo`, `key_data_wo`) and `PreferWriteOnlyAttribute` validators on the corresponding plain companions.
- Add `GetKbResourceConnectionBlock()` in `internal/schema/connection.go` using `resource/schema`. This mirrors `GetKbFWConnectionBlock()` but adds `password_wo`, `api_key_wo`, and `bearer_token_wo` companions and `PreferWriteOnlyAttribute` validators on the corresponding plain companions.
- Add `ElasticsearchResourceConnection` struct in `internal/clients/config/` with `_wo` fields alongside the existing fields.
- Add `KibanaResourceConnection` struct in `internal/clients/config/` with `_wo` fields.
- Update `GetElasticsearchClient` in `internal/clients/provider_client_factory.go` to decode the `types.List` into `[]ElasticsearchResourceConnection`, apply `_wo`-over-plain preference, and continue building the scoped client from the resolved connection. Existing callers pass the same `types.List` and require no signature changes.
- Update `GetKibanaClient` similarly to decode into `[]KibanaResourceConnection` and apply `_wo`-over-plain preference.
- Update `NewElasticsearchResource` envelope in `internal/entitycore/resource_envelope.go` to use `GetEsResourceConnectionBlock()` instead of `GetEsFWConnectionBlock()`.
- Update `NewKibanaResource` envelope in `internal/entitycore/kibana_resource_envelope.go` to use `GetKbResourceConnectionBlock()` instead of `GetKbFWConnectionBlock()`.
- Add new null-list helpers and object-type functions for the resource connection variants (`ElasticsearchResourceConnectionNullList`, `ElasticsearchResourceConnectionObjectType`, `KibanaResourceConnectionNullList`, `KibanaResourceConnectionObjectType`) so that ImportState and state upgraders that call `ElasticsearchConnectionNullList` / `KibanaConnectionNullList` can be updated or stay on the provider-schema variant as appropriate.
- Wire `writeonlyhash`-based `ModifyPlan` in both envelopes to detect silent in-config changes to write-only credentials and schedule updates without requiring user-managed version companions. No `_wo_version` attributes are added — the `writeonlyhash` mechanism is the sole drift-detection mechanism. Each concrete resource type uses its own per-type `Hasher` salt.

## Capabilities

### New Capabilities

- `elasticsearch-connection-writeonly`: Defines the extended per-resource `elasticsearch_connection` block schema with write-only credential companions and the `ElasticsearchResourceConnection` struct.
- `kibana-connection-writeonly`: Defines the extended per-resource `kibana_connection` block schema with write-only credential companions and the `KibanaResourceConnection` struct.

### Modified Capabilities

- `elasticsearch-provider-connection` (modified): Managed resources now expose the resource-schema `elasticsearch_connection` block variant with `_wo` companions; provider-level, data-source, ephemeral-resource, and action surfaces remain on the provider-schema variant.
- `provider-kibana-connection` (modified): Managed resources now expose the resource-schema `kibana_connection` block variant with `_wo` companions; provider-level, data-source, ephemeral-resource, and action surfaces remain on the provider-schema variant.
- `entitycore-resource-envelope` (modified): Switches `NewElasticsearchResource` to inject the resource-schema connection block and adds envelope-level `ModifyPlan` for `_wo` drift detection.
- `entitycore-kibana-resource-envelope` (modified): Switches `NewKibanaResource` to inject the resource-schema connection block and adds envelope-level `ModifyPlan` for `_wo` drift detection.

## Impact

All managed resources that embed `elasticsearch_connection` or `kibana_connection` via the `NewElasticsearchResource` / `NewKibanaResource` envelopes gain the new write-only attributes automatically. No existing attribute is removed or renamed. The plain `password`, `api_key`, `bearer_token`, `es_client_authentication`, and `key_data` attributes remain; their behavior is unchanged when the `_wo` companion is not set. When both are set, the `_wo` value is never actually reachable because the attributes hard-conflict, but the factory still defensively prefers `_wo` should a conflict validator be relaxed.

Practitioners currently relying on the plain attributes for resource-level connections are unaffected. No state migration is required. Only practitioners who switch to `_wo` attributes avoid storing plaintext credentials in state; plain attributes continue to be stored in state as before.

### Compatibility

- Backward-compatible: new attributes are optional. Existing configs do not need modification.
- No schema version bump required for provider-level blocks (unchanged).
- Resources using the envelope will have new optional attributes in their schema; since they are optional and the plain companions still exist, no state migration is needed.
