resource "elasticstack_kibana_dashboard" "my_dashboard" {
  title       = "My Dashboard"
  description = "A dashboard showing key metrics"

  time_range = {
    from = "now-15m"
    to   = "now"
  }

  refresh_interval = {
    pause = false
    value = 60000 # 60 seconds
  }

  query = {
    language = "kql"
    text     = "status:success"
  }

  # Optional tags
  tags = ["production", "monitoring"]
}

# Example with JSON query (mutually exclusive with query.text)
resource "elasticstack_kibana_dashboard" "my_dashboard_json" {
  title       = "My Dashboard with JSON Query"
  description = "A dashboard with a structured query"

  time_range = {
    from = "now-15m"
    to   = "now"
  }

  refresh_interval = {
    pause = false
    value = 60000 # 60 seconds
  }

  query = {
    language = "kql"
    json = jsonencode({
      bool = {
        must = [
          {
            match = {
              status = "success"
            }
          }
        ]
      }
    })
  }

  # Optional tags
  tags = ["production", "monitoring"]
}

# Inline Lens (`lens-dashboard-app`) panels: typed by-value metric charts (`metric_chart_config`, etc.;
# not `by_value.config_json`). These blocks intentionally omit chart-level presentation fields (REQ-039 on `vis`);
# panels with typed chart `*_config` nested under `viz_config.by_value` when you need `time_range`, `drilldowns`, etc.
resource "elasticstack_kibana_dashboard" "lens_app_typed_by_value" {
  title            = "Dashboard with lens-dashboard-app (typed by-value)"
  description      = "Example: two metric panels sharing dashboard time_range (no chart-level presentation fields)"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [
    {
      type = "lens-dashboard-app"
      grid = { x = 0, y = 0, w = 12, h = 15 }
      lens_dashboard_app_config = {
        by_value = {
          metric_chart_config = {
            data_source_json = jsonencode({
              type          = "data_view_spec"
              index_pattern = "metrics-*"
              time_field    = "@timestamp"
            })
            query = { expression = "" }
            metrics = [{
              config_json = jsonencode({
                type      = "primary"
                operation = "count"
                format    = { type = "number" }
              })
            }]
          }
        }
      }
    },
    {
      type = "lens-dashboard-app"
      grid = { x = 12, y = 0, w = 12, h = 15 }
      lens_dashboard_app_config = {
        by_value = {
          metric_chart_config = {
            data_source_json = jsonencode({
              type          = "data_view_spec"
              index_pattern = "metrics-*"
              time_field    = "@timestamp"
            })
            query = { expression = "" }
            metrics = [{
              config_json = jsonencode({
                type      = "primary"
                operation = "count"
                format    = { type = "number" }
              })
            }]
          }
        }
      }
    }
  ]
}

# Classic Lens (`vis`) panel: same typed metric chart as above, nested under `viz_config.by_value`.
# Contrasts with `lens-dashboard-app`: API sends `type = "vis"` and inline chart config.
resource "elasticstack_kibana_dashboard" "vis_typed_by_value" {
  title            = "Dashboard with vis (typed by-value)"
  description      = "Example: metric via viz_config.by_value.metric_chart_config"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [{
    type = "vis"
    grid = { x = 0, y = 0, w = 24, h = 15 }
    viz_config = {
      by_value = {
        metric_chart_config = {
          data_source_json = jsonencode({
            type          = "data_view_spec"
            index_pattern = "metrics-*"
            time_field    = "@timestamp"
          })
          query = { expression = "" }
          metrics = [{
            config_json = jsonencode({
              type      = "primary"
              operation = "count"
              format    = { type = "number" }
            })
          }]
        }
      }
    }
  }]
}

# Markdown panel: `markdown_config` is a union — use `by_value` (inline content) or `by_reference`
# (a library item via `ref_id`), not both. By-value panels require a `settings` object per the
# Kibana API (link behavior via `open_links_in_new_tab`).
resource "elasticstack_kibana_dashboard" "markdown_by_value" {
  title            = "Dashboard with markdown (by-value)"
  description      = "Example: markdown_config.by_value with settings"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [{
    type = "markdown"
    grid = { x = 0, y = 0, w = 24, h = 10 }
    markdown_config = {
      by_value = {
        content = "# Runbook\n\nLinks respect **open_links_in_new_tab**."
        title   = "On-call notes"
        settings = {
          open_links_in_new_tab = true
        }
      }
    }
  }]
}

# By-reference links a markdown *library* item. Create that saved object out-of-band (Kibana UI
# or API) and substitute a real id for `ref_id` — there is no dedicated Terraform resource for
# markdown library items today.
resource "elasticstack_kibana_dashboard" "markdown_by_reference" {
  title            = "Dashboard with markdown (by-reference)"
  description      = "Example: markdown_config.by_reference with a placeholder ref_id"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [{
    type = "markdown"
    grid = { x = 0, y = 0, w = 24, h = 10 }
    markdown_config = {
      by_reference = {
        ref_id = "REPLACE_WITH_MARKDOWN_LIBRARY_ITEM_ID"
        title  = "Title overlay for library markdown"
      }
    }
  }]
}
