variable "dashboard_title" {
  type = string
}

variable "markdown_lib_id" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "markdown by_reference acceptance"
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
    type = "markdown"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 10
    }
    markdown_config = {
      by_reference = {
        ref_id      = var.markdown_lib_id
        title       = "Overlay title for library markdown"
        description = "Overlay description for by-reference panel"
        hide_title  = true
        hide_border = false
      }
    }
  }]
}
