---
name: terraform-elasticstack-provider
description: >
  Author Terraform configuration that uses the elastic/elasticstack provider to manage Elasticsearch, Kibana, and Fleet
  resources and data sources. Use when writing, reviewing, or refactoring .tf files that reference
  elasticstack_* resources or data sources, or that configure the elasticstack provider block. Covers schema lookup,
  lifecycle semantics, provider connections, and common gotchas.
compatibility: "Terraform 1.0+, elasticstack ~> 0.14"
metadata:
  author: terraform-providers-admins
  version: "{{VERSION}}"
  visibility: internal
---

# Terraform elasticstack Provider

Author correct HCL for the `elastic/elasticstack` Terraform provider. The authoritative schema for each resource and data source lives under `references/` and loads on demand. Treat the per-entity reference file as the source of truth for attribute names, types, defaults, and lifecycle behavior.

## Workflow

Follow these steps, in order, before emitting HCL.

1. Capture context. Record Terraform version, target provider version, target stack flavor (Elasticsearch Cloud, Serverless, ECK, self-hosted), and which subsystems the user needs (Elasticsearch, Kibana, Fleet). See [references/context-checklist.md](references/context-checklist.md).
2. Pick the entity. Consult [references/index.md](references/index.md) to find the `elasticstack_*` resource or data source that matches the task.
3. Load the per-entity file from `references/resources/` or `references/data-sources/`. Read the schema block, the lifecycle notes, and the example. Do not guess attribute names.
4. Emit the provider block. Follow [references/provider.md](references/provider.md) for `required_providers`, auth environment variables, and per-subsystem connection blocks.
5. Cross-check against [references/gotchas.md](references/gotchas.md). Verify JSON attributes, deletion protection, connection precedence, and version gates before finalizing.
6. For content inside JSON-shaped attributes (mapping field types, query DSL, ingest processor bodies, detection rule params, ILM actions, Kibana rule params) consult [references/elastic-docs.md](references/elastic-docs.md). Do not guess JSON content.

## When to load which reference

| Scenario | Load |
|---|---|
| Starting a new configuration | `references/context-checklist.md`, then `references/provider.md` |
| Writing a specific resource (e.g. `elasticstack_kibana_alerting_rule`) | `references/resources/<short_name>.md` |
| Reading state with a data source | `references/data-sources/<short_name>.md` |
| Auth, connections, or multi-stack setup | `references/provider.md` |
| Error or unexpected plan diff | `references/gotchas.md` |
| Discovering what's available | `references/index.md` |
| Authoring JSON content (mappings, queries, rule params) inside HCL | `references/elastic-docs.md` |

## Examples

### User asks for an index with a mapping

> "Create an Elasticsearch index named `events` with a keyword field `user.id` and a text field `message`."

1. Load `references/resources/elasticsearch_index.md` for the schema.
2. Load `references/elastic-docs.md` only if the user's mapping requires field types you need to verify.
3. Emit HCL with `required_providers`, a `provider "elasticstack"` block, and a resource using `jsonencode` for `mappings`.

### User asks for a Kibana alerting rule

> "Create a Kibana alerting rule that fires when error count exceeds 10 in 5 minutes."

1. Load `references/resources/kibana_alerting_rule.md` for the schema.
2. Load `references/elastic-docs.md` to confirm the `params` JSON shape for the rule type.
3. Emit HCL with `jsonencode` around `params`.

### User hits `deletion_protection` error on destroy

> "Terraform refuses to destroy my `elasticstack_elasticsearch_index`."

1. Load `references/gotchas.md` (section on deletion protection).
2. Explain the two-step removal: set `deletion_protection = false` and apply, then remove the resource.

### User asks about an attribute that does not exist

> "What's the `retention` attribute on `elasticstack_elasticsearch_index`?"

1. Load `references/resources/elasticsearch_index.md` and confirm no such attribute exists.
2. Tell the user plainly. Suggest the closest real feature (e.g. ILM policy via `elasticstack_elasticsearch_index_lifecycle`) and link to its reference file.

## Guidelines

- Emit a `required_providers` block that pins `elastic/elasticstack` to an exact or `~>` version on every new configuration.
- Use `jsonencode({...})` for every JSON-shaped attribute (mappings, settings, queries, metadata, role `global` / `metadata`, ingest processors, alert params). Never embed raw JSON strings.
- Prefer the provider-level `elasticsearch {}` / `kibana {}` / `fleet {}` blocks. Do not mix them with per-resource `elasticsearch_connection {}` blocks — the latter is deprecated on most resources.
- Source credentials from environment variables or variables. Never inline secrets as string literals in HCL.
- Treat the per-entity reference files as authoritative for attribute names, types, defaults, and lifecycle. Do not guess from memory.
- When a user targets Elastic Cloud Serverless, verify the resource is supported there before emitting HCL. See `references/gotchas.md`.
- When a change will destroy data (force-new attributes on stateful resources like indices), call it out explicitly and recommend a reviewed `terraform plan` and a `moved` block or manual reindex.
- Return the output contract for every generated configuration: assumptions, required provider version, the HCL, and a `terraform init` + `terraform plan` validation plan.
