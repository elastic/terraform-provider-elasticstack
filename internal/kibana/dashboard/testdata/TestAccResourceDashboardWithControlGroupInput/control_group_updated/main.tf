variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Updated dashboard with modified control group input"

  time_from = "now-30m"
  time_to   = "now"

  refresh_interval_pause = false
  refresh_interval_value = 30000

  query_language = "kuery"
  query_text     = ""

  control_group_input = {
    auto_apply_selections = false
    chaining_system       = "NONE"
    label_position        = "twoLine"

    ignore_parent_settings = {
      ignore_filters     = false
      ignore_query       = true
      ignore_timerange   = false
      ignore_validations = true
    }

    controls = [
      {
        type  = "optionsListControl"
        order = 0
        width = "small"
        grow  = true
        control_config = jsonencode({
          dataViewId = "test-dataview"
          fieldName  = "category"
          title      = "Category Filter"
        })
      }
    ]
  }
}
