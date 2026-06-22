## 1. Prep and discovery

- [ ] 1.1 Confirm Plugin Framework version in `go.mod` supports `WriteOnly` on `resource/schema.StringAttribute` (requires PF ≥ 1.11); note the result in a code comment at the top of the new block function
- [ ] 1.2 Verify `stringvalidator.PreferWriteOnlyAttribute` is available in the current `terraform-plugin-framework-validators` version; confirm against the user schema precedent in `internal/elasticsearch/security/user/schema.go`
- [ ] 1.3 Confirm `internal/utils/writeonlyhash` package is present (introduced in the fleet-cloud-connector change); if absent, note a prerequisite dependency and do not inline the helper

## 2. `internal/schema/connection.go` — resource-schema block functions

- [ ] 2.1 Add `GetEsResourceConnectionBlock()` returning `resource/schema.Block` (using `"github.com/hashicorp/terraform-plugin-framework/resource/schema"`). Mirror every attribute from `GetEsFWConnectionBlock()` but switch to `resource/schema.StringAttribute` etc. Add `_wo` companions for: `password`, `api_key`, `bearer_token`, `es_client_authentication`, `key_data`
- [ ] 2.2 For each `_wo` attribute: set `Optional: true`, `Sensitive: true`, `WriteOnly: true`. Add ConflictsWith pointing at the plain companion.
- [ ] 2.3 For each plain companion that gains a `_wo` sibling (`password`, `api_key`, `bearer_token`, `es_client_authentication`, `key_data`): add `stringvalidator.PreferWriteOnlyAttribute(path.MatchRelative().AtParent().AtName("password_wo"))` (and similarly for the other fields) to the existing validators list
- [ ] 2.4 Add `GetKbResourceConnectionBlock()` returning `resource/schema.Block` mirroring `GetKbFWConnectionBlock()`. Add `password_wo`, `api_key_wo`, `bearer_token_wo` companions with the same treatment as 2.2 and 2.3.
- [ ] 2.5 Add `ElasticsearchResourceConnectionNullList()` and `ElasticsearchResourceConnectionObjectType()` helpers (analogous to the existing `ElasticsearchConnectionNullList` / `ElasticsearchConnectionObjectType`) that use the new resource-schema block's attribute set
- [ ] 2.6 Add `KibanaResourceConnectionNullList()` and `KibanaResourceConnectionObjectType()` helpers analogously
- [ ] 2.7 Add fallback `elasticsearchResourceConnectionBlockObjectAttrTypesFallback()` and `kibanaResourceConnectionBlockObjectAttrTypesFallback()` maps listing all attributes (including `_wo` fields, typed as `types.StringType`) to cover the sync.Once initialization path
- [ ] 2.8 Extend `fwAttributeToAttrType` (or add a resource-schema variant) to handle `resource/schema.StringAttribute`, `resource/schema.BoolAttribute`, etc., if the existing helper only handles `provider/schema` attribute types

## 3. `internal/clients/config/provider.go` — resource connection structs

- [ ] 3.1 Add `ElasticsearchResourceConnection` struct with all fields from `ElasticsearchConnection` plus `PasswordWo`, `APIKeyWo`, `BearerTokenWo`, `ESClientAuthenticationWo`, `KeyDataWo` as `types.String` with appropriate `tfsdk:"..._wo"` tags
- [ ] 3.2 Add `KibanaResourceConnection` struct with all fields from `KibanaConnection` plus `PasswordWo`, `APIKeyWo`, `BearerTokenWo` as `types.String` with appropriate `tfsdk:"..._wo"` tags

## 4. `internal/clients/config/` — factory functions for resource connections

- [ ] 4.1 Add `NewFromFrameworkElasticsearchResource(ctx, []ElasticsearchResourceConnection, version string) (*Config, diags)` (or extend an existing factory). Implement `_wo` preference: if `PasswordWo` is non-empty, use it; else use `Password`. Apply the same preference for `APIKeyWo`/`APIKey`, `BearerTokenWo`/`BearerToken`, `ESClientAuthenticationWo`/`ESClientAuthentication`, `KeyDataWo`/`KeyData`.
- [ ] 4.2 Add `NewFromFrameworkKibanaResource(ctx, []KibanaResourceConnection, version string) (*Config, diags)` with the same `_wo` preference for `password`, `api_key`, `bearer_token`.

