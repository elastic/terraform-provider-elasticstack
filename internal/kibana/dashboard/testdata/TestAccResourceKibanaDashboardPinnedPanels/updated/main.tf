variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_data_view" "test" {
  override = true
  data_view = {
    title          = "pinned-panels-acc-test-*"
    name           = "pinned-panels-acc-test"
    allow_no_index = true
  }
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with pinned options list + range slider controls"

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

  pinned_panels = [
    {
      type = "options_list_control"
      options_list_control_config = {
        data_view_id     = elasticstack_kibana_data_view.test.data_view.id
        field_name       = "host.name"
        search_technique = "wildcard"
        single_select    = false
        display_settings = {
          placeholder = "Pick a host..."
          hide_sort   = false
        }
      }
    },
    {
      type = "range_slider_control"
      range_slider_control_config = {
        data_view_id       = elasticstack_kibana_data_view.test.data_view.id
        field_name         = "source.bytes"
        title              = "Bytes Range Updated"
        use_global_filters = false
        ignore_validations = true
        value              = ["50", "400"]
        step               = 5
      }
    },
  ]
}
