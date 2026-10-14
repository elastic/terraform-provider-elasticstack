# Gotchas and anti-patterns

Load this when finalizing HCL or when diagnosing an unexpected plan diff.

## JSON-shaped attributes

Many resources accept JSON strings: `mappings`, `settings`, index/security role `query`, `metadata`, `global`, ingest pipeline `processors`, etc. Always use `jsonencode`:

```terraform
mappings = jsonencode({
  properties = {
    field1 = { type = "keyword" }
  }
})
```

The provider normalizes JSON server-side, so equivalent structures with different key order will not drift. But raw string literals with inconsistent whitespace WILL.

## Deletion protection

Resources like `elasticstack_elasticsearch_index` default `deletion_protection = true`. To actually delete the resource you must either:

1. Set `deletion_protection = false`, apply that change, then remove the resource and apply again, OR
2. Remove the resource from state with `terraform state rm` and delete the index out-of-band.

Never destroy with protection enabled.

## Forces-new attributes

Many Elasticsearch resources have static settings that cannot be changed in place: `number_of_shards`, `codec`, analyzers, `name`. Changing them triggers a destroy + create. For indices this means data loss. Always verify with `terraform plan` and recommend `moved` / manual reindex for production.

## Connection configuration precedence

Order of precedence (highest to lowest):

1. Provider alias (`provider = elasticstack.prod`) — preferred.
2. Provider-level `elasticsearch {}` / `kibana {}` / `fleet {}` block — preferred when you have one stack.
3. Per-resource `elasticsearch_connection {}` block — **deprecated** on most resources.

Never mix (2) and (3) in the same root module; configuration becomes ambiguous.

## Version gates

Specific attributes require minimum stack versions. Examples:

- `elasticstack_elasticsearch_security_role.description` — requires Elasticsearch >= 8.15.0.
- New Kibana alerting rule types — gated per Kibana minor.
- Fleet features — gated per Fleet major.

When the user is on an older stack, omit the attribute rather than setting it to null. The spec for each resource in `references/resources/<name>.md` lists the gates.

## Serverless caveats

Elastic Cloud Serverless does not expose every on-prem / ESS API. Common gaps:

- Cluster-level settings (`elasticsearch_cluster_settings`) are not user-configurable.
- ILM / snapshot repositories are managed by the platform.
- Some ML, Watcher, and node-level APIs are unavailable.

If the user targets Serverless and wants one of these, flag it rather than emitting HCL that will fail at apply time.

## Import IDs

Most resources use composite IDs. Patterns you will see:

- `<cluster_uuid>/<name>` for Elasticsearch resources scoped to a cluster.
- `<space_id>/<id>` for Kibana resources scoped to a space.
- Bare name for cluster-level singletons.

The per-entity spec documents the exact format; consult it before writing `terraform import` commands.

## Computed-vs-optional attributes

Many attributes are `optional+computed` with `UseStateForUnknown`. If you omit them, the server fills them in and Terraform tracks the value. Do NOT set them to empty strings / zero — that causes drift. Just omit the attribute.

## Reading what you write

For debugging use the matching `data.elasticstack_*` data source rather than hand-crafting API calls. Most resources have a paired data source (see `references/data-sources/`).
