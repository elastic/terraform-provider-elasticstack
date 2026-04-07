variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "apl-tcp-monitor-min" {
  name            = "TestMonitorResource Agent Policy - ${var.name}"
  namespace       = "testacc"
  description     = "TestMonitorResource Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "pl-tcp-monitor-min" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-tcp-monitor-min.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "tcp-monitor-min" {
  name              = "TestTcpMonitorResource - ${var.name}"
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-tcp-monitor-min.label]
  tcp = {
    host = "http://localhost:5601"
  }
}