## 5. `internal/clients/provider_client_factory.go` — typed resolution methods

- [ ] 5.1 Add `GetElasticsearchResourceClient(ctx, types.List) (*ElasticsearchScopedClient, diags)` that decodes `[]ElasticsearchResourceConnection` from the list and delegates to `NewFromFrameworkElasticsearchResource`. Falls back to the default provider client when the list is empty.
- [ ] 5.2 Add `GetKibanaResourceClient(ctx, types.List) (*KibanaScopedClient, diags)` analogously, decoding `[]KibanaResourceConnection`.

## 6. `internal/entitycore/resource_envelope.go` — ES envelope update

- [ ] 6.1 Update `NewElasticsearchResource` (and/or its `getSchema` equivalent) to inject `GetEsResourceConnectionBlock()` instead of `GetEsFWConnectionBlock()` into the resource schema
- [ ] 6.2 Update calls to `GetElasticsearchClient` → `GetElasticsearchResourceClient` in the envelope's Create/Read/Update/Delete/ImportState paths
- [ ] 6.3 Update any ImportState that builds a null connection list to use `ElasticsearchResourceConnectionNullList()`
- [ ] 6.4 Implement `ModifyPlan` on the envelope (or an envelope hook): for each write-only credential attribute path (`elasticsearch_connection[0].password_wo`, etc.), read the config value, look up the stored bcrypt hash from private state, compare with `writeonlyhash.Hasher.Matches`, and mark for update + emit a warning diagnostic on mismatch. Handle removal (config null → clear stored hash). Use `writeonlyhash.New("elasticstack_elasticsearch_resource")` as the hasher name.
- [ ] 6.5 After successful Create/Update in the envelope: for each `_wo` attribute that was set, compute and store the bcrypt hash via `hasher.Compute` + `resp.Private.SetKey`.
- [ ] 6.6 After successful Delete in the envelope: clear all `_wo` private-state keys.

## 7. `internal/entitycore/kibana_resource_envelope.go` — Kibana envelope update

- [ ] 7.1 Update `NewKibanaResource` to inject `GetKbResourceConnectionBlock()` instead of `GetKbFWConnectionBlock()`
- [ ] 7.2 Update calls to `GetKibanaClient` → `GetKibanaResourceClient` in the envelope's CRUD/ImportState paths
- [ ] 7.3 Update any ImportState that builds a null connection list to use `KibanaResourceConnectionNullList()`
- [ ] 7.4 Implement `ModifyPlan` on the Kibana envelope for the three `_wo` attributes (`kibana_connection[0].password_wo`, `kibana_connection[0].api_key_wo`, `kibana_connection[0].bearer_token_wo`). Same pattern as task 6.4 above. Use `writeonlyhash.New("elasticstack_kibana_resource")` as the hasher name.
- [ ] 7.5 Store/clear hashes on successful Create/Update/Delete as in tasks 6.5 and 6.6.

## 8. Documentation updates

- [ ] 8.1 Update the attribute descriptions for the plain credential attributes in both new block functions to note: "Prefer `password_wo` (write-only) when sourcing credentials from ephemeral secret stores."
- [ ] 8.2 Update the attribute descriptions for the `_wo` attributes to note that they are write-only and not stored in state.

## 9. Acceptance tests

- [ ] 9.1 Add acceptance tests for at least one ES-backed resource using `password_wo` in the `elasticsearch_connection` block: verify the resource creates successfully and that state does not contain the password value.
- [ ] 9.2 Add acceptance tests for at least one Kibana-backed resource using `api_key_wo` in the `kibana_connection` block: verify the resource creates successfully and that state does not contain the API key value.
- [ ] 9.3 Add acceptance tests for write-only drift detection: apply with a credential `_wo` value, then change the value in config, run a plan, and assert that a warning diagnostic is emitted and an update is scheduled.
