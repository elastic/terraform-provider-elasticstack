variable "dashboard_title" { type = string }
resource "elasticstack_kibana_dashboard" "test" {
  title            = var.dashboard_title
  description      = "plan: REQ-006 — type=vis with only lens_dashboard_app (no vis chart config)"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }
  panels = [{
    type = "vis"
    grid = { x = 0, y = 0, w = 4, h = 4 }
    lens_dashboard_app_config = {
      by_reference = {
        ref_id = "r"
        time_range = {
          from = "now-1h"
          to   = "now"
        }
      }
    }
  }]
}
