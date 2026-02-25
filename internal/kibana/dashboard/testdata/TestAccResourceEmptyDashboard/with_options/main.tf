variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Test dashboard with options"

  time_range = {
    from = "2024-01-01T00:00:00.000Z"
    to   = "2024-01-01T01:00:00.000Z"
    mode = "absolute"
  }

  refresh_interval = {
    pause = true
    value = 60000
  }

  query = {
    language = "kuery"
    text     = ""
  }

  options = {
    hide_panel_titles = true
    use_margins       = false
    sync_colors       = true
    sync_tooltips     = true
    sync_cursor       = false
  }
}
