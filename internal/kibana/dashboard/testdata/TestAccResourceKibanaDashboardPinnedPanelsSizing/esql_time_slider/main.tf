variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_data_view" "test" {
  override = true
  data_view = {
    title          = "pinned-panels-acc-sizing-*"
    name           = "pinned-panels-acc-sizing"
    allow_no_index = true
  }
}

resource "elasticstack_kibana_dashboard" "test" {
  title = var.dashboard_title

  time_range = {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval = {
    pause = true
    value = 0
  }
  query = {
    language   = "kql"
    expression = ""
  }

  pinned_panels = [
    {
      type  = "esql_control"
      width = "large"
      grow  = true
      esql_control_config = {
        selected_options  = ["option_a"]
        variable_name     = "my_var"
        variable_type     = "values"
        esql_query        = "FROM logs-* | STATS count = COUNT(*) BY host.name"
        control_type      = "STATIC_VALUES"
        available_options = ["option_a", "option_b"]
      }
    },
    {
      type  = "time_slider_control"
      width = "small"
      grow  = false
      time_slider_control_config = {
        start_percentage_of_time_range = 0.1
        end_percentage_of_time_range   = 0.9
        is_anchored                    = false
      }
    },
  ]
}
