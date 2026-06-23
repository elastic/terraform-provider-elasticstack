## 1. Prep and discovery

- [ ] 1.1 Confirm Plugin Framework version in `go.mod` supports `WriteOnly` on `resource/schema.StringAttribute` (requires PF ≥ 1.11); note the result in a code comment at the top of the new block function.
- [ ] 1.2 Verify `stringvalidator.PreferWriteOnlyAttribute` is available in the current `terraform-plugin-framework-validators` version; confirm against the security user schema precedent in `internal/elasticsearch/security/user/schema.go`.
- [ ] 1.3 Confirm `internal/utils/writeonlyhash` package is present (introduced in the fleet-cloud-connector change); if absent, note a prerequisite dependency and do not inline the helper.
- [ ] 1.4 Identify all resources that embed `ResourceBase` directly and call `GetElasticsearchClient`/`GetKibanaClient` directly (e.g., `internal/fleet/agentpolicy`, `internal/fleet/integration_policy`, `internal/apm/agent_configuration`, `internal/kibana/import_saved_objects`) and record them in the design as explicitly unaffected; skip any `_wo` changes there.
- [ ] 1.5 Identify all concrete resources with their own `ModifyPlan` method (e.g., `internal/elasticsearch/index/index/resource.go`, `internal/elasticsearch/index/template/resource.go`, `internal/kibana/spaces/resource.go`, `internal/kibana/prebuilt_rules/resource.go`, `internal/fleet/customintegration/resource.go`, `internal/elasticsearch/connector/resource/resource.go`) and create a tracking list for delegation to the envelope.

## 2. `internal/schema/connection.go` — resource-schema block functions

- [ ] 2.1 Add `GetEsResourceConnectionBlock()` returning `resource/schema.Block` (using `"github.com/hashicorp/terraform-plugin-framework/resource/schema"`). Mirror every attribute from `GetEsFWConnectionBlock()` but switch to `resource/schema.StringAttribute` etc. Add `_wo` companions for: `password`, `api_key`, `bearer_token`, `es_client_authentication`, `key_data`.
- [ ] 2.2 For each `_wo` attribute: set `Optional: true`, `Sensitive: true`, `WriteOnly: true`. Add `ConflictsWith` pointing at the plain companion.
- [ ] 2.3 For each plain companion that gains a `_wo` sibling (`password`, `api_key`, `bearer_token`, `es_client_authentication`, `key_data`): add `stringvalidator.PreferWriteOnlyAttribute(path.MatchRelative().AtParent().AtName("password_wo"))` (and similarly for the other fields) to the existing validators list, and add bidirectional `ConflictsWith` pointing at the `_wo` companion.
- [ ] 2.4 Add `GetKbResourceConnectionBlock()` returning `resource/schema.Block` mirroring `GetKbFWConnectionBlock()`. Add `password_wo`, `api_key_wo`, `bearer_token_wo` companions with the same treatment as 2.2 and 2.3.
- [ ] 2.5 Add `ElasticsearchResourceConnectionNullList()` and `ElasticsearchResourceConnectionObjectType()` helpers (analogous to the existing `ElasticsearchConnectionNullList`/`ElasticsearchConnectionObjectType`) that use the new resource-schema block's attribute set.
- [ ] 2.6 Add `KibanaResourceConnectionNullList()` and `KibanaResourceConnectionObjectType()` helpers analogously.
- [ ] 2.7 Add fallback `elasticsearchResourceConnectionBlockObjectAttrTypesFallback()` and `kibanaResourceConnectionBlockObjectAttrTypesFallback()` maps listing all attributes (including `_wo` fields, typed as `types.StringType`) to cover the sync.Once initialization path.
- [ ] 2.8 Add a `connectionBlockObjectAttrTypes` overload (e.g., `resourceConnectionBlockObjectAttrTypes`) and `rschemaAttributeToAttrType` that handle `resource/schema.StringAttribute`, `resource/schema.BoolAttribute`, `resource/schema.ListAttribute`, and `resource/schema.MapAttribute`, so the resource-variant object-type functions can compute from the actual block rather than only the fallback map.

## 3. `internal/clients/config/provider.go` — resource connection structs

