provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy with Advanced Settings"
  monitor_logs    = true
  monitor_metrics = true

  advanced_settings = {
    logging_level                  = "info"
    logging_to_files               = true
    logging_files_keepfiles        = 7
    logging_files_rotateeverybytes = 10485760
    go_max_procs                   = 4
    download_target_directory      = "/tmp/elastic-agent"
  }
}

