provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# By-reference `vis` panel: embed a saved Lens visualization via `viz_config.by_reference`.
# `references_json` wires the saved object ID to `ref_id`; structured `drilldowns` replaces legacy JSON.
resource "elasticstack_kibana_dashboard" "vis_by_reference_example" {
  title            = "Dashboard with vis (by-reference Lens)"
  description      = "Example: viz_config.by_reference with structured drilldowns"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [{
    type = "vis"
    grid = { x = 0, y = 0, w = 24, h = 12 }
    viz_config = {
      by_reference = {
        ref_id = "embedded_lens_ref"
        references_json = jsonencode([
          {
            id   = "00000000-0000-4000-8000-000000000001"
            name = "embedded_lens_ref"
            type = "lens"
          }
        ])
        time_range = {
          from = "now-7d"
          to   = "now"
          mode = "relative"
        }
        title = "Saved Lens visualization"

        drilldowns = [
          {
            dashboard = {
              dashboard_id    = "11111111-1111-4111-8111-111111111111"
              label           = "Open detail dashboard"
              use_filters     = false
              use_time_range  = true
              open_in_new_tab = true
            }
          },
          {
            url = {
              url     = "https://example.com/?value={{event.value}}"
              label   = "Open external link"
              trigger = "on_click_value"
            }
          },
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
