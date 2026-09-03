variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with a by-reference links panel"

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
    type = "links"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 10
    }
    links_config = {
      by_reference = {
        ref_id      = "links-ref-1"
        title       = "Linked links panel"
        description = "Linked links panel description"
        hide_title  = true
        hide_border = false
      }
    }
  }]
}
