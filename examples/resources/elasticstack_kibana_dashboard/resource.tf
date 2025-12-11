resource "elasticstack_kibana_dashboard" "my_dashboard" {
  title       = "My Dashboard"
  description = "A dashboard showing key metrics"

  # Time range
  time_from = "now-15m"
  time_to   = "now"

  # Refresh settings
  refresh_interval_pause = false
  refresh_interval_value = 60000 # 60 seconds

  # Query settings with text-based query (KQL or Lucene)
  query_language = "kuery"
  query_text     = "status:success"

  # Optional tags
  tags = ["production", "monitoring"]
}

# Example with JSON query (mutually exclusive with query_text)
resource "elasticstack_kibana_dashboard" "my_dashboard_json" {
  title       = "My Dashboard with JSON Query"
  description = "A dashboard with a structured query"

  # Time range
  time_from = "now-15m"
  time_to   = "now"

  # Refresh settings
  refresh_interval_pause = false
  refresh_interval_value = 60000 # 60 seconds

  # Query settings with JSON query object
  query_language = "kuery"
  query_json = jsonencode({
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

  # Optional tags
  tags = ["production", "monitoring"]
}

# Example with control group input for dashboard controls
resource "elasticstack_kibana_dashboard" "my_dashboard_with_controls" {
  title       = "My Dashboard with Controls"
  description = "A dashboard with interactive controls"

  # Time range
  time_from = "now-24h"
  time_to   = "now"

  # Refresh settings
  refresh_interval_pause = true
  refresh_interval_value = 90000

  # Query settings
  query_language = "kuery"
  query_text     = ""

  # Control group configuration
  control_group_input = {
    auto_apply_selections = true
    chaining_system       = "HIERARCHICAL"
    label_position        = "oneLine"

    # Settings to ignore global dashboard settings
    ignore_parent_settings = {
      ignore_filters     = false
      ignore_query       = false
      ignore_timerange   = false
      ignore_validations = false
    }

    # Individual control panels
    controls = [
      {
        type  = "optionsListControl"
        order = 0
        width = "medium"
        grow  = false
        control_config = jsonencode({
          dataViewId = "my-dataview-id"
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
          dataViewId = "my-dataview-id"
          fieldName  = "response_time"
          title      = "Response Time Range"
        })
      }
    ]
  }
}
