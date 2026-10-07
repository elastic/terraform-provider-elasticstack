## Why

`elasticstack_kibana_security_entity_store`'s (and its `_status` data source's) `status_json`
computed attribute is built by passing the raw `GET /api/security/entity_store/status` response
body straight into `jsontypes.NewNormalizedValue`. `jsontypes.Normalized` round-trips JSON through
`interface{}` for semantic equality, which makes object key order insignificant but leaves JSON
array order significant. Kibana can return the `engines` array in a different order across two
reads of the same logical state (`generic`/`user` swapped), so two semantically-identical responses
compare unequal.

This was caught as `TestAccResourceKibanaSecurityEntityStore_import`'s `ImportStateVerify` failure
against Kibana `9.6.0-SNAPSHOT` (apply-time read returned `generic` before `user`; import-time read
returned the reverse), observed on [PR #4890 CI run 34672008631](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/34672008631/job/103495094889).
Because that CI shard is `continue-on-error`, the run stayed green and the flake went unnoticed
until filed. The same instability is a latent real-user risk: a plain `terraform plan`/`refresh`
could show a spurious `status_json` diff whenever Kibana reorders `engines` between reads, since
the attribute's only plan modifier is `UseStateForUnknown()`.

The codebase already has a precedent for this exact class of problem: `flattenStatus` sorts
`entity_types` with `sort.Strings` ([`internal/kibana/security_entity_store/helpers.go:277`](../../../internal/kibana/security_entity_store/helpers.go))
specifically because `engines` order is not guaranteed. `status_json`'s raw-passthrough path has no
equivalent normalization.

## What Changes

- Add a shared helper that normalizes the raw `GET /api/security/entity_store/status` response body
  before it is stored in `status_json`: decode the body generically, stable-sort the `engines`
  array by each element's `type` field, and re-encode. All other fields, including unmodeled and
  nested fields, must be preserved semantically. Re-encoding may change formatting, object key order,
  or other incidental byte representation, so exact-byte equality is required only for otherwise
  identical responses whose `engines` order differs and for explicit passthrough paths that return
  the original body unchanged. The raw body may contain fields not modeled by the provider's typed
  structs, so normalization must operate on generically-decoded JSON rather than re-marshaling from
  the typed `entityStoreStatus` struct.
- Apply this normalization in both places that currently do
  `model.StatusJSON = jsontypes.NewNormalizedValue(string(rawBody))`:
  `readEntityStore` ([`internal/kibana/security_entity_store/read.go:55`](../../../internal/kibana/security_entity_store/read.go))
  and `readEntityStoreDataSource` ([`internal/kibana/security_entity_store/data_source_read.go:48`](../../../internal/kibana/security_entity_store/data_source_read.go)).
- No schema, API, or attribute-type changes. `status_json` remains a `jsontypes.Normalized` string
  attribute on both the resource and the data source.

## Capabilities

### Modified Capabilities

- `kibana-security-entity-store`: `status_json` (REQ-007) now normalizes `engines` array order
  before the value is stored, so two otherwise-identical reads produce the same `status_json` value
  regardless of the order Kibana returned `engines` in.
- `kibana-security-entity-store-status`: the `_status` data source's `status_json` attribute
  (REQ-001) gets the same normalization, since it is built from the same raw response body via the
  same code pattern.

## Impact

- `internal/kibana/security_entity_store/helpers.go` — add a shared `normalizeStatusJSON` (or
  equivalently named) helper and its unit tests.
- `internal/kibana/security_entity_store/read.go` — use the helper when building
  `model.StatusJSON` in `readEntityStore`.
- `internal/kibana/security_entity_store/data_source_read.go` — use the helper when building
  `model.StatusJSON` in `readEntityStoreDataSource`.
- `internal/kibana/security_entity_store/acc_test.go` — no required change;
  `TestAccResourceKibanaSecurityEntityStore_import` should pass once `status_json` is order-stable.
