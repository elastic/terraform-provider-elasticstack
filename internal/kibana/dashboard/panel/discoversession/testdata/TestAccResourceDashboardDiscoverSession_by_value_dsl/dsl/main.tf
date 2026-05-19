variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Discover session panel by_value DSL tab"

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
    type = "discover_session"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 12
    }
    discover_session_config = {
      title       = "DSL Discover"
      description = "by_value dsl tab"
      hide_title  = false
      hide_border = true
      by_value = {
        tab = {
          dsl = {
            query = {
              expression = "host.name : *"
              language   = "kql"
            }
            data_source_json = jsonencode({
              type   = "data_view_reference"
              ref_id = "kibana_sample_data_logs"
            })
            column_order = ["@timestamp", "message"]
            view_mode    = "documents"
          }
        }
      }
    }
  }]
}
