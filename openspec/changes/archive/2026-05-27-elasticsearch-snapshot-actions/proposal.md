## Why

SREs building HA/DR automation with Terraform cannot today trigger snapshot restores or on-demand snapshot creation through the provider. The provider supports snapshot repositories (`elasticstack_elasticsearch_snapshot_repository`) and SLM policies (`elasticstack_elasticsearch_snapshot_lifecycle`), but the actual restore and create-on-demand operations require falling out of IaC to call the Elasticsearch REST API directly. This gap forces teams to maintain additional tooling outside Terraform.

## What Changes

Add two Terraform provider-defined actions:

1. **`elasticstack_elasticsearch_snapshot_restore`** — invokes `POST /_snapshot/{repository}/{snapshot}/_restore` with user-configurable `wait_for_completion` and an `invoke` timeout.
2. **`elasticstack_elasticsearch_snapshot_create`** — invokes `POST /_snapshot/{repository}/{snapshot}` for on-demand snapshot creation with user-configurable `wait_for_completion` and an `invoke` timeout.

Both are implemented using `action.Action` from `terraform-plugin-framework v1.19.0` (already vendored). The provider struct gains a `provider.ProviderWithActions` implementation.

**Terraform 1.14+ is required** to use provider-defined actions.

### Restore action shape

```hcl
action "elasticstack_elasticsearch_snapshot_restore" "dr_restore" {
  repository           = elasticstack_elasticsearch_snapshot_repository.backup.name
  snapshot             = "my-snapshot-20240101"

  indices              = ["index-*"]
  include_global_state = false
  ignore_unavailable   = true
  partial              = false
  include_aliases      = true
  feature_states       = []

  rename_pattern        = "index-(.+)"
  rename_replacement    = "restored-index-$1"
  ignore_index_settings = ["index.refresh_interval"]
  index_settings        = jsonencode({ "index.number_of_replicas" = 0 })

  wait_for_completion = true
  timeouts {
    invoke = "30m"
  }

  elasticsearch_connection { ... }
}
```

### Create-snapshot action shape

```hcl
action "elasticstack_elasticsearch_snapshot_create" "nightly" {
  repository           = elasticstack_elasticsearch_snapshot_repository.backup.name
  snapshot             = "manual-snapshot-2024-01-01"

  indices              = ["index-*"]
  include_global_state = false
  ignore_unavailable   = true
  partial              = false
  expand_wildcards     = "open"
  metadata             = jsonencode({ created_by = "terraform" })

  wait_for_completion = true
  timeouts {
    invoke = "60m"
  }

  elasticsearch_connection { ... }
}
```

## Capabilities

### New Capabilities

- `elasticstack_elasticsearch_snapshot_restore`: provider-defined action for restoring an Elasticsearch snapshot
- `elasticstack_elasticsearch_snapshot_create`: provider-defined action for creating an Elasticsearch snapshot on demand

### Modified Capabilities

_(none)_

## Impact

- **`provider/plugin_framework.go`**: Implement `provider.ProviderWithActions`; add `res.ActionData = factory` in `Configure()`; register both actions in the `Actions()` method.
- **New packages**:
  - `internal/elasticsearch/cluster/snapshot_restore/` — restore action
  - `internal/elasticsearch/cluster/snapshot_create/` — create action
  - `internal/clients/elasticsearch/snapshot_restore.go` — ES client helper for restore
  - `internal/clients/elasticsearch/snapshot_create.go` — ES client helper for create
- **Minimum Terraform version**: 1.14+, documented in each action's generated docs page.
- **No breaking changes** to existing resources or data sources.
