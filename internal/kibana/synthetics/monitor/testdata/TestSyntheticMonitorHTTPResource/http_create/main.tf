variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "apl-http-monitor" {
  name            = "TestMonitorResource Agent Policy - ${var.name}"
  namespace       = "testacc"
  description     = "TestMonitorResource Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "pl-http-monitor" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-http-monitor.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "http-monitor" {
  name              = "TestHttpMonitorResource - ${var.name}"
  space_id          = "testacc"
  namespace         = "test_namespace"
  schedule          = 5
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-http-monitor.label]
  enabled           = true
  tags              = ["a", "b"]
  alert = {
    status = {
      enabled = true
    }
    tls = {
      enabled = true
    }
  }
  service_name = "test apm service"
  timeout      = 30
  http = {
    url  = "http://localhost:5601"
    mode = "any"
    ipv4 = true
    ipv6 = false
  }
}
