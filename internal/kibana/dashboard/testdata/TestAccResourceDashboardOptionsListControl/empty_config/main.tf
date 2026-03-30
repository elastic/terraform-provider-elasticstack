variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_data_view" "test" {
  override = true
  data_view = {
    title          = "options-list-control-acc-test-*"
    name           = "options-list-control-acc-test"
    allow_no_index = true
  }
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Options List Control Panel (required fields only)"

  time_from = "now-15m"
  time_to   = "now"

  refresh_interval_pause = true
  refresh_interval_value = 0

  query_language = "kuery"
  query_text     = ""

  panels = [{
    type = "options_list_control"
    grid = {
      x = 0
      y = 0
      w = 12
      h = 4
    }
    options_list_control_config = {
      data_view_id = elasticstack_kibana_data_view.test.id
      field_name   = "status"
    }
  }]
}
