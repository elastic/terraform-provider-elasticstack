variable "dashboard_title" {
  type = string
}

variable "discover_ref_id" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Discover session panel by_reference to search saved object"

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
    type = "discover_session"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 12
    }
    discover_session_config = {
      title       = "Saved search panel"
      description = "by_reference acceptance"
      hide_title  = true
      hide_border = false
      by_reference = {
        ref_id = var.discover_ref_id
        time_range = {
          from = "now-7d"
          to   = "now"
          mode = "relative"
        }
      }
    }
  }]
}