- [ ] 3.1 Add `ElasticsearchResourceConnection` struct with all fields from `ElasticsearchConnection` plus `PasswordWo`, `APIKeyWo`, `BearerTokenWo`, `ESClientAuthenticationWo`, `KeyDataWo` as `types.String` with `tfsdk:"password_wo"`, `tfsdk:"api_key_wo"`, `tfsdk:"bearer_token_wo"`, `tfsdk:"es_client_authentication_wo"`, `tfsdk:"key_data_wo"` tags.
- [ ] 3.2 Add `KibanaResourceConnection` struct with all fields from `KibanaConnection` plus `PasswordWo`, `APIKeyWo`, `BearerTokenWo` as `types.String` with `tfsdk:"password_wo"`, `tfsdk:"api_key_wo"`, `tfsdk:"bearer_token_wo"` tags.

## 4. `internal/clients/provider_client_factory.go` — in-place update of typed resolution methods

- [ ] 4.1 Update `GetElasticsearchClient(ctx, types.List)` in place:
  - Decode the list into `[]ElasticsearchResourceConnection`.
  - Apply `_wo` preference to produce `[]ElasticsearchConnection`: if `PasswordWo` is non-empty, copy it to `Password`; otherwise leave `Password` as-is. Repeat for `APIKeyWo`/`APIKey`, `BearerTokenWo`/`BearerToken`, `ESClientAuthenticationWo`/`ESClientAuthentication`, `KeyDataWo`/`KeyData`.
  - Pass the resolved `[]ElasticsearchConnection` to the existing `config.NewFromFramework` path.
  - Preserve the existing "at most one block" validation and default-client fallback.
- [ ] 4.2 Update `GetKibanaClient(ctx, types.List)` in place:
  - Decode the list into `[]KibanaResourceConnection`.
  - Apply the same `_wo` preference for `password`, `api_key`, `bearer_token` to produce `[]KibanaConnection`.
  - Pass the resolved `[]KibanaConnection` to the existing `config.NewFromFrameworkKibanaResource` path.
  - Preserve default-client fallback and endpoint validation.

All existing callers (envelopes, `ResourceBase`-only resources, data sources, ephemeral resources, actions) continue to invoke `GetElasticsearchClient` / `GetKibanaClient` without signature changes.

## 5. `internal/clients/config/` — helper for `_wo` preference

- [ ] 5.1 Add small package-local helpers (e.g., `ResolveElasticsearchResourceConnection` / `ResolveKibanaResourceConnection`) that convert `ElasticsearchResourceConnection` → `ElasticsearchConnection` and `KibanaResourceConnection` → `KibanaConnection` with `_wo`-over-plain substitution. Keep these helpers close to the structs so `provider_client_factory.go` stays readable.

## 6. `internal/entitycore/resource_envelope.go` — ES envelope update

- [ ] 6.1 Update `NewElasticsearchResource` (and/or its `getSchema` equivalent) to inject `GetEsResourceConnectionBlock()` instead of `GetEsFWConnectionBlock()` into the resource schema.
- [ ] 6.2 Verify the `getClient` closure and the direct call in `runWrite` both invoke `r.Client().GetElasticsearchClient()`. No code change is required unless a reviewer introduced a separate `GetElasticsearchResourceClient`; if so, revert to the existing method name.
- [ ] 6.3 Update any ImportState or state upgrader that builds a null connection list to use `ElasticsearchResourceConnectionNullList()`; at minimum update `internal/elasticsearch/ml/calendar_job/resource.go` ImportState and any state upgrades for resources on the new envelope. Record unchanged `ResourceBase`-only call sites.
- [ ] 6.4 Implement `ModifyPlan` on `ElasticsearchResource` (satisfy `resource.ResourceWithModifyPlan`). Store a per-resource-type `Hasher` on the envelope struct. For each `_wo` credential attribute path (`elasticsearch_connection[0].password_wo`, etc.):
  - 6.4.1 Decode the config connection list; if null, empty, or no `_wo` attribute set, skip drift logic and clear any stored `_wo` hashes.
  - 6.4.2 For each set `_wo` attribute, load the stored hash from private state using `hasher.PrivateStateKey(<attributePath>)`.
  - 6.4.3 If no stored hash exists, do nothing (first apply / post-import baseline).
  - 6.4.4 If the configured value does not match the stored hash via `hasher.Matches`, emit a warning diagnostic naming the attribute path only and call `resp.Plan.SetAttribute(ctx, path.Root("elasticsearch_connection").AtListIndex(0).AtName("<credential>_wo"), value)` to schedule an update.

  - 6.4.5 If a `_wo` attribute is removed from config, unconditionally clear its stored hash with `resp.Private.SetKey(ctx, key, nil)`.
  - 6.4.6 If the entire connection block is removed, clear every `_wo` private-state key for the block.
