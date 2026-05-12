variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "REQ-009: by_value settings present; open_links_in_new_tab omitted"
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
      by_value = {
        content = "Markdown with empty settings"
        # settings block required; omit open_links_in_new_tab so Kibana applies default (true).
        settings = {}
      }
    }
  }]
}
