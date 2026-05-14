variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_data_view" "test" {
  override = true
  data_view = {
    title          = "pinned-panels-acc-invalid-*"
    name           = "pinned-panels-acc-invalid"
    allow_no_index = true
  }
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Invalid pinned panel discriminator mismatch"

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
      type = "range_slider_control"
      options_list_control_config = {
        data_view_id = elasticstack_kibana_data_view.test.data_view.id
        field_name   = "status"
      }
    },
  ]
}
