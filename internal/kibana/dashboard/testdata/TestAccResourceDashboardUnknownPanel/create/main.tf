variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard for unknown panel preservation test"
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
    id   = "tf-acc-markdown-panel-uid"
    grid = {
      x = 0
      y = 0
      w = 48
      h = 15
    }
    markdown_config = {
      content = "Placeholder panel"
    }
  }]
}
