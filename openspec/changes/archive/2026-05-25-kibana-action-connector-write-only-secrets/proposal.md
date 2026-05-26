## Why

The `elasticstack_kibana_action_connector` resource stores connector secrets (API keys, routing keys, webhook auth tokens, etc.) in Terraform state as a sensitive-but-plaintext JSON string. Terraform 1.10+ supports write-only attributes that accept ephemeral values and are never persisted to state, enabling users to source secrets from external secret stores (e.g. Vault) without ever writing them to the state file.

Today, users who attempt to pass an ephemeral value into the `secrets` attribute receive:

```
│ Error: Invalid use of ephemeral value
│
│   with elasticstack_kibana_action_connector.example,
│   on connector.tf line 10, in resource "elasticstack_kibana_action_connector" "example":
│   10:   secrets = jsonencode({
│   11:     routingKey = ephemeral.vault_kv_secret_v2.my_service.data["api_key"]
│   12:   })
│
│ Ephemeral values are not valid for "secrets", because it is not a write-only
│ attribute and must be persisted to state.
```

This proposal follows the established provider pattern of `password_wo`/`password_wo_version` on `elasticsearch_security_user` (PR #1419) and brings the same capability to connector secrets.

## What Changes

Add two new optional attributes to `elasticstack_kibana_action_connector`:

- **`secrets_wo`** — write-only string attribute (accepts ephemeral values, never persisted to state). Mutually exclusive with `secrets`.
- **`secrets_wo_version`** — optional string attribute that the practitioner bumps when the secret rotates; triggers a re-send of `secrets_wo` on the next apply.

The existing `secrets` attribute is **unchanged and non-breaking**. Practitioners who already store secrets in state can migrate at their own pace. The `secrets` attribute gains a `PreferWriteOnlyAttribute` validator pointing to `secrets_wo` so Terraform can surface a deprecation advisory.

After this change, the following pattern works without any secrets landing in state:

```hcl
ephemeral "vault_kv_secret_v2" "my_service" {
  mount = "secret"
  name  = "myapp/my_service"
}

resource "elasticstack_kibana_action_connector" "example" {
  name              = "PagerDuty"
  connector_type_id = ".pagerduty"
  config            = jsonencode({ apiUrl = "https://events.pagerduty.com/v2/enqueue" })

  secrets_wo         = jsonencode({
    routingKey = ephemeral.vault_kv_secret_v2.my_service.data["api_key"]
  })
  secrets_wo_version = "1"
}
```

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-action-connector`: add `secrets_wo` and `secrets_wo_version` attributes; add `PreferWriteOnlyAttribute` validator to `secrets`.

## Impact

- `internal/kibana/connectors/schema.go` — new attributes + validators
- `internal/kibana/connectors/models.go` — new model fields + `toAPIModel()` logic
- `internal/kibana/connectors/create.go` — read `secrets_wo` from `request.Config`
- `internal/kibana/connectors/update.go` — read `secrets_wo` from `request.Config`
- Connector acceptance tests — cover the `secrets_wo` path
- Provider docs for `elasticstack_kibana_action_connector`
