provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name      = "Logstash Output (No SSL) ${var.policy_name}"
  type      = "logstash"
  output_id = "${var.policy_name}-logstash-output-no-ssl"

  default_integrations = false
  default_monitoring   = false
  hosts                = ["logstash:5044"]
}
