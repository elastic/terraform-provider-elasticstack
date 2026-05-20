## Why

The `elasticstack_elasticsearch_security_api_key` managed resource stores the generated `api_key` (raw secret) and `encoded` (Base64 `id:api_key`) values as sensitive attributes in Terraform state. Any operator with read access to the state file can extract these credentials. This violates the principle of least privilege for teams that only need the key during an apply run or want to store it in an external secret store (Vault, AWS SSM) without it ever touching Terraform state.

Terraform 1.10 introduced the **ephemeral resource** primitive, which holds values only in memory during an apply — credentials are never written to the state file. This proposal adds an ephemeral resource variant of `elasticstack_elasticsearch_security_api_key` so teams can adopt either a "persist externally" or "in-run only" pattern without storing credentials in state.

A separate comment from `@tobio` (2026-05-17) explicitly requested an `invalidate_on_close` attribute on the ephemeral resource to support in-run API keys that are automatically revoked after the Terraform run completes.

## What Changes

Add a new **ephemeral resource** `elasticstack_elasticsearch_security_api_key` implemented in `internal/elasticsearch/security/api_key/ephemeral_resource.go`. The resource:

- **`Open()`** — creates an API key (regular or cross-cluster) on every `terraform plan` and `apply` call, returning credentials in the ephemeral result. Credentials are never written to state.
- **`Close()`** — if `invalidate_on_close = true`, calls the Elasticsearch Invalidate API key API using the `key_id` stored in the result; otherwise, it is a no-op, leaving the key alive for external storage.
- **`Renew()`** — not implemented. Elasticsearch API keys cannot be refreshed server-side; a new key must be created each run.

Provider registration adds a new `EphemeralResources()` method on the provider type in `provider/plugin_framework.go`, registering the new factory.

No changes are made to the existing managed resource (`elasticstack_elasticsearch_security_api_key`) or its schema.

### Schema sketch

```hcl
ephemeral "elasticstack_elasticsearch_security_api_key" "example" {
  # --- Input attributes (same naming as managed resource) ---
  name       = "app-key"              # required, string, 1–1024 chars
  type       = "rest"                 # optional, "rest" (default) or "cross_cluster"
  expiration = "7d"                   # optional, string; strongly recommended when invalidate_on_close = false

  role_descriptors = jsonencode({     # optional, JSON string; REST keys only
    my-role = {
      cluster = ["monitor"]
      indices = [{ names = ["logs-*"], privileges = ["read"] }]
    }
  })

  metadata = jsonencode({ env = "prod" }) # optional, JSON string

  # Cross-cluster only — mirrors managed resource shape
  access {
    search      { names = ["remote-index-*"] }
    replication { names = ["remote-index-*"] }
  }

  # Lifecycle control (new attribute)
  invalidate_on_close = false         # optional, bool, default false

  # --- Computed outputs (in ephemeral result, never in state) ---
  # key_id               = <computed, string>
  # api_key              = <computed, sensitive string>
  # encoded              = <computed, sensitive string>
  # expiration_timestamp = <computed, int64>
}
```

### Usage patterns

**Persistent pattern (store in Vault, never in `.tfstate`)**

```hcl
ephemeral "elasticstack_elasticsearch_security_api_key" "vault_key" {
  name       = "app-key"
  expiration = "30d"
  role_descriptors = jsonencode({ ... })
  # invalidate_on_close defaults to false
}

resource "vault_kv_secret_v2" "creds" {
  data_json = jsonencode({
    encoded = ephemeral.elasticstack_elasticsearch_security_api_key.vault_key.encoded
  })
}
```

**In-run pattern (key valid only during this apply)**

```hcl
ephemeral "elasticstack_elasticsearch_security_api_key" "inrun_key" {
  name                = "seed-job-key"
  invalidate_on_close = true
  role_descriptors    = jsonencode({ ... })
}

resource "null_resource" "seed" {
  triggers = { always = timestamp() }
  provisioner "local-exec" {
    command     = "./seed.sh"
    environment = {
      ES_API_KEY = ephemeral.elasticstack_elasticsearch_security_api_key.inrun_key.encoded
    }
  }
}
```

### Attribute table

| Attribute | Kind | Sensitive | Notes |
|---|---|---|---|
| `name` | Input (required) | No | 1–1024 chars, Basic Latin printable |
| `type` | Input (optional) | No | `"rest"` (default) or `"cross_cluster"` |
| `role_descriptors` | Input (optional) | No | JSON; REST keys only |
| `expiration` | Input (optional) | No | Strongly recommended when `invalidate_on_close = false` |
| `metadata` | Input (optional) | No | Arbitrary JSON |
| `access` | Input (optional) | No | Cross-cluster keys only |
| `invalidate_on_close` | Input (optional) | No | Bool, default `false`; calls Invalidate API after apply |
| `key_id` | Result (computed) | No | Elasticsearch key ID; used by `Close()` |
| `api_key` | Result (computed) | Yes | Raw API secret |
| `encoded` | Result (computed) | Yes | Base64 `id:api_key` |
| `expiration_timestamp` | Result (computed) | No | Epoch-ms; 0 if no expiration |

### Elasticsearch API surface

| Operation | Endpoint |
|---|---|
| Create (REST) | `POST /_security/api_key` |
| Create (cross-cluster) | `POST /_security/cross_cluster/api_key` |
| Invalidate | `POST /_security/api_key/invalidate` |

### Documentation

Add a template under `templates/ephemeral-resources/elasticstack_elasticsearch_security_api_key.md.tmpl` and generated documentation at `docs/ephemeral-resources/elasticstack_elasticsearch_security_api_key.md`. The docs MUST include a prominent warning about the footgun of combining `invalidate_on_close = true` with a persistent secret store, and MUST strongly recommend setting `expiration` when `invalidate_on_close = false` to prevent key proliferation.

## Capabilities

### New Capabilities

- `elasticsearch-security-api-key-ephemeral`: Ephemeral resource `elasticstack_elasticsearch_security_api_key` with `Open()` (create), `Close()` (optional invalidation via `invalidate_on_close`), schema, acceptance tests, and documentation.

### Modified Capabilities

- _(none; existing managed resource is unchanged)_

## Impact

- **Specs**: Delta under `openspec/changes/elasticsearch-security-api-key-ephemeral/specs/elasticsearch-security-api-key-ephemeral/spec.md`.
- **Implementation (future)**: `internal/elasticsearch/security/api_key/ephemeral_resource.go`, `provider/plugin_framework.go` (add `EphemeralResources()`), `templates/ephemeral-resources/`, `docs/ephemeral-resources/`, acceptance tests in `internal/elasticsearch/security/api_key/`.
- **No breaking changes** to the existing managed resource.
