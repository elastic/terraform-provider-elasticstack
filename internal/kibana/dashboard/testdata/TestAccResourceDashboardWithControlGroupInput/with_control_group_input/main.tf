variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Test dashboard with control group input"

  time_from = "now-15m"
  time_to   = "now"

  refresh_interval_pause = true
  refresh_interval_value = 60000

  query_language = "kuery"
  query_text     = ""

  control_group_input = {
    auto_apply_selections = true
    chaining_system       = "HIERARCHICAL"
    label_position        = "oneLine"

    ignore_parent_settings = {
      ignore_filters     = true
      ignore_query       = false
      ignore_timerange   = true
      ignore_validations = false
    }

    controls = [
      {
        type  = "optionsListControl"
        order = 0
        width = "medium"
        grow  = false
        control_config = jsonencode({
          dataViewId = "test-dataview"
          fieldName  = "status"
          title      = "Status Filter"
        })
      },
      {
        type  = "rangeSliderControl"
        order = 1
        width = "large"
        grow  = true
        control_config = jsonencode({
          dataViewId = "test-dataview"
          fieldName  = "amount"
          title      = "Amount Range"
        })
      }
    ]
  }
}
