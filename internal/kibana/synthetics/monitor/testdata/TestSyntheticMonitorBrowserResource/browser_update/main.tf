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
  name              = "TestBrowserMonitorResource Updated - ${var.name}"
  schedule          = 10
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-browser-monitor.label]
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
  browser = {
    inline_script       = "step('Go to https://google.de', () => page.goto('https://www.google.de'))"
    synthetics_args     = ["--no-sandbox", "--disable-setuid-sandbox"]
    screenshots         = "off"
    ignore_https_errors = true
    playwright_options  = jsonencode({ "httpCredentials" : { "password" : "test", "username" : "test" }, "ignoreHTTPSErrors" : false })
  }
}
