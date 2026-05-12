variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "vis viz_config.by_reference mixed drilldowns"
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
    viz_config = {
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
            dashboard = {
              dashboard_id    = "22222222-2222-2222-2222-222222222222"
              label           = "Dashboard drill"
              use_filters     = false
              use_time_range  = true
              open_in_new_tab = true
            }
          },
          {
            url = {
              url             = "https://mixed.example/{{event.field}}"
              label           = "URL drill"
              trigger         = "on_open_panel_menu"
              encode_url      = true
              open_in_new_tab = true
            }
          },
          {
            discover = {
              label           = "Discover drill"
              open_in_new_tab = true
            }
          }
        ]
      }
    }
  }]
}
