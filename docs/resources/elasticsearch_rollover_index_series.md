---
page_title: "elasticstack_elasticsearch_rollover_index_series Resource - terraform-provider-elasticstack"
subcategory: "Index"
description: |-
  Manages one alias-owned Elasticsearch index series for ILM rollover.
  When rollover_alias is missing, Terraform creates <prefix>-000001 and attaches that alias as its only write alias. The create request does not set index settings or mappings. Put those on an index template that matches the series, including index.lifecycle.name and index.lifecycle.rollover_alias. The template must not also define rollover_alias in its aliases, because Elasticsearch then reports a duplicate alias during rollover.
  When the alias already exists, Terraform adopts it without an extra flag and without changing indices or aliases, but only when it is an ordinary index alias with one effective writer whose name ends in -<digits>. Adoption includes every index currently attached to the alias, including indices whose names do not start with prefix. Adopting the alias acquires deletion rights over every one of those indices once deletion_protection is disabled. deletion_protection defaults to true, and destroy is rejected while it remains enabled. Disabling it takes a prior, separate apply. Destroy and replacement both read that previously applied value.
  Keep rollover_alias for this series alone. Do not also manage that alias with elasticstack_elasticsearch_alias, and do not manage its member indices with elasticstack_elasticsearch_index. Those resources would compete to own the same alias and indices.
  The series exists only while the exact rollover_alias exists. Refresh reads that alias's current members and effective writer. It does not restore a write flag, move the alias onto another index, or put a removed member back. A missing effective writer, or a writer whose name does not end in -<digits>, produces a warning only: membership is refreshed, state is kept, and write_index is null only when no effective writer exists. The actual writer name is still reported when one exists.
  If that exact alias is absent, refresh removes this resource from state. It does not look up former members or names matching <prefix>-*. The next apply bootstraps <prefix>-000001 when that name is free. A detached index is left unchanged and is outside the new alias. This resource does not clean up detached indices. If <prefix>-000001 is still occupied, create fails and does not attach the alias or modify that index.
  Changing prefix or rollover_alias replaces the resource. The old instance must be destroyed before the replacement is created: destroy before create. Terraform Core controls that ordering. This provider does not enforce it, and it does not block create_before_destroy set on this resource or propagated through dependencies. If a replacement adopts the same alias before the old instance is destroyed, the old instance can delete the replacement's members. Setting deletion_protection = false in the same operation that replaces the resource does not disable protection, because destroy still uses the previously applied state. Apply deletion_protection = false first, then replace or destroy.
  Before a destroy, quiesce producers and let in-flight lifecycle replacements finish. Destroy then deletes only the concrete indices currently attached to rollover_alias, including hidden and closed members, and it rediscovers membership until the delete timeout. Those rechecks can pick up a rollover that attaches during destroy. They cannot guarantee discovery of a transformation target that is not yet attached to the alias, such as a shrink index or searchable snapshot ILM has not swapped onto the alias yet.
  Direct index deletion leaves searchable snapshots in their repositories. Snapshots and detached indices need separate cleanup. Detached indices, and other series that still reference the same ILM policy, can still block policy deletion. None of that authorizes broader cleanup: this resource does not delete snapshots, remove another series' policy reference, stop ILM, or change action.destructive_requires_name.
  Import an existing series with <cluster_uuid>/<rollover_alias>/<prefix>. Terraform stores id as <cluster_uuid>/<rollover_alias> and sets prefix from the import identifier. It does not store the three-part string as id, and it does not infer prefix from the current writer's name. Import initializes deletion_protection to true. Destroy stays blocked until a later apply sets deletion_protection to false. Import reads a healthy ordinary alias and does not mutate it. Create adopts a healthy alias the same way, with no extra flag.
---

# elasticstack_elasticsearch_rollover_index_series (Resource)

Manages one alias-owned Elasticsearch index series for ILM rollover.

