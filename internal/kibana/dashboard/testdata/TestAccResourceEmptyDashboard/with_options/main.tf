variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Test dashboard with options"

  time_from       = "2024-01-01T00:00:00.000Z"
  time_to         = "2024-01-01T01:00:00.000Z"
  time_range_mode = "absolute"

  refresh_interval_pause = true
  refresh_interval_value = 60000

  query_language = "kuery"
  query_text     = ""

  options = {
    auto_apply_filters = true
    hide_panel_titles = true
    use_margins       = false
    sync_colors       = true
    sync_tooltips     = true
    sync_cursor       = false
  }
}
