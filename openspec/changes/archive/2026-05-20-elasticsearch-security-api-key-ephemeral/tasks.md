## 1. Spec

- [x] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-security-api-key-ephemeral --type change` (or `make check-openspec` after sync).
- [x] 1.2 Resolve open question on whether to add a warning validator for missing `expiration` when `invalidate_on_close = false`; update delta spec if confirmed.
- [ ] 1.3 On completion of implementation, **sync** delta into `openspec/specs/elasticsearch-security-api-key/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [x] 2.1 Add `EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource` method to the `Provider` type in `provider/plugin_framework.go`. Register the new ephemeral resource factory. Import `"github.com/hashicorp/terraform-plugin-framework/ephemeral"`.
- [x] 2.2 Create `internal/elasticsearch/security/api_key/ephemeral_resource.go`:
  - Define an `ephemeralTfModel` struct for the ephemeral result (fields: `KeyID`, `Name`, `Type`, `RoleDescriptors`, `Expiration`, `ExpirationTimestamp`, `Metadata`, `Access`, `InvalidateOnClose`, `APIKey`, `Encoded`).
  - Implement `EphemeralResource` (metadata, schema), `EphemeralResourceWithConfigure` (receive provider client factory), and `EphemeralResourceWithClose` (invoke invalidation when `InvalidateOnClose = true`).
  - `Open()`: branch on `type`; call `elasticsearch.CreateAPIKey` or `elasticsearch.CreateCrossClusterAPIKey`; populate result with `key_id`, `api_key`, `encoded`, `expiration_timestamp`; set response via `resp.Result.Set(ctx, model)`.
  - `Close()`: read model from `req.Result`; if `InvalidateOnClose = true`, call `elasticsearch.DeleteAPIKey(ctx, client, keyID)`.
  - Expose a `NewEphemeralResource() ephemeral.EphemeralResource` constructor for registration.
- [x] 2.3 Define the ephemeral resource schema in `ephemeral_resource.go` (or a separate schema file if preferred):
  - Input attributes: `name` (required), `type` (optional, default `"rest"`), `role_descriptors` (optional, JSON, REST only), `expiration` (optional), `metadata` (optional, JSON), `access` (optional, cross-cluster only), `invalidate_on_close` (optional, bool, default `false`).
  - Result attributes (computed): `key_id` (string), `api_key` (string, sensitive), `encoded` (string, sensitive), `expiration_timestamp` (int64).
  - Validators: `name` length 1–1024, Basic Latin printable characters, matching the managed resource's current validator behavior (which currently allows leading/trailing whitespace); `type` one-of `rest`/`cross_cluster`; `role_descriptors` only with `type = "rest"`; `access` only with `type = "cross_cluster"`.
  - Do NOT include plan modifiers that apply to managed resources (`RequiresReplace`, `UseStateForUnknown`) — they are not applicable to ephemeral resources.
- [x] 2.4 Add version-gating for `type = "cross_cluster"` in `Open()`: check Elasticsearch version >= `8.10.0` before calling `CreateCrossClusterAPIKey`; return a clear error diagnostic if the version requirement is not met (mirrors `resource.go` logic).
- [x] 2.5 Add `elasticsearch_connection` block support to the ephemeral resource schema so the resource can be used with a non-default Elasticsearch connection (consistent with the managed resource).

## 3. Documentation

- [x] 3.1 Add template `templates/ephemeral-resources/elasticstack_elasticsearch_security_api_key.md.tmpl` covering:
  - Description of the ephemeral resource and when to use it vs. the managed resource.
  - Schema reference (input attributes and result outputs).
  - Usage example: persistent pattern (`invalidate_on_close = false`, store in Vault).
  - Usage example: in-run pattern (`invalidate_on_close = true`, provisioner).
  - **Warning**: combining `invalidate_on_close = true` with a persistent secret store results in immediate key invalidation after the run.
  - **Warning**: each plan/apply creates a new API key; strongly recommend setting `expiration` when `invalidate_on_close = false`.
  - **Note**: `Open()` is also called during `terraform plan`, not only apply.
  - **Note**: if Terraform is killed before `Close()` runs, the key may remain alive when `invalidate_on_close = true`.
- [x] 3.2 Run `make generate-docs` (or equivalent) to regenerate `docs/ephemeral-resources/elasticstack_elasticsearch_security_api_key.md`.

## 4. Testing

- [x] 4.1 Add acceptance test `TestAccEphemeralResourceSecurityAPIKey` verifying:
  - `api_key` and `encoded` are non-empty in the ephemeral result.
  - Neither `api_key` nor `encoded` appears in the Terraform state file after apply.
  - The Elasticsearch API key exists post-apply (confirm via a data source or direct API call) when `invalidate_on_close = false`.
- [x] 4.2 Add acceptance test `TestAccEphemeralResourceSecurityAPIKeyInvalidateOnClose` verifying:
  - With `invalidate_on_close = true`, the API key is invalidated after apply completes (confirm via Elasticsearch Get API key — key should be absent or marked as invalidated).
- [x] 4.3 Add acceptance test `TestAccEphemeralResourceSecurityAPIKeyWithExpiration` verifying:
  - Setting `expiration` populates `expiration_timestamp` in the result.
- [x] 4.4 Add acceptance test `TestAccEphemeralResourceSecurityAPIKeyCrossCluster` verifying:
  - `type = "cross_cluster"` creates a cross-cluster API key; `encoded` is non-empty; key appears in Elasticsearch.
  - Test is skipped when Elasticsearch version < `8.10.0`.
- [x] 4.5 Add unit tests for the ephemeral schema (attribute validation: `name` constraints, `type` one-of, `role_descriptors`/`access` mutual exclusion with type).
- [x] 4.6 Add unit tests for `Close()` logic: confirm `DeleteAPIKey` is called when `invalidate_on_close = true` and NOT called when `false`.
