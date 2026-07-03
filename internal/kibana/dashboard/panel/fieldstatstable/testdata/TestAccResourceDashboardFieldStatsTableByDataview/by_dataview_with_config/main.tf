variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_data_view" "test" {
  override = true
  data_view = {
    title          = "field-stats-table-acc-test-*"
    name           = "field-stats-table-acc-test"
    allow_no_index = true
  }
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with field_stats_table panel (by_dataview, with config)"

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
    type = "field_stats_table"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    field_stats_table_config = {
      by_dataview = {
        data_view_id       = elasticstack_kibana_data_view.test.data_view.id
        show_distributions = true
        title              = "Field statistics — logs view"
        description        = "Field stats table panel (dataview)"
        hide_title         = false
        hide_border        = true
        time_range = {
          from = "now-24h"
          to   = "now"
        }
      }
    }
  }]
}