When `rollover_alias` is missing, Terraform creates `<prefix>-000001` and attaches that alias as its only write alias. The create request does not set index settings or mappings. Put those on an index template that matches the series, including `index.lifecycle.name` and `index.lifecycle.rollover_alias`. The template must not also define `rollover_alias` in its aliases, because Elasticsearch then reports a duplicate alias during rollover.

When the alias already exists, Terraform adopts it without an extra flag and without changing indices or aliases, but only when it is an ordinary index alias with one effective writer whose name ends in `-<digits>`. Adoption includes every index currently attached to the alias, including indices whose names do not start with `prefix`. Adopting the alias acquires deletion rights over every one of those indices once `deletion_protection` is disabled. `deletion_protection` defaults to `true`, and destroy is rejected while it remains enabled. Disabling it takes a prior, separate apply. Destroy and replacement both read that previously applied value.

Keep `rollover_alias` for this series alone. Do not also manage that alias with `elasticstack_elasticsearch_alias`, and do not manage its member indices with `elasticstack_elasticsearch_index`. Those resources would compete to own the same alias and indices.

The series exists only while the exact `rollover_alias` exists. Refresh reads that alias's current members and effective writer. It does not restore a write flag, move the alias onto another index, or put a removed member back. A missing effective writer, or a writer whose name does not end in `-<digits>`, produces a warning only: membership is refreshed, state is kept, and `write_index` is null only when no effective writer exists. The actual writer name is still reported when one exists.

If that exact alias is absent, refresh removes this resource from state. It does not look up former members or names matching `<prefix>-*`. The next apply bootstraps `<prefix>-000001` when that name is free. A detached index is left unchanged and is outside the new alias. This resource does not clean up detached indices. If `<prefix>-000001` is still occupied, create fails and does not attach the alias or modify that index.

Changing `prefix` or `rollover_alias` replaces the resource. The old instance must be destroyed before the replacement is created: destroy before create. Terraform Core controls that ordering. This provider does not enforce it, and it does not block `create_before_destroy` set on this resource or propagated through dependencies. If a replacement adopts the same alias before the old instance is destroyed, the old instance can delete the replacement's members. Setting `deletion_protection = false` in the same operation that replaces the resource does not disable protection, because destroy still uses the previously applied state. Apply `deletion_protection = false` first, then replace or destroy.

Before a destroy, quiesce producers and let in-flight lifecycle replacements finish. Destroy then deletes only the concrete indices currently attached to `rollover_alias`, including hidden and closed members, and it rediscovers membership until the delete timeout. Those rechecks can pick up a rollover that attaches during destroy. They cannot guarantee discovery of a transformation target that is not yet attached to the alias, such as a shrink index or searchable snapshot ILM has not swapped onto the alias yet.

Direct index deletion leaves searchable snapshots in their repositories. Snapshots and detached indices need separate cleanup. Detached indices, and other series that still reference the same ILM policy, can still block policy deletion. None of that authorizes broader cleanup: this resource does not delete snapshots, remove another series' policy reference, stop ILM, or change `action.destructive_requires_name`.

Import an existing series with `<cluster_uuid>/<rollover_alias>/<prefix>`. Terraform stores `id` as `<cluster_uuid>/<rollover_alias>` and sets `prefix` from the import identifier. It does not store the three-part string as `id`, and it does not infer `prefix` from the current writer's name. Import initializes `deletion_protection` to `true`. Destroy stays blocked until a later apply sets `deletion_protection` to `false`. Import reads a healthy ordinary alias and does not mutate it. Create adopts a healthy alias the same way, with no extra flag.

## Example Usage

