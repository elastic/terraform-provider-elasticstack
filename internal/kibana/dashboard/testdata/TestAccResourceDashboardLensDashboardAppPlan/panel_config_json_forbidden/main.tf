variable "dashboard_title" { type = string }
resource "elasticstack_kibana_dashboard" "test" {
  title            = var.dashboard_title
  description      = "plan: panel-level config_json for lens (allowlist), REQ-006/6.17"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }
  panels = [{
    type = "lens-dashboard-app"
    grid = { x = 0, y = 0, w = 4, h = 4 }
    # Practitioner-authored: plan-time allowlist on panel-level `config_json`
    # (dedicated block omitted so we exercise schema-level rejection).
    config_json = jsonencode({ ref_id = "x" })
  }]
}
