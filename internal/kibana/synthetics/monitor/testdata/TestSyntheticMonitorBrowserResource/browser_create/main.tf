variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "apl-browser-monitor" {
  name            = "TestMonitorResource Agent Policy - ${var.name}"
  namespace       = "testacc"
  description     = "TestMonitorResource Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "pl-browser-monitor" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-browser-monitor.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "browser-monitor" {
  name              = "TestBrowserMonitorResource - ${var.name}"
  space_id          = "testacc"
  namespace         = "testacc_ns"
  schedule          = 5
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-browser-monitor.label]
  enabled           = true
  tags              = ["a", "b"]
  service_name      = "test apm service"
  timeout           = 30
  browser = {
    inline_script = "step('Go to https://google.com.co', () => page.goto('https://www.google.com'))"
  }
}
