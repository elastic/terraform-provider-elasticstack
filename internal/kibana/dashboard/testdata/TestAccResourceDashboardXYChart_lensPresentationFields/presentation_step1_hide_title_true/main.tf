variable "dashboard_title" {
  type = string
}

variable "target_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "target" {
  title       = var.target_title
  description = "acceptance target for xy chart dashboard drilldown"
  time_range = {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval = {
    pause = true
    value = 0
  }
  query = {
    language = "kql"
    text     = ""
  }
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with XY chart presentation fields acceptance"
  time_range = {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval = {
    pause = true
    value = 0
  }
  query = {
    language = "kql"
    text     = ""
  }

  panels = [{
    type = "vis"
    grid = { x = 0, y = 0, w = 24, h = 15 }
    viz_config = {
      by_value = {
        xy_chart_config = {
          hide_title  = true
          hide_border = true
          references_json = jsonencode([{
            name = "acc-ref-name"
            type = "index-pattern"
            id   = "acc-ref-id"
          }])
          drilldowns = [
            {
              dashboard_drilldown = {
                dashboard_id = elasticstack_kibana_dashboard.target.dashboard_id
                label        = "Go to target"
              }
            },
            {
              discover_drilldown = {
                label = "Open Discover"
              }
            },
            {
              url_drilldown = {
                url     = "https://example.com/{{event.field}}"
                label   = "External"
                trigger = "on_click_value"
              }
            }
          ]

          axis = {
            y = { domain_json = jsonencode({ type = "fit" }) }
          }
          legend      = {}
          fitting     = { type = "none" }
          decorations = {}
          query       = { expression = "" }
          layers = [{
            type = "area"
            data_layer = {
              data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "logs-*" })
              y = [{
                config_json = jsonencode({ operation = "count", empty_as_null = true })
              }]
            }
          }]
        }
      }
    }
  }]
}
