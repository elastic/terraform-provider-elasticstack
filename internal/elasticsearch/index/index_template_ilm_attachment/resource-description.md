Attaches an ILM policy to a Fleet-managed or externally-managed index template by creating/updating the `@custom` component template with the lifecycle setting.

**Important:** Do NOT use this resource for index templates already managed by Terraform. Instead, set `index.lifecycle.name` directly in the `elasticstack_elasticsearch_index_template` or `elasticstack_elasticsearch_component_template` resource settings.

See the [official Elastic documentation](https://www.elastic.co/guide/en/fleet/current/data-streams-scenario3.html) for more information about customizing Fleet-managed data streams.

## How It Works

This resource creates or updates a component template named `<index_template>@custom` with the `index.lifecycle.name` setting.

Fleet-managed index templates typically include `<template>@custom` in their `composed_of` list. When this component template exists, its settings are merged into the index template.

**Key behaviors:**

- **Non-destructive**: If the `@custom` component template already exists with other settings, this resource preserves them and only adds/updates the ILM setting.
- **Clean deletion**: When destroyed, this resource removes only the `index.lifecycle.name` setting. If other settings exist in the `@custom` template, they are preserved. If the template becomes empty, it is deleted.

**Limitations:**

- **Version field not updated**: This resource does not modify the component template's `version` field. If you rely on version tracking for change detection (e.g., external tooling that monitors template versions), consider using `elasticstack_elasticsearch_component_template` instead, which gives you full control over the template including its version.

## When to Use This Resource

Use this resource when:

- You want to attach an ILM policy to a Fleet-managed index template
- You want to attach an ILM policy to an externally-managed index template that includes `<template>@custom` in its `composed_of` list
- You don't want Terraform to manage the entire index template

Do NOT use this resource when:

- The index template is already managed by `elasticstack_elasticsearch_index_template` — set `index.lifecycle.name` directly in the template's settings instead
- The `@custom` component template is already managed by `elasticstack_elasticsearch_component_template` — set the ILM setting there instead

## Alternative Approach

You can achieve similar results using `elasticstack_elasticsearch_component_template` directly. **Comparison:**

| Approach | Terraform Manages | Behavior |
|----------|-------------------|----------|
| `elasticstack_elasticsearch_component_template` | Entire `@custom` template | Overwrites any external changes to the template |
| `elasticstack_elasticsearch_index_template_ilm_attachment` | Only the ILM setting | Preserves other settings in the template |

Use `elasticstack_elasticsearch_component_template` if you want full control over the `@custom` template. Use this resource if you only want to manage the ILM setting and preserve any other customizations.
