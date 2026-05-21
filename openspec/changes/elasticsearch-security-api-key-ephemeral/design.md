## Context

The existing `elasticstack_elasticsearch_security_api_key` managed resource computes `api_key` and `encoded` as `Sensitive: true` attributes, but they are still persisted in Terraform state. Any principal with read access to the state can extract raw credentials.

Terraform 1.10 introduced ephemeral resources: values returned by `Open()` live only in memory for the duration of the plan/apply. They are never written to the state file. The plugin framework (v1.10.0+, currently at v1.19.0 in this provider) supports this via the `ephemeral.EphemeralResource`, `EphemeralResourceWithConfigure`, and `EphemeralResourceWithClose` interfaces.

The existing security client functions `CreateAPIKey`, `CreateCrossClusterAPIKey`, and `DeleteAPIKey` in `internal/clients/elasticsearch/security.go` are ready to reuse without modification.

## Goals

- Expose a no-state-write path for Elasticsearch API key credentials via an ephemeral resource.
- Support both use cases with a single resource type: persistent-store pattern (`invalidate_on_close = false`, default) and in-run-only pattern (`invalidate_on_close = true`).
- Keep the existing managed resource entirely unchanged.
- Reuse existing Elasticsearch client helpers verbatim.

## Non-Goals

- Changes to the existing `elasticstack_elasticsearch_security_api_key` managed resource or its schema.
- `terraform import` support (not possible for ephemeral resources by design).
- A `Renew()` implementation (Elasticsearch API keys cannot be refreshed server-side; a new key is created each run).
- A companion ephemeral data source that reads an existing key by ID (impossible: the Elasticsearch Get API key endpoint does not return the raw `api_key` value).
- Automatic clean-up of previously-created ephemeral keys from prior runs.

## Decisions

| Topic | Decision |
|---|---|
| Resource count | Single resource type with `invalidate_on_close` attribute. Two separate types (Approach B) were evaluated and rejected: they duplicate `Open()` logic, proliferate provider surface area, and violate `@tobio`'s explicit request for a single attribute-controlled resource. |
| `invalidate_on_close` default | `false`. Ensures the dominant Vault/SSM persistent-store pattern works without the caller needing to remember to set the attribute. |
| `invalidate_on_close` required vs. optional | Optional with default `false`. Forcing an explicit choice adds friction for the common case and does not add safety beyond documentation. |
| `expiration` validation | Recommended (documentation warning), not required. Requiring `expiration` when `invalidate_on_close = false` would be more opinionated than the managed resource equivalent. A schema-level warning diagnostic is an option to revisit at implementation time. |
| `type = "cross_cluster"` scope | In scope. The code path is trivial (branch on type, call different client function); deferring would create an inconsistency with the managed resource and require a follow-up change. |
| Experimental flag | **Not** gated behind `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL`. The ephemeral resource contract is stable in the plugin framework. Gating it experimentally would delay adoption for the common use case. |
| `Close()` identity | `key_id` is stored in the result returned by `Open()`. The `Close()` callback receives the same result struct, so it can read `key_id` directly — no side-channel storage needed. |
| Provider registration pattern | Add an `EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource` method to the `Provider` type in `provider/plugin_framework.go`. Experimental/non-experimental split is not needed (see experimental flag decision above). |
| File location | `internal/elasticsearch/security/api_key/ephemeral_resource.go`. Keeps the new type co-located with the existing managed resource and shares package-level constants (`restAPIKeyType`, `crossClusterAPIKeyType`, `MinVersionWithCrossCluster`). |
| Acceptance test placement | `internal/elasticsearch/security/api_key/` alongside existing `acc_test.go`. |
| Documentation | `templates/ephemeral-resources/elasticstack_elasticsearch_security_api_key.md.tmpl`; generated to `docs/ephemeral-resources/elasticstack_elasticsearch_security_api_key.md`. |

## Non-Goals (implementation detail)

- Do not reuse the `tfModel` struct from the managed resource for the ephemeral resource — ephemeral result models use `ephemeral.OpenResponse.Result`, not `resource.State`, and have different plan modifier semantics.
- Do not add an `id` attribute (composite `<cluster_uuid>/<key_id>`) to the ephemeral result; `key_id` is sufficient for `Close()` operations and the composite format is a managed-resource artifact.

## Risks / Trade-offs

- **`invalidate_on_close = true` + persistent store**: If a caller sets `invalidate_on_close = true` and stores the credential in Vault or SSM, the key is immediately invalidated after the Terraform run, making the stored value useless. This is a footgun that must be documented prominently.
- **Key accumulation**: Each `terraform apply` (or even `plan`) calls `Open()` and creates a new API key. With `invalidate_on_close = false` and no `expiration` set, stale keys accumulate in Elasticsearch. Documentation should strongly recommend setting `expiration`.
- **Run interruption**: If Terraform is killed mid-apply with `invalidate_on_close = true`, `Close()` may not be called, leaving the key alive until it expires naturally (or never, if no expiration was set). This is a platform-level limitation of the ephemeral resource contract, not something the provider can address. It must be documented.
- **`Open()` called on plan**: Terraform calls `Open()` during `terraform plan` as well as `apply`. This means a new API key is created on every plan run. This is inherent to the Terraform ephemeral resource contract and not specific to this provider's implementation. Documentation should note this behavior.

## Open Questions

1. **Should `expiration` emit a warning validator when `invalidate_on_close = false`?** A `schema.Validator` that adds a warning diagnostic when `expiration` is absent and `invalidate_on_close = false` would help guide practitioners without being a hard error. Decision deferred to implementation.
2. **Should `invalidate_on_close` default to `false` or be an explicit required attribute?** Decided: optional with default `false`. Recorded here for auditability.
3. **Should cross-cluster API keys (`type = "cross_cluster"`) be in scope?** Decided: yes, in scope.
4. **Should the ephemeral resource be gated behind `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL`?** Decided: no, ship unconditionally.
5. **How should run interruption be documented for `invalidate_on_close = true`?** Document as a known limitation in the resource docs. No code-level mitigation is possible.

## Migration / State

Ephemeral resources have no persistent state by design. No state upgrade is needed. The existing managed resource is unaffected.
