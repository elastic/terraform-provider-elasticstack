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
    logging_level                  = "debug"
    logging_to_files               = true
    logging_files_interval         = "45s"
    logging_files_keepfiles        = 10
    logging_files_rotateeverybytes = 20971520
    logging_metrics_period         = "1m"
    go_max_procs                   = 2
    download_timeout               = "3h"
    download_target_directory      = "/tmp/agent-downloads"
  }
}

data "elasticstack_fleet_agent_policy" "test" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
