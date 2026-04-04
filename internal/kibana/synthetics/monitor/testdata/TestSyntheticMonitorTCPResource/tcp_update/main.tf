variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "apl-tcp-monitor" {
  name            = "TestMonitorResource Agent Policy - ${var.name}"
  namespace       = "testacc"
  description     = "TestMonitorResource Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "pl-tcp-monitor" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-tcp-monitor.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "tcp-monitor" {
  name              = "TestTcpMonitorResource Updated - ${var.name}"
  schedule          = 10
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-tcp-monitor.label]
  enabled           = false
  tags              = ["c", "d", "e"]
  alert = {
    status = {
      enabled = true
    }
    tls = {
      enabled = false
    }
  }
  service_name = "test apm service"
  timeout      = 30
  tcp = {
    host                     = "http://localhost:8080"
    ssl_verification_mode    = "full"
    ssl_supported_protocols  = ["TLSv1.2"]
    proxy_url                = "http://localhost"
    proxy_use_local_resolver = false
    check_send               = "Hello Updated"
    check_receive            = "World Updated"
  }
}