```terraform
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "orders" {
  name           = "orders"
  index_patterns = ["orders-*"]
  priority       = 500

  template {
    settings = jsonencode({
      number_of_shards   = 1
      number_of_replicas = 1
    })
  }
}

resource "elasticstack_elasticsearch_rollover_index_series" "orders" {
  prefix         = "orders"
  rollover_alias = "orders-write"

  # Defaults to true. A prior, separate apply must set this to false before destroy
  # or replacement. Setting it false in the same operation that replaces the resource
  # still uses the previously applied protection, so destroy deletes nothing.
  # After that apply, destroy deletes every current alias member by concrete name.
  # Quiesce producers and let in-flight replacements finish first. Membership
  # rechecks do not see a transformation target that is not yet attached.
  # Searchable snapshots and detached indices need separate cleanup, and other
  # series can still block deletion of a shared ILM policy.
  deletion_protection = false

  # Changing prefix or rollover_alias replaces the resource. Terraform Core orders
  # that replacement. Keep destroy-before-create. Do not set create_before_destroy
  # here or on a dependency that propagates it: the old instance can then delete
  # members the replacement just adopted.

  depends_on = [elasticstack_elasticsearch_index_template.orders]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `prefix` (String) Literal prefix used only to name the bootstrap index `<prefix>-000001`. Changing it forces replacement. Membership and deletion follow `rollover_alias`, not this prefix.
- `rollover_alias` (String) Dedicated ordinary-index alias. It is the series identity and the only membership boundary. Changing it forces replacement.

### Optional

- `deletion_protection` (Boolean) Whether Terraform may destroy this series. Defaults to `true`. While enabled, destroy is rejected and no index is deleted, including when a replacement destroys this resource. Set it to `false` in a prior, separate apply before a destroy or replacement. Setting it to `false` only in the same operation that replaces the resource leaves the previously applied protection in place, so destroy is still rejected. After that prior apply, destroy deletes every index currently attached to `rollover_alias`, including indices outside `prefix`.
- `elasticsearch_connection` (Block List) Elasticsearch connection configuration block. (see [below for nested schema](#nestedblock--elasticsearch_connection))
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Internal identifier in the form `<cluster_uuid>/<rollover_alias>`. It stays the same when the effective writer or alias membership changes.
- `indices` (Set of String) Concrete indices currently attached to `rollover_alias`, including members that do not match `prefix`.
- `write_index` (String) Concrete name of the effective write index. Null when the alias has no effective writer.

<a id="nestedblock--elasticsearch_connection"></a>
### Nested Schema for `elasticsearch_connection`

Optional:

- `api_key` (String, Sensitive) API Key to use for authentication to Elasticsearch
- `bearer_token` (String, Sensitive) Bearer Token to use for authentication to Elasticsearch
- `ca_data` (String) PEM-encoded custom Certificate Authority certificate
- `ca_file` (String) Path to a custom Certificate Authority certificate
- `ca_fingerprint` (String) SHA-256 hex fingerprint (64 hexadecimal characters, no colons or separators) of the server TLS certificate used to pin the connection instead of a full CA chain
- `cert_data` (String) PEM encoded certificate for client auth
- `cert_file` (String) Path to a file containing the PEM encoded certificate for client auth
- `endpoints` (List of String, Sensitive) A list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.
- `es_client_authentication` (String, Sensitive) ES Client Authentication field to be used with the JWT token
- `headers` (Map of String, Sensitive) A list of headers to be sent with each request to Elasticsearch.
- `insecure` (Boolean) Disable TLS certificate validation
- `key_data` (String, Sensitive) PEM encoded private key for client auth
- `key_file` (String) Path to a file containing the PEM encoded private key for client auth
- `password` (String, Sensitive) Password to use for API authentication to Elasticsearch.
- `username` (String) Username to use for API authentication to Elasticsearch.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
# Import identifier: <cluster_uuid>/<rollover_alias>/<prefix>
# State id is <cluster_uuid>/<rollover_alias>. The three-part import string is not stored as id.
# prefix is taken from this identifier and is not inferred from the current writer.
# deletion_protection is initialized to true. A later apply must set it to false before destroy.
# Healthy create and import read an ordinary alias and do not require an extra flag.
terraform import elasticstack_elasticsearch_rollover_index_series.orders '<cluster_uuid>/orders-write/orders'
```
