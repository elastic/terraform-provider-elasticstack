provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name             = "Test Policy"
  namespace        = "default"
  description      = "Test Agent Policy"
  sys_monitoring   = true
  monitor_logs     = true
  monitor_metrics  = true
  space_ids        = ["default"]
  host_name_format = "hostname" # or "fqdn" for fully qualified domain names

  global_data_tags = {
    first_tag = {
      string_value = "tag_value"
    },
    second_tag = {
      number_value = 1.2
    }
  }

  # Advanced monitoring options (requires Elastic Stack 8.16.0+)
  advanced_monitoring_options = {
    # HTTP monitoring endpoint for liveness probes / health checks
    http_monitoring_endpoint = {
      enabled        = true
      host           = "localhost"
      port           = 6791
      buffer_enabled = false
      pprof_enabled  = false # Enable for /debug/pprof/* profiling endpoints
    }

    # Diagnostic settings (optional - defaults: interval=1m, burst=1, init_duration=1s, backoff_duration=1m, max_retries=10)
    diagnostics = {
      rate_limits = {
        interval = "5m"
        burst    = 3
      }
      file_uploader = {
        init_duration    = "2s"
        backoff_duration = "2m"
        max_retries      = 5
      }
    }
  }
}
