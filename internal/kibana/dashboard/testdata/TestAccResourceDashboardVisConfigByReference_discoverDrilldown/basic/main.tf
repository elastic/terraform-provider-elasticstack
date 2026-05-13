variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "vis vis_config.by_reference discover drilldown"
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
    type = "vis"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 10
    }
    vis_config = {
      by_reference = {
        ref_id = "lensRef"
        references_json = jsonencode([
          {
            id   = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
            name = "lensRef"
            type = "lens"
          }
        ])
        time_range = {
          from = "now-7d"
          to   = "now"
          mode = "relative"
        }
        drilldowns = [
          {
            discover = {
              label           = "Open in Discover"
              open_in_new_tab = false
            }
          }
        ]
      }
    }
  }]
}
