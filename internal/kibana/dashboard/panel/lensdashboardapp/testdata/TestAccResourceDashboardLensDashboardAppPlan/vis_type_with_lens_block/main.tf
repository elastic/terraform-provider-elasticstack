variable "dashboard_title" { type = string }
resource "elasticstack_kibana_dashboard" "test" {
  title            = var.dashboard_title
  description      = "plan: vis + config_json + lens_dashboard_app (REQ-006 sibling conflict)"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }
  panels = [{
    type        = "vis"
    grid        = { x = 0, y = 0, w = 4, h = 4 }
    config_json = jsonencode({ title = "vis panel" })
    lens_dashboard_app_config = {
      by_reference = {
        ref_id = "r"
        time_range = {
          from = "a"
          to   = "b"
        }
      }
    }
  }]
}
