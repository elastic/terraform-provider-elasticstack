variable "name" {
  type = string
}

variable "space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-synthetics-browser-${var.space_id}"
  description = "Kibana space for synthetics browser monitor acceptance test"
}

resource "elasticstack_fleet_agent_policy" "apl-browser-monitor" {
  name            = "TestMonitorResource Agent Policy - ${var.name}"
  namespace       = replace(var.space_id, "-", "_")
  description     = "TestMonitorResource Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
  space_ids       = [elasticstack_kibana_space.test.space_id]
}

resource "elasticstack_kibana_synthetics_private_location" "pl-browser-monitor" {
  space_id        = elasticstack_kibana_space.test.space_id
  label           = "monitor-pll-ns-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-browser-monitor.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "browser-monitor" {
  space_id          = elasticstack_kibana_space.test.space_id
  name              = "TestBrowserMonitorResource - ${var.name}"
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
