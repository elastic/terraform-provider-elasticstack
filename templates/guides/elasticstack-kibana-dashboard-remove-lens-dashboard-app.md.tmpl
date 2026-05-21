---
subcategory: ""
page_title: "Migrating lens-dashboard-app panels to vis"
description: |-
  Upgrade guide for removing the lens-dashboard-app panel type from elasticstack_kibana_dashboard.
---
# Migrating `lens-dashboard-app` panels to `vis`

The `lens-dashboard-app` panel type was included in the Kibana Dashboard API spec by mistake and has been removed upstream. The provider no longer exposes `type = "lens-dashboard-app"` or the `lens_dashboard_app_config` block.

Migrate existing Terraform configurations to `type = "vis"` and `vis_config` before upgrading to a provider version that includes this change. The dashboard resource is in technical preview; this is a breaking change without a deprecation cycle.

## Attribute mapping

| Before | After |
|--------|-------|
| `type = "lens-dashboard-app"` | `type = "vis"` |
| `lens_dashboard_app_config.by_value.<chart>_config` | `vis_config.by_value.<chart>_config` |
| `lens_dashboard_app_config.by_value.config_json` | panel-level `config_json` (not under `vis_config`) |
| `lens_dashboard_app_config.by_reference.*` | `vis_config.by_reference.*` |

## Typed by-value chart

```hcl
# Before:
panels = [{
  type = "lens-dashboard-app"
  grid = { x = 0, y = 0, w = 12, h = 15 }
  lens_dashboard_app_config = {
    by_value = {
      metric_chart_config = { ... }
    }
  }
}]

# After:
panels = [{
  type = "vis"
  grid = { x = 0, y = 0, w = 12, h = 15 }
  vis_config = {
    by_value = {
      metric_chart_config = { ... }
    }
  }
}]
```

## `by_value.config_json`

When you used `lens_dashboard_app_config.by_value.config_json`, move the JSON to panel-level `config_json`:

```hcl
# Before:
panels = [{
  type = "lens-dashboard-app"
  grid = { x = 0, y = 0, w = 12, h = 15 }
  lens_dashboard_app_config = {
    by_value = {
      config_json = jsonencode({ ... })
    }
  }
}]

# After (config_json moves to panel level):
panels = [{
  type = "vis"
  grid = { x = 0, y = 0, w = 12, h = 15 }
  config_json = jsonencode({ ... })
}]
```

## By-reference panels

```hcl
# Before:
lens_dashboard_app_config = {
  by_reference = {
    ref_id          = "panel_0"
    time_range      = { from = "now-15m", to = "now" }
    references_json = jsonencode([...])
  }
}

# After:
vis_config = {
  by_reference = {
    ref_id          = "panel_0"
    time_range      = { from = "now-15m", to = "now" }
    references_json = jsonencode([...])
  }
}
```

## Existing Kibana dashboards

Kibana dashboards that still contain `lens-dashboard-app` panels at the API level remain readable after upgrade: the provider uses the unknown-panel fallback and stores the panel in `config_json`. Update your HCL to `type = "vis"` (or keep the panel as `config_json`-only until you migrate) and run `terraform apply` to reconcile state.
