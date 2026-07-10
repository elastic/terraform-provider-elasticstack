variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "target" {
  title       = "${var.dashboard_title} target"
  description = "Target dashboard for links panel acceptance"

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
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with a by-value links panel"

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
      by_value = {
        layout      = "vertical"
        title       = "Links panel title"
        description = "Links panel description"
        hide_title  = false
        hide_border = true
        links = [
          {
            type            = "dashboard"
            destination     = elasticstack_kibana_dashboard.target.dashboard_id
            label           = "Dashboard link"
            open_in_new_tab = false
            use_filters     = true
            use_time_range  = true
          },
          {
            type            = "external"
            destination     = "https://example.com"
            label           = "External link"
            open_in_new_tab = true
            encode_url      = true
          },
        ]
      }
    }
  }]

  depends_on = [elasticstack_kibana_dashboard.target]
}
