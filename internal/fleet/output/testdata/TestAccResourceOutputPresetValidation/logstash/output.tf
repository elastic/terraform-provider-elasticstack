provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name      = "Logstash Preset Output ${var.policy_name}"
  output_id = "${var.policy_name}-logstash-preset-output"
  type      = "logstash"
  preset    = "scale"

  hosts = [
    "logstash:5044",
  ]
}
