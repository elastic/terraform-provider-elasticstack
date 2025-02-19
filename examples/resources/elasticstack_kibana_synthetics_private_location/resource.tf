provider "elasticstack" {
  fleet {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "sample" {
  name            = "Sample Agent Policy"
  namespace       = "default"
  description     = "A sample agent policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "example" {
  label           = "example label"
  agent_policy_id = elasticstack_fleet_agent_policy.sample.policy_id
  tags            = ["tag-a", "tag-b"]
  geo = {
    lat = 40.7128
    lon = 74.0060
  }
}
