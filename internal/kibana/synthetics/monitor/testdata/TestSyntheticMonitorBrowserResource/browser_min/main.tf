variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "apl-browser-monitor-min" {
  name               = "TestMonitorResource Agent Policy - ${var.name}"
  namespace          = "testacc"
  description        = "TestMonitorResource Agent Policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Agent Download Source Browser Monitor ${var.name}"
  source_id = "agent-download-source-browser-monitor-${var.name}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
}

resource "elasticstack_kibana_synthetics_private_location" "pl-browser-monitor-min" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-browser-monitor-min.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "browser-monitor-min" {
  name              = "TestBrowserMonitorResource - ${var.name}"
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-browser-monitor-min.label]
  alert = {
    status = {
      enabled = true
    }
    tls = {
      enabled = true
    }
  }
  browser = {
    inline_script = "step('Go to https://google.com.co', () => page.goto('https://www.google.com'))"
  }
}