- [ ] 6.5 After successful Create/Update in the envelope: for each `_wo` attribute that was set, compute and store the bcrypt hash via `hasher.Compute` + `resp.Private.SetKey`.
- [ ] 6.6 After successful Delete in the envelope: clear all `_wo` private-state keys.
- [ ] 6.7 Update the 6–8 concrete resources identified in 1.5 that already implement `ModifyPlan` to call the envelope's `ModifyPlan` first (or chain the diagnostics).

## 7. `internal/entitycore/kibana_resource_envelope.go` — Kibana envelope update

- [ ] 7.1 Update `NewKibanaResource` to inject `GetKbResourceConnectionBlock()` instead of `GetKbFWConnectionBlock()`.
- [ ] 7.2 Verify the `getClient` closure and the direct call in `runKibanaWrite` both invoke `r.Client().GetKibanaClient()`. No code change is required unless a reviewer introduced a separate `GetKibanaResourceClient`; if so, revert to the existing method name.
- [ ] 7.3 Update any ImportState or state upgrader that builds a null connection list to use `KibanaResourceConnectionNullList()`; at minimum update `internal/kibana/dataview/resource.go`, `internal/fleet/integration/state_upgrade.go`, and `internal/fleet/integration_policy/schema_v1.go` / `schema_v2.go`. Record unchanged `ResourceBase`-only call sites.
- [ ] 7.4 Implement `ModifyPlan` on `KibanaResource` for the three `_wo` attributes (`kibana_connection[0].password_wo`, `kibana_connection[0].api_key_wo`, `kibana_connection[0].bearer_token_wo`) matching the pattern in task 6.4. Use a per-resource-type hasher name `elasticstack_kibana_<name>`.
- [ ] 7.5 Store/clear hashes on successful Create/Update/Delete as in tasks 6.5 and 6.6.
- [ ] 7.6 Update any concrete Kibana/Fleet envelope resources with their own `ModifyPlan` to chain to the envelope's `ModifyPlan`.

## 8. Documentation updates

- [ ] 8.1 Update the attribute descriptions for the plain credential attributes in both new block functions to note: "Prefer `password_wo` (write-only) when sourcing credentials from ephemeral secret stores."
- [ ] 8.2 Update the attribute descriptions for the `_wo` attributes to note that they are write-only and not stored in state.
- [ ] 8.3 Document that only the `_wo` attributes prevent plaintext storage in state; plain attributes remain stored as before.

## 9. Acceptance tests

- [ ] 9.1 Add acceptance tests for at least one ES-backed resource using `password_wo` in the `elasticsearch_connection` block: verify the resource creates successfully and that state does not contain the password value.
- [ ] 9.2 Add acceptance tests for at least one Kibana-backed resource using `api_key_wo` in the `kibana_connection` block: verify the resource creates successfully and that state does not contain the API key value.
- [ ] 9.3 Add acceptance tests for write-only drift detection: apply with a credential `_wo` value, then change the value in config, run a plan, and assert that a warning diagnostic is emitted and an update is scheduled.
- [ ] 9.4 Add acceptance tests asserting conflict rejection when both plain and `_wo` companions are set and asserting the `PreferWriteOnlyAttribute` warning when a plain companion is used.
- [ ] 9.5 Add acceptance tests asserting no drift warning when the same `_wo` value is used across consecutive applies and asserting hash refresh after Update.

## 10. Coverage test and canonical spec updates

- [ ] 10.1 Update `provider/connection_schema_test.go` so that managed resources using the resource envelope are compared against `GetEsResourceConnectionBlock()` / `GetKbResourceConnectionBlock()`, while provider/data-source/ephemeral/action surfaces continue to compare against `GetEsFWConnectionBlock()` / `GetKbFWConnectionBlock()`.
- [ ] 10.2 Add delta specs `openspec/changes/connection-block-writeonly-credentials/specs/elasticsearch-provider-connection/spec.md` and `specs/provider-kibana-connection/spec.md` documenting that managed resources expose the resource-schema block variant while other surfaces remain on the provider-schema variant.
- [ ] 10.3 Add delta specs `openspec/changes/connection-block-writeonly-credentials/specs/entitycore-resource-envelope/spec.md` and `specs/entitycore-kibana-resource-envelope/spec.md` documenting the `ModifyPlan` behavior and the connection-block helper switch.
