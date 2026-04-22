variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "apl-http-monitor" {
  name               = "TestMonitorResource Agent Policy - ${var.name}"
  namespace          = "testacc"
  description        = "TestMonitorResource Agent Policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Agent Download Source HTTP Monitor ${var.name}"
  source_id = "agent-download-source-http-monitor-${var.name}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
}

resource "elasticstack_kibana_synthetics_private_location" "pl-http-monitor" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-http-monitor.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "http-monitor" {
  name              = "TestHttpMonitorResource Updated - ${var.name}"
  schedule          = 10
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-http-monitor.label]
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
  service_name      = "test apm service"
  timeout           = 30
  retest_on_failure = false
  params = jsonencode({
    "param-name" = "param-value-updated"
  })
  http = {
    url                     = "http://localhost:8080"
    ssl_verification_mode   = "full"
    ssl_supported_protocols = ["TLSv1.2"]
    max_redirects           = 10
    mode                    = "all"
    ipv4                    = true
    ipv6                    = true
    proxy_url               = "http://localhost"
    proxy_header = jsonencode({
      "header-name" = "header-value-updated"
    })
    username = "testupdated"
    password = "testpassword-updated"
    check = jsonencode({
      "request" : {
        "method" : "POST",
        "headers" : {
          "Content-Type" : "application/x-www-form-urlencoded",
        },
        "body" : "name=first&email=someemail@someemailprovider.com",
      },
      "response" : {
        "status" : [200, 201, 301],
        "body" : {
          "positive" : ["foo", "bar"]
        }
      }
    })
    response = jsonencode({
      "include_body" : "never",
      "include_body_max_bytes" : "1024",
    })
  }
}
